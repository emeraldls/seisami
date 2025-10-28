package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"seisami/app/internal/repo"
	"seisami/app/internal/repo/sqlc/query"
	"seisami/app/types"
	"time"
)

type cloudFuncs struct {
	repo         repo.Repository
	sessionToken string
	ctx          context.Context
	cloudApiUrl  string
	httpClient   http.Client
}

func NewCloud(repo repo.Repository, sessionToken string, ctx context.Context, cloudApiUrl string) *cloudFuncs {
	httpClient := http.Client{Timeout: 30 * time.Second}

	return &cloudFuncs{
		repo,
		sessionToken,
		ctx,
		cloudApiUrl,
		httpClient,
	}
}

func (cf *cloudFuncs) GetAllOperations(tableName types.TableName) ([]types.OperationSync, error) {
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

func (cf cloudFuncs) PushRecord(payload types.OperationSync) HttpResponse {
	jBody, err := json.Marshal(payload)
	if err != nil {
		return HttpResponse{
			HasError: true,
			Message:  "unable to serialize data",
			Data:     err.Error(),
		}
	}

	// TODO: include path to api
	req, err := http.NewRequestWithContext(cf.ctx, "POST", cf.cloudApiUrl+"/sync/upload", bytes.NewBuffer(jBody))

	if err != nil {
		return HttpResponse{
			HasError: true,
			Message:  "unable to prepare api call",
			Data:     err.Error(),
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cf.sessionToken))

	res, err := cf.httpClient.Do(req)
	if err != nil {
		return HttpResponse{
			HasError: true,
			Message:  "unable to make api call",
			Data:     err.Error(),
		}
	}

	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return HttpResponse{
			HasError: true,
			Message:  "unable to read response body",
			Data:     err.Error(),
		}
	}

	if res.StatusCode != http.StatusCreated {
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

func (cf cloudFuncs) PullRecord(tableName types.TableName, syncState query.SyncState) (types.OperationSync, error) {
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
	url := fmt.Sprintf("%s/sync/pull/%s", cf.cloudApiUrl, tableName.String())

	req, err := http.NewRequestWithContext(cf.ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cf.sessionToken))

	res, err := cf.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to make request: %w", err)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api request failed with status %d: %s", res.StatusCode, string(resBody))
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

// TODO: work on this function to return HttpResponse also
func (cf cloudFuncs) GetSyncState(tableName types.TableName) (query.SyncState, error) {
	req, err := http.NewRequest("GET", cf.cloudApiUrl+"/state/"+tableName.String(), nil)
	if err != nil {
		return query.SyncState{}, fmt.Errorf("unable to prepare request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cf.sessionToken))

	res, err := cf.httpClient.Do(req)
	if err != nil {
		return query.SyncState{}, fmt.Errorf("unable to make request: %v", err)
	}

	var syncState types.SyncStatePayload

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return query.SyncState{}, fmt.Errorf("unable to read response body:%v", err)
	}

	err = json.Unmarshal(resBody, &syncState)
	if err != nil {
		return query.SyncState{}, fmt.Errorf("unable to deserialize data: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		fmt.Println("request wasnt successful: ", err)
		return query.SyncState{}, errors.New("something went wrong")
	}

	return query.SyncState(syncState), nil
}

func (cf cloudFuncs) UpdateSyncState(state query.SyncState) error {
	jBytes, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("unable to serialize data: %v", err)
	}

	req, err := http.NewRequest("POST", cf.cloudApiUrl+"/state", bytes.NewBuffer(jBytes))
	if err != nil {
		return fmt.Errorf("unable to prepare request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cf.sessionToken))

	res, err := cf.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to make request: %v", err)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		fmt.Println(string(resBody))
		return errors.New("something went wrong")
	}

	return nil
}
