package agents

import (
	"context"
	"testing"

	"github.com/lex00/wetwire-core-go/agent/results"
	"github.com/lex00/wetwire-core-go/mcp"
	"github.com/lex00/wetwire-core-go/providers"
)

// testAgentProvider is a test provider that returns predefined responses.
type testAgentProvider struct {
	responses []*providers.MessageResponse
	callCount int
}

func (m *testAgentProvider) CreateMessage(ctx context.Context, req providers.MessageRequest) (*providers.MessageResponse, error) {
	if m.callCount >= len(m.responses) {
		// Default to end turn
		return &providers.MessageResponse{
			Content:    []providers.ContentBlock{{Type: "text", Text: "Done"}},
			StopReason: providers.StopReasonEndTurn,
		}, nil
	}
	resp := m.responses[m.callCount]
	m.callCount++
	return resp, nil
}

func (m *testAgentProvider) StreamMessage(ctx context.Context, req providers.MessageRequest, handler providers.StreamHandler) (*providers.MessageResponse, error) {
	return m.CreateMessage(ctx, req)
}

func (m *testAgentProvider) Name() string {
	return "test-agent-mock"
}

// testAgentDeveloper is a test developer that returns predefined answers.
type testAgentDeveloper struct {
	answers []string
	calls   int
}

func (m *testAgentDeveloper) Respond(ctx context.Context, message string) (string, error) {
	if m.calls >= len(m.answers) {
		return "I don't know", nil
	}
	answer := m.answers[m.calls]
	m.calls++
	return answer, nil
}

func TestNewAgent(t *testing.T) {
	tests := []struct {
		name      string
		config    AgentConfig
		wantError bool
	}{
		{
			name: "valid config",
			config: AgentConfig{
				Provider:     &testAgentProvider{},
				MCPServer:    NewMCPServerAdapter(mcp.NewServer(mcp.Config{Name: "test"})),
				SystemPrompt: "test prompt",
			},
			wantError: false,
		},
		{
			name: "missing provider",
			config: AgentConfig{
				MCPServer:    NewMCPServerAdapter(mcp.NewServer(mcp.Config{Name: "test"})),
				SystemPrompt: "test prompt",
			},
			wantError: true,
		},
		{
			name: "missing mcp server",
			config: AgentConfig{
				Provider:     &testAgentProvider{},
				SystemPrompt: "test prompt",
			},
			wantError: true,
		},
		{
			name: "missing system prompt",
			config: AgentConfig{
				Provider:  &testAgentProvider{},
				MCPServer: NewMCPServerAdapter(mcp.NewServer(mcp.Config{Name: "test"})),
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := NewAgent(tt.config)
			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantError && agent == nil {
				t.Error("expected agent but got nil")
			}
		})
	}
}

func TestAgent_Run_Autonomous(t *testing.T) {
	// Create MCP server with a test tool
	server := mcp.NewServer(mcp.Config{Name: "test"})
	server.RegisterTool("test_tool", "A test tool", func(ctx context.Context, args map[string]any) (string, error) {
		return "tool result", nil
	})

	// Create provider that uses the tool then completes
	provider := &testAgentProvider{
		responses: []*providers.MessageResponse{
			{
				Content: []providers.ContentBlock{
					{Type: "text", Text: "I'll use the test tool"},
					{Type: "tool_use", ID: "1", Name: "test_tool", Input: []byte("{}")},
				},
				StopReason: providers.StopReasonToolUse,
			},
			{
				Content:    []providers.ContentBlock{{Type: "text", Text: "Done"}},
				StopReason: providers.StopReasonEndTurn,
			},
		},
	}

	agent, err := NewAgent(AgentConfig{
		Provider:     provider,
		MCPServer:    NewMCPServerAdapter(server),
		SystemPrompt: "You are a test agent",
	})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	ctx := context.Background()
	if err := agent.Run(ctx, "test prompt"); err != nil {
		t.Errorf("agent run failed: %v", err)
	}

	if provider.callCount != 2 {
		t.Errorf("expected 2 provider calls, got %d", provider.callCount)
	}
}

func TestAgent_Run_WithDeveloper(t *testing.T) {
	// Create MCP server with ask_developer tool
	server := mcp.NewServer(mcp.Config{Name: "test"})
	// Note: ask_developer is handled specially by Agent.executeTool, not registered in MCP

	developer := &testAgentDeveloper{
		answers: []string{"blue"},
	}

	// Create provider that asks a question
	provider := &testAgentProvider{
		responses: []*providers.MessageResponse{
			{
				Content: []providers.ContentBlock{
					{Type: "text", Text: "I need to ask a question"},
					{Type: "tool_use", ID: "1", Name: "ask_developer", Input: []byte(`{"question":"What color?"}`)},
				},
				StopReason: providers.StopReasonToolUse,
			},
			{
				Content:    []providers.ContentBlock{{Type: "text", Text: "The color is blue"}},
				StopReason: providers.StopReasonEndTurn,
			},
		},
	}

	session := results.NewSession("test", "test-scenario")
	agent, err := NewAgent(AgentConfig{
		Provider:     provider,
		MCPServer:    NewMCPServerAdapter(server),
		Developer:    developer,
		Session:      session,
		SystemPrompt: "You are a test agent",
	})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	ctx := context.Background()
	if err := agent.Run(ctx, "test prompt"); err != nil {
		t.Errorf("agent run failed: %v", err)
	}

	if developer.calls != 1 {
		t.Errorf("expected 1 developer call, got %d", developer.calls)
	}

	if len(session.Questions) != 1 {
		t.Errorf("expected 1 question in session, got %d", len(session.Questions))
	}
}

func TestAgent_Run_ContextCancellation(t *testing.T) {
	server := mcp.NewServer(mcp.Config{Name: "test"})

	// Provider that never ends
	provider := &testAgentProvider{
		responses: []*providers.MessageResponse{
			{
				Content: []providers.ContentBlock{
					{Type: "text", Text: "Looping forever"},
				},
				StopReason: providers.StopReasonToolUse, // Keep looping
			},
		},
	}

	agent, err := NewAgent(AgentConfig{
		Provider:     provider,
		MCPServer:    NewMCPServerAdapter(server),
		SystemPrompt: "You are a test agent",
	})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = agent.Run(ctx, "test prompt")
	if err != context.Canceled {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

func TestMCPServerAdapter(t *testing.T) {
	server := mcp.NewServer(mcp.Config{Name: "test"})
	server.RegisterToolWithSchema("test_tool", "A test tool", func(ctx context.Context, args map[string]any) (string, error) {
		return "result", nil
	}, map[string]any{
		"type": "object",
		"properties": map[string]any{
			"param": map[string]any{"type": "string"},
		},
	})

	adapter := NewMCPServerAdapter(server)

	// Test ExecuteTool
	result, err := adapter.ExecuteTool(context.Background(), "test_tool", map[string]any{"param": "value"})
	if err != nil {
		t.Errorf("ExecuteTool failed: %v", err)
	}
	if result != "result" {
		t.Errorf("expected 'result', got '%s'", result)
	}

	// Test GetTools
	tools := adapter.GetTools()
	if len(tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(tools))
	}
	if tools[0].Name != "test_tool" {
		t.Errorf("expected 'test_tool', got '%s'", tools[0].Name)
	}

	// Test non-existent tool
	_, err = adapter.ExecuteTool(context.Background(), "nonexistent", nil)
	if err == nil {
		t.Error("expected error for non-existent tool")
	}
}
