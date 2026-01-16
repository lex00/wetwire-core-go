package claude

import (
	"context"
	"os"
	"path/filepath"
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
	assert.Equal(t, "claude", p.Name())
}

func TestNewWithConfig(t *testing.T) {
	p, err := New(Config{
		WorkDir:        "/tmp/test",
		Model:          "sonnet",
		SystemPrompt:   "You are a helpful assistant",
		PermissionMode: "acceptEdits",
	})
	require.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, "/tmp/test", p.config.WorkDir)
	assert.Equal(t, "sonnet", p.config.Model)
}

func TestNewWithDefaults(t *testing.T) {
	p, err := New(Config{})
	require.NoError(t, err)
	assert.NotNil(t, p)
	// Should default to current working directory
	cwd, _ := os.Getwd()
	assert.Equal(t, cwd, p.config.WorkDir)
}

func TestCreateMessageClaudeUnavailable(t *testing.T) {
	// Save original PATH and set to empty to ensure claude is not found
	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", "/nonexistent")
	defer t.Setenv("PATH", originalPath)

	p := &Provider{
		config: Config{
			WorkDir: "/tmp/test",
		},
	}

	req := providers.MessageRequest{
		Messages: []providers.Message{providers.NewUserMessage("hello")},
	}

	_, err := p.CreateMessage(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "claude CLI not found")
}

func TestStreamMessageClaudeUnavailable(t *testing.T) {
	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", "/nonexistent")
	defer t.Setenv("PATH", originalPath)

	p := &Provider{
		config: Config{
			WorkDir: "/tmp/test",
		},
	}

	req := providers.MessageRequest{
		Messages: []providers.Message{providers.NewUserMessage("hello")},
	}

	handler := func(text string) {}

	_, err := p.StreamMessage(context.Background(), req, handler)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "claude CLI not found")
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
			name: "multiple user messages",
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

func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		req      providers.MessageRequest
		prompt   string
		contains []string
		excludes []string
	}{
		{
			name:   "basic args",
			config: Config{},
			req:    providers.MessageRequest{},
			prompt: "hello",
			contains: []string{
				"--print",
				"--output-format", "json",
				"hello",
			},
		},
		{
			name: "with model",
			config: Config{
				Model: "opus",
			},
			req:    providers.MessageRequest{},
			prompt: "hello",
			contains: []string{
				"--model", "opus",
			},
		},
		{
			name: "with system prompt",
			config: Config{
				SystemPrompt: "You are helpful",
			},
			req:    providers.MessageRequest{},
			prompt: "hello",
			contains: []string{
				"--system-prompt",
			},
		},
		{
			name: "with allowed tools",
			config: Config{
				AllowedTools: []string{"Bash", "Read"},
			},
			req:    providers.MessageRequest{},
			prompt: "hello",
			contains: []string{
				"--allowedTools", "Bash,Read",
			},
		},
		{
			name: "with permission mode",
			config: Config{
				PermissionMode: "acceptEdits",
			},
			req:    providers.MessageRequest{},
			prompt: "hello",
			contains: []string{
				"--permission-mode", "acceptEdits",
			},
		},
		{
			name: "with MCP config",
			config: Config{
				MCPConfigPath: "/path/to/mcp.json",
			},
			req:    providers.MessageRequest{},
			prompt: "hello",
			contains: []string{
				"--mcp-config", "/path/to/mcp.json",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := &Provider{config: tc.config}
			args := p.buildArgs(tc.req, tc.prompt)
			argsStr := joinArgs(args)

			for _, want := range tc.contains {
				assert.Contains(t, argsStr, want, "args should contain %q", want)
			}
			for _, exclude := range tc.excludes {
				assert.NotContains(t, argsStr, exclude, "args should not contain %q", exclude)
			}
		})
	}
}

func TestBuildStreamArgs(t *testing.T) {
	p := &Provider{config: Config{}}
	args := p.buildStreamArgs(providers.MessageRequest{}, "hello")
	argsStr := joinArgs(args)

	assert.Contains(t, argsStr, "--print")
	assert.Contains(t, argsStr, "--output-format")
	assert.Contains(t, argsStr, "stream-json")
	assert.Contains(t, argsStr, "--verbose")
}

func TestParseJSONOutput(t *testing.T) {
	p := &Provider{}

	tests := []struct {
		name     string
		output   string
		wantText string
		wantErr  bool
	}{
		{
			name:     "success result",
			output:   `{"type":"result","subtype":"success","is_error":false,"result":"Hello world","session_id":"abc123"}`,
			wantText: "Hello world",
			wantErr:  false,
		},
		{
			name:     "error result",
			output:   `{"type":"result","subtype":"error","is_error":true,"result":"Something went wrong","session_id":"abc123"}`,
			wantText: "Error: Something went wrong",
			wantErr:  false,
		},
		{
			name:     "empty result",
			output:   `{"type":"result","subtype":"success","is_error":false,"result":"","session_id":"abc123"}`,
			wantText: "",
			wantErr:  false,
		},
		{
			name:    "invalid json",
			output:  `not json`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := p.parseJSONOutput([]byte(tc.output))
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, providers.StopReasonEndTurn, result.StopReason)
			if tc.wantText != "" {
				require.NotEmpty(t, result.Content)
				assert.Equal(t, tc.wantText, result.Content[0].Text)
			}
		})
	}
}

func TestParseStreamEvent(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantType string
		wantErr  bool
	}{
		{
			name:     "system event",
			line:     `{"type":"system","subtype":"init","session_id":"abc123"}`,
			wantType: "system",
			wantErr:  false,
		},
		{
			name:     "assistant event",
			line:     `{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Hello"}]}}`,
			wantType: "assistant",
			wantErr:  false,
		},
		{
			name:     "result event",
			line:     `{"type":"result","subtype":"success","result":"Done"}`,
			wantType: "result",
			wantErr:  false,
		},
		{
			name:    "invalid json",
			line:    `not json`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			event, err := parseStreamEvent(tc.line)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantType, event.Type)
		})
	}
}

func TestWriteMCPConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "mcp.json")

	servers := map[string]MCPServerConfig{
		"mock-cli-a": {
			Command: "mock-cli-a-mcp",
			Args:    []string{"--debug"},
			Cwd:     "/workspace",
		},
	}

	err := WriteMCPConfig(configPath, servers)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// Read and verify content
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	assert.Contains(t, string(data), "mock-cli-a")
	assert.Contains(t, string(data), "mock-cli-a-mcp")
	assert.Contains(t, string(data), "--debug")
}

func TestWriteMCPConfigCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nested", "dir", "mcp.json")

	servers := map[string]MCPServerConfig{
		"test": {
			Command: "test-cmd",
		},
	}

	err := WriteMCPConfig(configPath, servers)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(configPath)
	require.NoError(t, err)
}

func TestAvailable(t *testing.T) {
	// This test checks the Available function works
	// The actual result depends on whether claude is installed
	result := Available()
	// Just verify it returns a boolean without panicking
	assert.IsType(t, true, result)
}

func TestSystemPromptCombination(t *testing.T) {
	tests := []struct {
		name          string
		configPrompt  string
		requestPrompt string
		wantContains  string
	}{
		{
			name:          "config only",
			configPrompt:  "You are helpful",
			requestPrompt: "",
			wantContains:  "You are helpful",
		},
		{
			name:          "request only",
			configPrompt:  "",
			requestPrompt: "Be concise",
			wantContains:  "Be concise",
		},
		{
			name:          "both combined",
			configPrompt:  "You are helpful",
			requestPrompt: "Be concise",
			wantContains:  "You are helpful",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := &Provider{config: Config{SystemPrompt: tc.configPrompt}}
			req := providers.MessageRequest{System: tc.requestPrompt}
			args := p.buildArgs(req, "test")
			argsStr := joinArgs(args)
			assert.Contains(t, argsStr, tc.wantContains)
		})
	}
}

// joinArgs joins args into a single string for easier assertion
func joinArgs(args []string) string {
	result := ""
	for _, arg := range args {
		result += arg + " "
	}
	return result
}
