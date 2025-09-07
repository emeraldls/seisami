package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"seisami/internal/repo"
	"seisami/internal/tools"
	"time"

	"github.com/sashabaranov/go-openai"
)

type Action struct {
	ctx          context.Context
	repo         repo.Repository
	tools        *tools.Tools
	openAiClient *openai.Client
}

func NewAction(ctx context.Context, repo repo.Repository) *Action {

	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		log.Fatal("OPENAI_API_KEY ENV NOT FOUND")
		return nil
	}

	openAiClient := openai.NewClient(openaiKey)
	tools := tools.NewTools(repo, *openAiClient, ctx)
	return &Action{
		ctx,
		repo,
		tools,
		openAiClient,
	}
}

func buildPromptTemplate(transcription string, boardId string) string {
	now := time.Now().Format(time.RFC3339)
	prompt := fmt.Sprintf(`You are Seisami AI — a minimalist, high-precision desktop-first productivity assistant. Your job is to convert a single user transcription (natural language) into a structured, extensible JSON payload of actions. Do this precisely and concisely: output ONLY valid JSON that matches the schema below. No explanation, no commentary, no markdown — just the raw JSON text.

IMPORTANT: DO NOT wrap the JSON in any Markdown or code fences. Do NOT include backticks before or after the JSON. The response must begin with "{" and end with "}". If you cannot produce valid JSON as specified, return a single action with "action_type": "error".

INJECTED VALUES
- Current time reference: %s  (RFC3339). Use this to resolve relative dates/times.
- Current Board ID: %s
- User transcription: %s

OUTPUT RULES
1. Always output a single JSON object with a top-level "actions" array.
2. Do NOT output text outside JSON. If you cannot comply, return a single action with "action_type": "error".
3. Do NOT invent facts or assume unstated details. If details required for execution are missing or ambiguous, return an "error" action describing the ambiguity and a suggested clarifying question.

CORE ACTION SCHEMA (each element in "actions"):
{
  "action_type": "add_task|update_task|delete_task|read|navigate|summarize|create_board|update_board|delete_board|list_boards|get_board|create_column|move_card|set_due|custom|error",
  "target": "task|event|calendar|summary|column|board|other",
  "title": "string (recommended for task/event-related actions)",
  "description": "string (optional: details from transcription)",
  "date": "YYYY-MM-DD (optional, ISO-8601 when present)",
  "time": "HH:MM (24-hour, optional)",
  "id": "string (optional: used for update/delete/move)",
  "priority": "low|medium|high (optional)",
  "tags": ["string", ...] (optional),
  "context": { "key": "value", ... } (optional, for board/column/project),
  "output": "string (optional, for summaries/read results)",
  "tool_call": {
    "tool_name": "task_manager|calendar_api|summarizer|board_manager|custom",
    "parameters": { ... mirror relevant action fields ... },
    "depends_on": [0,1]  // optional array of zero-based indices referencing previous actions for chaining
  }
}

KEY BEHAVIOR RULES
- Multi-step commands: break into sequential actions. If one action depends on another, use "tool_call.depends_on".
- Date/time resolution: use the injected current time to resolve relative expressions (e.g., "Saturday" => next calendar Saturday after current_time unless the user explicitly indicated a past date). Convert times like "5:30pm" to "17:30". Always return date in ISO format when resolved.
- Ambiguity: if the transcription lacks a required execution detail (e.g., missing target board when user says "move it"), produce an action with "action_type": "error" and "description" suggesting exactly what to ask next.
- No past context: do not reference or store prior interactions unless the transcription explicitly asks to "summarize my August 5" (then the model should output navigation/read/summarize actions, not assume the content).
- Extensibility: downstream systems may add custom keys; prefer adding them under "context" or allow "custom" action_type with descriptive parameters.
- Minimalism: outputs should be execution-ready minimal text, maximal structure.

ERROR ACTION FORMAT
{
  "action_type": "error",
  "target": "task|other (if applicable)",
  "description": "Explain missing/ambiguous fields and suggest the exact clarifying question to ask the user"
}

VALIDATION
- Ensure JSON is parseable and strictly conforms to the schema above.  
- If multiple plausible interpretations exist and you cannot decide safely, return an "error" action rather than guessing.

Remember: output ONLY the raw JSON object described above, starting with '{' and ending with '}'.`, now, boardId, transcription)

	return prompt
}

func (a *Action) ProcessTranscription(transcription string, boardId string) ([]any, error) {

	prompt := buildPromptTemplate(transcription, boardId)

	resp, err := a.openAiClient.CreateChatCompletion(a.ctx, openai.ChatCompletionRequest{
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

	// Handle the response
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned from OpenAI")
	}

	choice := resp.Choices[0]
	message := choice.Message

	// If there are tool calls, handle them
	if len(message.ToolCalls) > 0 {
		return a.handleToolCalls(message, prompt)
	}

	// If no tool calls, parse the JSON response directly
	var actions struct {
		Actions []any `json:"actions"`
	}

	if err := json.Unmarshal([]byte(message.Content), &actions); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return actions.Actions, nil
}

func (a *Action) handleToolCalls(message openai.ChatCompletionMessage, originalPrompt string) ([]any, error) {
	var toolMessages []openai.ChatCompletionMessage
	toolMessages = append(toolMessages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: originalPrompt,
	})
	toolMessages = append(toolMessages, message)

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

	resp, err := a.openAiClient.CreateChatCompletion(a.ctx, openai.ChatCompletionRequest{
		Model:       tools.TranscriptionModel,
		Messages:    toolMessages,
		Temperature: 0.1,
		Tools:       a.tools.AvailableTools(),
	})

	if err != nil {
		return nil, fmt.Errorf("unable to make follow-up OpenAI API call: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned from follow-up call")
	}

	var actions struct {
		Actions []any `json:"actions"`
	}

	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &actions); err != nil {
		return nil, fmt.Errorf("failed to parse final JSON response: %w", err)
	}

	return actions.Actions, nil
}
