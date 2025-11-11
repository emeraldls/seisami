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

func NewCloudFuncs(repo repo.Repository, sessionToken string, ctx context.Context, cloudApiUrl string) *cloudFuncs {
	httpClient := http.Client{Timeout: 30 * time.Second}

	return &cloudFuncs{
		repo,
		sessionToken,
		ctx,
		cloudApiUrl,
		httpClient,
	}
}

func (cf *cloudFuncs) GetAllOperations(tableName types.TableName, since int64) HttpResponse {
	// Get all operations without filters - this will return all operations for the user
	operations, err := cf.PullRecords(tableName, since)
	if err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "failed to get all operations",
		}
	}
	return HttpResponse{
		Message: "operations retrieved successfully",
		Data:    operations,
	}
}

type HttpResponse struct {
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (cf *cloudFuncs) buildURL(path string) string {
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

func (cf *cloudFuncs) UpdateSessionToken(token string) {
	cf.sessionToken = token
}

func (cf *cloudFuncs) doJSONRequest(method, path string, payload any) (int, []byte, error) {
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

func (cf *cloudFuncs) postSyncResource(path string, payload any) error {
	status, body, err := cf.doJSONRequest(http.MethodPost, path, payload)
	if err != nil {
		return err
	}

	if status != http.StatusOK && status != http.StatusCreated {
		return fmt.Errorf("sync api returned status %d: %s", status, string(body))
	}

	return nil
}

func (cf *cloudFuncs) PushRecord(payload types.OperationSync) HttpResponse {
	status, resBody, err := cf.doJSONRequest(http.MethodPost, "/sync/upload", payload)
	if err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to sync data",
		}
	}

	if status != http.StatusCreated {
		return HttpResponse{
			Error:   string(resBody),
			Message: "error syncing data",
		}
	}

	return HttpResponse{
		Error:   "",
		Message: "data synced successfully",
		Data:    resBody,
	}
}

func (cf *cloudFuncs) PullRecord(tableName types.TableName, since int64) HttpResponse {
	operations, err := cf.PullRecords(tableName, since)
	if err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to pull record",
		}
	}

	if len(operations) == 0 {
		return HttpResponse{
			Error:   fmt.Sprintf("no operations available for table %s", tableName.String()),
			Message: "no records found",
		}
	}

	return HttpResponse{
		Message: "record retrieved successfully",
		Data:    operations[0],
	}
}

func (cf *cloudFuncs) PullRecords(tableName types.TableName, since int64) ([]types.OperationSync, error) {
	status, resBody, err := cf.doJSONRequest(http.MethodGet, fmt.Sprintf("/sync/pull/%s?since=%d", tableName.String(), since), nil)
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

func (cf *cloudFuncs) PullRecordsV2(tableName types.TableName, since int64) HttpResponse {
	status, resBody, err := cf.doJSONRequest(http.MethodGet, fmt.Sprintf("/sync/pull/%s?since=%d", tableName.String(), since), nil)
	if err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to pull records",
		}
	}

	if status != http.StatusOK {
		return HttpResponse{
			Error:   string(resBody),
			Message: fmt.Sprintf("api request failed with status %d", status),
		}
	}

	var response struct {
		Table      string                `json:"table"`
		Count      int                   `json:"count"`
		Operations []types.OperationSync `json:"operations"`
	}

	if err := json.Unmarshal(resBody, &response); err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to decode response",
		}
	}

	return HttpResponse{
		Message: "records retrieved successfully",
		Data:    response.Operations,
	}
}

func (cf *cloudFuncs) GetSyncState(tableName types.TableName) HttpResponse {
	status, resBody, err := cf.doJSONRequest(http.MethodGet, "/sync/state/"+tableName.String(), nil)
	if err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to get sync state",
		}
	}

	if status != http.StatusOK {
		return HttpResponse{
			Error:   string(resBody),
			Message: fmt.Sprintf("api request failed with status %d", status),
		}
	}

	var response HttpResponse
	if err := json.Unmarshal(resBody, &response); err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to deserialize data",
		}
	}

	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to re-marshal data",
		}
	}

	var syncState types.SyncStatePayload
	if err := json.Unmarshal(dataBytes, &syncState); err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to decode sync state",
		}
	}

	return HttpResponse{
		Message: "sync state retrieved successfully",
		Data:    query.SyncState(syncState),
	}
}

func (cf *cloudFuncs) UpdateSyncState(state query.SyncState) HttpResponse {
	payload := types.SyncStatePayload{
		TableName:      state.TableName,
		LastSyncedAt:   state.LastSyncedAt,
		LastSyncedOpID: state.LastSyncedOpID,
	}

	if err := cf.postSyncResource("/sync/state", payload); err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to update sync state",
		}
	}

	return HttpResponse{
		Message: "sync state updated successfully",
	}
}

func (cf *cloudFuncs) UpsertBoard(board types.ExportedBoard) HttpResponse {
	if err := cf.postSyncResource("/sync/board", board); err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to upsert board",
		}
	}

	return HttpResponse{
		Message: "board upserted successfully",
	}
}

func (cf *cloudFuncs) UpsertColumn(column types.ExportedColumn) HttpResponse {
	if err := cf.postSyncResource("/sync/column", column); err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to upsert column",
		}
	}

	return HttpResponse{
		Message: "column upserted successfully",
	}
}

func (cf *cloudFuncs) UpsertCard(card types.ExportedCard) HttpResponse {
	if err := cf.postSyncResource("/sync/card", card); err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to upsert card",
		}
	}

	return HttpResponse{
		Message: "card upserted successfully",
	}
}

func (cf *cloudFuncs) InitializeSyncStateForUser() HttpResponse {
	if err := cf.postSyncResource("/sync/init", nil); err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to init sync state",
		}
	}

	return HttpResponse{
		Message: "sync state initialized successfully",
	}
}

/**
So now the workflow is when you invite someone to the board, like to the Kanban board, what you have to do is:

When they click the link to open the board on your app, it will pull all the info for that particular board. It's going to pull all the data for that particular board.

Then when it has pulled all the data for a particular board, the user will now update their local database to what has been pulled. It's going to create a new table of boards, insert the board data and every other information.

I think this is perfect.
*/

func (cf *cloudFuncs) ImportBoardData(boardId string) HttpResponse {
	status, body, err := cf.doJSONRequest("GET", "/sync/export/"+boardId, nil)

	if err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to import board data",
		}
	}

	var httpResp HttpResponse

	if err = json.Unmarshal(body, &httpResp); err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to unmarshal body",
		}
	}

	if status != http.StatusOK && status != http.StatusCreated {
		return HttpResponse{
			Error:   string(body),
			Message: fmt.Sprintf("sync api returned status %d", status),
		}
	}

	dataBytes, err := json.Marshal(httpResp.Data)
	if err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to re-marshal data",
		}
	}

	var data types.ImportUserBoardData

	err = json.Unmarshal(dataBytes, &data)
	if err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to decode board data",
		}
	}

	return HttpResponse{
		Message: "board data imported successfully",
		Data:    data,
	}
}

func (cf *cloudFuncs) ImportAllUserData() HttpResponse {
	status, body, err := cf.doJSONRequest("GET", "/sync/export/", nil)

	if err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to import user data",
		}
	}

	var httpResp HttpResponse

	if err = json.Unmarshal(body, &httpResp); err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to unmarshal body",
		}
	}

	if status != http.StatusOK && status != http.StatusCreated {
		return HttpResponse{
			Error:   string(body),
			Message: fmt.Sprintf("sync api returned status %d", status),
		}
	}

	dataBytes, err := json.Marshal(httpResp.Data)
	if err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to re-marshal data",
		}
	}

	var data types.ExportedData

	err = json.Unmarshal(dataBytes, &data)
	if err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to decode user data",
		}
	}

	return HttpResponse{
		Message: "user data imported successfully",
		Data:    data,
	}
}

func (cf *cloudFuncs) InitCloud() HttpResponse {
	if err := cf.postSyncResource("/init/cloud", nil); err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to init cloud status",
		}
	}

	return HttpResponse{
		Message: "cloud initialized successfully",
	}
}

func (cf *cloudFuncs) FetchAppVersion() HttpResponse {
	status, body, err := cf.doJSONRequest("GET", "/updates/latest", nil)

	if err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to fetch app version",
		}
	}

	if status != http.StatusOK && status != http.StatusCreated {
		return HttpResponse{
			Error:   string(body),
			Message: fmt.Sprintf("updates api returned status %d", status),
		}
	}

	var httpResp HttpResponse

	if err = json.Unmarshal(body, &httpResp); err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to unmarshal body",
		}
	}

	dataBytes, err := json.Marshal(httpResp.Data)
	if err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to re-marshal data",
		}
	}

	var appVersion types.AppVersion

	err = json.Unmarshal(dataBytes, &appVersion)
	if err != nil {
		return HttpResponse{
			Error:   err.Error(),
			Message: "unable to decode app version",
		}
	}

	return HttpResponse{
		Message: "app version fetched successfully",
		Data:    appVersion,
	}
}
