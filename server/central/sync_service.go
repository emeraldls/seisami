package central

// TODO: implement soft delete for sync operations instead
// this service talks directly to the DB, which doesnt really make sense, there should be an intermediary, fix later

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"seisami/server/centraldb"
	"seisami/server/types"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SyncService struct {
	pool    *pgxpool.Pool
	queries *centraldb.Queries
}

func NewSyncService(pool *pgxpool.Pool, queries *centraldb.Queries) *SyncService {

	return &SyncService{pool, queries}
}

type SyncOperation struct {
	ID            string `json:"id" validate:"required"`
	TableName     string `json:"table_name" validate:"required"`
	RecordID      string `json:"record_id" validate:"required"`
	OperationType string `json:"operation_type" validate:"required"`
	DeviceID      string `json:"device_id"`
	Payload       string `json:"payload"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type SyncStatus struct {
	Boards         int   `json:"boards"`
	Columns        int   `json:"columns"`
	Cards          int   `json:"cards"`
	Transcriptions int   `json:"transcriptions"`
	LastUpdated    int64 `json:"last_updated"`
}

var (
	errUnsupportedTable     = errors.New("unsupported sync table")
	errUnsupportedOperation = errors.New("unsupported sync operation")
)

func (s *SyncService) ProcessOperation(ctx context.Context, userID string, op SyncOperation) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}

	switch strings.ToLower(op.TableName) {
	case "boards":
		return s.handleBoardOperation(ctx, userUUID, op)
	case "columns":
		return s.handleColumnOperation(ctx, userUUID, op)
	case "cards":
		return s.handleCardOperation(ctx, userUUID, op)
	case "transcriptions":
		return s.handleTranscriptionOperation(ctx, userUUID, op)
	default:
		return fmt.Errorf("%w: %s", errUnsupportedTable, op.TableName)
	}
}

func (s *SyncService) handleBoardOperation(ctx context.Context, userUUID uuid.UUID, op SyncOperation) error {
	switch strings.ToLower(op.OperationType) {
	case "create", "update":
		var payload boardPayload
		if strings.TrimSpace(op.Payload) != "" {
			if err := json.Unmarshal([]byte(op.Payload), &payload); err != nil {
				return fmt.Errorf("decode board payload: %w", err)
			}
		}

		if payload.ID == "" {
			payload.ID = op.RecordID
		}
		if payload.ID == "" {
			return fmt.Errorf("board payload missing id")
		}
		if payload.Name == "" {
			payload.Name = "Untitled Board"
		}

		createdAt := selectTimestamp(payload.CreatedAt, op.CreatedAt)
		updatedAt := selectTimestamp(payload.UpdatedAt, op.UpdatedAt)

		_, err := s.pool.Exec(ctx, upsertBoardSQL, payload.ID, userUUID, payload.Name, createdAt, updatedAt)
		return err
	case "delete":
		_, err := s.pool.Exec(ctx, deleteBoardSQL, op.RecordID, userUUID)
		return err
	default:
		return fmt.Errorf("%w: %s on boards", errUnsupportedOperation, op.OperationType)
	}
}

func (s *SyncService) handleColumnOperation(ctx context.Context, userUUID uuid.UUID, op SyncOperation) error {
	switch strings.ToLower(op.OperationType) {
	case "create", "update":
		var payload columnPayload
		if strings.TrimSpace(op.Payload) == "" {
			return fmt.Errorf("column payload missing data")
		}
		if err := json.Unmarshal([]byte(op.Payload), &payload); err != nil {
			return fmt.Errorf("decode column payload: %w", err)
		}
		if payload.ID == "" {
			payload.ID = op.RecordID
		}
		if payload.ID == "" || payload.BoardID == "" {
			return fmt.Errorf("column payload missing identifiers")
		}

		if err := s.ensureBoardOwnership(ctx, payload.BoardID, userUUID); err != nil {
			return err
		}

		createdAt := selectTimestamp(payload.CreatedAt, op.CreatedAt)
		updatedAt := selectTimestamp(payload.UpdatedAt, op.UpdatedAt)

		err := s.queries.SyncUpsertColumn(ctx, centraldb.SyncUpsertColumnParams{
			ID:       payload.ID,
			BoardID:  payload.BoardID,
			Name:     payload.Name,
			Position: payload.Position,
			CreatedAt: pgtype.Timestamptz{
				Time:  createdAt,
				Valid: true,
			},
			UpdatedAt: pgtype.Timestamptz{
				Time:  updatedAt,
				Valid: true,
			},
		})

		return err
	case "delete":
		err := s.queries.SyncDeleteColumn(ctx, centraldb.SyncDeleteColumnParams{
			ID: op.RecordID,
			UserID: pgtype.UUID{
				Bytes: userUUID,
				Valid: true,
			},
		})

		return err
	default:
		return fmt.Errorf("%w: %s on columns", errUnsupportedOperation, op.OperationType)
	}
}

func (s *SyncService) handleCardOperation(ctx context.Context, userUUID uuid.UUID, op SyncOperation) error {
	switch strings.ToLower(op.OperationType) {
	case "create", "update":
		var payload cardPayload

		cardID := payload.ID
		if cardID == "" {
			cardID = op.RecordID
		}
		if cardID == "" {
			return fmt.Errorf("card payload missing id")
		}

		columnID := payload.ColumnID
		if columnID == "" {
			columnID = payload.ColumnID
		}
		if columnID == "" {
			return fmt.Errorf("card payload missing column id")
		}

		if err := s.ensureColumnOwnership(ctx, columnID, userUUID); err != nil {
			return err
		}

		createdAt := selectTimestamp(payload.CreatedAt, op.CreatedAt)
		updatedAt := selectTimestamp(payload.UpdatedAt, op.UpdatedAt)

		err := s.queries.SyncUpsertCard(ctx, centraldb.SyncUpsertCardParams{
			ID:       cardID,
			ColumnID: columnID,
			Title:    payload.Name,
			Description: pgtype.Text{
				String: payload.Description,
				Valid:  true,
			},
			CreatedAt: pgtype.Timestamptz{
				Time:  createdAt,
				Valid: true,
			},
			UpdatedAt: pgtype.Timestamptz{
				Time:  updatedAt,
				Valid: true,
			},
			// TODO: fix when attachments is working
		})

		return err
	case "delete":
		err := s.queries.SyncDeleteCard(ctx, centraldb.SyncDeleteCardParams{
			ID: op.RecordID,
			UserID: pgtype.UUID{
				Bytes: userUUID,
				Valid: true,
			},
		})

		return err
	case "update-card-column":
		var payload cardColumnPayload
		if payload.CardID == "" {
			payload.CardID = op.RecordID
		}
		if payload.CardID == "" || payload.NewColumn.ID == "" {
			return fmt.Errorf("card column payload missing identifiers")
		}

		if err := s.ensureColumnOwnership(ctx, payload.NewColumn.ID, userUUID); err != nil {
			return err
		}

		updatedAt := selectTimestamp(op.UpdatedAt, op.CreatedAt)

		err := s.queries.SyncUpdateCardColumn(ctx, centraldb.SyncUpdateCardColumnParams{
			ColumnID: payload.NewColumn.ID,
			UpdatedAt: pgtype.Timestamptz{
				Time:  updatedAt,
				Valid: true,
			},
			ID:     payload.CardID,
			UserID: pgtype.UUID{Bytes: userUUID, Valid: true},
		})

		return err
	default:
		return fmt.Errorf("%w: %s on cards", errUnsupportedOperation, op.OperationType)
	}
}

func (s *SyncService) handleTranscriptionOperation(ctx context.Context, userUUID uuid.UUID, op SyncOperation) error {
	switch strings.ToLower(op.OperationType) {
	case "create", "update":
		var payload transcriptionPayload
		if strings.TrimSpace(op.Payload) == "" {
			return fmt.Errorf("transcription payload missing data")
		}
		if err := json.Unmarshal([]byte(op.Payload), &payload); err != nil {
			return fmt.Errorf("decode transcription payload: %w", err)
		}
		if payload.ID == "" {
			payload.ID = op.RecordID
		}
		if payload.ID == "" || payload.BoardID == "" {
			return fmt.Errorf("transcription payload missing identifiers")
		}

		if err := s.ensureBoardOwnership(ctx, payload.BoardID, userUUID); err != nil {
			return err
		}

		createdAt := selectTimestamp(payload.CreatedAt, op.CreatedAt)
		updatedAt := selectTimestamp(payload.UpdatedAt, op.UpdatedAt)

		_, err := s.pool.Exec(ctx, upsertTranscriptionSQL,
			payload.ID,
			payload.BoardID,
			payload.Transcription,
			nullableString(payload.RecordingPath),
			nullableString(payload.Intent),
			nullableString(payload.AssistantResponse),
			createdAt,
			updatedAt,
		)
		return err
	case "delete":
		_, err := s.pool.Exec(ctx, deleteTranscriptionSQL, op.RecordID, userUUID)
		return err
	default:
		return fmt.Errorf("%w: %s on transcriptions", errUnsupportedOperation, op.OperationType)
	}
}

func (s *SyncService) PullOperations(ctx context.Context, userID, tableName string) ([]SyncOperation, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	validTables := map[string]bool{
		"boards": true, "columns": true, "cards": true, "transcriptions": true,
	}
	if !validTables[strings.ToLower(tableName)] {
		return nil, fmt.Errorf("invalid table name: %s", tableName)
	}

	switch strings.ToLower(tableName) {
	case "boards":

	case "columns":
		return s.pullColumnOperations(ctx, userUUID)
	case "cards":
		return s.pullCardOperations(ctx, userUUID)
	case "transcriptions":

	default:
		return nil, fmt.Errorf("unsupported table: %s", tableName)
	}

	return nil, fmt.Errorf("blah blah blah")
}

func (s *SyncService) pullColumnOperations(ctx context.Context, userUUID uuid.UUID) ([]SyncOperation, error) {

	userOperations, err := s.queries.GetAllOperations(ctx, centraldb.GetAllOperationsParams{
		TableName:   "columns",
		TableName_2: "columns",
		TableName_3: "columns",
		UserID:      pgtype.UUID{Bytes: userUUID, Valid: true},
	})

	if err != nil {
		return nil, fmt.Errorf("unable to get all operations: %v", err)
	}

	var operations []SyncOperation

	for _, userOp := range userOperations {
		var op = SyncOperation{
			ID:            userOp.ID,
			TableName:     userOp.TableName,
			RecordID:      userOp.RecordID,
			OperationType: userOp.OperationType,
			DeviceID:      userOp.DeviceID.String,
			Payload:       userOp.Payload,
			CreatedAt:     userOp.CreatedAt.String,
			UpdatedAt:     userOp.UpdatedAt.String,
		}

		operations = append(operations, op)
	}

	return operations, nil
}

func (s *SyncService) pullCardOperations(ctx context.Context, userUUID uuid.UUID) ([]SyncOperation, error) {
	userOperations, err := s.queries.GetAllOperations(ctx, centraldb.GetAllOperationsParams{
		TableName:   "cards",
		TableName_2: "cards",
		TableName_3: "cards",
		UserID:      pgtype.UUID{Bytes: userUUID, Valid: true},
	})

	if err != nil {
		return nil, fmt.Errorf("unable to get all operations: %v", err)
	}

	var operations []SyncOperation

	for _, userOp := range userOperations {
		var op = SyncOperation{
			ID:            userOp.ID,
			TableName:     userOp.TableName,
			RecordID:      userOp.RecordID,
			OperationType: userOp.OperationType,
			DeviceID:      userOp.DeviceID.String,
			Payload:       userOp.Payload,
			CreatedAt:     userOp.CreatedAt.String,
			UpdatedAt:     userOp.UpdatedAt.String,
		}

		operations = append(operations, op)
	}

	return operations, nil
}

func (s *SyncService) ensureBoardOwnership(ctx context.Context, boardID string, userUUID uuid.UUID) error {

	var exists bool
	if err := s.pool.QueryRow(ctx, boardOwnershipSQL, boardID, userUUID).Scan(&exists); err != nil {
		return fmt.Errorf("check board ownership: %w", err)
	}
	if !exists {
		return fmt.Errorf("board %s not found for user", boardID)
	}
	return nil
}

// This is only temporary, we will need a list of team members that has access to the column instead, likewise for board & card
func (s *SyncService) ensureColumnOwnership(ctx context.Context, columnID string, userUUID uuid.UUID) error {
	var exists bool
	if err := s.pool.QueryRow(ctx, columnOwnershipSQL, columnID, userUUID).Scan(&exists); err != nil {
		return fmt.Errorf("check column ownership: %w", err)
	}
	if !exists {
		return fmt.Errorf("column %s not found for user", columnID)
	}
	return nil
}

func (s *SyncService) initCloud(ctx context.Context, userUUID uuid.UUID) error {
	status, err := s.queries.GetCloudInitStatus(ctx, pgtype.UUID{Bytes: userUUID, Valid: true})
	if err != nil {
		return fmt.Errorf("unable to get cloud init status: %v", err)
	}

	if !status.Bool && status.Valid {
		err = s.queries.InitCloud(ctx, pgtype.UUID{Bytes: userUUID, Valid: true})
		if err != nil {
			return fmt.Errorf("unable to init cloud status: %v", err)
		}
	}

	return nil
}

func (s *SyncService) getCloudStatus(ctx context.Context, userUUID uuid.UUID) (bool, error) {
	status, err := s.queries.GetCloudInitStatus(ctx, pgtype.UUID{Bytes: userUUID, Valid: true})
	if err != nil {
		return false, fmt.Errorf("unable to get cloud init status: %v", err)
	}

	return status.Bool, nil
}

const layout = "2006-01-02 15:04:05"

func (s *SyncService) upsertBoard(ctx context.Context, userUUID uuid.UUID, board boardPayload) error {
	createdAt, err := time.Parse(layout, board.CreatedAt)
	if err != nil {
		return fmt.Errorf("unable to parse created_art into that layout: %v", err)
	}
	updatedAt, err := time.Parse(layout, board.UpdatedAt)
	if err != nil {
		return fmt.Errorf("unable to parse updated_at into that layout: %v", err)
	}
	err = s.queries.SyncUpsertBoard(ctx, centraldb.SyncUpsertBoardParams{
		ID:   board.ID,
		Name: board.Name,
		UserID: pgtype.UUID{
			Bytes: userUUID,
			Valid: true,
		},
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  updatedAt,
			Valid: true,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to upsert board: %v", err)
	}

	return nil
}

func (s *SyncService) upsertColumn(ctx context.Context, userUUID uuid.UUID, column columnPayload) error {
	createdAt, err := time.Parse(layout, column.CreatedAt)
	if err != nil {
		return fmt.Errorf("unable to parse created_art into that layout: %v", err)
	}
	updatedAt, err := time.Parse(layout, column.UpdatedAt)
	if err != nil {
		return fmt.Errorf("unable to parse updated_at into that layout: %v", err)
	}

	err = s.queries.SyncUpsertColumn(ctx, centraldb.SyncUpsertColumnParams{
		ID:       column.ID,
		BoardID:  column.BoardID,
		Name:     column.Name,
		Position: column.Position,
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  updatedAt,
			Valid: true,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to upsert column: %v", err)
	}

	return nil
}

func (s *SyncService) upsertCard(ctx context.Context, card cardPayload) error {
	createdAt, err := time.Parse(layout, card.CreatedAt)
	if err != nil {
		return fmt.Errorf("unable to parse created_art into that layout: %v", err)
	}
	updatedAt, err := time.Parse(layout, card.UpdatedAt)
	if err != nil {
		return fmt.Errorf("unable to parse updated_at into that layout: %v", err)
	}

	err = s.queries.SyncUpsertCard(ctx, centraldb.SyncUpsertCardParams{
		ID:       card.ID,
		ColumnID: card.ColumnID,
		Title:    card.Name,
		Description: pgtype.Text{
			String: card.Description,
			Valid:  true,
		},
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  updatedAt,
			Valid: true,
		},
		// TODO: attachments isnt here yet
	})

	if err != nil {
		return fmt.Errorf("unable to upsert card: %v", err)
	}

	return nil
}

func (s *SyncService) initializeSyncState(ctx context.Context, userUUID uuid.UUID) error {
	return s.queries.InitializeSyncStateForUser(ctx, pgtype.UUID{Bytes: userUUID, Valid: true})
}

func (s *SyncService) getSyncState(ctx context.Context, userUUID uuid.UUID, tableName string) (types.SyncStatePayload, error) {
	syncState, err := s.queries.GetSyncState(ctx, centraldb.GetSyncStateParams{
		TableName: tableName,
		UserID: pgtype.UUID{
			Bytes: userUUID,
			Valid: true,
		},
	})

	if err != nil {
		return types.SyncStatePayload{}, fmt.Errorf("error occured getting sync state: %v", err)
	}

	resp := types.SyncStatePayload{
		TableName:      syncState.TableName,
		LastSyncedAt:   syncState.LastSyncedAt,
		LastSyncedOpID: syncState.LastSyncedOpID,
	}

	return resp, nil
}

func (s *SyncService) updateSyncState(ctx context.Context, userUUID uuid.UUID, payload types.SyncStatePayload) error {
	err := s.queries.UpdateSyncState(ctx, centraldb.UpdateSyncStateParams{
		UserID: pgtype.UUID{
			Bytes: userUUID,
			Valid: true,
		},
		LastSyncedAt:   payload.LastSyncedAt,
		LastSyncedOpID: payload.LastSyncedOpID,
		TableName:      payload.TableName,
	})
	if err != nil {
		return fmt.Errorf("error occured updating sync state: %v", err)
	}

	return nil
}

func selectTimestamp(values ...string) time.Time {
	for _, value := range values {
		if t, ok := parseTimestamp(value); ok {
			return t
		}
	}
	return time.Now().UTC()
}

func parseTimestamp(value string) (time.Time, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, false
	}

	layouts := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, trimmed); err == nil {
			return t.UTC(), true
		}
	}

	if unix, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return time.Unix(unix, 0).UTC(), true
	}

	return time.Time{}, false
}

func nullableString(value string) interface{} {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func maxTime(values ...time.Time) time.Time {
	latest := time.Time{}
	for _, value := range values {
		if value.After(latest) {
			latest = value
		}
	}
	return latest
}

type boardPayload struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type columnPayload struct {
	ID        string `json:"id"`
	BoardID   string `json:"board_id"`
	Name      string `json:"name"`
	Position  int32  `json:"position"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type cardPayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ColumnID    string `json:"column_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type cardColumnPayload struct {
	CardID    string `json:"card_id"`
	NewColumn struct {
		ID string `json:"id"`
	} `json:"new_column"`
}

type transcriptionPayload struct {
	ID                string `json:"id"`
	BoardID           string `json:"board_id"`
	Transcription     string `json:"transcription"`
	RecordingPath     string `json:"recording_path"`
	Intent            string `json:"intent"`
	AssistantResponse string `json:"assistant_response"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

const (
	upsertBoardSQL = `
INSERT INTO boards (id, user_id, name, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    updated_at = EXCLUDED.updated_at
`

	deleteBoardSQL = `
DELETE FROM boards
WHERE id = $1
  AND user_id = $2
`

	upsertTranscriptionSQL = `
INSERT INTO transcriptions (id, board_id, transcription, recording_path, intent, assistant_response, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (id) DO UPDATE SET
    board_id = EXCLUDED.board_id,
    transcription = EXCLUDED.transcription,
    recording_path = EXCLUDED.recording_path,
    intent = EXCLUDED.intent,
    assistant_response = EXCLUDED.assistant_response,
    updated_at = EXCLUDED.updated_at
`

	deleteTranscriptionSQL = `
DELETE FROM transcriptions t
USING boards b
WHERE t.id = $1
  AND t.board_id = b.id
  AND b.user_id = $2
`

	boardOwnershipSQL = `
SELECT EXISTS (
    SELECT 1
    FROM boards
    WHERE id = $1
      AND user_id = $2
)
`

	columnOwnershipSQL = `
SELECT EXISTS (
    SELECT 1
    FROM columns c
    JOIN boards b ON b.id = c.board_id
    WHERE c.id = $1
      AND b.user_id = $2
)
`

	boardsCountSQL = `
SELECT COUNT(*)
FROM boards
WHERE user_id = $1
`

	columnsCountSQL = `
SELECT COUNT(*)
FROM columns c
JOIN boards b ON b.id = c.board_id
WHERE b.user_id = $1
`

	cardsCountSQL = `
SELECT COUNT(*)
FROM cards c
JOIN columns col ON col.id = c.column_id
JOIN boards b ON b.id = col.board_id
WHERE b.user_id = $1
`

	transcriptionsCountSQL = `
SELECT COUNT(*)
FROM transcriptions t
JOIN boards b ON b.id = t.board_id
WHERE b.user_id = $1
`

	boardsLastUpdatedSQL = `
SELECT COALESCE(MAX(updated_at), to_timestamp(0))
FROM boards
WHERE user_id = $1
`

	columnsLastUpdatedSQL = `
SELECT COALESCE(MAX(c.updated_at), to_timestamp(0))
FROM columns c
JOIN boards b ON b.id = c.board_id
WHERE b.user_id = $1
`

	cardsLastUpdatedSQL = `
SELECT COALESCE(MAX(c.updated_at), to_timestamp(0))
FROM cards c
JOIN columns col ON col.id = c.column_id
JOIN boards b ON b.id = col.board_id
WHERE b.user_id = $1
`

	transcriptionsLastUpdatedSQL = `
SELECT COALESCE(MAX(t.updated_at), to_timestamp(0))
FROM transcriptions t
JOIN boards b ON b.id = t.board_id
WHERE b.user_id = $1
`
)
