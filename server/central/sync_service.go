package central

// TODO: implement soft delete for sync operations instead
// this service talks directly to the DB, which doesnt really make sense, there should be an intermediary, fix later

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
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

type ImportSummary struct {
	Boards         int `json:"boards"`
	Columns        int `json:"columns"`
	Cards          int `json:"cards"`
	Transcriptions int `json:"transcriptions"`
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

		boardID, err := uuid.Parse(payload.ID)
		if err != nil {
			return fmt.Errorf("unable to parse data type into uuid: %v", err)
		}

		id := pgtype.UUID{
			Bytes: boardID,
			Valid: true,
		}

		err = s.queries.SyncUpsertBoard(ctx, centraldb.SyncUpsertBoardParams{
			ID:   id,
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

		boardUUID, errr := uuid.Parse(payload.ID)
		if errr != nil {
			return fmt.Errorf("unable to convert board id to uuid: %v", err)
		}

		// TODO: would be best to use transactions to ensure atomicity between boards & board_memebers table

		if op.OperationType == "insert" {
			err = s.queries.InsertBoardMember(ctx, centraldb.InsertBoardMemberParams{
				BoardID: pgtype.UUID{
					Bytes: boardUUID,
					Valid: true,
				},
				UserID: pgtype.UUID{
					Bytes: userUUID,
					Valid: true,
				},
				Role: pgtype.Text{
					String: types.BoardOwnerRole.String(),
					Valid:  true,
				},
			})

			if err != nil {
				return fmt.Errorf("unable to create board member: %v", err)
			}
		}

	case "delete":
		boardID, err := uuid.Parse(op.RecordID)
		if err != nil {
			return fmt.Errorf("unable to parse data type into uuid: %v", err)
		}

		id := pgtype.UUID{
			Bytes: boardID,
			Valid: true,
		}
		err = s.queries.SyncDeleteBoard(ctx, centraldb.SyncDeleteBoardParams{
			ID: id,
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

		boardID, err := uuid.Parse(payload.BoardID)
		if err != nil {
			return fmt.Errorf("unable to parse board id into uuid: %v", err)
		}

		err = s.ensureBoardAccess(ctx, boardID, userUUID)
		if err != nil {
			return err
		}

		id := pgtype.UUID{
			Bytes: boardID,
			Valid: true,
		}

		err = s.queries.SyncUpsertColumn(ctx, centraldb.SyncUpsertColumnParams{
			ID:       payload.ID,
			BoardID:  id,
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

		b, _ := json.MarshalIndent(payload, "", " ")
		fmt.Println(string(b))

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

		column, err := s.queries.GetColumnByID(ctx, payload.Column.ID)
		if err != nil {
			return fmt.Errorf("card with column id (%s) doesnt exist: %v", payload.Column.ID, err)
		}

		boardID, err := uuid.FromBytes(column.BoardID.Bytes[:])
		if err != nil {
			return fmt.Errorf("unable to parse board id into uuid: %v", err)
		}

		err = s.ensureBoardAccess(ctx, boardID, userUUID)
		if err != nil {
			return err
		}

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

		b, _ := json.MarshalIndent(payload, "", " ")
		fmt.Println(string(b))

		if payload.CardID == "" {
			payload.CardID = op.RecordID
		}
		if payload.CardID == "" || payload.NewColumn.ID == "" {
			return fmt.Errorf("card column payload missing identifiers")
		}

		// if err := s.ensureColumnOwnership(ctx, payload.NewColumn.ID, userUUID); err != nil {
		// 	return err
		// }

		column, err := s.queries.GetColumnByID(ctx, payload.NewColumn.ID)
		if err != nil {
			return fmt.Errorf("card with column id (%s) doesnt exist: %v", payload.NewColumn.ID, err)
		}

		boardID, err := uuid.FromBytes(column.BoardID.Bytes[:])
		if err != nil {
			return fmt.Errorf("unable to parse board id into uuid: %v", err)
		}

		err = s.ensureBoardAccess(ctx, boardID, userUUID)
		if err != nil {
			return err
		}

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

		boardID, err := uuid.Parse(payload.ID)
		if err != nil {
			return fmt.Errorf("unable to convert string to uuid: %v", err)
		}

		createdAt := selectTimestamp(payload.CreatedAt, op.CreatedAt)
		updatedAt := selectTimestamp(payload.UpdatedAt, op.UpdatedAt)

		err = s.queries.SyncUpsertTranscription(ctx, centraldb.SyncUpsertTranscriptionParams{
			ID: payload.ID,
			BoardID: pgtype.UUID{
				Bytes: boardID,
				Valid: true,
			},
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

func (s *SyncService) upsertBoard(ctx context.Context, userUUID uuid.UUID, board boardPayload) error {
	createdAt, ok := parseTimestamp(board.CreatedAt)
	if !ok {
		return fmt.Errorf("unable to parse board created_at value %q", board.CreatedAt)
	}
	updatedAt, ok := parseTimestamp(board.UpdatedAt)
	if !ok {
		return fmt.Errorf("unable to parse board updated_at value %q", board.UpdatedAt)
	}

	boardID, err := uuid.Parse(board.ID)
	if err != nil {
		return fmt.Errorf("unable to parse data type into uuid: %v", err)
	}

	id := pgtype.UUID{
		Bytes: boardID,
		Valid: true,
	}

	if err := s.queries.SyncUpsertBoard(ctx, centraldb.SyncUpsertBoardParams{
		ID:   id,
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
	}); err != nil {
		return fmt.Errorf("unable to upsert board: %v", err)
	}

	boardUUID, err := uuid.Parse(board.ID)
	if err != nil {
		return fmt.Errorf("unable to parse board id into uuid: %v", err)
	}

	if err := s.queries.InsertBoardMember(ctx, centraldb.InsertBoardMemberParams{
		BoardID: pgtype.UUID{
			Bytes: boardUUID,
			Valid: true,
		},
		UserID: pgtype.UUID{
			Bytes: userUUID,
			Valid: true,
		},
		Role: pgtype.Text{
			String: types.BoardOwnerRole.String(),
			Valid:  true,
		},
	}); err != nil {
		return fmt.Errorf("unable to create board member: %v", err)
	}

	return nil
}

func (s *SyncService) upsertColumn(ctx context.Context, userUUID uuid.UUID, column columnPayload) error {
	createdAt, ok := parseTimestamp(column.CreatedAt)
	if !ok {
		return fmt.Errorf("unable to parse column created_at value %q", column.CreatedAt)
	}
	updatedAt, ok := parseTimestamp(column.UpdatedAt)
	if !ok {
		return fmt.Errorf("unable to parse column updated_at value %q", column.UpdatedAt)
	}

	boardID, err := uuid.Parse(column.BoardID)
	if err != nil {
		return fmt.Errorf("unable to parse board id into uuid: %v", err)
	}

	// TODO: errors should be returned as a value so it can return different error codes
	if err := s.ensureBoardAccess(ctx, boardID, userUUID); err != nil {
		return err
	}

	id := pgtype.UUID{
		Bytes: boardID,
		Valid: true,
	}

	if err := s.queries.SyncUpsertColumn(ctx, centraldb.SyncUpsertColumnParams{
		ID:       column.ID,
		BoardID:  id,
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
	}); err != nil {
		return fmt.Errorf("unable to upsert column: %v", err)
	}

	return nil
}

func (s *SyncService) upsertCard(ctx context.Context, userUUID uuid.UUID, card cardPayload) error {
	b, _ := json.MarshalIndent(card, "", " ")
	fmt.Println(string(b))
	createdAt, ok := parseTimestamp(card.CreatedAt)
	if !ok {
		return fmt.Errorf("unable to parse card created_at value %q", card.CreatedAt)
	}
	updatedAt, ok := parseTimestamp(card.UpdatedAt)
	if !ok {
		return fmt.Errorf("unable to parse card updated_at value %q", card.UpdatedAt)
	}

	column, err := s.queries.GetColumnByID(ctx, card.ColumnID)
	if err != nil {
		return fmt.Errorf("card with column id (%s) doesnt exist: %v", card.ColumnID, err)
	}

	boardID, err := uuid.FromBytes(column.BoardID.Bytes[:])
	if err != nil {
		return fmt.Errorf("unable to parse board id into uuid: %v", err)
	}

	if err := s.ensureBoardAccess(ctx, boardID, userUUID); err != nil {
		return err
	}

	if err := s.queries.SyncUpsertCard(ctx, centraldb.SyncUpsertCardParams{
		ID:       card.ID,
		ColumnID: card.ColumnID,
		Title:    card.Title,
		Description: func() pgtype.Text {
			if strings.TrimSpace(card.Description) == "" {
				return pgtype.Text{}
			}
			return pgtype.Text{String: card.Description, Valid: true}
		}(),
		Attachments: func() pgtype.Text {
			if strings.TrimSpace(card.Attachments) == "" {
				return pgtype.Text{}
			}
			return pgtype.Text{String: card.Attachments, Valid: true}
		}(),
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  updatedAt,
			Valid: true,
		},
	}); err != nil {
		return fmt.Errorf("unable to upsert card: %v", err)
	}

	return nil
}

func (s *SyncService) upsertTranscription(ctx context.Context, userUUID uuid.UUID, transcription transcriptionPayload) error {
	createdAt, ok := parseTimestamp(transcription.CreatedAt)
	if !ok {
		return fmt.Errorf("unable to parse transcription created_at value %q", transcription.CreatedAt)
	}
	updatedAt, ok := parseTimestamp(transcription.UpdatedAt)
	if !ok {
		return fmt.Errorf("unable to parse transcription updated_at value %q", transcription.UpdatedAt)
	}

	boardID, err := uuid.Parse(transcription.BoardID)
	if err != nil {
		return fmt.Errorf("unable to parse board id into uuid: %w", err)
	}

	if err := s.ensureBoardAccess(ctx, boardID, userUUID); err != nil {
		return err
	}

	recordingPath := pgtype.Text{}
	if strings.TrimSpace(transcription.RecordingPath) != "" {
		recordingPath = pgtype.Text{String: transcription.RecordingPath, Valid: true}
	}

	intent := pgtype.Text{}
	if strings.TrimSpace(transcription.Intent) != "" {
		intent = pgtype.Text{String: transcription.Intent, Valid: true}
	}

	assistantResponse := pgtype.Text{}
	if strings.TrimSpace(transcription.AssistantResponse) != "" {
		assistantResponse = pgtype.Text{String: transcription.AssistantResponse, Valid: true}
	}

	if err := s.queries.SyncUpsertTranscription(ctx, centraldb.SyncUpsertTranscriptionParams{
		ID: transcription.ID,
		BoardID: pgtype.UUID{
			Bytes: boardID,
			Valid: true,
		},
		Transcription:     transcription.Transcription,
		RecordingPath:     recordingPath,
		Intent:            intent,
		AssistantResponse: assistantResponse,
		CreatedAt: pgtype.Timestamptz{
			Time:  createdAt,
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamptz{
			Time:  updatedAt,
			Valid: true,
		},
	}); err != nil {
		return fmt.Errorf("unable to upsert transcription: %v", err)
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

// TODO: errors should be returned as a value so it can return different error codes
func (s *SyncService) ensureBoardAccess(ctx context.Context, boardID, userID uuid.UUID) error {
	hasAccess, err := s.queries.ValidateBoardAccess(ctx, centraldb.ValidateBoardAccessParams{
		BoardID: pgtype.UUID{Bytes: boardID, Valid: true},
		UserID:  pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("error validating board access: %v", err)
	}

	if !hasAccess {
		return fmt.Errorf("user %s does not have access to board %s", userID, boardID)
	}

	return nil
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

// These functions are not to be in sync service
func (s *SyncService) inviteUserToBoard(ctx context.Context, userID uuid.UUID, payload boardMemberActionPayload) error {
	boardID, err := uuid.Parse(payload.BoardID)
	if err != nil {
		return fmt.Errorf("unable to parse board id to uuid: %v", err)
	}
	err = s.ensureBoardOwner(ctx, boardID, userID)
	if err != nil {
		return err
	}

	user, err := s.queries.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		return fmt.Errorf("unable to get user by email %v: %v", payload.Email, err.Error())
	}

	id := pgtype.UUID{
		Bytes: boardID,
		Valid: true,
	}

	_, err = s.queries.GetBoardByID(ctx, id)
	if err != nil {
		return fmt.Errorf("unable to get board with id %s: %v", payload.BoardID, err)
	}

	err = s.queries.InsertBoardMember(ctx, centraldb.InsertBoardMemberParams{
		Role: pgtype.Text{
			String: types.BoardMemberRole.String(),
			Valid:  true,
		},
		UserID: user.ID,
		BoardID: pgtype.UUID{
			Bytes: boardID,
			Valid: true,
		},
	})

	if err != nil {
		return fmt.Errorf("unable to invite member to board: %v", err)
	}

	return nil
}

func (s *SyncService) removeUserFromBoard(ctx context.Context, userUUID uuid.UUID, payload boardMemberActionPayload) error {
	boardID, err := uuid.Parse(payload.BoardID)
	if err != nil {
		return fmt.Errorf("unable to parse board id to uuid: %v", err)
	}

	if err := s.ensureBoardOwner(ctx, boardID, userUUID); err != nil {
		return err
	}

	userToRemoveID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return err
	}

	targetUser, err := s.queries.GetUserByID(ctx, pgtype.UUID{
		Bytes: userToRemoveID,
		Valid: true,
	})

	if err != nil {
		return fmt.Errorf("unable to get user by email %s: %v", payload.Email, err)
	}

	err = s.queries.RemoveBoardMember(ctx, centraldb.RemoveBoardMemberParams{
		BoardID: pgtype.UUID{
			Bytes: boardID,
			Valid: true,
		},
		UserID: targetUser.ID,
	})
	if err != nil {
		return fmt.Errorf("unable to remove user from board: %v", err)
	}

	return nil
}

func (s *SyncService) getBoardMembers(ctx context.Context, boardID, userId uuid.UUID) ([]types.BoardMember, error) {
	if err := s.ensureBoardAccess(ctx, boardID, userId); err != nil {
		return nil, err
	}

	members, err := s.queries.ListBoardMembers(ctx, pgtype.UUID{Bytes: boardID, Valid: true})
	if err != nil {
		return nil, err
	}

	boardMembers := make([]types.BoardMember, len(members))

	//TODO: should include user email

	for i, member := range members {
		boardMembers[i] = types.BoardMember{
			UserID:   member.UserID.String(),
			Role:     member.Role.String,
			JoinedAt: member.JoinedAt.Time.String(),
			Email:    member.Email,
		}
	}

	return boardMembers, nil

}

func (s *SyncService) getBoardMetadata(ctx context.Context, boardID, userID uuid.UUID) (*types.BoardMetadata, error) {
	if err := s.ensureBoardAccess(ctx, boardID, userID); err != nil {
		return nil, err
	}

	metadata, err := s.queries.GetBoardMetadata(ctx, pgtype.UUID{Bytes: boardID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("unable to get board metadata: %v", err)
	}

	return &types.BoardMetadata{
		ID:                  metadata.ID.String(),
		Name:                metadata.Name,
		CreatedAt:           metadata.CreatedAt.Time.String(),
		UpdatedAt:           metadata.UpdatedAt.Time.String(),
		ColumnsCount:        int(metadata.ColumnsCount),
		CardsCount:          int(metadata.CardsCount),
		TranscriptionsCount: int(metadata.TranscriptionsCount),
	}, nil
}

func (s *SyncService) ensureBoardOwner(ctx context.Context, boardID, userID uuid.UUID) error {
	isOwner, err := s.queries.EnsureBoardOwner(ctx, centraldb.EnsureBoardOwnerParams{
		BoardID: pgtype.UUID{Bytes: boardID, Valid: true},
		UserID:  pgtype.UUID{Bytes: userID, Valid: true},
	})

	if err != nil {
		return fmt.Errorf("error checking board ownership: %v", err)
	}

	if !isOwner {
		return fmt.Errorf("user %s is not the owner of board %s", userID, boardID)
	}

	return nil
}

func (s *SyncService) ExportAllData(ctx context.Context, userID uuid.UUID, boardID string) (*types.ExportedData, error) {

	boardUUID, err := uuid.Parse(boardID)
	if err != nil {
		return nil, fmt.Errorf("unable to parse data type into uuid: %v", err)
	}

	id := pgtype.UUID{
		Bytes: boardUUID,
		Valid: true,
	}

	board, err := s.queries.GetBoardByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch board with id (%s): %v", boardID, err)
	}

	columns, err := s.queries.GetBoardColumns(ctx, centraldb.GetBoardColumnsParams{
		BoardID: id,
		UserID: pgtype.UUID{
			Bytes: userID,
			Valid: true,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("unable to fetch board columns with id (%s): %v", boardID, err)
	}

	exportedColumns := make([]types.ExportedColumn, len(columns))
	for i, c := range columns {
		exportedColumns[i] = types.ExportedColumn{
			ID:        c.ID,
			BoardID:   c.BoardID.String(),
			Name:      c.Name,
			Position:  int64(c.Position),
			CreatedAt: c.CreatedAt.Time.String(),
			UpdatedAt: c.UpdatedAt.Time.String(),
		}
	}

	cards, err := s.queries.ListBoardsCards(ctx, centraldb.ListBoardsCardsParams{
		UserID:  board.UserID,
		BoardID: id,
	})

	if err != nil {
		return nil, fmt.Errorf("unable to fetch board cards with id %s boards: %v", boardID, err)
	}

	exportedCards := make([]types.ExportedCard, len(cards))
	for i, card := range cards {
		exportedCards[i] = types.ExportedCard{
			ID:          card.ID,
			ColumnID:    card.ColumnID,
			Title:       card.Title,
			Description: card.Description.String,
			Attachments: card.Attachments.String,
			CreatedAt:   card.CreatedAt.Time.String(),
			UpdatedAt:   card.UpdatedAt.Time.String(),
		}
	}

	transcriptions, err := s.queries.ListBoardTranscriptions(ctx, centraldb.ListBoardTranscriptionsParams{
		Limit:  1000,
		Offset: 0,
		ID:     id,
		UserID: pgtype.UUID{
			Bytes: userID,
			Valid: true,
		}})

	if err != nil {
		return nil, fmt.Errorf("error exporting transcriptions: %v", err)
	}

	exportedTranscriptions := make([]types.ExportedTranscription, len(transcriptions))
	for i, t := range transcriptions {
		exportedTranscriptions[i] = types.ExportedTranscription{
			ID:                t.ID,
			BoardID:           id.String(),
			Transcription:     t.Transcription,
			RecordingPath:     t.RecordingPath.String,
			Intent:            t.Intent.String,
			AssistantResponse: t.AssistantResponse.String,
			CreatedAt:         t.CreatedAt.Time.String(),
			UpdatedAt:         t.UpdatedAt.Time.String(),
		}
	}

	exportedBoard := types.ExportedBoard{
		ID:        board.ID.String(),
		Name:      board.Name,
		CreatedAt: board.CreatedAt.Time.String(),
		UpdatedAt: board.UpdatedAt.Time.String(),
	}

	return &types.ExportedData{
		Board:          exportedBoard,
		Columns:        exportedColumns,
		Cards:          exportedCards,
		Transcriptions: exportedTranscriptions,
	}, nil
}

func (s *SyncService) getAppLatestVersion(ctx context.Context) (types.AppVersion, error) {
	version, err := s.queries.GetLatestAppVersion(ctx)
	if err != nil {
		return types.AppVersion{}, fmt.Errorf("unable to get app version: %v", err)
	}

	appVersion := types.AppVersion{
		Version: version.Version,
		Notes:   version.Notes.String,
		URL:     version.Url,
		Sha256:  version.Sha256.String,
	}

	return appVersion, nil
}

func (s *SyncService) createAppNewVersion(ctx context.Context, releaseUrl string) error {
	// eg url :// https://github.com/emeraldls/seisami/releases/download/test-build6/Seisami-test-build6-macos.zip
	// download build

	f, err := os.Create("release.zip")
	if err != nil {
		return fmt.Errorf("unable to create file: %v", err)
	}

	resp, err := http.Get(releaseUrl)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(f, resp.Body)

	if err != nil {
		return err
	}

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return err
	}

	return nil

	// hash := hasher.Sum(nil)
	// hex := hex.EncodeToString(hash[:])

	// 	v, err := s.queries.CreateAppVersion(ctx, centraldb.CreateAppVersionParams{
	//     Version: "1.0.5",
	//     URL: releaseURL,
	//     Notes: "Auto-uploaded from build pipeline",
	//     Sha256: hashHex,
	// })

	// compute sha456
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

type boardMemberActionPayload struct {
	Email   string `json:"email"`
	BoardID string `json:"board_id" validate:"required"`
	UserID  string `json:"user_id"`
}
