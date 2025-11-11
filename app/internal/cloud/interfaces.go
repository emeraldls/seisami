package cloud

import (
	"seisami/app/internal/repo/sqlc/query"
	"seisami/app/types"
)

type Cloud interface {
	GetAllOperations(tableName types.TableName, since int64) HttpResponse
	PullRecord(tableName types.TableName, since int64) HttpResponse
	PullRecords(tableName types.TableName, since int64) ([]types.OperationSync, error)
	PushRecord(op types.OperationSync) HttpResponse
	UpdateSyncState(state query.SyncState) HttpResponse
	GetSyncState(tableName types.TableName) HttpResponse

	UpsertBoard(types.ExportedBoard) HttpResponse
	UpsertColumn(types.ExportedColumn) HttpResponse
	UpsertCard(types.ExportedCard) HttpResponse
	InitializeSyncStateForUser() HttpResponse

	ImportBoardData(boardId string) HttpResponse
	UpdateSessionToken(token string)

	InitCloud() HttpResponse
	FetchAppVersion() HttpResponse

	ImportAllUserData() HttpResponse
}
