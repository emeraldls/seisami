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

func buildPromptTemplate(transcription string, boardId string) string {
	now := time.Now().Format(time.RFC3339)
	prompt := fmt.Sprintf(`You are Seisami AI — a minimalist, high-precision desktop-first productivity assistant. You help users manage their boards and tasks through voice commands.

CONTEXT:
- Current date time: %s (RFC3339)
- Current Board ID: %s
- User said: "%s"

IMPORTANT CONCEPTS:
- Tasks = Cards in this system. When users mention "tasks", they mean cards.
- If a column for a task doesn't exist, create it first, then add the card to that column.
- Extract multiple tasks from a single transcription when mentioned.
- Use appropriate column names like "To Do", "In Progress", "Done", "Backlog", etc.
- Context Awareness: Use the current datetime (provided as {{%s}} RFC3339) to resolve relative times like "Saturday" or "next week." Handle multi-step commands by sequencing actions logically.

CRITICAL DATA VALIDATION:
- column_id is a UUID (e.g., "550e8400-e29b-41d4-a716-446655440000")
- column name is the human-readable text (e.g., "To Do", "In Progress")
- NEVER use column name as column_id
- ALWAYS use list_columns_by_board first to get actual column IDs
- When creating a card, you MUST provide the UUID column_id, not the column name
- If you need to create a column, the create_column tool will return a UUID that you must use
- board_id is also a UUID - use the one provided in the context

WORKFLOW FOR CREATING CARDS:
1. Use list_columns_by_board to get existing columns with their UUIDs
2. If the column doesn't exist, use create_column to create it (you'll get a UUID back)
3. Use the UUID from step 1 or 2 as the column_id when creating a card
4. NEVER make up or guess column IDs - always fetch or create them

INSTRUCTIONS:
1. Use available tools when you need to fetch or modify data
2. After using tools (or if no tools needed), provide a structured response that includes:
   - What you understood from the user's request
   - What actions were taken (if any)
   - The results or next steps

RESPONSE FORMAT:
Always respond with a JSON object containing:
{
  "intent": "string - what the user wanted (e.g., 'create_task', 'read_board', 'move_task')",
  "understood": "string - natural language summary of what you understood",
  "actions_taken": ["array of actions performed"],
  "result": "string - the main result or answer",
  "data": {} // optional - any structured data returned from tools
}

EXAMPLE RESPONSE:
{
 "intent": "create_task",
 "understood": "The user wants to schedule several tasks: visit their girlfriend by 5pm tomorrow, watch a play before a 10am meeting with Oluwasamwe and Smart, and complete house chores before these events.",
 "actions_taken": [
  "Listed existing columns to get their UUIDs",
  "Created 3 cards in the appropriate column using the column UUID"
 ],
 "result": "Created tasks for visiting girlfriend, watching a play, and completing house chores. All tasks have been added to the 'To Do' column.",
}

IMPORTANT: This is NOT a chat interface. The response will be saved as a summary with the transcription. Focus on providing a clear, actionable summary of what was accomplished.`, now, boardId, transcription, now)

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
