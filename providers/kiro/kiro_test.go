package kiro

import (
	"context"
	"os"
	"testing"

	"github.com/lex00/wetwire-core-go/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderImplementsInterface(t *testing.T) {
	var _ providers.Provider = (*Provider)(nil)
}

func TestProviderName(t *testing.T) {
	p := &Provider{}
	assert.Equal(t, "kiro", p.Name())
}

func TestNewWithConfig(t *testing.T) {
	p, err := New(Config{
		AgentName:   "test-agent",
		AgentPrompt: "Test prompt",
		MCPCommand:  "test-mcp",
		WorkDir:     "/tmp/test",
	})
	require.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, "test-agent", p.config.AgentName)
}

func TestNewWithDefaults(t *testing.T) {
	p, err := New(Config{})
	require.NoError(t, err)
	assert.NotNil(t, p)
}

func TestCreateMessageKiroUnavailable(t *testing.T) {
	// Save original PATH and set to empty to ensure kiro is not found
	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", "/nonexistent")
	defer t.Setenv("PATH", originalPath)

	p := &Provider{
		config: Config{
			AgentName: "test-agent",
		},
	}

	req := providers.MessageRequest{
		Messages: []providers.Message{providers.NewUserMessage("hello")},
	}

	_, err := p.CreateMessage(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kiro")
}

func TestStreamMessageKiroUnavailable(t *testing.T) {
	// Save original PATH and set to empty to ensure kiro is not found
	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", "/nonexistent")
	defer t.Setenv("PATH", originalPath)

	p := &Provider{
		config: Config{
			AgentName: "test-agent",
		},
	}

	req := providers.MessageRequest{
		Messages: []providers.Message{providers.NewUserMessage("hello")},
	}

	handler := func(text string) {}

	_, err := p.StreamMessage(context.Background(), req, handler)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kiro")
}

func TestBuildPromptFromMessages(t *testing.T) {
	p := &Provider{}

	tests := []struct {
		name     string
		messages []providers.Message
		expected string
	}{
		{
			name: "single user message",
			messages: []providers.Message{
				providers.NewUserMessage("Create an S3 bucket"),
			},
			expected: "Create an S3 bucket",
		},
		{
			name: "multiple messages",
			messages: []providers.Message{
				providers.NewUserMessage("Create an S3 bucket"),
				providers.NewAssistantMessage([]providers.ContentBlock{
					{Type: "text", Text: "I'll create the bucket"},
				}),
				providers.NewUserMessage("with versioning enabled"),
			},
			expected: "Create an S3 bucket\n\nwith versioning enabled",
		},
		{
			name:     "empty messages",
			messages: []providers.Message{},
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := providers.MessageRequest{Messages: tc.messages}
			result := p.buildPrompt(req)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseKiroOutput(t *testing.T) {
	p := &Provider{}

	tests := []struct {
		name       string
		output     string
		expectText bool
		expectErr  bool
	}{
		{
			name:       "simple text output",
			output:     "Generated successfully: test.go",
			expectText: true,
			expectErr:  false,
		},
		{
			name:       "empty output",
			output:     "",
			expectText: false,
			expectErr:  false,
		},
		{
			name:       "multiline output",
			output:     "Line 1\nLine 2\nLine 3",
			expectText: true,
			expectErr:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := p.parseOutput(tc.output, 0)
			if tc.expectText {
				require.NotEmpty(t, result.Content)
				assert.Equal(t, "text", result.Content[0].Type)
				assert.Equal(t, tc.output, result.Content[0].Text)
			} else {
				assert.Empty(t, result.Content)
			}
			assert.Equal(t, providers.StopReasonEndTurn, result.StopReason)
		})
	}
}

func TestParseKiroOutputWithExitCode(t *testing.T) {
	p := &Provider{}

	// Non-zero exit code should still parse output but may indicate error
	result := p.parseOutput("Error occurred", 1)
	require.NotEmpty(t, result.Content)
	assert.Equal(t, "Error occurred", result.Content[0].Text)
	assert.Equal(t, providers.StopReasonEndTurn, result.StopReason)
}
