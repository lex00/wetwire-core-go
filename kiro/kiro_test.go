package kiro

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateMCPConfig(t *testing.T) {
	config := Config{
		AgentName:   "test-agent",
		AgentPrompt: "Test prompt",
		MCPCommand:  "test-mcp",
		MCPArgs:     []string{"--verbose"},
	}

	mcpConfig := GenerateMCPConfig(config)

	if len(mcpConfig.MCPServers) != 1 {
		t.Errorf("expected 1 MCP server, got %d", len(mcpConfig.MCPServers))
	}

	server, ok := mcpConfig.MCPServers["test-mcp"]
	if !ok {
		t.Error("expected test-mcp server in config")
	}

	if server.Command != "uvx" {
		t.Errorf("expected command 'uvx', got %q", server.Command)
	}

	expectedArgs := []string{"test-mcp", "--verbose"}
	if len(server.Args) != len(expectedArgs) {
		t.Errorf("expected %d args, got %d", len(expectedArgs), len(server.Args))
	}
	for i, arg := range expectedArgs {
		if server.Args[i] != arg {
			t.Errorf("expected arg[%d] = %q, got %q", i, arg, server.Args[i])
		}
	}
}

func TestGenerateAgentConfig(t *testing.T) {
	config := Config{
		AgentName:   "test-agent",
		AgentPrompt: "Test system prompt",
		MCPCommand:  "test-mcp",
	}

	agentConfig := GenerateAgentConfig(config)

	if agentConfig.Name != "test-agent" {
		t.Errorf("expected name 'test-agent', got %q", agentConfig.Name)
	}

	if agentConfig.SystemPrompt != "Test system prompt" {
		t.Errorf("expected system prompt 'Test system prompt', got %q", agentConfig.SystemPrompt)
	}

	if len(agentConfig.MCPServers) != 1 || agentConfig.MCPServers[0] != "test-mcp" {
		t.Errorf("expected MCPServers = [test-mcp], got %v", agentConfig.MCPServers)
	}
}

func TestInstall(t *testing.T) {
	// Create temp directories
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "project")
	homeDir := filepath.Join(tmpDir, "home")

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Override home directory for test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer os.Setenv("HOME", originalHome)

	config := Config{
		AgentName:   "test-agent",
		AgentPrompt: "Test prompt",
		MCPCommand:  "test-mcp",
		WorkDir:     projectDir,
	}

	err := Install(config)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Check MCP config was created
	mcpConfigPath := filepath.Join(projectDir, ".kiro", "mcp.json")
	if _, err := os.Stat(mcpConfigPath); os.IsNotExist(err) {
		t.Error("mcp.json was not created")
	}

	// Check agent config was created
	agentConfigPath := filepath.Join(homeDir, ".kiro", "agents", "test-agent.json")
	if _, err := os.Stat(agentConfigPath); os.IsNotExist(err) {
		t.Error("agent config was not created")
	}
}

func TestBuildCommand(t *testing.T) {
	// Skip if kiro is not installed
	if !KiroAvailable() {
		t.Skip("kiro-cli not installed")
	}

	tests := []struct {
		name           string
		agentName      string
		prompt         string
		nonInteractive bool
		wantLen        int
	}{
		{
			name:           "interactive mode",
			agentName:      "test-agent",
			prompt:         "test prompt",
			nonInteractive: false,
			wantLen:        6, // kiro-cli chat --agent test-agent --prompt "test prompt"
		},
		{
			name:           "non-interactive mode",
			agentName:      "test-agent",
			prompt:         "test prompt",
			nonInteractive: true,
			wantLen:        7, // kiro-cli chat --agent test-agent --non-interactive --prompt "test prompt"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, err := BuildCommand(tt.agentName, tt.prompt, tt.nonInteractive)
			if err != nil {
				t.Fatalf("BuildCommand failed: %v", err)
			}

			if len(args) != tt.wantLen {
				t.Errorf("expected %d args, got %d: %v", tt.wantLen, len(args), args)
			}

			// Verify agent name is in args
			found := false
			for i, arg := range args {
				if arg == "--agent" && i+1 < len(args) && args[i+1] == tt.agentName {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("agent name not found in args: %v", args)
			}

			// Verify non-interactive flag
			if tt.nonInteractive {
				found = false
				for _, arg := range args {
					if arg == "--non-interactive" {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("--non-interactive not found in args: %v", args)
				}
			}
		})
	}
}

func TestBuildCommandNoKiro(t *testing.T) {
	// This test verifies behavior when kiro is not installed
	// We can't easily test this without mocking exec.LookPath
	// So we just verify the error message format
	if KiroAvailable() {
		t.Skip("kiro-cli is installed, cannot test missing kiro case")
	}

	_, err := BuildCommand("test", "prompt", false)
	if err == nil {
		t.Error("expected error when kiro is not installed")
	}
}

func TestKiroAvailable(t *testing.T) {
	// Just verify it returns a boolean without panicking
	_ = KiroAvailable()
}
