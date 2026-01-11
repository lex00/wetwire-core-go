// Package providers defines the interface for AI backend providers.
//
// This package provides a provider abstraction layer that allows switching
// between different AI backends (Anthropic, Kiro, etc.) without changing
// the agent code.
package providers

import (
	"context"
	"encoding/json"
)

// Provider is the interface that all AI backend providers must implement.
type Provider interface {
	// CreateMessage sends a message request and returns the complete response.
	CreateMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error)

	// StreamMessage sends a message request and streams the response via the handler.
	// The handler is called for each text chunk as it is generated.
	// Returns the complete response after streaming completes.
	StreamMessage(ctx context.Context, req MessageRequest, handler StreamHandler) (*MessageResponse, error)

	// Name returns the provider name (e.g., "anthropic", "kiro").
	Name() string
}

// StreamHandler is called for each text chunk during streaming.
type StreamHandler func(text string)

// MessageRequest contains the parameters for creating a message.
type MessageRequest struct {
	// Model identifier (e.g., "claude-sonnet-4-20250514")
	Model string

	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int

	// System is the system prompt
	System string

	// Messages is the conversation history
	Messages []Message

	// Tools available to the model
	Tools []Tool
}

// MessageResponse contains the response from the AI.
type MessageResponse struct {
	// Content contains the response blocks (text and tool use)
	Content []ContentBlock

	// StopReason indicates why the model stopped generating
	StopReason StopReason
}

// Message represents a conversation message.
type Message struct {
	// Role is "user" or "assistant"
	Role string

	// Content is the message content
	Content []ContentBlock
}

// ContentBlock represents a content block in a message.
type ContentBlock struct {
	// Type is "text" or "tool_use" or "tool_result"
	Type string

	// Text content (for Type="text")
	Text string

	// Tool use fields (for Type="tool_use")
	ID    string
	Name  string
	Input json.RawMessage

	// Tool result fields (for Type="tool_result")
	ToolUseID string
	Content   string
	IsError   bool
}

// Tool defines a tool that can be used by the model.
type Tool struct {
	// Name of the tool
	Name string

	// Description of what the tool does
	Description string

	// InputSchema defines the JSON schema for tool input
	InputSchema ToolInputSchema
}

// ToolInputSchema defines the JSON schema for tool parameters.
type ToolInputSchema struct {
	// Properties defines the input parameters
	Properties map[string]any

	// Required lists required parameter names
	Required []string
}

// StopReason indicates why the model stopped generating.
type StopReason string

const (
	// StopReasonEndTurn indicates the model completed its response.
	StopReasonEndTurn StopReason = "end_turn"

	// StopReasonToolUse indicates the model wants to use a tool.
	StopReasonToolUse StopReason = "tool_use"

	// StopReasonMaxTokens indicates the response was truncated due to max tokens.
	StopReasonMaxTokens StopReason = "max_tokens"

	// StopReasonStopSequence indicates a stop sequence was encountered.
	StopReasonStopSequence StopReason = "stop_sequence"
)

// NewUserMessage creates a new user message with text content.
func NewUserMessage(text string) Message {
	return Message{
		Role: "user",
		Content: []ContentBlock{
			{Type: "text", Text: text},
		},
	}
}

// NewAssistantMessage creates a new assistant message from content blocks.
func NewAssistantMessage(blocks []ContentBlock) Message {
	return Message{
		Role:    "assistant",
		Content: blocks,
	}
}

// NewToolResultMessage creates a new user message containing tool results.
func NewToolResultMessage(results []ContentBlock) Message {
	return Message{
		Role:    "user",
		Content: results,
	}
}

// NewToolResult creates a tool result content block.
func NewToolResult(toolUseID, content string, isError bool) ContentBlock {
	return ContentBlock{
		Type:      "tool_result",
		ToolUseID: toolUseID,
		Content:   content,
		IsError:   isError,
	}
}
