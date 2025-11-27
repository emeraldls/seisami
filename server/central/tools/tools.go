package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"seisami/server/centraldb"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sashabaranov/go-openai"
)

var TranscriptionModel = openai.GPT4o

type ToolHandler func(args json.RawMessage, queries *centraldb.Queries, ctx context.Context, userID, boardID uuid.UUID) (string, error)

type Tools struct {
	openAiTools   []openai.Tool
	ctx           context.Context
	queries       *centraldb.Queries
	userID        uuid.UUID
	boardID       uuid.UUID
	toolsRegistry map[string]ToolHandler
	NotifySync    func(userID, tableName string)
}

func NewTools(queries *centraldb.Queries, ctx context.Context, userID, boardID uuid.UUID) *Tools {
	t := &Tools{
		queries:       queries,
		ctx:           ctx,
		userID:        userID,
		boardID:       boardID,
		openAiTools:   make([]openai.Tool, 0),
		toolsRegistry: make(map[string]ToolHandler),
	}

	t.registerTools()

	return t
}

func (t *Tools) AvailableTools() []openai.Tool {
	return t.openAiTools
}

func (t *Tools) ExecuteTool(toolCall openai.ToolCall) (string, error) {
	handler, ok := t.toolsRegistry[toolCall.Function.Name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", toolCall.Function.Name)
	}

	return handler(json.RawMessage(toolCall.Function.Arguments), t.queries, t.ctx, t.userID, t.boardID)
}

func (t *Tools) registerTools() {
	t.HandleListColumns()
	t.HandleCreateColumn()
	t.HandleCreateCard()
	t.HandleUpdateCard()
	t.HandleMoveCard()
}

// Tool parameter types
type listColumnsParameter struct {
	BoardID string `json:"board_id"`
}

type createColumnParameter struct {
	BoardID    string `json:"board_id"`
	ColumnName string `json:"column_name"`
}

type createCardParameter struct {
	ColumnID    string `json:"column_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type updateCardParameter struct {
	CardID      string `json:"card_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type moveCardParameter struct {
	CardID   string `json:"card_id"`
	ColumnID string `json:"column_id"`
}

func (t *Tools) HandleListColumns() {
	handler := func(args json.RawMessage, queries *centraldb.Queries, ctx context.Context, userID, boardID uuid.UUID) (string, error) {
		var params listColumnsParameter
		if err := json.Unmarshal(args, &params); err != nil {
			return "", err
		}

		targetBoardID := boardID
		if params.BoardID != "" {
			parsed, err := uuid.Parse(params.BoardID)
			if err != nil {
				return "", fmt.Errorf("invalid board_id: %w", err)
			}
			targetBoardID = parsed
		}

		columns, err := queries.GetBoardColumns(ctx, pgtype.UUID{Bytes: targetBoardID, Valid: true})
		if err != nil {
			return "", err
		}

		res, _ := json.MarshalIndent(columns, "", " ")
		return string(res), nil
	}

	t.toolsRegistry["list_columns_by_board"] = handler

	listColumnsTool := openai.Tool{
		Type: "function",
		Function: &openai.FunctionDefinition{
			Name:        "list_columns_by_board",
			Description: "List all columns in a board",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"board_id": map[string]any{
						"type":        "string",
						"description": "The ID of the board (optional, defaults to current board)",
					},
				},
			},
		},
	}

	t.openAiTools = append(t.openAiTools, listColumnsTool)
}

func (t *Tools) HandleCreateColumn() {
	fmt.Printf("\n\ncreating column\n\n....")
	handler := func(args json.RawMessage, queries *centraldb.Queries, ctx context.Context, userID, boardID uuid.UUID) (string, error) {
		var params createColumnParameter
		if err := json.Unmarshal(args, &params); err != nil {
			return "", err
		}

		targetBoardID := boardID
		if params.BoardID != "" {
			parsed, err := uuid.Parse(params.BoardID)
			if err != nil {
				return "", fmt.Errorf("invalid board_id: %w", err)
			}
			targetBoardID = parsed
		}

		// Get max position
		columns, err := queries.GetBoardColumns(ctx, pgtype.UUID{Bytes: targetBoardID, Valid: true})
		if err != nil {
			return "", err
		}

		maxPosition := int32(0)
		for _, col := range columns {
			if col.Position > maxPosition {
				maxPosition = col.Position
			}
		}

		id := uuid.New().String()
		now := time.Now()

		column, err := queries.CreateColumn(ctx, centraldb.CreateColumnParams{
			Name:     params.ColumnName,
			Position: maxPosition + 1,
			BoardID:  pgtype.UUID{Bytes: targetBoardID, Valid: true},
			ID:       id,
			CreatedAt: pgtype.Timestamptz{
				Time:  now,
				Valid: true,
			},
			UpdatedAt: pgtype.Timestamptz{
				Time:  now,
				Valid: true,
			},
		})
		if err != nil {
			return "", err
		}

		payload, _ := json.Marshal(map[string]interface{}{
			"id":         id,
			"board_id":   targetBoardID.String(),
			"name":       params.ColumnName,
			"position":   maxPosition + 1,
			"created_at": now.Format("2006-01-02 15:04:05"),
			"updated_at": now.Format("2006-01-02 15:04:05"),
		})

		_ = queries.CreateOperation(ctx, centraldb.CreateOperationParams{
			ID:            uuid.New().String(),
			TableName:     "columns",
			RecordID:      id,
			OperationType: "insert",
			DeviceID:      pgtype.Text{String: "cloud", Valid: true},
			Payload:       string(payload),
			CreatedAt:     pgtype.Text{String: now.Format("2006-01-02 15:04:05"), Valid: true},
			UpdatedAt:     pgtype.Text{String: now.Format("2006-01-02 15:04:05"), Valid: true},
		})

		t.NotifySync(userID.String(), "columns")

		res, _ := json.MarshalIndent(column, "", " ")
		return string(res), nil
	}

	t.toolsRegistry["create_column"] = handler

	createColumnTool := openai.Tool{
		Type: "function",
		Function: &openai.FunctionDefinition{
			Name:        "create_column",
			Description: "Create a new column in a board",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"board_id": map[string]any{
						"type":        "string",
						"description": "The ID of the board (optional, defaults to current board)",
					},
					"column_name": map[string]any{
						"type":        "string",
						"description": "The name of the new column",
					},
				},
				"required": []string{"column_name"},
			},
		},
	}

	t.openAiTools = append(t.openAiTools, createColumnTool)
}

func (t *Tools) HandleCreateCard() {
	fmt.Println("creating card....")
	handler := func(args json.RawMessage, queries *centraldb.Queries, ctx context.Context, userID, boardID uuid.UUID) (string, error) {
		var params createCardParameter
		if err := json.Unmarshal(args, &params); err != nil {
			return "", err
		}

		id := uuid.New().String()
		now := time.Now()

		card, err := queries.CreateCard(ctx, centraldb.CreateCardParams{
			Title:       params.Title,
			Description: pgtype.Text{String: params.Description, Valid: params.Description != ""},
			ColumnID:    params.ColumnID,
			CreatedAt: pgtype.Timestamptz{
				Time:  now,
				Valid: true,
			},
			UpdatedAt: pgtype.Timestamptz{
				Time:  now,
				Valid: true,
			},
			ID: id,
		})
		if err != nil {
			return "", err
		}

		payload, _ := json.Marshal(map[string]interface{}{
			"id":          id,
			"column_id":   params.ColumnID,
			"title":       params.Title,
			"description": params.Description,
			"created_at":  now.Format("2006-01-02 15:04:05"),
			"updated_at":  now.Format("2006-01-02 15:04:05"),
		})

		fmt.Println("Now: ", now.Format("2006-01-02 15:04:05"))

		_ = queries.CreateOperation(ctx, centraldb.CreateOperationParams{
			ID:            uuid.New().String(),
			TableName:     "cards",
			RecordID:      id,
			OperationType: "insert",
			DeviceID:      pgtype.Text{String: "cloud", Valid: true},
			Payload:       string(payload),
			CreatedAt:     pgtype.Text{String: now.Format("2006-01-02 15:04:05"), Valid: true},
			UpdatedAt:     pgtype.Text{String: now.Format("2006-01-02 15:04:05"), Valid: true},
		})

		t.NotifySync(userID.String(), "cards")

		res, _ := json.MarshalIndent(card, "", " ")
		return string(res), nil
	}

	t.toolsRegistry["create_card"] = handler

	createCardTool := openai.Tool{
		Type: "function",
		Function: &openai.FunctionDefinition{
			Name:        "create_card",
			Description: "Create a new card in a column",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"column_id": map[string]any{
						"type":        "string",
						"description": "The ID of the column",
					},
					"title": map[string]any{
						"type":        "string",
						"description": "The title of the card",
					},
					"description": map[string]any{
						"type":        "string",
						"description": "The description of the card",
					},
				},
				"required": []string{"column_id", "title", "description"},
			},
		},
	}

	t.openAiTools = append(t.openAiTools, createCardTool)
}

func (t *Tools) HandleUpdateCard() {
	handler := func(args json.RawMessage, queries *centraldb.Queries, ctx context.Context, userID, boardID uuid.UUID) (string, error) {
		var params updateCardParameter
		if err := json.Unmarshal(args, &params); err != nil {
			return "", err
		}

		now := time.Now()

		err := queries.SyncUpsertCard(ctx, centraldb.SyncUpsertCardParams{
			ID:          params.CardID,
			Title:       params.Title,
			Description: pgtype.Text{String: params.Description, Valid: params.Description != ""},
			UpdatedAt: pgtype.Timestamptz{
				Time:  now,
				Valid: true,
			},
		})
		if err != nil {
			return "", err
		}

		payload, _ := json.Marshal(map[string]interface{}{
			"id":          params.CardID,
			"title":       params.Title,
			"description": params.Description,
			"updated_at":  now.Format("2006-01-02 15:04:05"),
		})

		_ = queries.CreateOperation(ctx, centraldb.CreateOperationParams{
			ID:            uuid.New().String(),
			TableName:     "cards",
			RecordID:      params.CardID,
			OperationType: "update",
			DeviceID:      pgtype.Text{String: "cloud", Valid: true},
			Payload:       string(payload),
			CreatedAt:     pgtype.Text{String: now.Format("2006-01-02 15:04:05"), Valid: true},
			UpdatedAt:     pgtype.Text{String: now.Format("2006-01-02 15:04:05"), Valid: true},
		})

		t.NotifySync(userID.String(), "cards")

		return fmt.Sprintf(`{"card_id": "%s", "status": "updated"}`, params.CardID), nil
	}

	t.toolsRegistry["update_card"] = handler

	updateCardTool := openai.Tool{
		Type: "function",
		Function: &openai.FunctionDefinition{
			Name:        "update_card",
			Description: "Update a card's title and description",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"card_id": map[string]any{
						"type":        "string",
						"description": "The ID of the card to update",
					},
					"title": map[string]any{
						"type":        "string",
						"description": "The new title",
					},
					"description": map[string]any{
						"type":        "string",
						"description": "The new description",
					},
				},
				"required": []string{"card_id", "title", "description"},
			},
		},
	}

	t.openAiTools = append(t.openAiTools, updateCardTool)
}

func (t *Tools) HandleMoveCard() {
	handler := func(args json.RawMessage, queries *centraldb.Queries, ctx context.Context, userID, boardID uuid.UUID) (string, error) {
		var params moveCardParameter
		if err := json.Unmarshal(args, &params); err != nil {
			return "", err
		}

		now := time.Now()

		err := queries.SyncUpdateCardColumn(ctx, centraldb.SyncUpdateCardColumnParams{
			ID:       params.CardID,
			ColumnID: params.ColumnID,
			UpdatedAt: pgtype.Timestamptz{
				Time:  now,
				Valid: true,
			},
		})
		if err != nil {
			return "", err
		}

		payload, _ := json.Marshal(map[string]interface{}{
			"card_id": params.CardID,
			"new_column": map[string]string{
				"id": params.ColumnID,
			},
		})

		_ = queries.CreateOperation(ctx, centraldb.CreateOperationParams{
			ID:            uuid.New().String(),
			TableName:     "cards",
			RecordID:      params.CardID,
			OperationType: "update-card-column",
			DeviceID:      pgtype.Text{String: "cloud", Valid: true},
			Payload:       string(payload),
			CreatedAt:     pgtype.Text{String: now.Format("2006-01-02 15:04:05"), Valid: true},
			UpdatedAt:     pgtype.Text{String: now.Format("2006-01-02 15:04:05"), Valid: true},
		})

		t.NotifySync(userID.String(), "cards")

		return fmt.Sprintf(`{"card_id": "%s", "column_id": "%s", "status": "moved"}`, params.CardID, params.ColumnID), nil
	}

	t.toolsRegistry["move_card"] = handler

	moveCardTool := openai.Tool{
		Type: "function",
		Function: &openai.FunctionDefinition{
			Name:        "move_card",
			Description: "Move a card to a different column",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"card_id": map[string]any{
						"type":        "string",
						"description": "The ID of the card to move",
					},
					"column_id": map[string]any{
						"type":        "string",
						"description": "The ID of the target column",
					},
				},
				"required": []string{"card_id", "column_id"},
			},
		},
	}

	t.openAiTools = append(t.openAiTools, moveCardTool)
}
