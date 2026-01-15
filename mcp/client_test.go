package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClientRequiresCommand(t *testing.T) {
	ctx := context.Background()

	_, err := NewClient(ctx, ClientConfig{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "command is required")
}

func TestNewClientWithInvalidCommand(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := NewClient(ctx, ClientConfig{
		Command: "/nonexistent/command",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start")
}

// TestClientWithMockServer tests the client using a real MCP server.
// This requires a mock server script or binary.
func TestClientToolsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// This test would require a mock MCP server
	// For now, we test the basic structures work
	t.Skip("requires mock MCP server")
}

func TestToolInfoStructure(t *testing.T) {
	info := ToolInfo{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"arg1": map[string]any{"type": "string"},
			},
		},
	}

	assert.Equal(t, "test_tool", info.Name)
	assert.Equal(t, "A test tool", info.Description)
	assert.NotNil(t, info.InputSchema)
}

func TestToolCallParamsStructure(t *testing.T) {
	params := ToolCallParams{
		Name: "test_tool",
		Arguments: map[string]any{
			"arg1": "value1",
			"arg2": 42,
		},
	}

	assert.Equal(t, "test_tool", params.Name)
	assert.Equal(t, "value1", params.Arguments["arg1"])
	assert.Equal(t, 42, params.Arguments["arg2"])
}

func TestToolCallResultStructure(t *testing.T) {
	result := ToolCallResult{
		Content: []ContentBlock{
			{Type: "text", Text: "Success"},
		},
		IsError: false,
	}

	assert.False(t, result.IsError)
	require.Len(t, result.Content, 1)
	assert.Equal(t, "text", result.Content[0].Type)
	assert.Equal(t, "Success", result.Content[0].Text)
}

func TestToolCallResultWithError(t *testing.T) {
	result := ToolCallResult{
		Content: []ContentBlock{
			{Type: "text", Text: "Error: something went wrong"},
		},
		IsError: true,
	}

	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].Text, "Error")
}
