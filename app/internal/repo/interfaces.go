package repo

import (
	"seisami/app/internal/repo/sqlc/query"
	"seisami/app/types"
)

type Repository interface {
	CreateBoard(name string) (query.Board, error)
	DeleteBoard(id string) error
	GetBoard(id string) (query.Board, error)
	// will update this later to include query params
	GetAllBoards(page int64, pageSize int64) ([]query.Board, error)
	UpdateBoard(id string, name string) (query.Board, error)

	CreateColumn(boardId string, columnName string) (query.Column, error)
	DeleteColumn(id string) error
	GetColumn(id string) (query.Column, error)
	ListColumnsByBoard(boardId string) ([]query.Column, error)
	UpdateColumn(id string, name string) (query.Column, error)

	CreateCard(columnId string, title string, description string) (query.Card, error)
	DeleteCard(id string) error
	GetCard(id string) (query.Card, error)
	ListCardsByColumn(columnId string) ([]query.Card, error)
	UpdateCard(id string, title string, description string) (query.Card, error)
	UpdateCardColumn(CardId string, columnId string) (query.Card, error)

	AddTransscription(boardId string, transcription string, recordingPath string) (query.Transcription, error)
	GetTranscriptions(boardId string, page, pageSize int64) ([]query.Transcription, error)
	GetTranscriptionByID(transcriptionId string) (query.Transcription, error)
	UpdateTranscriptionIntent(transcriptionId string, intent string) error
	UpdateTranscriptionResponse(transcriptionId string, response string) error

	GetSettings() (query.Setting, error)
	CreateOrUpdateSettings(transcriptionMethod string, whisperBinaryPath *string, whisperModelPath *string, openaiApiKey *string) (query.Setting, error)

	SearchColumnsByBoardAndName(boardId, searchQuery string) ([]query.Column, error)

	CreateOperation(tableName types.TableName, recordId, payload string, opType types.Operation) (query.Operation, error)
	GetAllOperations(tableName types.TableName) ([]query.Operation, error)
	UpsertSyncState(tableName types.TableName, lastOpID string, lastSyncedAt int64) error
	GetSyncState(tableName types.TableName) (query.SyncState, error)
	UpdateSyncState(tableName types.TableName, lastOpID string, lastSyncedAt int64) error

	ExportAllData() (*types.ExportedData, error)

	ImportBoard(id, name, createdAt, updatedAt string) (query.Board, error)
	ImportColumn(id, boardId, name string, position int64, createdAt, updatedAt string) (query.Column, error)
	ImportCard(id, columnId, title, description, attachments, createdAt, updatedAt string) (query.Card, error)
	ImportTranscription(id, boardId, transcription, recordingPath, intent, assistantResponse, createdAt, updatedAt string) (query.Transcription, error)
}
