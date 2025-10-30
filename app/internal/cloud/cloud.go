package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"seisami/app/internal/repo"
	"seisami/app/internal/repo/sqlc/query"
	"seisami/app/types"
	"strings"
	"time"
)

type cloudFuncs struct {
	repo         repo.Repository
	sessionToken string
	ctx          context.Context
	cloudApiUrl  string
	httpClient   http.Client
}

func NewCloudFuncs(repo repo.Repository, sessionToken string, ctx context.Context, cloudApiUrl string) cloudFuncs {
	httpClient := http.Client{Timeout: 30 * time.Second}

	return cloudFuncs{
		repo,
		sessionToken,
		ctx,
		cloudApiUrl,
		httpClient,
	}
}

func (cf cloudFuncs) GetAllOperations(tableName types.TableName) ([]types.OperationSync, error) {
	// Get all operations without filters - this will return all operations for the user
	operations, err := cf.PullRecords(tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get all operations: %w", err)
	}
	return operations, nil
}

type HttpResponse struct {
	HasError bool
	Message  string
	Data     any
}

func (cf cloudFuncs) buildURL(path string) string {
	if strings.HasPrefix(path, "http") {
		return path
	}

	base := strings.TrimRight(cf.cloudApiUrl, "/")
	suffix := strings.TrimLeft(path, "/")
	if suffix == "" {
		return base
	}

	return base + "/" + suffix
}

func (cf cloudFuncs) doJSONRequest(method, path string, payload any) (int, []byte, error) {
	ctx := cf.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return 0, nil, fmt.Errorf("serialize payload: %w", err)
		}
		body = bytes.NewBuffer(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, cf.buildURL(path), body)
	if err != nil {
		return 0, nil, fmt.Errorf("prepare request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cf.sessionToken))

	res, err := cf.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("execute request: %w", err)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, nil, fmt.Errorf("read response body: %w", err)
	}

	return res.StatusCode, resBody, nil
}

func (cf cloudFuncs) postSyncResource(path string, payload any) error {
	status, body, err := cf.doJSONRequest(http.MethodPost, path, payload)
	if err != nil {
		return err
	}

	if status != http.StatusOK && status != http.StatusCreated {
		return fmt.Errorf("sync api returned status %d: %s", status, string(body))
	}

	return nil
}

func (cf cloudFuncs) PushRecord(payload types.OperationSync) HttpResponse {
	status, resBody, err := cf.doJSONRequest(http.MethodPost, "/sync/upload", payload)
	if err != nil {
		return HttpResponse{
			HasError: true,
			Message:  "unable to sync data",
			Data:     err.Error(),
		}
	}

	if status != http.StatusCreated {
		return HttpResponse{
			HasError: true,
			Message:  "error syncing data",
			Data:     string(resBody),
		}
	}

	return HttpResponse{
		HasError: false,
		Message:  "data synced successfully",
		Data:     resBody,
	}
}

func (cf cloudFuncs) PullRecord(tableName types.TableName) (types.OperationSync, error) {
	operations, err := cf.PullRecords(tableName)
	if err != nil {
		return types.OperationSync{}, err
	}

	if len(operations) == 0 {
		return types.OperationSync{}, fmt.Errorf("no operations available for table %s", tableName.String())
	}

	return operations[0], nil
}

func (cf cloudFuncs) PullRecords(tableName types.TableName) ([]types.OperationSync, error) {
	status, resBody, err := cf.doJSONRequest(http.MethodGet, fmt.Sprintf("/sync/pull/%s", tableName.String()), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to pull records: %w", err)
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("api request failed with status %d: %s", status, string(resBody))
	}

	var response struct {
		Table      string                `json:"table"`
		Count      int                   `json:"count"`
		Operations []types.OperationSync `json:"operations"`
	}

	if err := json.Unmarshal(resBody, &response); err != nil {
		return nil, fmt.Errorf("unable to decode response: %w", err)
	}

	return response.Operations, nil
}

// TODO: work on this function  in cloud & return HttpResponse also
func (cf cloudFuncs) GetSyncState(tableName types.TableName) (query.SyncState, error) {
	status, resBody, err := cf.doJSONRequest(http.MethodGet, "/sync/state/"+tableName.String(), nil)
	if err != nil {
		return query.SyncState{}, fmt.Errorf("unable to get sync state: %w", err)
	}

	if status != http.StatusOK {
		return query.SyncState{}, fmt.Errorf("api request failed with status %d: %s", status, string(resBody))
	}

	var response HttpResponse
	if err := json.Unmarshal(resBody, &response); err != nil {
		return query.SyncState{}, fmt.Errorf("unable to deserialize data: %w", err)
	}

	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return query.SyncState{}, fmt.Errorf("unable to re-marshal data: %w", err)
	}

	var syncState types.SyncStatePayload
	if err := json.Unmarshal(dataBytes, &syncState); err != nil {
		return query.SyncState{}, fmt.Errorf("unable to decode sync state: %w", err)
	}

	return query.SyncState(syncState), nil
}

func (cf cloudFuncs) UpdateSyncState(state query.SyncState) error {
	if err := cf.postSyncResource("/sync/state", state); err != nil {
		return fmt.Errorf("unable to update sync state: %w", err)
	}

	return nil
}

func (cf cloudFuncs) UpsertBoard(board types.ExportedBoard) error {
	if err := cf.postSyncResource("/sync/board", board); err != nil {
		return fmt.Errorf("unable to upsert board: %w", err)
	}

	return nil
}

func (cf cloudFuncs) UpsertColumn(column types.ExportedColumn) error {
	if err := cf.postSyncResource("/sync/column", column); err != nil {
		return fmt.Errorf("unable to upsert column: %w", err)
	}

	return nil
}

func (cf cloudFuncs) UpsertCard(card types.ExportedCard) error {
	if err := cf.postSyncResource("/sync/card", card); err != nil {
		return fmt.Errorf("unable to upsert card: %w", err)
	}

	return nil
}

func (cf cloudFuncs) InitializeSyncStateForUser() error {
	if err := cf.postSyncResource("/sync/init", nil); err != nil {
		return fmt.Errorf("unable to init sync state: %w", err)
	}

	return nil
}
