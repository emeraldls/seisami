package local

import (
	"encoding/json"
	"fmt"
	"seisami/app/internal/repo"
	"seisami/app/internal/repo/sqlc/query"
	"seisami/app/types"
)

type localFuncs struct {
	repo repo.Repository
}

func NewLocalFuncs(repo repo.Repository) localFuncs {
	return localFuncs{
		repo,
	}
}

func (lf localFuncs) GetAllOperations(tableName types.TableName) ([]types.OperationSync, error) {
	ops, err := lf.repo.GetAllOperations(tableName)
	if err != nil {
		return nil, err
	}

	var operationPayload = make([]types.OperationSync, 0)

	for _, op := range ops {
		payload := types.OperationSync{
			ID:            op.ID,
			OperationType: op.OperationType,
			TableName:     op.TableName,
			RecordID:      op.RecordID,
			DeviceID:      op.DeviceID.String,
			PayloadData:   op.Payload,
			CreatedAt:     op.CreatedAt.String,
			UpdatedAt:     op.UpdatedAt.String,
		}

		operationPayload = append(operationPayload, payload)
	}

	return operationPayload, nil
}

func (lf localFuncs) UpdateLocalDB(op types.OperationSync) error {
	// Parse the table name
	tableName, err := types.TableNameFromString(op.TableName)
	if err != nil {
		return fmt.Errorf("invalid table name: %v", err)
	}

	switch tableName {
	case types.BoardTable:
		return lf.updateBoardFromOperation(op)
	case types.ColumnTable:
		return lf.updateColumnFromOperation(op)
	case types.CardTable:
		return lf.updateCardFromOperation(op)
	case types.TranscriptionTable:
		return lf.updateTranscriptionFromOperation(op)
	default:
		return fmt.Errorf("unsupported table: %s", op.TableName)
	}
}

func (lf localFuncs) UpdateSyncState(state query.SyncState) error {
	tableName, err := types.TableNameFromString(state.TableName)
	if err != nil {
		return err
	}
	return lf.repo.UpsertSyncState(tableName, state.LastSyncedOpID, state.LastSyncedAt)
}

func (lf localFuncs) UpsertSyncState(state query.SyncState) error {
	tableName, err := types.TableNameFromString(state.TableName)
	if err != nil {
		return err
	}
	return lf.repo.UpsertSyncState(tableName, state.LastSyncedOpID, state.LastSyncedAt)
}

func (lf localFuncs) updateBoardFromOperation(op types.OperationSync) error {
	var payload struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	if err := json.Unmarshal([]byte(op.PayloadData), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal board payload: %v", err)
	}

	switch op.OperationType {
	case "insert", "update":

		_, err := lf.repo.ImportBoard(payload.ID, payload.Name, payload.CreatedAt, payload.UpdatedAt)
		return err
	case "delete":
		return lf.repo.DeleteBoard(payload.ID)
	default:
		return fmt.Errorf("unsupported operation type: %s", op.OperationType)
	}
}

func (lf localFuncs) updateColumnFromOperation(op types.OperationSync) error {
	var payload struct {
		ID        string `json:"id"`
		BoardID   string `json:"board_id"`
		Name      string `json:"name"`
		Position  int64  `json:"position"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	if err := json.Unmarshal([]byte(op.PayloadData), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal column payload: %v", err)
	}

	switch op.OperationType {
	case "insert", "update":
		_, err := lf.repo.ImportColumn(payload.ID, payload.BoardID, payload.Name, payload.Position, payload.CreatedAt, payload.UpdatedAt)
		return err
	case "delete":
		return lf.repo.DeleteColumn(payload.ID)
	default:
		return fmt.Errorf("unsupported operation type: %s", op.OperationType)
	}
}

func (lf localFuncs) updateCardFromOperation(op types.OperationSync) error {
	var payload struct {
		ID          string `json:"id"`
		ColumnID    string `json:"column_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Attachments string `json:"attachments"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
	}

	if err := json.Unmarshal([]byte(op.PayloadData), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal card payload: %v", err)
	}

	switch op.OperationType {
	case "insert", "update":
		_, err := lf.repo.ImportCard(payload.ID, payload.ColumnID, payload.Title, payload.Description, payload.Attachments, payload.CreatedAt, payload.UpdatedAt)
		return err
	case "delete":
		return lf.repo.DeleteCard(payload.ID)
	case "update-card-column":
		_, err := lf.repo.UpdateCardColumn(payload.ID, payload.ColumnID)
		return err
	default:
		return fmt.Errorf("unsupported operation type: %s", op.OperationType)
	}
}

func (lf localFuncs) updateTranscriptionFromOperation(op types.OperationSync) error {
	var payload struct {
		ID                string `json:"id"`
		BoardID           string `json:"board_id"`
		Transcription     string `json:"transcription"`
		RecordingPath     string `json:"recording_path"`
		Intent            string `json:"intent"`
		AssistantResponse string `json:"assistant_response"`
		CreatedAt         string `json:"created_at"`
		UpdatedAt         string `json:"updated_at"`
	}

	if err := json.Unmarshal([]byte(op.PayloadData), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal transcription payload: %v", err)
	}

	switch op.OperationType {
	case "insert", "update":
		_, err := lf.repo.ImportTranscription(payload.ID, payload.BoardID, payload.Transcription, payload.RecordingPath, payload.Intent, payload.AssistantResponse, payload.CreatedAt, payload.UpdatedAt)
		return err
	default:
		return fmt.Errorf("unsupported operation type: %s for transcriptions", op.OperationType)
	}
}
