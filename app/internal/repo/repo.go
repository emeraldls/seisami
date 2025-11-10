package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"seisami/app/internal/repo/sqlc/query"
	"seisami/app/types"

	_ "embed"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed sqlc/schema.sql
var Schema string

type repo struct {
	queries *query.Queries
	ctx     context.Context
}

func NewRepo(db *sql.DB, ctx context.Context) *repo {
	queries := query.New(db)

	return &repo{
		queries,
		ctx,
	}
}

func (r *repo) GetBoard(boardId string) (query.Board, error) {
	board, err := r.queries.GetBoard(r.ctx, boardId)
	if err != nil {
		return query.Board{}, fmt.Errorf("error occured fetching board: (%s)... %v", boardId, err)
	}

	return board, nil
}

func (r *repo) CreateBoard(name string) (query.Board, error) {
	id := uuid.New().String()
	board, err := r.queries.CreateBoard(r.ctx, query.CreateBoardParams{ID: id, Name: name})
	if err != nil {
		return query.Board{}, fmt.Errorf("error occured creating board: %v", err)
	}

	return board, nil
}

func (r *repo) DeleteBoard(boardId string) error {
	if err := r.queries.DeleteBoard(r.ctx, boardId); err != nil {
		fmt.Println("deleting board...", boardId)
		return fmt.Errorf("error occured deleting board: (%s)...%v", boardId, err)
	}

	return nil
}

func (r *repo) GetAllBoards(page int64, pageSize int64) ([]query.Board, error) {
	boards, err := r.queries.ListBoards(r.ctx, query.ListBoardsParams{Limit: 10, Offset: (page - 1) * pageSize})
	if err != nil {
		return nil, fmt.Errorf("error occured getting all boards: %v", err)
	}

	if boards == nil {
		return []query.Board{}, nil
	}

	return boards, nil
}

func (r *repo) UpdateBoard(boardId string, name string) (query.Board, error) {
	return r.queries.UpdateBoard(r.ctx, query.UpdateBoardParams{ID: boardId, Name: name})
}

func (r *repo) CreateColumn(boardId string, columnName string) (query.Column, error) {
	id := uuid.New().String()

	columns, err := r.ListColumnsByBoard(boardId)
	if err != nil {
		return query.Column{}, fmt.Errorf("error getting columns to determine position: %v", err)
	}
	position := len(columns)

	column, err := r.queries.CreateColumn(r.ctx, query.CreateColumnParams{
		ID:       id,
		BoardID:  boardId,
		Name:     columnName,
		Position: int64(position),
	})
	if err != nil {
		return query.Column{}, fmt.Errorf("error creating column: %v", err)
	}
	return column, nil
}

func (r *repo) DeleteColumn(columnId string) error {
	if err := r.queries.DeleteColumn(r.ctx, columnId); err != nil {
		return fmt.Errorf("error deleting column: %v", err)
	}
	return nil
}

func (r *repo) GetColumn(columnId string) (query.Column, error) {
	column, err := r.queries.GetColumn(r.ctx, columnId)
	if err != nil {
		return query.Column{}, fmt.Errorf("error getting column: %v", err)
	}
	return column, nil
}

func (r *repo) ListColumnsByBoard(boardId string) ([]query.Column, error) {
	columns, err := r.queries.ListColumnsByBoard(r.ctx, boardId)
	if err != nil {
		return nil, fmt.Errorf("error listing columns by board: %v", err)
	}
	if columns == nil {
		return []query.Column{}, nil
	}
	return columns, nil
}

func (r *repo) UpdateColumn(columnId string, name string) (query.Column, error) {
	column, err := r.queries.UpdateColumn(r.ctx, query.UpdateColumnParams{
		ID:   columnId,
		Name: name,
	})
	if err != nil {
		return query.Column{}, fmt.Errorf("error updating column: %v", err)
	}
	return column, nil
}

func (r *repo) CreateCard(columnId string, title string, description string) (query.Card, error) {
	id := uuid.New().String()
	card, err := r.queries.CreateCard(r.ctx, query.CreateCardParams{
		ID:       id,
		ColumnID: columnId,
		Title:    title,
		Description: sql.NullString{
			String: description,
			Valid:  true,
		},
		// I'll make this files names separated by comma or space
		Attachments: sql.NullString{
			String: "",
			Valid:  true,
		},
	})
	if err != nil {
		return query.Card{}, fmt.Errorf("error creating card: %v", err)
	}
	return card, nil
}

func (r *repo) DeleteCard(cardId string) error {
	if err := r.queries.DeleteCard(r.ctx, cardId); err != nil {
		return fmt.Errorf("error deleting card: %v", err)
	}
	return nil
}

func (r *repo) GetCard(cardId string) (query.Card, error) {
	card, err := r.queries.GetCard(r.ctx, cardId)
	if err != nil {
		return query.Card{}, fmt.Errorf("error getting card: %v", err)
	}
	return card, nil
}

func (r *repo) ListCardsByColumn(columnId string) ([]query.Card, error) {
	cards, err := r.queries.ListCardsByColumn(r.ctx, columnId)
	if err != nil {
		return nil, fmt.Errorf("error listing cards by column: %v", err)
	}

	if cards == nil {
		return []query.Card{}, nil
	}
	return cards, nil
}

func (r *repo) UpdateCard(cardId string, title string, description string) (query.Card, error) {
	card, err := r.queries.UpdateCard(r.ctx, query.UpdateCardParams{
		ID:    cardId,
		Title: title,
		Description: sql.NullString{
			String: description,
			Valid:  true,
		},
	})
	if err != nil {
		return query.Card{}, fmt.Errorf("error updating card: %v", err)
	}
	return card, nil
}

func (r *repo) UpdateCardColumn(cardId string, columnId string) (query.Card, error) {
	card, err := r.queries.UpdateCardColumn(r.ctx, query.UpdateCardColumnParams{
		ID:       cardId,
		ColumnID: columnId,
	})
	if err != nil {
		return query.Card{}, fmt.Errorf("error updating card column: %v", err)
	}
	return card, nil
}

func (r *repo) AddTransscription(boardId string, transcription string, recordingPath string) (query.Transcription, error) {
	Id := uuid.New().String()
	data, err := r.queries.CreateTranscription(r.ctx, query.CreateTranscriptionParams{
		ID:            Id,
		BoardID:       boardId,
		Transcription: transcription,
		RecordingPath: sql.NullString{String: recordingPath, Valid: true},
	})
	if err != nil {
		return query.Transcription{}, fmt.Errorf("unable to create transcription: %w", err)
	}

	return data, nil
}

func (r *repo) GetTranscriptions(boardId string, page, pageSize int64) ([]query.Transcription, error) {
	transcriptions, err := r.queries.ListTranscriptionsByBoard(r.ctx, boardId)
	if err != nil {
		return nil, err
	}

	if transcriptions == nil {
		return []query.Transcription{}, nil
	}

	return transcriptions, nil

}

func (r *repo) GetTranscriptionByID(transcriptionId string) (query.Transcription, error) {
	transcription, err := r.queries.GetTranscription(r.ctx, transcriptionId)
	if err != nil {
		return query.Transcription{}, fmt.Errorf("unable to get transcription: %w", err)
	}

	return transcription, nil
}

func (r *repo) UpdateTranscriptionIntent(transcriptionId string, intent string) error {
	_, err := r.queries.UpdateTranscriptionIntent(r.ctx, query.UpdateTranscriptionIntentParams{
		ID: transcriptionId,
		Intent: sql.NullString{
			String: intent,
			Valid:  true,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to update transcription intent: %w", err)
	}
	return nil
}

func (r *repo) UpdateTranscriptionResponse(transcriptionId string, response string) error {
	_, err := r.queries.UpdateTranscriptionResponse(r.ctx, query.UpdateTranscriptionResponseParams{
		ID: transcriptionId,
		AssistantResponse: sql.NullString{
			String: response,
			Valid:  true,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to update transcription response: %w", err)
	}
	return nil
}

func (r *repo) GetSettings() (query.Setting, error) {
	settings, err := r.queries.GetSettings(r.ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return query.Setting{
				ID:                  1,
				TranscriptionMethod: "cloud",
			}, nil
		}
		return query.Setting{}, fmt.Errorf("unable to get settings: %w", err)
	}
	return settings, nil
}

func (r *repo) CreateOrUpdateSettings(transcriptionMethod string, whisperBinaryPath *string, whisperModelPath *string, openaiApiKey *string) (query.Setting, error) {
	_, err := r.queries.GetSettings(r.ctx)

	var binaryPath, modelPath, apiKey sql.NullString

	if whisperBinaryPath != nil {
		binaryPath = sql.NullString{String: *whisperBinaryPath, Valid: true}
	}
	if whisperModelPath != nil {
		modelPath = sql.NullString{String: *whisperModelPath, Valid: true}
	}
	if openaiApiKey != nil {
		apiKey = sql.NullString{String: *openaiApiKey, Valid: true}
	}

	if err == sql.ErrNoRows {
		return r.queries.CreateSettings(r.ctx, query.CreateSettingsParams{
			TranscriptionMethod: transcriptionMethod,
			WhisperBinaryPath:   binaryPath,
			WhisperModelPath:    modelPath,
			OpenaiApiKey:        apiKey,
		})
	} else if err != nil {
		return query.Setting{}, fmt.Errorf("unable to check existing settings: %w", err)
	}

	return r.queries.UpdateSettings(r.ctx, query.UpdateSettingsParams{
		TranscriptionMethod: transcriptionMethod,
		WhisperBinaryPath:   binaryPath,
		WhisperModelPath:    modelPath,
		OpenaiApiKey:        apiKey,
	})
}

func (r *repo) SearchColumnsByBoardAndName(boardId, searchQuery string) ([]query.Column, error) {
	columns, err := r.queries.SearchColumnsByBoardAndName(r.ctx, query.SearchColumnsByBoardAndNameParams{
		BoardID: boardId,
		Column2: sql.NullString{
			String: searchQuery,
			Valid:  true,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("unable to search board for that column name: %v", err)
	}

	if columns == nil {
		return []query.Column{}, nil
	}

	return columns, nil
}

/*
	When syncing data, we should filter by record_id & select the most recent record for that particular id within the timestamp
*/

func (r *repo) CreateOperation(tableName types.TableName, recordId, payload string, opType types.Operation) (query.Operation, error) {
	id := uuid.New().String()
	operation, err := r.queries.CreateOperation(r.ctx, query.CreateOperationParams{
		ID:            id,
		OperationType: opType.String(),
		TableName:     tableName.String(),
		RecordID:      recordId,
		Payload:       payload,
	})

	if err != nil {
		return query.Operation{}, fmt.Errorf("unable to create operation: %v", err)
	}

	return operation, nil
}

// Operations has to be uploaded to the cloud

/* we're out of version if local sync_state last_synced_at or last_synced_op_id is not the same with cloud sync_state
then we pull the cloud operations based on it last_synced_at & local operations based on local last_synced_at

This is where the hard problem lies, validating each row

// This will store only
changed_operations = []
unchanged_operations = []


This would only work if both lengths are equal

loop over both cloud & local
if cloud.op_id == local.op_id {
	push into unchanged
} else {
	push into changed. even index would be cloud, odd would be local
}

what if i checked for the one with higher length, cloud or local
then the one with shorter length is inner & i check if it exist in the one with higher length

or how tf do i do this, im confused
*/

// the timestamp validation with sync is in the sqlc query
func (r *repo) GetAllOperations(tableName types.TableName) ([]query.Operation, error) {
	ops, err := r.queries.GetAllOperations(r.ctx, query.GetAllOperationsParams{TableName: tableName.String(), TableName_2: tableName.String(), TableName_3: tableName.String()})
	if err != nil {
		return nil, err
	}

	return ops, nil
}

func (r *repo) UpsertSyncState(tableName types.TableName, lastOpID string, lastSyncedAt int64) error {
	return r.queries.UpsertSyncState(r.ctx, query.UpsertSyncStateParams{
		TableName:      tableName.String(),
		LastSyncedAt:   lastSyncedAt,
		LastSyncedOpID: lastOpID,
	})
}

func (r *repo) GetSyncState(tableName types.TableName) (query.SyncState, error) {
	syncState, err := r.queries.GetSyncState(r.ctx, tableName.String())
	if err != nil {
		return query.SyncState{}, fmt.Errorf("unable to get sync state: %v", err)
	}

	return syncState, nil
}

func (r *repo) UpdateSyncState(tableName types.TableName, lastOpID string, lastSyncedAt int64) error {
	err := r.queries.UpdateSyncState(r.ctx, query.UpdateSyncStateParams{
		TableName:      tableName.String(),
		LastSyncedAt:   lastSyncedAt,
		LastSyncedOpID: lastOpID,
	})

	if err != nil {
		return fmt.Errorf("unable to update sync state: %v", err)
	}

	return nil
}

func (r *repo) ExportAllData() (*types.ExportedData, error) {
	boards, err := r.queries.ListBoards(r.ctx, query.ListBoardsParams{Limit: 1000, Offset: 0})
	if err != nil {
		return nil, fmt.Errorf("error exporting boards: %v", err)
	}

	exportedBoards := make([]types.ExportedBoard, len(boards))
	for i, b := range boards {
		exportedBoards[i] = types.ExportedBoard{
			ID:        b.ID,
			Name:      b.Name,
			CreatedAt: b.CreatedAt.String,
			UpdatedAt: b.UpdatedAt.String,
		}
	}

	columns, err := r.queries.ListAllColumns(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("error exporting columns: %v", err)
	}

	exportedColumns := make([]types.ExportedColumn, len(columns))
	for i, c := range columns {
		exportedColumns[i] = types.ExportedColumn{
			ID:        c.ID,
			BoardID:   c.BoardID,
			Name:      c.Name,
			Position:  c.Position,
			CreatedAt: c.CreatedAt.String,
			UpdatedAt: c.UpdatedAt.String,
		}
	}

	// Get all cards
	cards, err := r.queries.ListAllCards(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("error exporting cards: %v", err)
	}

	exportedCards := make([]types.ExportedCard, len(cards))
	for i, card := range cards {
		exportedCards[i] = types.ExportedCard{
			ID:          card.ID,
			ColumnID:    card.ColumnID,
			Title:       card.Title,
			Description: card.Description.String,
			Attachments: card.Attachments.String,
			CreatedAt:   card.CreatedAt.String,
			UpdatedAt:   card.UpdatedAt.String,
		}
	}

	transcriptions, err := r.queries.ListAllTranscriptions(r.ctx, query.ListAllTranscriptionsParams{Limit: 1000, Offset: 0})
	if err != nil {
		return nil, fmt.Errorf("error exporting transcriptions: %v", err)
	}

	exportedTranscriptions := make([]types.ExportedTranscription, len(transcriptions))
	for i, t := range transcriptions {
		exportedTranscriptions[i] = types.ExportedTranscription{
			ID:                t.ID,
			BoardID:           t.BoardID,
			Transcription:     t.Transcription,
			RecordingPath:     t.RecordingPath.String,
			Intent:            t.Intent.String,
			AssistantResponse: t.AssistantResponse.String,
			CreatedAt:         t.CreatedAt.String,
			UpdatedAt:         t.UpdatedAt.String,
		}
	}

	return &types.ExportedData{
		Boards:         exportedBoards,
		Columns:        exportedColumns,
		Cards:          exportedCards,
		Transcriptions: exportedTranscriptions,
	}, nil
}

func (r *repo) ImportBoard(id, name, createdAt, updatedAt string) (query.Board, error) {
	board, err := r.queries.ImportBoard(r.ctx, query.ImportBoardParams{
		ID:        id,
		Name:      name,
		CreatedAt: sql.NullString{String: createdAt, Valid: true},
		UpdatedAt: sql.NullString{String: updatedAt, Valid: true},
	})
	if err != nil {
		return query.Board{}, fmt.Errorf("unable to import board: %v", err)
	}
	return board, nil
}

func (r *repo) ImportColumn(id, boardId, name string, position int64, createdAt, updatedAt string) (query.Column, error) {
	column, err := r.queries.ImportColumn(r.ctx, query.ImportColumnParams{
		ID:        id,
		BoardID:   boardId,
		Name:      name,
		Position:  position,
		CreatedAt: sql.NullString{String: createdAt, Valid: true},
		UpdatedAt: sql.NullString{String: updatedAt, Valid: true},
	})
	if err != nil {
		return query.Column{}, fmt.Errorf("unable to import column: %v", err)
	}
	return column, nil
}

func (r *repo) ImportCard(id, columnId, title, description, attachments, createdAt, updatedAt string) (query.Card, error) {
	card, err := r.queries.ImportCard(r.ctx, query.ImportCardParams{
		ID:       id,
		ColumnID: columnId,
		Title:    title,
		Description: sql.NullString{
			String: description,
			Valid:  description != "",
		},
		Attachments: sql.NullString{
			String: attachments,
			Valid:  attachments != "",
		},
		CreatedAt: sql.NullString{String: createdAt, Valid: true},
		UpdatedAt: sql.NullString{String: updatedAt, Valid: true},
	})
	if err != nil {
		return query.Card{}, fmt.Errorf("unable to import card: %v", err)
	}
	return card, nil
}

func (r *repo) ImportTranscription(id, boardId, transcription, recordingPath, intent, assistantResponse, createdAt, updatedAt string) (query.Transcription, error) {
	t, err := r.queries.ImportTranscription(r.ctx, query.ImportTranscriptionParams{
		ID:            id,
		BoardID:       boardId,
		Transcription: transcription,
		RecordingPath: sql.NullString{
			String: recordingPath,
			Valid:  recordingPath != "",
		},
		Intent: sql.NullString{
			String: intent,
			Valid:  intent != "",
		},
		AssistantResponse: sql.NullString{
			String: assistantResponse,
			Valid:  assistantResponse != "",
		},
		CreatedAt: sql.NullString{String: createdAt, Valid: true},
		UpdatedAt: sql.NullString{String: updatedAt, Valid: true},
	})
	if err != nil {
		return query.Transcription{}, fmt.Errorf("unable to import transcription: %v", err)
	}
	return t, nil
}

func (r *repo) UpdateLocalVersion(version string) error {
	return r.queries.UpsertAppMeta(r.ctx, query.UpsertAppMetaParams{
		Key: "local_version",
		Value: sql.NullString{
			String: version,
			Valid:  true,
		},
	})
}

func (r *repo) GetLocalVersion() (string, error) {
	meta, err := r.queries.GetAppMeta(r.ctx, "local_version")
	localVersion := meta.String
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("no local version found, setting to 0.0.0")
			localVersion = "0.0.0"

			if err := r.UpdateLocalVersion(localVersion); err != nil {

				return "", fmt.Errorf("failed to set default version: %v", err)
			}

		} else {

			return "", fmt.Errorf("unable to get local version: %v", err)
		}
	}

	return localVersion, nil

}
