package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"seisami/app/internal/repo/sqlc/query"
	"seisami/app/types"
	"testing"

	_ "embed"
)

func setupTestDB(t *testing.T) *repo {

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	if _, err := db.Exec(Schema); err != nil {
		t.Fatalf("failed to exec schema: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return NewRepo(db, context.Background())
}

func TestBoard(t *testing.T) {

	t.Run("create_board", func(t *testing.T) {
		repo := setupTestDB(t)
		board, err := repo.CreateBoard("Test Board")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		got, err := repo.GetBoard(board.ID)
		if err != nil {
			t.Fatalf("failed to get board: %v", err)
		}
		if got.Name != "Test Board" {
			t.Errorf("expected board name 'Test Board', got '%s'", got.Name)
		}
	})

	t.Run("get_board", func(t *testing.T) {
		repo := setupTestDB(t)
		board, err := repo.CreateBoard("Test Get Board")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		got, err := repo.GetBoard(board.ID)
		if err != nil {
			t.Fatalf("failed to get board: %v", err)
		}
		if got.Name != "Test Get Board" {
			t.Errorf("expected board name 'Test Get Board', got '%s'", got.Name)
		}
	})

	t.Run("list_boards", func(t *testing.T) {
		repo := setupTestDB(t)
		testBoards := []query.Board{}
		for i := 1; i <= 3; i++ {
			board, err := repo.CreateBoard(fmt.Sprintf("Board %d", i))
			if err != nil {
				t.Fatalf("failed to create board: %v", err)
			}
			testBoards = append(testBoards, board)
		}

		got, err := repo.GetAllBoards(1, 10)
		if err != nil {
			t.Fatalf("failed to get all boards: %v", err)
		}

		// this includes the board created in previous test (create_board)
		if len(got) != len(testBoards) {
			t.Errorf("expected %d boards, got %d", len(testBoards), len(got))
		}

		_ = testBoards
	})

	t.Run("delete_board", func(t *testing.T) {
		repo := setupTestDB(t)
		board, err := repo.CreateBoard("To Be Deleted")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		err = repo.DeleteBoard(board.ID)
		if err != nil {
			t.Fatalf("failed to delete board: %v", err)
		}

		_, err = repo.GetBoard(board.ID)
		if err == nil {
			t.Errorf("expected error when getting deleted board, got nil")
		}
	})

	t.Run("update_board", func(t *testing.T) {
		repo := setupTestDB(t)
		board, err := repo.CreateBoard("Test Board")
		if err != nil {
			t.Fatalf("unable to create board: %v", err)
		}

		updatedBoard, err := repo.UpdateBoard(board.ID, "Updated Board Name")
		if err != nil {
			t.Fatalf("unable to update board: %v", err)
		}

		if updatedBoard.Name != "Updated Board Name" {
			t.Errorf("expected board name 'Updated Board Name', got '%s'", updatedBoard.Name)
		}
	})

}

func TestColumn(t *testing.T) {

	t.Run("create_column", func(t *testing.T) {
		repo := setupTestDB(t)
		board, err := repo.CreateBoard("Board for Column")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		column, err := repo.CreateColumn(board.ID, "Test Column")
		if err != nil {
			t.Fatalf("failed to create column: %v", err)
		}

		if column.Name != "Test Column" {
			t.Errorf("expected column name 'Test Column', got '%s'", column.Name)
		}
	})

	t.Run("list_columns", func(t *testing.T) {
		repo := setupTestDB(t)
		board, err := repo.CreateBoard("Board for Columns")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		for i := 1; i <= 3; i++ {
			_, err := repo.CreateColumn(board.ID, "Column "+string(rune(i)))
			if err != nil {
				t.Fatalf("failed to create column: %v", err)
			}
		}

		columns, err := repo.ListColumnsByBoard(board.ID)
		if err != nil {
			t.Fatalf("failed to get columns: %v", err)
		}

		fmt.Println(len(columns))

		if len(columns) != 3 {
			t.Errorf("expected 3 columns, got %d", len(columns))
		}
	})

	t.Run("delete_column", func(t *testing.T) {
		repo := setupTestDB(t)
		board, err := repo.CreateBoard("Board for Deleting Column")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		column, err := repo.CreateColumn(board.ID, "Column to Delete")
		if err != nil {
			t.Fatalf("failed to create column: %v", err)
		}

		err = repo.DeleteColumn(column.ID)
		if err != nil {
			t.Fatalf("failed to delete column: %v", err)
		}

		_, err = repo.GetColumn(column.ID)
		if err == nil {
			t.Errorf("expected error when getting deleted column, got nil")
		}
	})

	t.Run("get_column", func(t *testing.T) {
		repo := setupTestDB(t)
		board, err := repo.CreateBoard("Board for Get Column")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		column, err := repo.CreateColumn(board.ID, "Column to Get")
		if err != nil {
			t.Fatalf("failed to create column: %v", err)
		}

		got, err := repo.GetColumn(column.ID)
		if err != nil {
			t.Fatalf("failed to get column: %v", err)
		}

		if got.Name != "Column to Get" {
			t.Errorf("expected column name 'Column to Get', got '%s'", got.Name)
		}
	})

	t.Run("update_column", func(t *testing.T) {
		repo := setupTestDB(t)
		board, err := repo.CreateBoard("Board for Update Column")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		column, err := repo.CreateColumn(board.ID, "Column to Update")
		if err != nil {
			t.Fatalf("failed to create column: %v", err)
		}

		updatedColumn, err := repo.UpdateColumn(column.ID, "Updated Column Name")
		if err != nil {
			t.Fatalf("failed to update column: %v", err)
		}

		if updatedColumn.Name != "Updated Column Name" {
			t.Errorf("expected column name 'Updated Column Name', got '%s'", updatedColumn.Name)
		}
	})
}

func TestCard(t *testing.T) {
	t.Run("create_card", func(t *testing.T) {
		repo := setupTestDB(t)

		board, err := repo.CreateBoard("Test Board")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		column, err := repo.CreateColumn(board.ID, "Test Column")
		if err != nil {
			t.Fatalf("failed to create column: %v", err)
		}

		card, err := repo.CreateCard(column.ID, "Test Title", "Test Description")
		if err != nil {
			t.Fatalf("failed to create card: %v", err)
		}

		got, err := repo.GetCard(card.ID)
		if err != nil {
			t.Fatalf("failed to retrieve card: %v", err)
		}

		if card.Title != got.Title {
			t.Errorf("expected title '%s', got '%s'", card.Title, got.Title)
		}

		if card.Description.String != got.Description.String {
			t.Errorf("expected description '%s', got '%s'", card.Description.String, got.Description.String)
		}

		if card.Attachments != got.Attachments {
			t.Errorf("expected attachments '%s', got '%s'", card.Attachments.String, got.Attachments.String)
		}
	})

	t.Run("delete_card", func(t *testing.T) {
		repo := setupTestDB(t)

		board, err := repo.CreateBoard("Test Board")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		column, err := repo.CreateColumn(board.ID, "Test Column")
		if err != nil {
			t.Fatalf("failed to create column: %v", err)
		}

		card, err := repo.CreateCard(column.ID, "Test Title", "Test Description")
		if err != nil {
			t.Fatalf("failed to create card: %v", err)
		}

		err = repo.DeleteCard(card.ID)
		if err != nil {
			t.Fatalf("failed to delete card: %v", err)
		}

		_, err = repo.GetCard(card.ID)
		if err == nil {
			t.Errorf("expected error when getting deleted card, got nil")
		}
	})

	t.Run("update_card", func(t *testing.T) {
		repo := setupTestDB(t)

		board, err := repo.CreateBoard("Test Board")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		column, err := repo.CreateColumn(board.ID, "Test Column")
		if err != nil {
			t.Fatalf("failed to create column: %v", err)
		}

		card, err := repo.CreateCard(column.ID, "Test Title", "Test Description")
		if err != nil {
			t.Fatalf("failed to create card: %v", err)
		}

		updatedCard, err := repo.UpdateCard(card.ID, "Updated Title", "Updated Description")
		if err != nil {
			t.Fatalf("failed to update card: %v", err)
		}

		if updatedCard.Title != "Updated Title" {
			t.Errorf("expected title 'Updated Title', got '%s'", updatedCard.Title)
		}

		if updatedCard.Description.String != "Updated Description" {
			t.Errorf("expected description 'Updated Description', got '%s'", updatedCard.Description.String)
		}
	})

	t.Run("update_card_column", func(t *testing.T) {
		repo := setupTestDB(t)

		board, err := repo.CreateBoard("Test Board")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		column, err := repo.CreateColumn(board.ID, "Test Column")
		if err != nil {
			t.Fatalf("failed to create column: %v", err)
		}

		card, err := repo.CreateCard(column.ID, "Test Title", "Test Description")
		if err != nil {
			t.Fatalf("failed to create card: %v", err)
		}

		newColumn, err := repo.CreateColumn(board.ID, "New Column")
		if err != nil {
			t.Fatalf("failed to create new column: %v", err)
		}

		_, err = repo.UpdateCardColumn(card.ID, newColumn.ID)
		if err != nil {
			t.Fatalf("failed to update card column: %v", err)
		}

		updatedCard, err := repo.GetCard(card.ID)
		if err != nil {
			t.Fatalf("failed to get updated card: %v", err)
		}

		if updatedCard.ColumnID != newColumn.ID {
			t.Errorf("expected: %v, got: %v", newColumn.ID, updatedCard.ColumnID)
		}
	})

	t.Run("list_column_cards", func(t *testing.T) {
		repo := setupTestDB(t)

		board, err := repo.CreateBoard("Test Board")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		column, err := repo.CreateColumn(board.ID, "Test Column")
		if err != nil {
			t.Fatalf("failed to create column: %v", err)
		}

		column2, err := repo.CreateColumn(board.ID, "Test Column")
		if err != nil {
			t.Fatalf("failed to create column 2: %v", err)
		}

		for i := range 5 {
			_, err := repo.CreateCard(column.ID, fmt.Sprintf("Test Title: %d", i), fmt.Sprintf("Test Description: %d", i))
			if err != nil {
				t.Fatalf("failed to create card: %v", err)
			}
		}

		for i := range 2 {
			_, err := repo.CreateCard(column2.ID, fmt.Sprintf("Test Title: %d", i), fmt.Sprintf("Test Description: %d", i))
			if err != nil {
				t.Fatalf("failed to create card: %v", err)
			}
		}

		cards, err := repo.ListCardsByColumn(column.ID)
		if err != nil {
			t.Fatalf("failed to list column cards: %v", err)
		}

		if len(cards) != 5 {
			t.Errorf("expected 5 cards, got :%d", len(cards))
		}
	})

}

func TestTranscription(t *testing.T) {
	t.Run("add_transcription", func(t *testing.T) {
		repo := setupTestDB(t)

		board, err := repo.CreateBoard("Test Board")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		transcription, err := repo.AddTransscription(board.ID, "This is a test transcription", "/path/to/recording.wav")
		if err != nil {
			t.Fatalf("failed to add transcription: %v", err)
		}

		if transcription.Transcription != "This is a test transcription" {
			t.Errorf("expected transcription 'This is a test transcription', got '%s'", transcription.Transcription)
		}

		if transcription.RecordingPath.String != "/path/to/recording.wav" {
			t.Errorf("expected recording path '/path/to/recording.wav', got '%s'", transcription.RecordingPath.String)
		}

		if transcription.BoardID != board.ID {
			t.Errorf("expected board ID '%s', got '%s'", board.ID, transcription.BoardID)
		}
	})

	t.Run("get_transcriptions", func(t *testing.T) {
		repo := setupTestDB(t)

		board, err := repo.CreateBoard("Test Board")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		for i := 1; i <= 5; i++ {
			_, err := repo.AddTransscription(
				board.ID,
				fmt.Sprintf("Transcription %d", i),
				fmt.Sprintf("/path/to/recording%d.wav", i),
			)
			if err != nil {
				t.Fatalf("failed to add transcription: %v", err)
			}
		}

		transcriptions, err := repo.GetTranscriptions(board.ID, 1, 10)
		if err != nil {
			t.Fatalf("failed to get transcriptions: %v", err)
		}

		if len(transcriptions) != 5 {
			t.Errorf("expected 5 transcriptions, got %d", len(transcriptions))
		}
	})

	t.Run("get_transcriptions_empty", func(t *testing.T) {
		repo := setupTestDB(t)

		board, err := repo.CreateBoard("Empty Board")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		transcriptions, err := repo.GetTranscriptions(board.ID, 1, 10)
		if err != nil {
			t.Fatalf("failed to get transcriptions: %v", err)
		}

		if len(transcriptions) != 0 {
			t.Errorf("expected 0 transcriptions, got %d", len(transcriptions))
		}
	})

	t.Run("get_transcription_by_id", func(t *testing.T) {
		repo := setupTestDB(t)

		board, err := repo.CreateBoard("Test Board")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		transcription, err := repo.AddTransscription(board.ID, "Test transcription by ID", "/path/to/recording.wav")
		if err != nil {
			t.Fatalf("failed to add transcription: %v", err)
		}

		got, err := repo.GetTranscriptionByID(transcription.ID)
		if err != nil {
			t.Fatalf("failed to get transcription by ID: %v", err)
		}

		if got.ID != transcription.ID {
			t.Errorf("expected ID '%s', got '%s'", transcription.ID, got.ID)
		}

		if got.Transcription != "Test transcription by ID" {
			t.Errorf("expected transcription 'Test transcription by ID', got '%s'", got.Transcription)
		}
	})

	t.Run("update_transcription_intent", func(t *testing.T) {
		repo := setupTestDB(t)

		board, err := repo.CreateBoard("Test Board")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		transcription, err := repo.AddTransscription(board.ID, "Test transcription", "/path/to/recording.wav")
		if err != nil {
			t.Fatalf("failed to add transcription: %v", err)
		}

		err = repo.UpdateTranscriptionIntent(transcription.ID, "Create a new task")
		if err != nil {
			t.Fatalf("failed to update transcription intent: %v", err)
		}

		updated, err := repo.GetTranscriptionByID(transcription.ID)
		if err != nil {
			t.Fatalf("failed to get updated transcription: %v", err)
		}

		if updated.Intent.String != "Create a new task" {
			t.Errorf("expected intent 'Create a new task', got '%s'", updated.Intent.String)
		}
	})

	t.Run("update_transcription_response", func(t *testing.T) {
		repo := setupTestDB(t)

		board, err := repo.CreateBoard("Test Board")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		transcription, err := repo.AddTransscription(board.ID, "Test transcription", "/path/to/recording.wav")
		if err != nil {
			t.Fatalf("failed to add transcription: %v", err)
		}

		err = repo.UpdateTranscriptionResponse(transcription.ID, "Task created successfully")
		if err != nil {
			t.Fatalf("failed to update transcription response: %v", err)
		}

		updated, err := repo.GetTranscriptionByID(transcription.ID)
		if err != nil {
			t.Fatalf("failed to get updated transcription: %v", err)
		}

		if updated.AssistantResponse.String != "Task created successfully" {
			t.Errorf("expected response 'Task created successfully', got '%s'", updated.AssistantResponse.String)
		}
	})

	t.Run("update_both_intent_and_response", func(t *testing.T) {
		repo := setupTestDB(t)

		board, err := repo.CreateBoard("Test Board")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		transcription, err := repo.AddTransscription(board.ID, "Test transcription", "/path/to/recording.wav")
		if err != nil {
			t.Fatalf("failed to add transcription: %v", err)
		}

		err = repo.UpdateTranscriptionIntent(transcription.ID, "Delete a card")
		if err != nil {
			t.Fatalf("failed to update transcription intent: %v", err)
		}

		err = repo.UpdateTranscriptionResponse(transcription.ID, "Card deleted successfully")
		if err != nil {
			t.Fatalf("failed to update transcription response: %v", err)
		}

		updated, err := repo.GetTranscriptionByID(transcription.ID)
		if err != nil {
			t.Fatalf("failed to get updated transcription: %v", err)
		}

		if updated.Intent.String != "Delete a card" {
			t.Errorf("expected intent 'Delete a card', got '%s'", updated.Intent.String)
		}

		if updated.AssistantResponse.String != "Card deleted successfully" {
			t.Errorf("expected response 'Card deleted successfully', got '%s'", updated.AssistantResponse.String)
		}
	})

	t.Run("get_transcription_by_nonexistent_id", func(t *testing.T) {
		repo := setupTestDB(t)

		_, err := repo.GetTranscriptionByID("nonexistent-id")
		if err == nil {
			t.Errorf("expected error when getting nonexistent transcription, got nil")
		}
	})

	t.Run("multiple_boards_with_transcriptions", func(t *testing.T) {
		repo := setupTestDB(t)

		board1, err := repo.CreateBoard("Board 1")
		if err != nil {
			t.Fatalf("failed to create board 1: %v", err)
		}

		board2, err := repo.CreateBoard("Board 2")
		if err != nil {
			t.Fatalf("failed to create board 2: %v", err)
		}

		for i := 1; i <= 3; i++ {
			_, err := repo.AddTransscription(board1.ID, fmt.Sprintf("Board 1 Transcription %d", i), "/path/recording.wav")
			if err != nil {
				t.Fatalf("failed to add transcription to board 1: %v", err)
			}
		}

		for i := 1; i <= 2; i++ {
			_, err := repo.AddTransscription(board2.ID, fmt.Sprintf("Board 2 Transcription %d", i), "/path/recording.wav")
			if err != nil {
				t.Fatalf("failed to add transcription to board 2: %v", err)
			}
		}

		board1Transcriptions, err := repo.GetTranscriptions(board1.ID, 1, 10)
		if err != nil {
			t.Fatalf("failed to get board 1 transcriptions: %v", err)
		}

		board2Transcriptions, err := repo.GetTranscriptions(board2.ID, 1, 10)
		if err != nil {
			t.Fatalf("failed to get board 2 transcriptions: %v", err)
		}

		if len(board1Transcriptions) != 3 {
			t.Errorf("expected 3 transcriptions for board 1, got %d", len(board1Transcriptions))
		}

		if len(board2Transcriptions) != 2 {
			t.Errorf("expected 2 transcriptions for board 2, got %d", len(board2Transcriptions))
		}
	})
}

func TestSettings(t *testing.T) {
	t.Run("get_settings_default", func(t *testing.T) {
		repo := setupTestDB(t)

		settings, err := repo.GetSettings()
		if err != nil {
			t.Fatalf("failed to get settings: %v", err)
		}

		if settings.TranscriptionMethod != "cloud" {
			t.Errorf("expected default transcription method 'cloud', got '%s'", settings.TranscriptionMethod)
		}

		if settings.ID != 1 {
			t.Errorf("expected ID 1, got %d", settings.ID)
		}
	})

	t.Run("create_settings", func(t *testing.T) {
		repo := setupTestDB(t)

		binaryPath := "/path/to/whisper"
		modelPath := "/path/to/model"
		apiKey := "test-api-key"

		settings, err := repo.CreateOrUpdateSettings("local", &binaryPath, &modelPath, &apiKey)
		if err != nil {
			t.Fatalf("failed to create settings: %v", err)
		}

		if settings.TranscriptionMethod != "local" {
			t.Errorf("expected transcription method 'local', got '%s'", settings.TranscriptionMethod)
		}

		if settings.WhisperBinaryPath.String != binaryPath {
			t.Errorf("expected binary path '%s', got '%s'", binaryPath, settings.WhisperBinaryPath.String)
		}

		if settings.WhisperModelPath.String != modelPath {
			t.Errorf("expected model path '%s', got '%s'", modelPath, settings.WhisperModelPath.String)
		}

		if settings.OpenaiApiKey.String != apiKey {
			t.Errorf("expected API key '%s', got '%s'", apiKey, settings.OpenaiApiKey.String)
		}
	})

	t.Run("update_settings", func(t *testing.T) {
		repo := setupTestDB(t)

		binaryPath := "/path/to/whisper"
		modelPath := "/path/to/model"
		apiKey := "test-api-key"

		_, err := repo.CreateOrUpdateSettings("local", &binaryPath, &modelPath, &apiKey)
		if err != nil {
			t.Fatalf("failed to create settings: %v", err)
		}

		newBinaryPath := "/new/path/to/whisper"
		newModelPath := "/new/path/to/model"
		newApiKey := "new-api-key"

		updated, err := repo.CreateOrUpdateSettings("cloud", &newBinaryPath, &newModelPath, &newApiKey)
		if err != nil {
			t.Fatalf("failed to update settings: %v", err)
		}

		if updated.TranscriptionMethod != "cloud" {
			t.Errorf("expected transcription method 'cloud', got '%s'", updated.TranscriptionMethod)
		}

		if updated.WhisperBinaryPath.String != newBinaryPath {
			t.Errorf("expected binary path '%s', got '%s'", newBinaryPath, updated.WhisperBinaryPath.String)
		}

		if updated.WhisperModelPath.String != newModelPath {
			t.Errorf("expected model path '%s', got '%s'", newModelPath, updated.WhisperModelPath.String)
		}

		if updated.OpenaiApiKey.String != newApiKey {
			t.Errorf("expected API key '%s', got '%s'", newApiKey, updated.OpenaiApiKey.String)
		}
	})

	t.Run("create_settings_with_nil_values", func(t *testing.T) {
		repo := setupTestDB(t)

		settings, err := repo.CreateOrUpdateSettings("cloud", nil, nil, nil)
		if err != nil {
			t.Fatalf("failed to create settings with nil values: %v", err)
		}

		if settings.TranscriptionMethod != "cloud" {
			t.Errorf("expected transcription method 'cloud', got '%s'", settings.TranscriptionMethod)
		}

		if settings.WhisperBinaryPath.Valid {
			t.Errorf("expected WhisperBinaryPath to be null, got valid value: '%s'", settings.WhisperBinaryPath.String)
		}

		if settings.WhisperModelPath.Valid {
			t.Errorf("expected WhisperModelPath to be null, got valid value: '%s'", settings.WhisperModelPath.String)
		}

		if settings.OpenaiApiKey.Valid {
			t.Errorf("expected OpenaiApiKey to be null, got valid value: '%s'", settings.OpenaiApiKey.String)
		}
	})

	t.Run("update_settings_partial", func(t *testing.T) {
		repo := setupTestDB(t)

		binaryPath := "/path/to/whisper"
		modelPath := "/path/to/model"
		apiKey := "test-api-key"

		_, err := repo.CreateOrUpdateSettings("local", &binaryPath, &modelPath, &apiKey)
		if err != nil {
			t.Fatalf("failed to create settings: %v", err)
		}

		updated, err := repo.CreateOrUpdateSettings("cloud", nil, nil, nil)
		if err != nil {
			t.Fatalf("failed to update settings: %v", err)
		}

		if updated.TranscriptionMethod != "cloud" {
			t.Errorf("expected transcription method 'cloud', got '%s'", updated.TranscriptionMethod)
		}

		if updated.WhisperBinaryPath.Valid {
			t.Errorf("expected WhisperBinaryPath to be null after update with nil")
		}
	})

	t.Run("get_settings_after_create", func(t *testing.T) {
		repo := setupTestDB(t)

		binaryPath := "/path/to/whisper"
		apiKey := "my-api-key"

		_, err := repo.CreateOrUpdateSettings("custom", &binaryPath, nil, &apiKey)
		if err != nil {
			t.Fatalf("failed to create settings: %v", err)
		}

		settings, err := repo.GetSettings()
		if err != nil {
			t.Fatalf("failed to get settings: %v", err)
		}

		if settings.TranscriptionMethod != "custom" {
			t.Errorf("expected transcription method 'custom', got '%s'", settings.TranscriptionMethod)
		}

		if settings.WhisperBinaryPath.String != binaryPath {
			t.Errorf("expected binary path '%s', got '%s'", binaryPath, settings.WhisperBinaryPath.String)
		}

		if settings.WhisperModelPath.Valid {
			t.Errorf("expected WhisperModelPath to be null, got '%s'", settings.WhisperModelPath.String)
		}

		if settings.OpenaiApiKey.String != apiKey {
			t.Errorf("expected API key '%s', got '%s'", apiKey, settings.OpenaiApiKey.String)
		}
	})

}

func TestOperations(t *testing.T) {
	t.Run("create_operation", func(t *testing.T) {
		repo := setupTestDB(t)

		board, err := repo.CreateBoard("Test Board")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		operation, err := repo.CreateOperation(
			types.BoardTable,
			board.ID,
			`{"name":"Test Board"}`,
			types.InsertOperation,
		)
		if err != nil {
			t.Fatalf("failed to create operation: %v", err)
		}

		if operation.TableName != "boards" {
			t.Errorf("expected table name 'boards', got '%s'", operation.TableName)
		}

		if operation.RecordID != board.ID {
			t.Errorf("expected record ID '%s', got '%s'", board.ID, operation.RecordID)
		}

		if operation.OperationType != "insert" {
			t.Errorf("expected operation type 'insert', got '%s'", operation.OperationType)
		}
	})

	t.Run("get_all_operations", func(t *testing.T) {
		repo := setupTestDB(t)

		board, err := repo.CreateBoard("Test Board")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		for i := 0; i < 3; i++ {
			_, err := repo.CreateOperation(
				types.BoardTable,
				board.ID,
				fmt.Sprintf(`{"name":"Board %d"}`, i),
				types.UpdateOperation,
			)
			if err != nil {
				t.Fatalf("failed to create operation: %v", err)
			}
		}

		operations, err := repo.GetAllOperations(types.BoardTable)
		if err != nil {
			t.Fatalf("failed to get operations: %v", err)
		}

		if len(operations) != 3 {
			t.Errorf("expected 3 operations, got %d", len(operations))
		}
	})

	t.Run("operations_different_tables", func(t *testing.T) {
		repo := setupTestDB(t)

		board, err := repo.CreateBoard("Test Board")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		column, err := repo.CreateColumn(board.ID, "Test Column")
		if err != nil {
			t.Fatalf("failed to create column: %v", err)
		}

		_, err = repo.CreateOperation(types.BoardTable, board.ID, `{"name":"Board"}`, types.InsertOperation)
		if err != nil {
			t.Fatalf("failed to create board operation: %v", err)
		}

		_, err = repo.CreateOperation(types.ColumnTable, column.ID, `{"name":"Column"}`, types.InsertOperation)
		if err != nil {
			t.Fatalf("failed to create column operation: %v", err)
		}

		boardOps, err := repo.GetAllOperations(types.BoardTable)
		if err != nil {
			t.Fatalf("failed to get board operations: %v", err)
		}

		columnOps, err := repo.GetAllOperations(types.ColumnTable)
		if err != nil {
			t.Fatalf("failed to get column operations: %v", err)
		}

		if len(boardOps) != 1 {
			t.Errorf("expected 1 board operation, got %d", len(boardOps))
		}

		if len(columnOps) != 1 {
			t.Errorf("expected 1 column operation, got %d", len(columnOps))
		}
	})

	type boardOperationPayload struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	type columnOperationPayload struct {
		RoomID    string `json:"room_id"`
		ID        string `json:"id"`
		BoardID   string `json:"board_id"`
		Name      string `json:"name"`
		Position  int    `json:"position"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	t.Run("get_all_board_operations", func(t *testing.T) {
		repo := setupTestDB(t)

		payload := boardOperationPayload{
			ID:        "1234",
			Name:      "lawrence",
			CreatedAt: "2024-01-01T00:00:00Z",
			UpdatedAt: "2024-01-01T00:00:00Z",
		}

		pBytes, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}

		op, err := repo.CreateOperation(types.BoardTable, "1234", string(pBytes), types.InsertOperation)
		if err != nil {
			t.Fatalf("CreateOperation failed: %v", err)
		}

		ops, err := repo.GetAllOperations(types.BoardTable)
		if err != nil {
			t.Fatalf("GetAllOperations failed: %v", err)
		}

		if len(ops) != 1 {
			t.Fatalf("expected 1 operation, got %d", len(ops))
		}

		if ops[0].ID != op.ID {
			t.Fatalf("expected operation ID %s, got %s", op.ID, ops[0].ID)
		}

		gottenOperation := ops[0]
		var gottenPayload boardOperationPayload
		if err := json.Unmarshal([]byte(gottenOperation.Payload), &gottenPayload); err != nil {
			t.Fatalf("failed to unmarshal payload: %v", err)
		}

		if gottenPayload.ID != payload.ID {
			t.Fatalf("expected payload ID %s, got %s", payload.ID, gottenPayload.ID)
		}
		if gottenPayload.Name != payload.Name {
			t.Fatalf("expected payload Name %s, got %s", payload.Name, gottenPayload.Name)
		}
		if gottenPayload.CreatedAt != payload.CreatedAt {
			t.Fatalf("expected payload CreatedAt %s, got %s", payload.CreatedAt, gottenPayload.CreatedAt)
		}
		if gottenPayload.UpdatedAt != payload.UpdatedAt {
			t.Fatalf("expected payload UpdatedAt %s, got %s", payload.UpdatedAt, gottenPayload.UpdatedAt)
		}
	})

	t.Run("get_all_column_operations", func(t *testing.T) {
		repo := setupTestDB(t)

		payload := columnOperationPayload{
			RoomID:    "room_1234",
			ID:        "col_1234",
			BoardID:   "board_1234",
			Name:      "To Do",
			Position:  1,
			CreatedAt: "2024-01-01T00:00:00Z",
			UpdatedAt: "2024-01-01T00:00:00Z",
		}

		pBytes, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}

		op, err := repo.CreateOperation(types.ColumnTable, "col_1234", string(pBytes), types.InsertOperation)
		if err != nil {
			t.Fatalf("CreateOperation failed: %v", err)
		}

		ops, err := repo.GetAllOperations(types.ColumnTable)
		if err != nil {
			t.Fatalf("GetAllOperations failed: %v", err)
		}

		if len(ops) != 1 {
			t.Fatalf("expected 1 operation, got %d", len(ops))
		}

		if ops[0].ID != op.ID {
			t.Fatalf("expected operation ID %s, got %s", op.ID, ops[0].ID)
		}

		gottenOperation := ops[0]
		var gottenPayload columnOperationPayload
		if err := json.Unmarshal([]byte(gottenOperation.Payload), &gottenPayload); err != nil {
			t.Fatalf("failed to unmarshal payload: %v", err)
		}

		if gottenPayload.ID != payload.ID {
			t.Fatalf("expected payload ID %s, got %s", payload.ID, gottenPayload.ID)
		}
		if gottenPayload.Name != payload.Name {
			t.Fatalf("expected payload Name %s, got %s", payload.Name, gottenPayload.Name)
		}
		if gottenPayload.BoardID != payload.BoardID {
			t.Fatalf("expected payload BoardID %s, got %s", payload.BoardID, gottenPayload.BoardID)
		}
		if gottenPayload.Position != payload.Position {
			t.Fatalf("expected payload Position %d, got %d", payload.Position, gottenPayload.Position)
		}
		if gottenPayload.CreatedAt != payload.CreatedAt {
			t.Fatalf("expected payload CreatedAt %s, got %s", payload.CreatedAt, gottenPayload.CreatedAt)
		}
		if gottenPayload.UpdatedAt != payload.UpdatedAt {
			t.Fatalf("expected payload UpdatedAt %s, got %s", payload.UpdatedAt, gottenPayload.UpdatedAt)
		}
	})
}

func TestSyncState(t *testing.T) {
	t.Run("upsert_sync_state", func(t *testing.T) {
		repo := setupTestDB(t)

		err := repo.UpsertSyncState(types.BoardTable, "op-123", 1234567890)
		if err != nil {
			t.Fatalf("failed to upsert sync state: %v", err)
		}

		syncState, err := repo.GetSyncState(types.BoardTable)
		if err != nil {
			t.Fatalf("failed to get sync state: %v", err)
		}

		if syncState.LastSyncedOpID != "op-123" {
			t.Errorf("expected last synced op ID 'op-123', got '%s'", syncState.LastSyncedOpID)
		}

		if syncState.LastSyncedAt != 1234567890 {
			t.Errorf("expected last synced at 1234567890, got %d", syncState.LastSyncedAt)
		}
	})

	t.Run("update_sync_state", func(t *testing.T) {
		repo := setupTestDB(t)

		err := repo.UpsertSyncState(types.BoardTable, "op-123", 1234567890)
		if err != nil {
			t.Fatalf("failed to upsert sync state: %v", err)
		}

		err = repo.UpdateSyncState(types.BoardTable, "op-456", 9876543210)
		if err != nil {
			t.Fatalf("failed to update sync state: %v", err)
		}

		syncState, err := repo.GetSyncState(types.BoardTable)
		if err != nil {
			t.Fatalf("failed to get sync state: %v", err)
		}

		if syncState.LastSyncedOpID != "op-456" {
			t.Errorf("expected last synced op ID 'op-456', got '%s'", syncState.LastSyncedOpID)
		}

		if syncState.LastSyncedAt != 9876543210 {
			t.Errorf("expected last synced at 9876543210, got %d", syncState.LastSyncedAt)
		}
	})

	t.Run("sync_state_different_tables", func(t *testing.T) {
		repo := setupTestDB(t)

		err := repo.UpsertSyncState(types.BoardTable, "board-op-1", 1000)
		if err != nil {
			t.Fatalf("failed to upsert board sync state: %v", err)
		}

		err = repo.UpsertSyncState(types.CardTable, "card-op-1", 2000)
		if err != nil {
			t.Fatalf("failed to upsert card sync state: %v", err)
		}

		boardSync, err := repo.GetSyncState(types.BoardTable)
		if err != nil {
			t.Fatalf("failed to get board sync state: %v", err)
		}

		cardSync, err := repo.GetSyncState(types.CardTable)
		if err != nil {
			t.Fatalf("failed to get card sync state: %v", err)
		}

		if boardSync.LastSyncedAt != 1000 {
			t.Errorf("expected board last synced at 1000, got %d", boardSync.LastSyncedAt)
		}

		if cardSync.LastSyncedAt != 2000 {
			t.Errorf("expected card last synced at 2000, got %d", cardSync.LastSyncedAt)
		}
	})
}

func TestAppVersion(t *testing.T) {
	t.Run("get_local_version_default", func(t *testing.T) {
		repo := setupTestDB(t)

		version, err := repo.GetLocalVersion()
		if err != nil {
			t.Fatalf("failed to get local version: %v", err)
		}

		if version != "0.0.0" {
			t.Errorf("expected default version '0.0.0', got '%s'", version)
		}
	})

	t.Run("update_local_version", func(t *testing.T) {
		repo := setupTestDB(t)

		err := repo.UpdateLocalVersion("1.2.3")
		if err != nil {
			t.Fatalf("failed to update local version: %v", err)
		}

		version, err := repo.GetLocalVersion()
		if err != nil {
			t.Fatalf("failed to get local version: %v", err)
		}

		if version != "1.2.3" {
			t.Errorf("expected version '1.2.3', got '%s'", version)
		}
	})

}

func TestExportData(t *testing.T) {
	t.Run("export_all_data", func(t *testing.T) {
		repo := setupTestDB(t)

		board, err := repo.CreateBoard("Test Board")
		if err != nil {
			t.Fatalf("failed to create board: %v", err)
		}

		column, err := repo.CreateColumn(board.ID, "Test Column")
		if err != nil {
			t.Fatalf("failed to create column: %v", err)
		}

		card, err := repo.CreateCard(column.ID, "Test Card", "Test Description")
		if err != nil {
			t.Fatalf("failed to create card: %v", err)
		}

		transcription, err := repo.AddTransscription(board.ID, "Test transcription", "/path/recording.wav")
		if err != nil {
			t.Fatalf("failed to add transcription: %v", err)
		}

		exportedData, err := repo.ExportAllData()
		if err != nil {
			t.Fatalf("failed to export data: %v", err)
		}

		if len(exportedData.Boards) != 1 {
			t.Errorf("expected 1 board, got %d", len(exportedData.Boards))
		}

		if len(exportedData.Columns) != 1 {
			t.Errorf("expected 1 column, got %d", len(exportedData.Columns))
		}

		if len(exportedData.Cards) != 1 {
			t.Errorf("expected 1 card, got %d", len(exportedData.Cards))
		}

		if len(exportedData.Transcriptions) != 1 {
			t.Errorf("expected 1 transcription, got %d", len(exportedData.Transcriptions))
		}

		if exportedData.Boards[0].ID != board.ID {
			t.Errorf("expected board ID '%s', got '%s'", board.ID, exportedData.Boards[0].ID)
		}

		if exportedData.Cards[0].Title != card.Title {
			t.Errorf("expected card title '%s', got '%s'", card.Title, exportedData.Cards[0].Title)
		}

		if exportedData.Transcriptions[0].Transcription != transcription.Transcription {
			t.Errorf("expected transcription '%s', got '%s'", transcription.Transcription, exportedData.Transcriptions[0].Transcription)
		}
	})

	t.Run("export_empty_database", func(t *testing.T) {
		repo := setupTestDB(t)

		exportedData, err := repo.ExportAllData()
		if err != nil {
			t.Fatalf("failed to export empty data: %v", err)
		}

		if len(exportedData.Boards) != 0 {
			t.Errorf("expected 0 boards, got %d", len(exportedData.Boards))
		}

		if len(exportedData.Columns) != 0 {
			t.Errorf("expected 0 columns, got %d", len(exportedData.Columns))
		}

		if len(exportedData.Cards) != 0 {
			t.Errorf("expected 0 cards, got %d", len(exportedData.Cards))
		}

		if len(exportedData.Transcriptions) != 0 {
			t.Errorf("expected 0 transcriptions, got %d", len(exportedData.Transcriptions))
		}
	})
}

func TestImportData(t *testing.T) {
	t.Run("import_board", func(t *testing.T) {
		repo := setupTestDB(t)

		board, err := repo.ImportBoard("board-123", "Imported Board", "2024-01-01", "2024-01-02")
		if err != nil {
			t.Fatalf("failed to import board: %v", err)
		}

		if board.ID != "board-123" {
			t.Errorf("expected board ID 'board-123', got '%s'", board.ID)
		}

		if board.Name != "Imported Board" {
			t.Errorf("expected board name 'Imported Board', got '%s'", board.Name)
		}
	})

	t.Run("import_column", func(t *testing.T) {
		repo := setupTestDB(t)

		board, err := repo.ImportBoard("board-123", "Test Board", "2024-01-01", "2024-01-02")
		if err != nil {
			t.Fatalf("failed to import board: %v", err)
		}

		column, err := repo.ImportColumn("column-123", board.ID, "Imported Column", 0, "2024-01-01", "2024-01-02")
		if err != nil {
			t.Fatalf("failed to import column: %v", err)
		}

		if column.ID != "column-123" {
			t.Errorf("expected column ID 'column-123', got '%s'", column.ID)
		}

		if column.Name != "Imported Column" {
			t.Errorf("expected column name 'Imported Column', got '%s'", column.Name)
		}

		if column.BoardID != board.ID {
			t.Errorf("expected board ID '%s', got '%s'", board.ID, column.BoardID)
		}
	})

	t.Run("import_card", func(t *testing.T) {
		repo := setupTestDB(t)

		board, err := repo.ImportBoard("board-123", "Test Board", "2024-01-01", "2024-01-02")
		if err != nil {
			t.Fatalf("failed to import board: %v", err)
		}

		column, err := repo.ImportColumn("column-123", board.ID, "Test Column", 0, "2024-01-01", "2024-01-02")
		if err != nil {
			t.Fatalf("failed to import column: %v", err)
		}

		card, err := repo.ImportCard("card-123", column.ID, "Imported Card", "Card description", "", "2024-01-01", "2024-01-02")
		if err != nil {
			t.Fatalf("failed to import card: %v", err)
		}

		if card.ID != "card-123" {
			t.Errorf("expected card ID 'card-123', got '%s'", card.ID)
		}

		if card.Title != "Imported Card" {
			t.Errorf("expected card title 'Imported Card', got '%s'", card.Title)
		}

		if card.Description.String != "Card description" {
			t.Errorf("expected description 'Card description', got '%s'", card.Description.String)
		}
	})

	t.Run("import_transcription", func(t *testing.T) {
		repo := setupTestDB(t)

		board, err := repo.ImportBoard("board-123", "Test Board", "2024-01-01", "2024-01-02")
		if err != nil {
			t.Fatalf("failed to import board: %v", err)
		}

		transcription, err := repo.ImportTranscription(
			"trans-123",
			board.ID,
			"Imported transcription",
			"/path/recording.wav",
			"create task",
			"Task created",
			"2024-01-01",
			"2024-01-02",
		)
		if err != nil {
			t.Fatalf("failed to import transcription: %v", err)
		}

		if transcription.ID != "trans-123" {
			t.Errorf("expected transcription ID 'trans-123', got '%s'", transcription.ID)
		}

		if transcription.Transcription != "Imported transcription" {
			t.Errorf("expected transcription 'Imported transcription', got '%s'", transcription.Transcription)
		}

		if transcription.Intent.String != "create task" {
			t.Errorf("expected intent 'create task', got '%s'", transcription.Intent.String)
		}
	})

	t.Run("import_complete_board_data", func(t *testing.T) {
		repo := setupTestDB(t)

		board, err := repo.ImportBoard("board-456", "Complete Board", "2024-01-01", "2024-01-02")
		if err != nil {
			t.Fatalf("failed to import board: %v", err)
		}

		column1, err := repo.ImportColumn("col-1", board.ID, "Column 1", 0, "2024-01-01", "2024-01-02")
		if err != nil {
			t.Fatalf("failed to import column 1: %v", err)
		}

		column2, err := repo.ImportColumn("col-2", board.ID, "Column 2", 1, "2024-01-01", "2024-01-02")
		if err != nil {
			t.Fatalf("failed to import column 2: %v", err)
		}

		_, err = repo.ImportCard("card-1", column1.ID, "Card 1", "Description 1", "", "2024-01-01", "2024-01-02")
		if err != nil {
			t.Fatalf("failed to import card 1: %v", err)
		}

		_, err = repo.ImportCard("card-2", column2.ID, "Card 2", "Description 2", "", "2024-01-01", "2024-01-02")
		if err != nil {
			t.Fatalf("failed to import card 2: %v", err)
		}

		_, err = repo.ImportTranscription("trans-1", board.ID, "Test transcription", "", "", "", "2024-01-01", "2024-01-02")
		if err != nil {
			t.Fatalf("failed to import transcription: %v", err)
		}

		columns, err := repo.ListColumnsByBoard(board.ID)
		if err != nil {
			t.Fatalf("failed to list columns: %v", err)
		}

		if len(columns) != 2 {
			t.Errorf("expected 2 columns, got %d", len(columns))
		}

		cards1, err := repo.ListCardsByColumn(column1.ID)
		if err != nil {
			t.Fatalf("failed to list cards for column 1: %v", err)
		}

		if len(cards1) != 1 {
			t.Errorf("expected 1 card in column 1, got %d", len(cards1))
		}

		transcriptions, err := repo.GetTranscriptions(board.ID, 1, 10)
		if err != nil {
			t.Fatalf("failed to get transcriptions: %v", err)
		}

		if len(transcriptions) != 1 {
			t.Errorf("expected 1 transcription, got %d", len(transcriptions))
		}
	})
}
