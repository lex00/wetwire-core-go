package mcp

import (
	"context"
	"testing"
)

// TestServer_ExecuteTool tests the in-process tool execution method.
func TestServer_ExecuteTool(t *testing.T) {
	server := NewServer(Config{Name: "test"})

	// Register a test tool
	called := false
	server.RegisterToolWithSchema(
		"test_tool",
		"A test tool",
		func(ctx context.Context, args map[string]any) (string, error) {
			called = true
			if val, ok := args["param"].(string); ok && val == "expected" {
				return "success", nil
			}
			return "unexpected param", nil
		},
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"param": map[string]any{"type": "string"},
			},
		},
	)

	// Test successful execution
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "test_tool", map[string]any{
		"param": "expected",
	})

	if err != nil {
		t.Fatalf("ExecuteTool failed: %v", err)
	}

	if result != "success" {
		t.Errorf("expected 'success', got '%s'", result)
	}

	if !called {
		t.Error("tool handler was not called")
	}

	// Test non-existent tool
	_, err = server.ExecuteTool(ctx, "nonexistent", nil)
	if err == nil {
		t.Error("expected error for non-existent tool")
	}
}

// TestServer_GetTools tests the tool listing method.
func TestServer_GetTools(t *testing.T) {
	server := NewServer(Config{Name: "test"})

	// Register multiple tools
	server.RegisterToolWithSchema(
		"tool1",
		"First tool",
		func(ctx context.Context, args map[string]any) (string, error) {
			return "result1", nil
		},
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"param1": map[string]any{"type": "string"},
			},
		},
	)

	server.RegisterTool(
		"tool2",
		"Second tool",
		func(ctx context.Context, args map[string]any) (string, error) {
			return "result2", nil
		},
	)

	// Get tools
	tools := server.GetTools()

	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}

	// Verify tool1
	var tool1, tool2 *ToolInfo
	for i := range tools {
		if tools[i].Name == "tool1" {
			tool1 = &tools[i]
		} else if tools[i].Name == "tool2" {
			tool2 = &tools[i]
		}
	}

	if tool1 == nil {
		t.Error("tool1 not found")
	} else {
		if tool1.Description != "First tool" {
			t.Errorf("tool1 description mismatch: %s", tool1.Description)
		}
		if props, ok := tool1.InputSchema["properties"].(map[string]any); !ok || len(props) != 1 {
			t.Error("tool1 schema not preserved")
		}
	}

	if tool2 == nil {
		t.Error("tool2 not found")
	} else {
		if tool2.Description != "Second tool" {
			t.Errorf("tool2 description mismatch: %s", tool2.Description)
		}
		// tool2 should have default empty schema
		if tool2.InputSchema["type"] != "object" {
			t.Error("tool2 should have default object schema")
		}
	}
}
