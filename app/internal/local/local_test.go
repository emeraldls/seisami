package local

import (
	"context"
	"database/sql"
	"encoding/json"
	"seisami/app/internal/repo"
	"seisami/app/types"
	"testing"

	_ "embed"
)

func setupTestDB(t *testing.T) repo.Repository {

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	if _, err := db.Exec(repo.Schema); err != nil {
		t.Fatalf("failed to exec schema: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return repo.NewRepo(db, context.Background())
}

func TestLocal(t *testing.T) {
	t.Run("get_all_operations_empty", func(t *testing.T) {
		repo := setupTestDB(t)
		lf := NewLocalFuncs(repo)

		ops, err := lf.GetAllOperations(types.BoardTable)
		if err != nil {
			t.Fatalf("GetAllOperations failed: %v", err)
		}

		if len(ops) != 0 {
			t.Fatalf("expected 0 operations, got %d", len(ops))
		}
	})

	type boardOperationPayload struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	t.Run("get_all_operations", func(t *testing.T) {
		repo := setupTestDB(t)
		lf := NewLocalFuncs(repo)

		payload := boardOperationPayload{
			ID:        "1234",
			Name:      "Test Board",
			CreatedAt: "2023-01-01T00:00:00Z",
			UpdatedAt: "2023-01-01T00:00:00Z",
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}

		if _, err := repo.CreateOperation(types.BoardTable, "1234", string(payloadBytes), types.InsertOperation); err != nil {
			t.Fatalf("CreateOperation failed: %v", err)
		}

		ops, err := lf.GetAllOperations(types.BoardTable)
		if err != nil {
			t.Fatalf("GetAllOperations failed: %v", err)
		}

		if len(ops) != 1 {
			t.Fatalf("expected 1 operation, got %d", len(ops))
		}

		gottenOperation := ops[0]
		var gottenPayload boardOperationPayload
		if err := json.Unmarshal([]byte(gottenOperation.PayloadData), &gottenPayload); err != nil {
			t.Fatalf("failed to unmarshal payload: %v", err)
		}

		if gottenPayload.ID != payload.ID || gottenPayload.Name != payload.Name {
			t.Fatalf("gotten payload does not match expected")
		}
	})

	t.Run("update_local_db_board_insert", func(t *testing.T) {
		repo := setupTestDB(t)
		lf := NewLocalFuncs(repo)

		payload := boardOperationPayload{
			ID:        "board_123",
			Name:      "Test Board",
			CreatedAt: "2023-01-01 00:00:00",
			UpdatedAt: "2023-01-01 00:00:00",
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}

		operation := types.OperationSync{
			ID:            "op_123",
			TableName:     "boards",
			RecordID:      "board_123",
			OperationType: "insert",
			DeviceID:      "device_123",
			PayloadData:   string(payloadBytes),
			CreatedAt:     "2023-01-01 00:00:00",
			UpdatedAt:     "2023-01-01 00:00:00",
		}

		err = lf.UpdateLocalDB(operation)
		if err != nil {
			t.Fatalf("UpdateLocalDB failed: %v", err)
		}

		board, err := repo.GetBoard("board_123")
		if err != nil {
			t.Fatalf("GetBoard failed: %v", err)
		}

		if board.Name != "Test Board" {
			t.Fatalf("expected board name 'Test Board', got '%s'", board.Name)
		}
	})

	t.Run("update_local_db_board_update", func(t *testing.T) {
		repo := setupTestDB(t)
		lf := NewLocalFuncs(repo)

		_, err := repo.CreateBoard("Original Name")
		if err != nil {
			t.Fatalf("CreateBoard failed: %v", err)
		}

		boards, err := repo.GetAllBoards(1, 10)
		if err != nil || len(boards) == 0 {
			t.Fatalf("GetAllBoards failed: %v", err)
		}
		boardID := boards[0].ID

		payload := boardOperationPayload{
			ID:        boardID,
			Name:      "Updated Name",
			CreatedAt: "2023-01-01 00:00:00",
			UpdatedAt: "2023-01-02 00:00:00",
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}

		operation := types.OperationSync{
			TableName:     "boards",
			RecordID:      boardID,
			OperationType: "update",
			PayloadData:   string(payloadBytes),
		}

		err = lf.UpdateLocalDB(operation)
		if err != nil {
			t.Fatalf("UpdateLocalDB failed: %v", err)
		}

		board, err := repo.GetBoard(boardID)
		if err != nil {
			t.Fatalf("GetBoard failed: %v", err)
		}

		if board.Name != "Updated Name" {
			t.Fatalf("expected board name 'Updated Name', got '%s'", board.Name)
		}
	})

	t.Run("update_local_db_board_delete", func(t *testing.T) {
		repo := setupTestDB(t)
		lf := NewLocalFuncs(repo)

		board, err := repo.CreateBoard("Test Board")
		if err != nil {
			t.Fatalf("CreateBoard failed: %v", err)
		}

		payload := boardOperationPayload{
			ID:        board.ID,
			Name:      "Test Board",
			CreatedAt: "2023-01-01 00:00:00",
			UpdatedAt: "2023-01-01 00:00:00",
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}

		operation := types.OperationSync{
			TableName:     "boards",
			RecordID:      board.ID,
			OperationType: "delete",
			PayloadData:   string(payloadBytes),
		}

		err = lf.UpdateLocalDB(operation)
		if err != nil {
			t.Fatalf("UpdateLocalDB failed: %v", err)
		}

		_, err = repo.GetBoard(board.ID)
		if err == nil {
			t.Fatalf("expected error when getting deleted board, got nil")
		}
	})

	t.Run("update_local_db_column_insert", func(t *testing.T) {
		repo := setupTestDB(t)
		lf := NewLocalFuncs(repo)

		board, err := repo.CreateBoard("Test Board")
		if err != nil {
			t.Fatalf("CreateBoard failed: %v", err)
		}

		type columnPayload struct {
			ID        string `json:"id"`
			BoardID   string `json:"board_id"`
			Name      string `json:"name"`
			Position  int64  `json:"position"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
		}

		payload := columnPayload{
			ID:        "col_123",
			BoardID:   board.ID,
			Name:      "To Do",
			Position:  1,
			CreatedAt: "2023-01-01 00:00:00",
			UpdatedAt: "2023-01-01 00:00:00",
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}

		operation := types.OperationSync{
			TableName:     "columns",
			RecordID:      "col_123",
			OperationType: "insert",
			PayloadData:   string(payloadBytes),
		}

		err = lf.UpdateLocalDB(operation)
		if err != nil {
			t.Fatalf("UpdateLocalDB failed: %v", err)
		}

		column, err := repo.GetColumn("col_123")
		if err != nil {
			t.Fatalf("GetColumn failed: %v", err)
		}

		if column.Name != "To Do" {
			t.Fatalf("expected column name 'To Do', got '%s'", column.Name)
		}
		if column.Position != 1 {
			t.Fatalf("expected column position 1, got %d", column.Position)
		}
	})

	t.Run("update_local_db_card_insert", func(t *testing.T) {
		repo := setupTestDB(t)
		lf := NewLocalFuncs(repo)

		board, err := repo.CreateBoard("Test Board")
		if err != nil {
			t.Fatalf("CreateBoard failed: %v", err)
		}

		column, err := repo.CreateColumn(board.ID, "To Do")
		if err != nil {
			t.Fatalf("CreateColumn failed: %v", err)
		}

		type cardPayload struct {
			ID          string `json:"id"`
			ColumnID    string `json:"column_id"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Attachments string `json:"attachments"`
			CreatedAt   string `json:"created_at"`
			UpdatedAt   string `json:"updated_at"`
		}

		payload := cardPayload{
			ID:          "card_123",
			ColumnID:    column.ID,
			Title:       "Test Card",
			Description: "Test Description",
			Attachments: "",
			CreatedAt:   "2023-01-01 00:00:00",
			UpdatedAt:   "2023-01-01 00:00:00",
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}

		operation := types.OperationSync{
			TableName:     "cards",
			RecordID:      "card_123",
			OperationType: "insert",
			PayloadData:   string(payloadBytes),
		}

		err = lf.UpdateLocalDB(operation)
		if err != nil {
			t.Fatalf("UpdateLocalDB failed: %v", err)
		}

		card, err := repo.GetCard("card_123")
		if err != nil {
			t.Fatalf("GetCard failed: %v", err)
		}

		if card.Title != "Test Card" {
			t.Fatalf("expected card title 'Test Card', got '%s'", card.Title)
		}
		if card.Description.String != "Test Description" {
			t.Fatalf("expected card description 'Test Description', got '%s'", card.Description.String)
		}
	})

	t.Run("update_local_db_card_update_column", func(t *testing.T) {
		repo := setupTestDB(t)
		lf := NewLocalFuncs(repo)

		board, err := repo.CreateBoard("Test Board")
		if err != nil {
			t.Fatalf("CreateBoard failed: %v", err)
		}

		column1, err := repo.CreateColumn(board.ID, "To Do")
		if err != nil {
			t.Fatalf("CreateColumn failed: %v", err)
		}

		column2, err := repo.CreateColumn(board.ID, "Done")
		if err != nil {
			t.Fatalf("CreateColumn failed: %v", err)
		}

		card, err := repo.CreateCard(column1.ID, "Test Card", "Description")
		if err != nil {
			t.Fatalf("CreateCard failed: %v", err)
		}

		type cardPayload struct {
			ID          string `json:"id"`
			ColumnID    string `json:"column_id"`
			Title       string `json:"title"`
			Description string `json:"description"`
			CreatedAt   string `json:"created_at"`
			UpdatedAt   string `json:"updated_at"`
		}

		payload := cardPayload{
			ID:          card.ID,
			ColumnID:    column2.ID,
			Title:       "Test Card",
			Description: "Description",
			CreatedAt:   "2023-01-01 00:00:00",
			UpdatedAt:   "2023-01-02 00:00:00",
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}

		operation := types.OperationSync{
			TableName:     "cards",
			RecordID:      card.ID,
			OperationType: "update-card-column",
			PayloadData:   string(payloadBytes),
		}

		err = lf.UpdateLocalDB(operation)
		if err != nil {
			t.Fatalf("UpdateLocalDB failed: %v", err)
		}

		updatedCard, err := repo.GetCard(card.ID)
		if err != nil {
			t.Fatalf("GetCard failed: %v", err)
		}

		if updatedCard.ColumnID != column2.ID {
			t.Fatalf("expected card column_id '%s', got '%s'", column2.ID, updatedCard.ColumnID)
		}
	})

	t.Run("update_local_db_invalid_table", func(t *testing.T) {
		repo := setupTestDB(t)
		lf := NewLocalFuncs(repo)

		operation := types.OperationSync{
			TableName:     "invalid_table",
			RecordID:      "123",
			OperationType: "insert",
			PayloadData:   "{}",
		}

		err := lf.UpdateLocalDB(operation)
		if err == nil {
			t.Fatalf("expected error for invalid table name, got nil")
		}
	})

	t.Run("update_local_db_invalid_json", func(t *testing.T) {
		repo := setupTestDB(t)
		lf := NewLocalFuncs(repo)

		operation := types.OperationSync{
			TableName:     "boards",
			RecordID:      "123",
			OperationType: "insert",
			PayloadData:   "invalid json",
		}

		err := lf.UpdateLocalDB(operation)
		if err == nil {
			t.Fatalf("expected error for invalid JSON, got nil")
		}
	})

}
