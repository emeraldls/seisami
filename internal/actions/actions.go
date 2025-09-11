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

// StructuredResponse represents the AI's structured response
type StructuredResponse struct {
	Intent       string                 `json:"intent"`
	Understood   string                 `json:"understood"`
	ActionsTaken []string               `json:"actions_taken"`
	Result       string                 `json:"result"`
	Data         map[string]interface{} `json:"data,omitempty"`
	Suggestions  []string               `json:"suggestions,omitempty"`
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
	prompt := fmt.Sprintf(`You are Seisami AI — a minimalist, high-precision desktop-first productivity assistant. You help users manage their boards and tasks through voice commands.

CONTEXT:
- Current time: %s (RFC3339)
- Current Board ID: %s
- User said: "%s"

INSTRUCTIONS:
1. Use available tools when you need to fetch or modify data (read_board, list_boards, etc.)
2. After using tools (or if no tools needed), provide a structured response that includes:
   - What you understood from the user's request
   - What actions were taken (if any)
   - The results or next steps

RESPONSE FORMAT:
Always respond with a JSON object containing:
{
  "intent": "string - what the user wanted (e.g., 'read_board', 'create_task', 'get_summary')",
  "understood": "string - natural language summary of what you understood",
  "actions_taken": ["array of actions performed"],
  "result": "string - the main result or answer",
  "data": {} // optional - any structured data returned from tools,
  "suggestions": ["array of suggested next actions"] // optional
}

IMPORTANT: Use tools when needed to get current data, then format the response with that data included.`, now, boardId, transcription)

	return prompt
}

func (a *Action) ProcessTranscription(transcription string, boardId string) (*StructuredResponse, error) {
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

	// Ask OpenAI to provide a final structured response based on the tool results
	resp, err := a.openAiClient.CreateChatCompletion(a.ctx, openai.ChatCompletionRequest{
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
