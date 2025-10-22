package tools

import (
	"encoding/json"
	"fmt"
	"seisami/app/internal/repo"

	"github.com/sashabaranov/go-openai"
	"golang.org/x/net/context"
)

var TranscriptionModel = openai.GPT4Dot1

type ToolHandler func(args json.RawMessage, repo repo.Repository) (string, error)

type Tools struct {
	openAiTools   []openai.Tool
	ctx           context.Context
	repo          repo.Repository
	toolsRegistry map[string]ToolHandler
}

func NewTools(repo repo.Repository, ctx context.Context) *Tools {
	t := &Tools{
		repo:          repo,
		ctx:           ctx,
		openAiTools:   make([]openai.Tool, 0),
		toolsRegistry: make(map[string]ToolHandler),
	}

	t.registerTools()

	return t
}

func (t *Tools) AvailableTools() []openai.Tool {
	return t.openAiTools
}

func (t *Tools) registerTools() {
	t.HandleReadBoard()
	t.HandleListBoards()
	t.HandleSearchColumns()
	t.HandleListCards()
	t.HandleMoveCard()
	t.HandleCreateCard()
	t.HandleUpdateCard()
	t.HandleCreateColumn()
}

type readBoardParameter struct {
	BoardID string `json:"board_id" validate:"required"`
}

type listBoardsParameter struct {
	Page     int64 `json:"page,omitempty"`
	PageSize int64 `json:"page_size,omitempty"`
}

type searchColumnsParameter struct {
	BoardID     string `json:"board_id" validate:"required"`
	SearchQuery string `json:"search_query" validate:"required"`
}

type listCardsParameter struct {
	ColumnID string `json:"column_id" validate:"required"`
}

type moveCardParameter struct {
	CardID   string `json:"card_id" validate:"required"`
	ColumnID string `json:"column_id" validate:"required"`
}

type createCardParameter struct {
	ColumnID    string `json:"column_id" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
}

type updateCardParameter struct {
	CardID      string `json:"card_id" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
}

type createColumnParameter struct {
	BoardID    string `json:"board_id" validate:"required"`
	ColumnName string `json:"column_name" validate:"required"`
}

func (t *Tools) HandleReadBoard() {
	handler := func(args json.RawMessage, repo repo.Repository) (string, error) {
		var params readBoardParameter
		if err := json.Unmarshal(args, &params); err != nil {
			return "", err
		}

		columns, err := repo.ListColumnsByBoard(params.BoardID)
		if err != nil {
			return "", err
		}

		res, _ := json.MarshalIndent(columns, "", " ")
		fmt.Println("Columns: ", string(res))

		return string(res), nil
	}

	t.toolsRegistry["read_board"] = handler

	readboardTool := openai.Tool{
		Type: "function",
		Function: &openai.FunctionDefinition{
			Name:        "read_board",
			Description: "Fetch board details by ID",
			Strict:      false,
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"board_id": map[string]any{
						"type":        "string",
						"description": "The ID of the board to fetch",
					},
				},
				"required": []string{"board_id"},
			},
		},
	}

	t.openAiTools = append(t.openAiTools, readboardTool)
}

func (t *Tools) HandleListBoards() {
	handler := func(args json.RawMessage, repo repo.Repository) (string, error) {
		var params listBoardsParameter
		if err := json.Unmarshal(args, &params); err != nil {
			return "", err
		}

		if params.Page == 0 {
			params.Page = 1
		}
		if params.PageSize == 0 {
			params.PageSize = 10
		}

		boards, err := repo.GetAllBoards(params.Page, params.PageSize)
		if err != nil {
			return "", err
		}

		res, _ := json.Marshal(boards)
		return string(res), nil
	}

	t.toolsRegistry["list_boards"] = handler

	listBoardsTool := openai.Tool{
		Type: "function",
		Function: &openai.FunctionDefinition{
			Name:        "list_boards",
			Description: "List all available boards with optional pagination",
			Strict:      false,
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"page": map[string]any{
						"type":        "integer",
						"description": "Page number (default: 1)",
					},
					"page_size": map[string]any{
						"type":        "integer",
						"description": "Number of boards per page (default: 10)",
					},
				},
				"required": []string{},
			},
		},
	}

	t.openAiTools = append(t.openAiTools, listBoardsTool)
}

func (t *Tools) HandleSearchColumns() {
	handler := func(args json.RawMessage, repo repo.Repository) (string, error) {
		var params searchColumnsParameter
		if err := json.Unmarshal(args, &params); err != nil {
			return "", err
		}

		columns, err := repo.SearchColumnsByBoardAndName(params.BoardID, params.SearchQuery)
		if err != nil {
			return "", err
		}

		res, _ := json.MarshalIndent(columns, "", " ")
		return string(res), nil
	}

	t.toolsRegistry["search_columns"] = handler

	searchColumnsTool := openai.Tool{
		Type: "function",
		Function: &openai.FunctionDefinition{
			Name:        "search_columns",
			Description: "Search for columns by name within a specific board",
			Strict:      false,
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"board_id": map[string]any{
						"type":        "string",
						"description": "The ID of the board to search in",
					},
					"search_query": map[string]any{
						"type":        "string",
						"description": "The search query to match column names",
					},
				},
				"required": []string{"board_id", "search_query"},
			},
		},
	}

	t.openAiTools = append(t.openAiTools, searchColumnsTool)
}

func (t *Tools) HandleListCards() {
	handler := func(args json.RawMessage, repo repo.Repository) (string, error) {
		var params listCardsParameter
		if err := json.Unmarshal(args, &params); err != nil {
			return "", err
		}

		cards, err := repo.ListCardsByColumn(params.ColumnID)
		if err != nil {
			return "", err
		}

		res, _ := json.MarshalIndent(cards, "", " ")
		return string(res), nil
	}

	t.toolsRegistry["list_cards"] = handler

	listCardsTool := openai.Tool{
		Type: "function",
		Function: &openai.FunctionDefinition{
			Name:        "list_cards",
			Description: "List all cards in a specific column",
			Strict:      false,
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"column_id": map[string]any{
						"type":        "string",
						"description": "The ID of the column to list cards from",
					},
				},
				"required": []string{"column_id"},
			},
		},
	}

	t.openAiTools = append(t.openAiTools, listCardsTool)
}

func (t *Tools) HandleMoveCard() {
	handler := func(args json.RawMessage, repo repo.Repository) (string, error) {
		var params moveCardParameter
		if err := json.Unmarshal(args, &params); err != nil {
			return "", err
		}

		updatedCard, err := repo.UpdateCardColumn(params.CardID, params.ColumnID)
		if err != nil {
			return "", err
		}

		res, _ := json.MarshalIndent(updatedCard, "", " ")
		return string(res), nil
	}

	t.toolsRegistry["move_card"] = handler

	moveCardTool := openai.Tool{
		Type: "function",
		Function: &openai.FunctionDefinition{
			Name:        "move_card",
			Description: "Move a card to a different column",
			Strict:      false,
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"card_id": map[string]any{
						"type":        "string",
						"description": "The ID of the card to move",
					},
					"column_id": map[string]any{
						"type":        "string",
						"description": "The ID of the target column to move the card to",
					},
				},
				"required": []string{"card_id", "column_id"},
			},
		},
	}

	t.openAiTools = append(t.openAiTools, moveCardTool)
}

func (t *Tools) HandleCreateCard() {
	handler := func(args json.RawMessage, repo repo.Repository) (string, error) {
		var params createCardParameter
		if err := json.Unmarshal(args, &params); err != nil {
			return "", err
		}

		newCard, err := repo.CreateCard(params.ColumnID, params.Title, params.Description)
		if err != nil {
			return "", err
		}

		res, _ := json.MarshalIndent(newCard, "", " ")
		return string(res), nil
	}

	t.toolsRegistry["create_card"] = handler

	createCardTool := openai.Tool{
		Type: "function",
		Function: &openai.FunctionDefinition{
			Name:        "create_card",
			Description: "Create a new card in a specific column",
			Strict:      false,
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"column_id": map[string]any{
						"type":        "string",
						"description": "The ID of the column to create the card in",
					},
					"title": map[string]any{
						"type":        "string",
						"description": "The title of the new card",
					},
					"description": map[string]any{
						"type":        "string",
						"description": "Well detailed explanation of the new card",
					},
				},
				"required": []string{"column_id", "title", "description"},
			},
		},
	}

	t.openAiTools = append(t.openAiTools, createCardTool)
}

func (t *Tools) HandleUpdateCard() {
	handler := func(args json.RawMessage, repo repo.Repository) (string, error) {
		var params updateCardParameter
		if err := json.Unmarshal(args, &params); err != nil {
			return "", err
		}

		updatedCard, err := repo.UpdateCard(params.CardID, params.Title, params.Description)
		if err != nil {
			return "", err
		}

		res, _ := json.MarshalIndent(updatedCard, "", " ")
		return string(res), nil
	}

	t.toolsRegistry["update_card"] = handler

	updateCardTool := openai.Tool{
		Type: "function",
		Function: &openai.FunctionDefinition{
			Name:        "update_card",
			Description: "Update an existing card's title and description",
			Strict:      false,
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"card_id": map[string]any{
						"type":        "string",
						"description": "The ID of the card to update",
					},
					"title": map[string]any{
						"type":        "string",
						"description": "The new title for the card",
					},
					"description": map[string]any{
						"type":        "string",
						"description": "The new description for the card",
					},
				},
				"required": []string{"card_id", "title", "description"},
			},
		},
	}

	t.openAiTools = append(t.openAiTools, updateCardTool)
}

func (t *Tools) HandleCreateColumn() {
	handler := func(args json.RawMessage, repo repo.Repository) (string, error) {
		var params createColumnParameter
		if err := json.Unmarshal(args, &params); err != nil {
			return "", err
		}

		newColumn, err := repo.CreateColumn(params.BoardID, params.ColumnName)
		if err != nil {
			return "", err
		}

		res, _ := json.MarshalIndent(newColumn, "", " ")
		return string(res), nil
	}

	t.toolsRegistry["create_column"] = handler

	createColumnTool := openai.Tool{
		Type: "function",
		Function: &openai.FunctionDefinition{
			Name:        "create_column",
			Description: "Create a new column in a specific board",
			Strict:      false,
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"board_id": map[string]any{
						"type":        "string",
						"description": "The ID of the board to create the column in",
					},
					"column_name": map[string]any{
						"type":        "string",
						"description": "The name of the new column",
					},
				},
				"required": []string{"board_id", "column_name"},
			},
		},
	}

	t.openAiTools = append(t.openAiTools, createColumnTool)
}

func (t *Tools) ExecuteTool(toolCall openai.ToolCall) (string, error) {
	handler, exists := t.toolsRegistry[toolCall.Function.Name]
	if !exists {
		return "", fmt.Errorf("unknown tool: %s", toolCall.Function.Name)
	}

	return handler([]byte(toolCall.Function.Arguments), t.repo)
}
