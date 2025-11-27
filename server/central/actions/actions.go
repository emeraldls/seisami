package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"seisami/server/central/tools"
	"seisami/server/centraldb"
	"seisami/server/synchub"
	"seisami/server/types"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sashabaranov/go-openai"
)

type Action struct {
	openAIAPIKey string
	queries      *centraldb.Queries
}

func NewAction(apiKey string, queries *centraldb.Queries) *Action {
	return &Action{openAIAPIKey: apiKey, queries: queries}
}

// SSE event types
type SSEEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type SSEWriter interface {
	WriteSSE(event string, data interface{}) error
	Flush() error
}

func (a *Action) TranscribeAudio(ctx context.Context, audioData []byte) (string, error) {
	if a.openAIAPIKey == "" {
		return "", fmt.Errorf("OpenAI API key not configured on server")
	}

	client := openai.NewClient(a.openAIAPIKey)

	tmpFile, err := os.CreateTemp("", "recording-*.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(audioData); err != nil {
		return "", fmt.Errorf("failed to write audio data: %w", err)
	}

	tmpFile.Close()

	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: tmpFile.Name(),
		Language: "en",
	}

	resp, err := client.CreateTranscription(ctx, req)
	if err != nil {
		return "", fmt.Errorf("OpenAI transcription failed: %v", err)
	}

	return resp.Text, nil
}

func buildPromptTemplate(transcription string, boardID string) string {
	now := time.Now().Format(time.RFC3339)
	prompt := fmt.Sprintf(`You are Seisami AI — a voice-driven assistant that interprets the user's transcription and produces a structured JSON summary of what happened. You rely entirely on available tools to interact with boards, columns, and cards.

CONTEXT:
- timestamp (RFC3339): {{%v}}
- board_id (UUID): {{%s}}
- transcription: "{{%s}}"

CORE PRINCIPLES (MUST NEVER BE BROKEN):
- "Task" means "card".
- You do not invent IDs. 'column_id' and 'board_id' must be valid UUIDs.
- You never substitute a column name where a column_id is required.
- If a card requires a column: the column must exist. If it doesn't, create it.
- You may extract multiple tasks from a single transcription.
- If dates or times are mentioned, interpret them using the provided timestamp.

TOOL RULES:
- When you need column IDs, use 'list_columns_by_board'.
- When you need a new column, call 'create_column' and use its returned UUID.
- When creating or modifying cards, you must supply a real column_id.

RESPONSE CONTRACT:
You must return valid JSON:

{
  "intent": "string",
  "understood": "natural language interpretation of the transcription",
  "actions_taken": ["list of actions performed"],
  "result": "summary of what the system accomplished",
  "data": {} 
}

NOTES:
- You are not chatting; you are generating state change summaries.
- No invented data, no missing UUIDs, no invalid JSON.
`, now, boardID, transcription)

	return prompt
}

func (a *Action) ProcessTranscriptionWithSSE(ctx context.Context, userID uuid.UUID, transcription, boardID string, writer SSEWriter) (*types.ProcessTranscriptionResponse, error) {
	boardUUID, err := uuid.Parse(boardID)
	if err != nil {
		return nil, fmt.Errorf("invalid board ID: %w", err)
	}

	err = a.ensureBoardAccess(ctx, boardUUID, userID)
	if err != nil {
		return nil, fmt.Errorf("board access denied: %w", err)
	}

	_ = writer.WriteSSE("ai:processing_start", map[string]interface{}{
		"transcription": transcription,
		"boardId":       boardID,
		"timestamp":     time.Now().Format(time.RFC3339),
	})
	_ = writer.Flush()

	prompt := buildPromptTemplate(transcription, boardID)

	openAiClient := openai.NewClient(a.openAIAPIKey)

	toolsInstance := tools.NewTools(a.queries, ctx, userID, boardUUID)

	toolsInstance.NotifySync = func(uid, tableName string) {
		if hub := synchub.Get(); hub != nil {
			hub.NotifyUserSync(uid, tableName)
		}
	}

	resp, err := openAiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: tools.TranscriptionModel,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.1,
		Tools:       toolsInstance.AvailableTools(),
	})

	if err != nil {
		_ = writer.WriteSSE("ai:error", map[string]interface{}{
			"error":     fmt.Sprintf("OpenAI API call failed: %v", err),
			"timestamp": time.Now().Format(time.RFC3339),
		})
		_ = writer.Flush()
		return nil, fmt.Errorf("unable to make OpenAI API call: %w", err)
	}

	if len(resp.Choices) == 0 {
		_ = writer.WriteSSE("ai:error", map[string]interface{}{
			"error":     "No response choices returned from OpenAI",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		_ = writer.Flush()
		return nil, fmt.Errorf("no response choices returned from OpenAI")
	}

	choice := resp.Choices[0]
	message := choice.Message

	var finalResponse string

	if len(message.ToolCalls) > 0 {
		finalResponse, err = a.handleToolCallsWithSSE(ctx, openAiClient, message, prompt, toolsInstance, writer)
		if err != nil {
			_ = writer.WriteSSE("ai:error", map[string]interface{}{
				"error":     fmt.Sprintf("Tool execution failed: %v", err),
				"timestamp": time.Now().Format(time.RFC3339),
			})
			_ = writer.Flush()
			return nil, err
		}
	} else {
		finalResponse = message.Content
	}

	var structuredResp types.ProcessTranscriptionResponse
	if err := json.Unmarshal([]byte(finalResponse), &structuredResp); err != nil {
		structuredResp = types.ProcessTranscriptionResponse{
			Intent:     "unknown",
			Understood: transcription,
			Result:     finalResponse,
		}
	}

	_ = writer.WriteSSE("ai:processing_complete", map[string]interface{}{
		"intent":       structuredResp.Intent,
		"actionsTaken": structuredResp.ActionsTaken,
		"result":       structuredResp.Result,
		"timestamp":    time.Now().Format(time.RFC3339),
	})
	_ = writer.Flush()

	return &structuredResp, nil
}

func (a *Action) handleToolCallsWithSSE(ctx context.Context, client *openai.Client, message openai.ChatCompletionMessage, originalPrompt string, toolsInstance *tools.Tools, writer SSEWriter) (string, error) {
	var toolMessages []openai.ChatCompletionMessage
	toolMessages = append(toolMessages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: originalPrompt,
	})
	toolMessages = append(toolMessages, message)

	currentMessage := message
	maxIterations := 10

	for iteration := 0; iteration < maxIterations && len(currentMessage.ToolCalls) > 0; iteration++ {
		for _, toolCall := range currentMessage.ToolCalls {
			result, err := toolsInstance.ExecuteTool(toolCall)
			if err != nil {
				result = fmt.Sprintf("Error executing tool: %s", err.Error())

				_ = writer.WriteSSE("ai:tool_error", map[string]interface{}{
					"toolName":  toolCall.Function.Name,
					"error":     err.Error(),
					"iteration": iteration + 1,
					"timestamp": time.Now().Format(time.RFC3339),
				})
				_ = writer.Flush()
			} else {
				_ = writer.WriteSSE("ai:tool_complete", map[string]interface{}{
					"toolName":  toolCall.Function.Name,
					"result":    result,
					"iteration": iteration + 1,
					"timestamp": time.Now().Format(time.RFC3339),
				})
				_ = writer.Flush()
			}

			toolMessages = append(toolMessages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				ToolCallID: toolCall.ID,
				Content:    result,
			})
		}

		resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:       tools.TranscriptionModel,
			Messages:    toolMessages,
			Temperature: 0.1,
			Tools:       toolsInstance.AvailableTools(),
		})

		if err != nil {
			return "", fmt.Errorf("unable to make follow-up OpenAI API call: %w", err)
		}

		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("no response choices returned from follow-up call")
		}

		currentMessage = resp.Choices[0].Message
		toolMessages = append(toolMessages, currentMessage)

		if len(currentMessage.ToolCalls) == 0 {
			return currentMessage.Content, nil
		}
	}

	return currentMessage.Content, nil
}

// TODO: copied this from sync_service, unify things later
func (a *Action) ensureBoardAccess(ctx context.Context, boardID, userID uuid.UUID) error {
	hasAccess, err := a.queries.ValidateBoardAccess(ctx, centraldb.ValidateBoardAccessParams{
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
