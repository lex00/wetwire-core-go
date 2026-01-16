package agents

import (
	"context"
	"testing"

	"github.com/lex00/wetwire-core-go/agent/results"
	"github.com/lex00/wetwire-core-go/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProviderForCapture implements providers.Provider for testing message capture
type mockProviderForCapture struct {
	responses []*providers.MessageResponse
	callCount int
}

func (m *mockProviderForCapture) Name() string {
	return "mock-capture"
}

func (m *mockProviderForCapture) CreateMessage(ctx context.Context, req providers.MessageRequest) (*providers.MessageResponse, error) {
	if m.callCount >= len(m.responses) {
		return &providers.MessageResponse{
			Content:    []providers.ContentBlock{{Type: "text", Text: "Done"}},
			StopReason: providers.StopReasonEndTurn,
		}, nil
	}
	resp := m.responses[m.callCount]
	m.callCount++
	return resp, nil
}

func (m *mockProviderForCapture) StreamMessage(ctx context.Context, req providers.MessageRequest, handler providers.StreamHandler) (*providers.MessageResponse, error) {
	return m.CreateMessage(ctx, req)
}

func TestRunnerAgent_CapturesMessagesToSession(t *testing.T) {
	// Create a session to capture messages
	session := results.NewSession("test", "test-scenario")

	// Mock provider that returns one text response then stops
	mockProvider := &mockProviderForCapture{
		responses: []*providers.MessageResponse{
			{
				Content: []providers.ContentBlock{
					{Type: "text", Text: "I will create the file"},
				},
				StopReason: providers.StopReasonEndTurn,
			},
		},
	}

	// Create RunnerAgent with session
	runner := &RunnerAgent{
		provider: mockProvider,
		model:    "test-model",
		domain: DomainConfig{
			Name:         "test",
			SystemPrompt: "Test system prompt",
		},
		session: session,
	}

	// Run the agent
	err := runner.Run(context.Background(), "Create a test file")
	require.NoError(t, err)

	// Verify that messages were captured to the session
	assert.NotEmpty(t, session.Messages, "Session should have captured messages")

	// Should have at least one runner message
	var hasRunnerMessage bool
	for _, msg := range session.Messages {
		if msg.Role == "runner" {
			hasRunnerMessage = true
			assert.Contains(t, msg.Content, "I will create the file")
		}
	}
	assert.True(t, hasRunnerMessage, "Session should contain at least one runner message")
}

func TestRunnerAgent_CapturesMultipleMessages(t *testing.T) {
	session := results.NewSession("test", "test-scenario")

	// Mock provider with multiple turns (all ending with EndTurn to avoid tool execution)
	mockProvider := &mockProviderForCapture{
		responses: []*providers.MessageResponse{
			{
				Content: []providers.ContentBlock{
					{Type: "text", Text: "First response"},
				},
				StopReason: providers.StopReasonEndTurn,
			},
		},
	}

	runner := &RunnerAgent{
		provider: mockProvider,
		model:    "test-model",
		domain: DomainConfig{
			Name:         "test",
			SystemPrompt: "Test system prompt",
		},
		session: session,
	}

	err := runner.Run(context.Background(), "Do something")
	require.NoError(t, err)

	// Should have at least one runner message
	runnerMessageCount := 0
	for _, msg := range session.Messages {
		if msg.Role == "runner" {
			runnerMessageCount++
		}
	}

	assert.GreaterOrEqual(t, runnerMessageCount, 1, "Should capture at least one runner response")
}

func TestRunnerAgent_CapturesTextFromMixedContent(t *testing.T) {
	session := results.NewSession("test", "test-scenario")

	// Mock provider that returns mixed content (text + other blocks)
	// We simulate the case where response has both text and non-text content
	mockProvider := &mockProviderForCapture{
		responses: []*providers.MessageResponse{
			{
				Content: []providers.ContentBlock{
					{Type: "text", Text: "Here is my response"},
					{Type: "text", Text: "Additional text"},
				},
				StopReason: providers.StopReasonEndTurn,
			},
		},
	}

	runner := &RunnerAgent{
		provider: mockProvider,
		model:    "test-model",
		domain: DomainConfig{
			Name:         "test",
			SystemPrompt: "Test system prompt",
		},
		session: session,
	}

	err := runner.Run(context.Background(), "Test")
	require.NoError(t, err)

	// Should capture concatenated text content
	var foundMessage bool
	for _, msg := range session.Messages {
		if msg.Role == "runner" && msg.Content == "Here is my response\nAdditional text" {
			foundMessage = true
		}
	}
	assert.True(t, foundMessage, "Should concatenate multiple text blocks")
}

func TestRunnerAgent_NilSession_NoError(t *testing.T) {
	// Mock provider
	mockProvider := &mockProviderForCapture{
		responses: []*providers.MessageResponse{
			{
				Content: []providers.ContentBlock{
					{Type: "text", Text: "Response"},
				},
				StopReason: providers.StopReasonEndTurn,
			},
		},
	}

	// Create RunnerAgent with nil session
	runner := &RunnerAgent{
		provider: mockProvider,
		model:    "test-model",
		domain: DomainConfig{
			Name:         "test",
			SystemPrompt: "Test system prompt",
		},
		session: nil, // No session
	}

	// Should not panic or error
	err := runner.Run(context.Background(), "Test")
	assert.NoError(t, err)
}

func TestExtractTextContent_OnlyText(t *testing.T) {
	content := []providers.ContentBlock{
		{Type: "text", Text: "Hello world"},
	}
	result := extractTextContent(content)
	assert.Equal(t, "Hello world", result)
}

func TestExtractTextContent_MultipleText(t *testing.T) {
	content := []providers.ContentBlock{
		{Type: "text", Text: "First"},
		{Type: "text", Text: "Second"},
	}
	result := extractTextContent(content)
	assert.Equal(t, "First\nSecond", result)
}

func TestExtractTextContent_MixedTypes(t *testing.T) {
	content := []providers.ContentBlock{
		{Type: "text", Text: "Text content"},
		{Type: "tool_use", Name: "some_tool"},
		{Type: "text", Text: "More text"},
	}
	result := extractTextContent(content)
	assert.Equal(t, "Text content\nMore text", result)
}

func TestExtractTextContent_NoText(t *testing.T) {
	content := []providers.ContentBlock{
		{Type: "tool_use", Name: "some_tool"},
	}
	result := extractTextContent(content)
	assert.Equal(t, "", result)
}

func TestExtractTextContent_EmptyText(t *testing.T) {
	content := []providers.ContentBlock{
		{Type: "text", Text: ""},
		{Type: "text", Text: "Not empty"},
	}
	result := extractTextContent(content)
	assert.Equal(t, "Not empty", result)
}
