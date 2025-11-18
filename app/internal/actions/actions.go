package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"seisami/app/internal/repo"
	"seisami/app/internal/tools"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Action struct {
	ctx   context.Context
	repo  repo.Repository
	tools *tools.Tools
}

type StructuredResponse struct {
	Intent       string                 `json:"intent"`
	Understood   string                 `json:"understood"`
	ActionsTaken []string               `json:"actions_taken"`
	Result       string                 `json:"result"`
	Data         map[string]interface{} `json:"data,omitempty"`
}

func NewAction(ctx context.Context, repo repo.Repository) *Action {

	tools := tools.NewTools(repo, ctx)
	return &Action{
		ctx,
		repo,
		tools,
	}
}

/*
	im forcing the UUID as the column_id, because i notice an error, might be because of

small model that's executing the actions

perhaps when a better model is used, it'll know a column_id

TODO: bug fix
I mentioned bug in my input & no bug column was created, rather it just transcribed & did nothing, i should look into this also
*/
func buildPromptTemplate(transcription string, boardId string) string {
	now := time.Now().Format(time.RFC3339)
	prompt := fmt.Sprintf(`You are Seisami AI — a voice-driven assistant that interprets the user's transcription and produces a structured JSON summary of what happened. You rely entirely on available tools to interact with boards, columns, and cards.

CONTEXT:
- timestamp (RFC3339): {{%v}}
- board_id (UUID): {{%s}}
- transcription: "{{%s}}"

CORE PRINCIPLES (MUST NEVER BE BROKEN):
- “Task” means “card”.
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
`, now, boardId, transcription)

	return prompt
}

// TODO: implemented process transcription with cloud api
func (a *Action) ProcessTranscription(transcription string, boardId string) (*StructuredResponse, error) {
	runtime.EventsEmit(a.ctx, "ai:processing_start", map[string]interface{}{
		"transcription": transcription,
		"boardId":       boardId,
		"timestamp":     time.Now().Format(time.RFC3339),
	})

	prompt := buildPromptTemplate(transcription, boardId)

	settings, err := a.repo.GetSettings()
	if err != nil {
		fmt.Printf("Error getting settings, using default transcription: %v\n", err)
	}

	if !settings.OpenaiApiKey.Valid || settings.OpenaiApiKey.String == "" {
		runtime.EventsEmit(a.ctx, "ai:error", map[string]interface{}{
			"error":     "OpenAI API key not configured",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	openAiClient := openai.NewClient(settings.OpenaiApiKey.String)

	resp, err := openAiClient.CreateChatCompletion(a.ctx, openai.ChatCompletionRequest{
		Model: tools.TranscriptionModel,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.1,
		Tools:       a.tools.AvailableTools(),
	})

	if err != nil {
		runtime.EventsEmit(a.ctx, "ai:error", map[string]interface{}{
			"error":     fmt.Sprintf("OpenAI API call failed: %v", err),
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return nil, fmt.Errorf("unable to make OpenAI API call: %w", err)
	}

	if len(resp.Choices) == 0 {
		runtime.EventsEmit(a.ctx, "ai:error", map[string]interface{}{
			"error":     "No response choices returned from OpenAI",
			"timestamp": time.Now().Format(time.RFC3339),
		})
		return nil, fmt.Errorf("no response choices returned from OpenAI")
	}

	choice := resp.Choices[0]
	message := choice.Message

	var finalResponse string

	if len(message.ToolCalls) > 0 {

		finalResponse, err = a.handleToolCalls(message, prompt)
		if err != nil {
			runtime.EventsEmit(a.ctx, "ai:error", map[string]interface{}{
				"error":     fmt.Sprintf("Tool execution failed: %v", err),
				"timestamp": time.Now().Format(time.RFC3339),
			})
			return nil, err
		}
	} else {
		finalResponse = message.Content
	}

	var structuredResp StructuredResponse
	if err := json.Unmarshal([]byte(finalResponse), &structuredResp); err != nil {
		structuredResp = StructuredResponse{
			Intent:     "unknown",
			Understood: transcription,
			Result:     finalResponse,
		}
	}

	runtime.EventsEmit(a.ctx, "ai:processing_complete", map[string]interface{}{
		"intent":       structuredResp.Intent,
		"actionsTaken": structuredResp.ActionsTaken,
		"result":       structuredResp.Result,
		"timestamp":    time.Now().Format(time.RFC3339),
	})

	return &structuredResp, nil
}

func (a *Action) handleToolCalls(message openai.ChatCompletionMessage, originalPrompt string) (string, error) {
	var toolMessages []openai.ChatCompletionMessage
	toolMessages = append(toolMessages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: originalPrompt,
	})
	toolMessages = append(toolMessages, message)

	settings, err := a.repo.GetSettings()
	if err != nil {
		fmt.Printf("Error getting settings, using default transcription: %v\n", err)
	}

	if !settings.OpenaiApiKey.Valid || settings.OpenaiApiKey.String == "" {
		return "", fmt.Errorf("OpenAI API key not configured")
	}

	openAiClient := openai.NewClient(settings.OpenaiApiKey.String)
	currentMessage := message
	maxIterations := 10

	for iteration := 0; iteration < maxIterations && len(currentMessage.ToolCalls) > 0; iteration++ {
		fmt.Printf("Tool call iteration %d with %d tool calls\n", iteration+1, len(currentMessage.ToolCalls))

		for _, toolCall := range currentMessage.ToolCalls {
			fmt.Printf("Executing tool: %s\n", toolCall.Function.Name)

			result, err := a.tools.ExecuteTool(toolCall)
			if err != nil {
				result = fmt.Sprintf("Error executing tool: %s", err.Error())

				runtime.EventsEmit(a.ctx, "ai:tool_error", map[string]interface{}{
					"toolName":  toolCall.Function.Name,
					"error":     err.Error(),
					"iteration": iteration + 1,
					"timestamp": time.Now().Format(time.RFC3339),
				})
			} else {

				runtime.EventsEmit(a.ctx, "ai:tool_complete", map[string]interface{}{
					"toolName":  toolCall.Function.Name,
					"result":    result,
					"iteration": iteration + 1,
					"timestamp": time.Now().Format(time.RFC3339),
				})
			}

			toolMessages = append(toolMessages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				ToolCallID: toolCall.ID,
				Content:    result,
			})
		}

		resp, err := openAiClient.CreateChatCompletion(a.ctx, openai.ChatCompletionRequest{
			Model:       tools.TranscriptionModel,
			Messages:    toolMessages,
			Temperature: 0.1,
			Tools:       a.tools.AvailableTools(),
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
			fmt.Printf("No more tool calls, returning final response\n")
			return currentMessage.Content, nil
		}
	}

	if len(currentMessage.ToolCalls) > 0 {
		fmt.Printf("Hit max iterations (%d), stopping tool call chain\n", maxIterations)
	}

	return currentMessage.Content, nil
}
