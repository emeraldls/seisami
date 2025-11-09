package cloud

import (
	"seisami/app/internal/repo/sqlc/query"
	"seisami/app/types"
)

type Cloud interface {
	GetAllOperations(tableName types.TableName, since int64) ([]types.OperationSync, error)
	PullRecord(tableName types.TableName, since int64) (types.OperationSync, error)
	PullRecords(tableName types.TableName, since int64) ([]types.OperationSync, error)
	PushRecord(op types.OperationSync) HttpResponse
	UpdateSyncState(state query.SyncState) error
	GetSyncState(tableName types.TableName) (query.SyncState, error)

	UpsertBoard(types.ExportedBoard) error
	UpsertColumn(types.ExportedColumn) error
	UpsertCard(types.ExportedCard) error
	InitializeSyncStateForUser() error

	ImportBoardData(boardId string) (types.ImportUserBoardData, error)
	UpdateSessionToken(token string)

	InitCloud() error
	FetchAppVersion() (types.AppVersion, error)

	ImportAllUserData() (types.ExportedData, error)
}
