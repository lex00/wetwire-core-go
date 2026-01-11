package anthropic

import (
	"encoding/json"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/lex00/wetwire-core-go/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderImplementsInterface(t *testing.T) {
	var _ providers.Provider = (*Provider)(nil)
}

func TestProviderName(t *testing.T) {
	// We can't create a real provider without an API key, but we can test the method
	p := &Provider{}
	assert.Equal(t, "anthropic", p.Name())
}

func TestNewRequiresAPIKey(t *testing.T) {
	// Clear env var for test
	t.Setenv("ANTHROPIC_API_KEY", "")

	_, err := New(Config{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ANTHROPIC_API_KEY")
}

func TestNewWithConfigAPIKey(t *testing.T) {
	p, err := New(Config{APIKey: "test-key"})
	require.NoError(t, err)
	assert.NotNil(t, p)
}

func TestNewWithEnvAPIKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "test-env-key")

	p, err := New(Config{})
	require.NoError(t, err)
	assert.NotNil(t, p)
}

func TestBuildParamsDefaults(t *testing.T) {
	p := &Provider{}

	req := providers.MessageRequest{
		Messages: []providers.Message{providers.NewUserMessage("hello")},
	}

	params := p.buildParams(req)

	assert.Equal(t, anthropic.Model(DefaultModel), params.Model)
	assert.Equal(t, int64(4096), params.MaxTokens)
}

func TestBuildParamsWithValues(t *testing.T) {
	p := &Provider{}

	req := providers.MessageRequest{
		Model:     "claude-opus-4-20250514",
		MaxTokens: 8192,
		System:    "You are helpful",
		Messages:  []providers.Message{providers.NewUserMessage("hello")},
	}

	params := p.buildParams(req)

	assert.Equal(t, anthropic.Model("claude-opus-4-20250514"), params.Model)
	assert.Equal(t, int64(8192), params.MaxTokens)
	require.Len(t, params.System, 1)
	assert.Equal(t, "You are helpful", params.System[0].Text)
}

func TestConvertMessagesTextOnly(t *testing.T) {
	p := &Provider{}

	msgs := []providers.Message{
		providers.NewUserMessage("hello"),
		providers.NewAssistantMessage([]providers.ContentBlock{
			{Type: "text", Text: "hi there"},
		}),
		providers.NewUserMessage("how are you"),
	}

	result := p.convertMessages(msgs)

	require.Len(t, result, 3)
}

func TestConvertMessagesWithToolUse(t *testing.T) {
	p := &Provider{}

	msgs := []providers.Message{
		providers.NewUserMessage("write a file"),
		{
			Role: "assistant",
			Content: []providers.ContentBlock{
				{
					Type:  "tool_use",
					ID:    "tool-123",
					Name:  "write_file",
					Input: json.RawMessage(`{"path":"test.go"}`),
				},
			},
		},
		{
			Role: "user",
			Content: []providers.ContentBlock{
				providers.NewToolResult("tool-123", "file written", false),
			},
		},
	}

	result := p.convertMessages(msgs)

	require.Len(t, result, 3)
}

func TestConvertTools(t *testing.T) {
	p := &Provider{}

	tools := []providers.Tool{
		{
			Name:        "read_file",
			Description: "Read a file's contents",
			InputSchema: providers.ToolInputSchema{
				Properties: map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "File path",
					},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "write_file",
			Description: "Write content to a file",
			InputSchema: providers.ToolInputSchema{
				Properties: map[string]any{
					"path":    map[string]any{"type": "string"},
					"content": map[string]any{"type": "string"},
				},
				Required: []string{"path", "content"},
			},
		},
	}

	result := p.convertTools(tools)

	require.Len(t, result, 2)
	assert.Equal(t, "read_file", result[0].OfTool.Name)
	assert.Equal(t, "write_file", result[1].OfTool.Name)
}

func TestConvertResponse(t *testing.T) {
	p := &Provider{}

	resp := &anthropic.Message{
		Content: []anthropic.ContentBlockUnion{
			{Type: "text", Text: "Hello, I can help"},
			{Type: "tool_use", ID: "tool-1", Name: "read_file", Input: json.RawMessage(`{"path":"test.go"}`)},
		},
		StopReason: anthropic.StopReasonToolUse,
	}

	result := p.convertResponse(resp)

	assert.Equal(t, providers.StopReasonToolUse, result.StopReason)
	require.Len(t, result.Content, 2)

	assert.Equal(t, "text", result.Content[0].Type)
	assert.Equal(t, "Hello, I can help", result.Content[0].Text)

	assert.Equal(t, "tool_use", result.Content[1].Type)
	assert.Equal(t, "tool-1", result.Content[1].ID)
	assert.Equal(t, "read_file", result.Content[1].Name)
}

func TestConvertResponseNil(t *testing.T) {
	p := &Provider{}

	result := p.convertResponse(nil)

	assert.NotNil(t, result)
	assert.Empty(t, result.Content)
}

func TestConvertStopReason(t *testing.T) {
	tests := []struct {
		input    anthropic.StopReason
		expected providers.StopReason
	}{
		{anthropic.StopReasonEndTurn, providers.StopReasonEndTurn},
		{anthropic.StopReasonToolUse, providers.StopReasonToolUse},
		{anthropic.StopReasonMaxTokens, providers.StopReasonMaxTokens},
		{anthropic.StopReasonStopSequence, providers.StopReasonStopSequence},
	}

	for _, tc := range tests {
		t.Run(string(tc.input), func(t *testing.T) {
			result := convertStopReason(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDefaultModel(t *testing.T) {
	assert.Equal(t, "claude-sonnet-4-20250514", DefaultModel)
}
