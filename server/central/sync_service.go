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
	case "insert", "update":
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

		err := s.queries.SyncUpsertBoard(ctx, centraldb.SyncUpsertBoardParams{
			ID:   payload.ID,
			Name: payload.Name,
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

	case "delete":
		err := s.queries.SyncDeleteBoard(ctx, centraldb.SyncDeleteBoardParams{
			ID: op.RecordID,
			UserID: pgtype.UUID{
				Bytes: userUUID,
				Valid: true,
			},
		})

		if err != nil {
			return fmt.Errorf("unable to delete board: %v", err)
		}

	default:
		return fmt.Errorf("%w: %s on boards", errUnsupportedOperation, op.OperationType)
	}

	return s.queries.CreateOperation(ctx, centraldb.CreateOperationParams{
		ID:            op.ID,
		TableName:     op.TableName,
		RecordID:      op.RecordID,
		OperationType: op.OperationType,
		DeviceID: pgtype.Text{
			String: op.DeviceID,
			Valid:  true,
		},
		Payload: op.Payload,
		CreatedAt: pgtype.Text{
			String: op.CreatedAt,
			Valid:  true,
		},
		UpdatedAt: pgtype.Text{
			String: op.UpdatedAt,
			Valid:  true,
		},
	})

}

func (s *SyncService) handleColumnOperation(ctx context.Context, userUUID uuid.UUID, op SyncOperation) error {
	switch strings.ToLower(op.OperationType) {
	case "insert", "update":
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

		if err != nil {
			return err
		}
	case "delete":
		err := s.queries.SyncDeleteColumn(ctx, centraldb.SyncDeleteColumnParams{
			ID: op.RecordID,
			UserID: pgtype.UUID{
				Bytes: userUUID,
				Valid: true,
			},
		})

		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("%w: %s on columns", errUnsupportedOperation, op.OperationType)
	}

	return s.queries.CreateOperation(ctx, centraldb.CreateOperationParams{
		ID:            op.ID,
		TableName:     op.TableName,
		RecordID:      op.RecordID,
		OperationType: op.OperationType,
		DeviceID: pgtype.Text{
			String: op.DeviceID,
			Valid:  true,
		},
		Payload: op.Payload,
		CreatedAt: pgtype.Text{
			String: op.CreatedAt,
			Valid:  true,
		},
		UpdatedAt: pgtype.Text{
			String: op.UpdatedAt,
			Valid:  true,
		},
	})
}

func (s *SyncService) handleCardOperation(ctx context.Context, userUUID uuid.UUID, op SyncOperation) error {
	switch strings.ToLower(op.OperationType) {
	case "insert", "update":
		var payload cardOperationPayload

		err := json.Unmarshal([]byte(op.Payload), &payload)
		if err != nil {
			fmt.Println("unable to unmarshal payload: ", err)
			return err
		}

		cardID := payload.Card.ID
		if cardID == "" {
			cardID = op.RecordID
		}
		if cardID == "" {
			return fmt.Errorf("card payload missing id")
		}

		columnID := payload.Column.ID
		if columnID == "" {
			columnID = payload.Column.ID
		}
		if columnID == "" {
			return fmt.Errorf("card payload missing column id")
		}

		// if err := s.ensureColumnOwnership(ctx, columnID, userUUID); err != nil {
		// 	return err
		// }

		createdAt := selectTimestamp(payload.Card.CreatedAt, op.CreatedAt)
		updatedAt := selectTimestamp(payload.Card.UpdatedAt, op.UpdatedAt)

		err = s.queries.SyncUpsertCard(ctx, centraldb.SyncUpsertCardParams{
			ID:       cardID,
			ColumnID: columnID,
			Title:    payload.Card.Name,
			Description: pgtype.Text{
				String: payload.Card.Description,
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

		if err != nil {
			return err
		}
	case "delete":
		err := s.queries.SyncDeleteCard(ctx, centraldb.SyncDeleteCardParams{
			ID: op.RecordID,
			UserID: pgtype.UUID{
				Bytes: userUUID,
				Valid: true,
			},
		})

		if err != nil {
			return err
		}
	case "update-card-column":
		var payload cardColumnPayload

		err := json.Unmarshal([]byte(op.Payload), &payload)
		if err != nil {
			return fmt.Errorf("unable to unmarhsal data: %v", err)
		}

		if payload.CardID == "" {
			payload.CardID = op.RecordID
		}
		if payload.CardID == "" || payload.NewColumn.ID == "" {
			return fmt.Errorf("card column payload missing identifiers")
		}

		// if err := s.ensureColumnOwnership(ctx, payload.NewColumn.ID, userUUID); err != nil {
		// 	return err
		// }

		updatedAt := selectTimestamp(op.UpdatedAt, op.CreatedAt)

		err = s.queries.SyncUpdateCardColumn(ctx, centraldb.SyncUpdateCardColumnParams{
			ColumnID: payload.NewColumn.ID,
			UpdatedAt: pgtype.Timestamptz{
				Time:  updatedAt,
				Valid: true,
			},
			ID:     payload.CardID,
			UserID: pgtype.UUID{Bytes: userUUID, Valid: true},
		})

		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("%w: %s on cards", errUnsupportedOperation, op.OperationType)
	}

	return s.queries.CreateOperation(ctx, centraldb.CreateOperationParams{
		ID:            op.ID,
		TableName:     op.TableName,
		RecordID:      op.RecordID,
		OperationType: op.OperationType,
		DeviceID: pgtype.Text{
			String: op.DeviceID,
			Valid:  true,
		},
		Payload: op.Payload,
		CreatedAt: pgtype.Text{
			String: op.CreatedAt,
			Valid:  true,
		},
		UpdatedAt: pgtype.Text{
			String: op.UpdatedAt,
			Valid:  true,
		},
	})
}

func (s *SyncService) handleTranscriptionOperation(ctx context.Context, userUUID uuid.UUID, op SyncOperation) error {
	switch strings.ToLower(op.OperationType) {
	case "insert", "update":
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

		createdAt := selectTimestamp(payload.CreatedAt, op.CreatedAt)
		updatedAt := selectTimestamp(payload.UpdatedAt, op.UpdatedAt)

		err := s.queries.SyncUpsertTranscription(ctx, centraldb.SyncUpsertTranscriptionParams{
			ID:            payload.ID,
			BoardID:       payload.BoardID,
			Transcription: payload.Transcription,
			RecordingPath: pgtype.Text{
				String: payload.RecordingPath,
				Valid:  true,
			},
			Intent: pgtype.Text{
				String: payload.Intent,
				Valid:  true,
			},
			AssistantResponse: pgtype.Text{
				String: payload.AssistantResponse,
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
		})

		if err != nil {
			return err
		}

	case "delete":
		err := s.queries.SyncDeleteTranscription(ctx, centraldb.SyncDeleteTranscriptionParams{
			ID: op.RecordID,
			UserID: pgtype.UUID{
				Bytes: userUUID,
				Valid: true,
			},
		})
		if err != nil {
			return err
		}

		return err
	default:
		return fmt.Errorf("%w: %s on transcriptions", errUnsupportedOperation, op.OperationType)
	}

	return s.queries.CreateOperation(ctx, centraldb.CreateOperationParams{
		ID:            op.ID,
		TableName:     op.TableName,
		RecordID:      op.RecordID,
		OperationType: op.OperationType,
		DeviceID: pgtype.Text{
			String: op.DeviceID,
			Valid:  true,
		},
		Payload: op.Payload,
		CreatedAt: pgtype.Text{
			String: op.CreatedAt,
			Valid:  true,
		},
		UpdatedAt: pgtype.Text{
			String: op.UpdatedAt,
			Valid:  true,
		},
	})
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
		s.pullBoardOperations(ctx, userUUID)
	case "columns":
		return s.pullColumnOperations(ctx, userUUID)
	case "cards":
		return s.pullCardOperations(ctx, userUUID)
	case "transcriptions":

	default:
		return nil, fmt.Errorf("unsupported table: %s", tableName)
	}

	return []SyncOperation{}, nil
}

func (s *SyncService) pullBoardOperations(ctx context.Context, userUUID uuid.UUID) ([]SyncOperation, error) {
	userOperations, err := s.queries.GetAllOperations(ctx, centraldb.GetAllOperationsParams{
		TableName:   "boards",
		TableName_2: "boards",
		TableName_3: "boards",
		UserID:      pgtype.UUID{Bytes: userUUID, Valid: true},
	})

	if err != nil {
		fmt.Println("Board error: ", err)
		return nil, fmt.Errorf("unable to get board operations: %v", err)
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

	// b, _ := json.MarshalIndent(operations, "", " ")
	// fmt.Println(string(b))

	return operations, nil
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
	b, _ := json.MarshalIndent(card, "", " ")
	fmt.Println(string(b))
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
		Title:    card.Title,
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
		LastSyncedOpID: syncState.LastSyncedOpID.String,
	}

	fmt.Println(resp)

	return resp, nil
}

func (s *SyncService) updateSyncState(ctx context.Context, userUUID uuid.UUID, payload types.SyncStatePayload) error {
	err := s.queries.UpdateSyncState(ctx, centraldb.UpdateSyncStateParams{
		UserID: pgtype.UUID{
			Bytes: userUUID,
			Valid: true,
		},
		LastSyncedAt: payload.LastSyncedAt,
		LastSyncedOpID: pgtype.Text{
			String: payload.LastSyncedOpID,
			Valid:  true,
		},
		TableName: payload.TableName,
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

type cardOperationPayload struct {
	Column struct {
		ID string `json:"id"`
	} `json:"column"`
	Card struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		ColumnID    string `json:"column_id"`
		Index       int    `json:"index"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
	}
}

type cardPayload struct {
	ID          string `json:"id"`
	ColumnID    string `json:"column_id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Attachments string `json:"attachments,omitempty"`
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
