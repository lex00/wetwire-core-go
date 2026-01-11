package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	config := Config{
		Name:    "test-server",
		Version: "1.0.0",
	}

	server := NewServer(config)

	if server.Name() != "test-server" {
		t.Errorf("expected name 'test-server', got '%s'", server.Name())
	}
}

func TestRegisterTool(t *testing.T) {
	server := NewServer(Config{Name: "test"})

	handler := func(ctx context.Context, args map[string]any) (string, error) {
		return "result", nil
	}

	server.RegisterTool("my-tool", "A test tool", handler)

	if len(server.tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(server.tools))
	}

	tool, exists := server.tools["my-tool"]
	if !exists {
		t.Fatal("tool 'my-tool' not found")
	}

	if tool.Description != "A test tool" {
		t.Errorf("expected description 'A test tool', got '%s'", tool.Description)
	}
}

func TestRegisterToolWithSchema(t *testing.T) {
	server := NewServer(Config{Name: "test"})

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "The file path",
			},
		},
		"required": []string{"path"},
	}

	server.RegisterToolWithSchema("build", "Build the project", nil, schema)

	tool := server.tools["build"]
	if tool.InputSchema == nil {
		t.Fatal("expected input schema to be set")
	}

	if tool.InputSchema["type"] != "object" {
		t.Errorf("expected schema type 'object', got '%v'", tool.InputSchema["type"])
	}
}

func TestHandleInitialize(t *testing.T) {
	server := NewServer(Config{
		Name:    "test-server",
		Version: "1.0.0",
	})

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "initialize",
		ID:      1,
	}

	resp, err := server.handleInitialize(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("unexpected error in response: %v", resp.Error)
	}

	result, ok := resp.Result.(InitializeResult)
	if !ok {
		t.Fatal("expected InitializeResult")
	}

	if result.ServerInfo.Name != "test-server" {
		t.Errorf("expected server name 'test-server', got '%s'", result.ServerInfo.Name)
	}

	if result.ServerInfo.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", result.ServerInfo.Version)
	}

	if result.Capabilities.Tools == nil {
		t.Error("expected tools capability to be set")
	}
}

func TestHandleToolsList(t *testing.T) {
	server := NewServer(Config{Name: "test"})

	server.RegisterTool("tool1", "First tool", nil)
	server.RegisterTool("tool2", "Second tool", nil)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/list",
		ID:      1,
	}

	resp, err := server.handleToolsList(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, ok := resp.Result.(ToolsListResult)
	if !ok {
		t.Fatal("expected ToolsListResult")
	}

	if len(result.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(result.Tools))
	}
}

func TestHandleToolsCall(t *testing.T) {
	server := NewServer(Config{Name: "test"})

	server.RegisterTool("echo", "Echo the input", func(ctx context.Context, args map[string]any) (string, error) {
		msg, _ := args["message"].(string)
		return "Echo: " + msg, nil
	})

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: map[string]any{
			"name": "echo",
			"arguments": map[string]any{
				"message": "hello",
			},
		},
		ID: 1,
	}

	resp, err := server.handleToolsCall(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, ok := resp.Result.(ToolCallResult)
	if !ok {
		t.Fatal("expected ToolCallResult")
	}

	if result.IsError {
		t.Error("expected IsError to be false")
	}

	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(result.Content))
	}

	if result.Content[0].Text != "Echo: hello" {
		t.Errorf("expected 'Echo: hello', got '%s'", result.Content[0].Text)
	}
}

func TestHandleToolsCallNotFound(t *testing.T) {
	server := NewServer(Config{Name: "test"})

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: map[string]any{
			"name": "nonexistent",
		},
		ID: 1,
	}

	resp, err := server.handleToolsCall(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("expected error response")
	}

	if resp.Error.Code != InvalidParams {
		t.Errorf("expected InvalidParams error code, got %d", resp.Error.Code)
	}
}

func TestHandleToolsCallError(t *testing.T) {
	server := NewServer(Config{Name: "test"})

	server.RegisterTool("failing", "A failing tool", func(ctx context.Context, args map[string]any) (string, error) {
		return "", context.DeadlineExceeded
	})

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: map[string]any{
			"name": "failing",
		},
		ID: 1,
	}

	resp, err := server.handleToolsCall(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, ok := resp.Result.(ToolCallResult)
	if !ok {
		t.Fatal("expected ToolCallResult")
	}

	if !result.IsError {
		t.Error("expected IsError to be true")
	}

	if !strings.Contains(result.Content[0].Text, "Error:") {
		t.Errorf("expected error message, got '%s'", result.Content[0].Text)
	}
}

func TestHandleUnknownMethod(t *testing.T) {
	server := NewServer(Config{Name: "test"})

	data := []byte(`{"jsonrpc":"2.0","method":"unknown/method","id":1}`)

	resp, err := server.handleMessage(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("expected error response")
	}

	if resp.Error.Code != MethodNotFound {
		t.Errorf("expected MethodNotFound error code, got %d", resp.Error.Code)
	}
}

func TestHandleParseError(t *testing.T) {
	server := NewServer(Config{Name: "test"})

	data := []byte(`{invalid json}`)

	resp, err := server.handleMessage(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("expected error response")
	}

	if resp.Error.Code != ParseError {
		t.Errorf("expected ParseError code, got %d", resp.Error.Code)
	}
}

func TestServerStartWithStdio(t *testing.T) {
	server := NewServer(Config{Name: "test"})
	server.RegisterTool("ping", "Ping tool", func(ctx context.Context, args map[string]any) (string, error) {
		return "pong", nil
	})

	// Create test input/output
	input := `{"jsonrpc":"2.0","method":"initialize","id":1}
{"jsonrpc":"2.0","method":"tools/list","id":2}
{"jsonrpc":"2.0","method":"tools/call","params":{"name":"ping"},"id":3}
`
	reader := strings.NewReader(input)
	var output bytes.Buffer

	server.reader = reader
	server.writer = &output

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := server.Start(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse responses
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 responses, got %d: %s", len(lines), output.String())
	}

	// Check initialize response
	var initResp JSONRPCResponse
	if err := json.Unmarshal([]byte(lines[0]), &initResp); err != nil {
		t.Fatalf("failed to parse initialize response: %v", err)
	}
	if initResp.Error != nil {
		t.Errorf("unexpected error in initialize: %v", initResp.Error)
	}

	// Check tools/list response
	var listResp JSONRPCResponse
	if err := json.Unmarshal([]byte(lines[1]), &listResp); err != nil {
		t.Fatalf("failed to parse tools/list response: %v", err)
	}
	if listResp.Error != nil {
		t.Errorf("unexpected error in tools/list: %v", listResp.Error)
	}

	// Check tools/call response
	var callResp JSONRPCResponse
	if err := json.Unmarshal([]byte(lines[2]), &callResp); err != nil {
		t.Fatalf("failed to parse tools/call response: %v", err)
	}
	if callResp.Error != nil {
		t.Errorf("unexpected error in tools/call: %v", callResp.Error)
	}
}

func TestGetInstallInstructions(t *testing.T) {
	instructions := GetInstallInstructions("wetwire-azure", "wetwire-azure-mcp")

	if !strings.Contains(instructions, "wetwire-azure") {
		t.Error("expected instructions to contain server name")
	}

	if !strings.Contains(instructions, "wetwire-azure-mcp") {
		t.Error("expected instructions to contain binary name")
	}

	if !strings.Contains(instructions, "mcpServers") {
		t.Error("expected instructions to contain mcpServers config")
	}
}

func TestNotificationNoResponse(t *testing.T) {
	server := NewServer(Config{Name: "test"})

	data := []byte(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)

	resp, err := server.handleMessage(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp != nil {
		t.Error("expected no response for notification")
	}
}
