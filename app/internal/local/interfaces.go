package local

import (
	"seisami/app/internal/repo/sqlc/query"
	"seisami/app/types"
)

type Local interface {
	GetAllOperations(tableName types.TableName) ([]types.OperationSync, error)
	UpdateSyncState(state query.SyncState) error
	// This would be an upsert query
	UpdateLocalDB(op types.OperationSync) error
}
