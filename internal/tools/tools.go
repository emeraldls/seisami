package tools

import (
	"encoding/json"
	"fmt"
	"seisami/internal/repo"

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
}

type readBoardParameter struct {
	BoardID string `json:"board_id" validate:"required"`
}

type listBoardsParameter struct {
	Page     int64 `json:"page,omitempty"`
	PageSize int64 `json:"page_size,omitempty"`
}

func (t *Tools) HandleReadBoard() {
	handler := func(args json.RawMessage, repo repo.Repository) (string, error) {
		var params readBoardParameter
		if err := json.Unmarshal(args, &params); err != nil {
			return "", err
		}

		board, err := repo.GetBoard(params.BoardID)
		if err != nil {
			return "", err
		}

		res, _ := json.MarshalIndent(board, "", " ")
		fmt.Println("Board: ", string(res))

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

func (t *Tools) ExecuteTool(toolCall openai.ToolCall) (string, error) {
	handler, exists := t.toolsRegistry[toolCall.Function.Name]
	if !exists {
		return "", fmt.Errorf("unknown tool: %s", toolCall.Function.Name)
	}

	return handler([]byte(toolCall.Function.Arguments), t.repo)
}
