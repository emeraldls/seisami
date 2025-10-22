package repo

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"seisami/app/internal/repo/sqlc/query"

	_ "embed"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type repo struct {
	queries *query.Queries
	ctx     context.Context
}

//go:embed sqlc/schema.sql
var schema string

func dbPath() string {
	home, _ := os.UserHomeDir()
	appDir := filepath.Join(home, "Library", "Application Support", "Seisami")
	os.MkdirAll(appDir, 0755)
	return filepath.Join(appDir, "seisami.db")
}

func NewRepo() *repo {
	ctx := context.Background()
	db, err := sql.Open("sqlite3", dbPath())
	if err != nil {
		log.Fatalf("unable to setup sqlite: %v\n", err)
	}

	if _, err := db.ExecContext(ctx, schema); err != nil {
		log.Fatalf("unable to create tables: %v\n", err)
	}

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
