package cloud

import (
	"seisami/app/internal/repo/sqlc/query"
	"seisami/app/types"
)

type Cloud interface {
	GetAllOperations(tableName types.TableName) ([]types.OperationSync, error)
	PullRecord(tableName types.TableName) (types.OperationSync, error)
	PullRecords(tableName types.TableName, recordID string, since string) ([]types.OperationSync, error)
	PushRecord(op types.OperationSync) HttpResponse
	UpdateSyncState(state query.SyncState) error
	GetSyncState(tableName types.TableName) (query.SyncState, error)
}
