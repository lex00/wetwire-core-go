package providers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserMessage(t *testing.T) {
	msg := NewUserMessage("hello world")

	assert.Equal(t, "user", msg.Role)
	require.Len(t, msg.Content, 1)
	assert.Equal(t, "text", msg.Content[0].Type)
	assert.Equal(t, "hello world", msg.Content[0].Text)
}

func TestNewAssistantMessage(t *testing.T) {
	blocks := []ContentBlock{
		{Type: "text", Text: "response"},
		{Type: "tool_use", ID: "123", Name: "test_tool"},
	}
	msg := NewAssistantMessage(blocks)

	assert.Equal(t, "assistant", msg.Role)
	require.Len(t, msg.Content, 2)
	assert.Equal(t, "text", msg.Content[0].Type)
	assert.Equal(t, "tool_use", msg.Content[1].Type)
}

func TestNewToolResult(t *testing.T) {
	result := NewToolResult("tool-123", "success output", false)

	assert.Equal(t, "tool_result", result.Type)
	assert.Equal(t, "tool-123", result.ToolUseID)
	assert.Equal(t, "success output", result.Content)
	assert.False(t, result.IsError)
}

func TestNewToolResultWithError(t *testing.T) {
	result := NewToolResult("tool-456", "error: failed", true)

	assert.Equal(t, "tool_result", result.Type)
	assert.Equal(t, "tool-456", result.ToolUseID)
	assert.Equal(t, "error: failed", result.Content)
	assert.True(t, result.IsError)
}

func TestNewToolResultMessage(t *testing.T) {
	results := []ContentBlock{
		NewToolResult("tool-1", "result1", false),
		NewToolResult("tool-2", "result2", false),
	}
	msg := NewToolResultMessage(results)

	assert.Equal(t, "user", msg.Role)
	require.Len(t, msg.Content, 2)
	assert.Equal(t, "tool_result", msg.Content[0].Type)
	assert.Equal(t, "tool_result", msg.Content[1].Type)
}

func TestStopReasonConstants(t *testing.T) {
	assert.Equal(t, StopReason("end_turn"), StopReasonEndTurn)
	assert.Equal(t, StopReason("tool_use"), StopReasonToolUse)
	assert.Equal(t, StopReason("max_tokens"), StopReasonMaxTokens)
	assert.Equal(t, StopReason("stop_sequence"), StopReasonStopSequence)
}

func TestContentBlockJSON(t *testing.T) {
	input := json.RawMessage(`{"path": "/tmp/test.go"}`)
	block := ContentBlock{
		Type:  "tool_use",
		ID:    "tool-123",
		Name:  "write_file",
		Input: input,
	}

	// Verify Input can be unmarshaled
	var params map[string]string
	err := json.Unmarshal(block.Input, &params)
	require.NoError(t, err)
	assert.Equal(t, "/tmp/test.go", params["path"])
}

func TestToolDefinition(t *testing.T) {
	tool := Tool{
		Name:        "write_file",
		Description: "Write content to a file",
		InputSchema: ToolInputSchema{
			Properties: map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "File path",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "File content",
				},
			},
			Required: []string{"path", "content"},
		},
	}

	assert.Equal(t, "write_file", tool.Name)
	assert.Contains(t, tool.InputSchema.Required, "path")
	assert.Contains(t, tool.InputSchema.Required, "content")
	assert.NotNil(t, tool.InputSchema.Properties["path"])
}

// MockProvider is a test implementation of the Provider interface.
type MockProvider struct {
	name               string
	createMessageFunc  func(ctx context.Context, req MessageRequest) (*MessageResponse, error)
	streamMessageFunc  func(ctx context.Context, req MessageRequest, handler StreamHandler) (*MessageResponse, error)
	createMessageCalls []MessageRequest
	streamMessageCalls []MessageRequest
}

func NewMockProvider(name string) *MockProvider {
	return &MockProvider{name: name}
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) CreateMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error) {
	m.createMessageCalls = append(m.createMessageCalls, req)
	if m.createMessageFunc != nil {
		return m.createMessageFunc(ctx, req)
	}
	return &MessageResponse{
		Content: []ContentBlock{
			{Type: "text", Text: "mock response"},
		},
		StopReason: StopReasonEndTurn,
	}, nil
}

func (m *MockProvider) StreamMessage(ctx context.Context, req MessageRequest, handler StreamHandler) (*MessageResponse, error) {
	m.streamMessageCalls = append(m.streamMessageCalls, req)
	if m.streamMessageFunc != nil {
		return m.streamMessageFunc(ctx, req, handler)
	}
	// Simulate streaming by calling handler with chunks
	chunks := []string{"mock ", "streaming ", "response"}
	for _, chunk := range chunks {
		handler(chunk)
	}
	return &MessageResponse{
		Content: []ContentBlock{
			{Type: "text", Text: "mock streaming response"},
		},
		StopReason: StopReasonEndTurn,
	}, nil
}

func TestMockProviderImplementsInterface(t *testing.T) {
	var _ Provider = (*MockProvider)(nil)
}

func TestMockProviderCreateMessage(t *testing.T) {
	mock := NewMockProvider("test")

	req := MessageRequest{
		Model:     "test-model",
		MaxTokens: 1024,
		System:    "You are a helpful assistant",
		Messages:  []Message{NewUserMessage("hello")},
	}

	resp, err := mock.CreateMessage(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, StopReasonEndTurn, resp.StopReason)
	require.Len(t, resp.Content, 1)
	assert.Equal(t, "mock response", resp.Content[0].Text)
	require.Len(t, mock.createMessageCalls, 1)
	assert.Equal(t, "test-model", mock.createMessageCalls[0].Model)
}

func TestMockProviderStreamMessage(t *testing.T) {
	mock := NewMockProvider("test")

	req := MessageRequest{
		Model:     "test-model",
		MaxTokens: 1024,
		Messages:  []Message{NewUserMessage("hello")},
	}

	var streamedText string
	handler := func(text string) {
		streamedText += text
	}

	resp, err := mock.StreamMessage(context.Background(), req, handler)

	require.NoError(t, err)
	assert.Equal(t, "mock streaming response", streamedText)
	assert.Equal(t, StopReasonEndTurn, resp.StopReason)
}

func TestMessageRequestWithTools(t *testing.T) {
	tools := []Tool{
		{
			Name:        "read_file",
			Description: "Read a file",
			InputSchema: ToolInputSchema{
				Properties: map[string]any{
					"path": map[string]any{"type": "string"},
				},
				Required: []string{"path"},
			},
		},
	}

	req := MessageRequest{
		Model:     "test-model",
		MaxTokens: 4096,
		System:    "System prompt",
		Messages:  []Message{NewUserMessage("read test.go")},
		Tools:     tools,
	}

	assert.Equal(t, "test-model", req.Model)
	assert.Equal(t, 4096, req.MaxTokens)
	require.Len(t, req.Tools, 1)
	assert.Equal(t, "read_file", req.Tools[0].Name)
}
