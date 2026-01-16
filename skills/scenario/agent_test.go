package scenario

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lex00/wetwire-core-go/mcp"
	"github.com/lex00/wetwire-core-go/providers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider implements providers.Provider for testing
type mockProvider struct {
	responses []*providers.MessageResponse
	callCount int
}

func (m *mockProvider) CreateMessage(ctx context.Context, req providers.MessageRequest) (*providers.MessageResponse, error) {
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

func (m *mockProvider) StreamMessage(ctx context.Context, req providers.MessageRequest, handler providers.StreamHandler) (*providers.MessageResponse, error) {
	return m.CreateMessage(ctx, req)
}

func (m *mockProvider) Name() string {
	return "mock"
}

func TestNewScenarioAgent(t *testing.T) {
	provider := &mockProvider{}
	server := mcp.NewServer(mcp.Config{Name: "test"})

	agent := NewScenarioAgent(ScenarioAgentConfig{
		Provider:  provider,
		MCPServer: server,
	})

	assert.NotNil(t, agent)
	assert.Equal(t, "claude-sonnet-4-20250514", agent.model)
}

func TestNewScenarioAgent_CustomModel(t *testing.T) {
	provider := &mockProvider{}
	server := mcp.NewServer(mcp.Config{Name: "test"})

	agent := NewScenarioAgent(ScenarioAgentConfig{
		Provider:  provider,
		MCPServer: server,
		Model:     "custom-model",
	})

	assert.NotNil(t, agent)
	assert.Equal(t, "custom-model", agent.model)
}

func TestScenarioAgent_Run_NoTools(t *testing.T) {
	provider := &mockProvider{
		responses: []*providers.MessageResponse{
			{
				Content: []providers.ContentBlock{
					{Type: "text", Text: "Task completed"},
				},
				StopReason: providers.StopReasonEndTurn,
			},
		},
	}
	server := mcp.NewServer(mcp.Config{Name: "test"})

	agent := NewScenarioAgent(ScenarioAgentConfig{
		Provider:  provider,
		MCPServer: server,
	})

	err := agent.Run(context.Background(), "test prompt")
	require.NoError(t, err)
	assert.Equal(t, 1, provider.callCount)
}

func TestScenarioAgent_Run_WithToolCalls(t *testing.T) {
	provider := &mockProvider{
		responses: []*providers.MessageResponse{
			// First response: agent requests to use a tool
			{
				Content: []providers.ContentBlock{
					{Type: "text", Text: "I'll initialize the package"},
					{
						Type:  "tool_use",
						ID:    "tool-1",
						Name:  "test_tool",
						Input: json.RawMessage(`{"param": "value"}`),
					},
				},
				StopReason: providers.StopReasonToolUse,
			},
			// Second response: agent finishes after getting tool result
			{
				Content: []providers.ContentBlock{
					{Type: "text", Text: "Done"},
				},
				StopReason: providers.StopReasonEndTurn,
			},
		},
	}

	server := mcp.NewServer(mcp.Config{Name: "test"})
	server.RegisterTool("test_tool", "Test tool", func(ctx context.Context, args map[string]any) (string, error) {
		return "tool result", nil
	})

	agent := NewScenarioAgent(ScenarioAgentConfig{
		Provider:  provider,
		MCPServer: server,
	})

	err := agent.Run(context.Background(), "test prompt")
	require.NoError(t, err)
	assert.Equal(t, 2, provider.callCount)
}

func TestScenarioAgent_GetMCPTools(t *testing.T) {
	provider := &mockProvider{}
	server := mcp.NewServer(mcp.Config{Name: "test"})

	// Register some test tools
	server.RegisterToolWithSchema("tool1", "First tool", func(ctx context.Context, args map[string]any) (string, error) {
		return "result1", nil
	}, map[string]any{
		"type": "object",
		"properties": map[string]any{
			"param1": map[string]any{
				"type":        "string",
				"description": "Test parameter",
			},
		},
		"required": []any{"param1"},
	})

	server.RegisterTool("tool2", "Second tool", func(ctx context.Context, args map[string]any) (string, error) {
		return "result2", nil
	})

	agent := NewScenarioAgent(ScenarioAgentConfig{
		Provider:  provider,
		MCPServer: server,
	})

	tools := agent.getMCPTools()

	assert.Len(t, tools, 2)

	// Find tools by name (order is not deterministic)
	toolMap := make(map[string]providers.Tool)
	for _, tool := range tools {
		toolMap[tool.Name] = tool
	}

	tool1, ok := toolMap["tool1"]
	assert.True(t, ok, "tool1 should exist")
	assert.Equal(t, "First tool", tool1.Description)
	assert.NotNil(t, tool1.InputSchema.Properties)
	assert.Contains(t, tool1.InputSchema.Required, "param1")

	tool2, ok := toolMap["tool2"]
	assert.True(t, ok, "tool2 should exist")
	assert.Equal(t, "Second tool", tool2.Description)
}

func TestScenarioAgent_ExecuteMCPTool(t *testing.T) {
	provider := &mockProvider{}
	server := mcp.NewServer(mcp.Config{Name: "test"})

	called := false
	server.RegisterTool("test_tool", "Test tool", func(ctx context.Context, args map[string]any) (string, error) {
		called = true
		assert.Equal(t, "test_value", args["test_param"])
		return "success", nil
	})

	agent := NewScenarioAgent(ScenarioAgentConfig{
		Provider:  provider,
		MCPServer: server,
	})

	result, err := agent.executeMCPTool(
		context.Background(),
		"test_tool",
		json.RawMessage(`{"test_param": "test_value"}`),
	)

	require.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.True(t, called)
}

func TestScenarioAgent_ContextCancellation(t *testing.T) {
	provider := &mockProvider{
		responses: []*providers.MessageResponse{},
	}
	server := mcp.NewServer(mcp.Config{Name: "test"})

	agent := NewScenarioAgent(ScenarioAgentConfig{
		Provider:  provider,
		MCPServer: server,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := agent.Run(ctx, "test prompt")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}
