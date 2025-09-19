package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"seisami/internal/repo"
	"seisami/internal/tools"
	"time"

	"github.com/sashabaranov/go-openai"
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
- Current time: %s (RFC3339)
- Current Board ID: %s
- User said: "%s"

IMPORTANT CONCEPTS:
- Tasks = Cards in this system. When users mention "tasks", they mean cards.
- If a column for a task doesn't exist, create it first, then add the card to that column.
- Extract multiple tasks from a single transcription when mentioned.
- Use appropriate column names like "To Do", "In Progress", "Done", "Backlog", etc.

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
  "Fetched current board to prepare for task creation"
 ],
 "result": "Ready to create tasks for visiting girlfriend, watching a play, attending a meeting, and completing house chores as per the user's schedule.",
 "data": {
  "tasks_to_create": [
   {
    "due": "2025-09-20T17:00:00+01:00",
    "title": "Visit girlfriend"
   },
   {
    "due": "2025-09-20T09:30:00+01:00",
    "title": "Watch a play before 10am meeting"
   },
   {
    "due": "2025-09-20T10:00:00+01:00",
    "title": "Meeting with Oluwasamwe and Smart"
   },
   {
    "due": "2025-09-20T09:00:00+01:00",
    "title": "Complete house chores before play and meeting"
   }
  ]
 }
}

IMPORTANT: This is NOT a chat interface. The response will be saved as a summary with the transcription. Focus on providing a clear, actionable summary of what was accomplished.`, now, boardId, transcription)

	return prompt
}

// TODO: implemented process transcription with cloud api
func (a *Action) ProcessTranscription(transcription string, boardId string) (*StructuredResponse, error) {
	prompt := buildPromptTemplate(transcription, boardId)

	settings, err := a.repo.GetSettings()
	if err != nil {
		fmt.Printf("Error getting settings, using default transcription: %v\n", err)
	}

	if !settings.OpenaiApiKey.Valid || settings.OpenaiApiKey.String == "" {
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
		return nil, fmt.Errorf("unable to make OpenAI API call: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned from OpenAI")
	}

	choice := resp.Choices[0]
	message := choice.Message

	var finalResponse string

	if len(message.ToolCalls) > 0 {
		finalResponse, err = a.handleToolCalls(message, prompt)
		if err != nil {
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

	return &structuredResp, nil
}

func (a *Action) handleToolCalls(message openai.ChatCompletionMessage, originalPrompt string) (string, error) {
	var toolMessages []openai.ChatCompletionMessage
	toolMessages = append(toolMessages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: originalPrompt,
	})
	toolMessages = append(toolMessages, message)

	fmt.Println("Tool calls:", message.ToolCalls)

	for _, toolCall := range message.ToolCalls {
		fmt.Printf("Executing tool: %s\n", toolCall.Function.Name)

		result, err := a.tools.ExecuteTool(toolCall)
		if err != nil {
			result = fmt.Sprintf("Error executing tool: %s", err.Error())
		}

		toolMessages = append(toolMessages, openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			ToolCallID: toolCall.ID,
			Content:    result,
		})
	}

	settings, err := a.repo.GetSettings()
	if err != nil {
		fmt.Printf("Error getting settings, using default transcription: %v\n", err)
	}

	if !settings.OpenaiApiKey.Valid || settings.OpenaiApiKey.String == "" {
		return "", fmt.Errorf("OpenAI API key not configured")
	}

	openAiClient := openai.NewClient(settings.OpenaiApiKey.String)

	// Ask OpenAI to provide a final structured response based on the tool results
	resp, err := openAiClient.CreateChatCompletion(a.ctx, openai.ChatCompletionRequest{
		Model:       tools.TranscriptionModel,
		Messages:    toolMessages,
		Temperature: 0.1,
	})

	if err != nil {
		return "", fmt.Errorf("unable to make follow-up OpenAI API call: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned from follow-up call")
	}

	return resp.Choices[0].Message.Content, nil
}
