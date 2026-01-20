package kiro

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGenerateMCPConfig_NilArgs(t *testing.T) {
	config := Config{
		MCPCommand: "test-mcp",
		MCPArgs:    nil, // Explicitly nil
	}

	result := GenerateMCPConfig(config)

	// Marshal to JSON
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	jsonStr := string(data)

	// Verify args is [] not null
	if strings.Contains(jsonStr, `"args":null`) {
		t.Errorf("Expected args to be [], got null. JSON: %s", jsonStr)
	}

	if !strings.Contains(jsonStr, `"args":[]`) {
		t.Errorf("Expected args to contain [], JSON: %s", jsonStr)
	}
}

func TestGenerateAgentConfig_NilArgs(t *testing.T) {
	config := Config{
		AgentName:   "test-agent",
		AgentPrompt: "test prompt",
		MCPCommand:  "test-mcp",
		MCPArgs:     nil, // Explicitly nil
		WorkDir:     "/tmp",
	}

	result := GenerateAgentConfig(config)

	// Marshal to JSON
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	jsonStr := string(data)

	// Verify args is [] not null
	if strings.Contains(jsonStr, `"args":null`) {
		t.Errorf("Expected args to be [], got null. JSON: %s", jsonStr)
	}

	if !strings.Contains(jsonStr, `"args":[]`) {
		t.Errorf("Expected args to contain [], JSON: %s", jsonStr)
	}
}

func TestGenerateMCPConfig_WithArgs(t *testing.T) {
	config := Config{
		MCPCommand: "test-mcp",
		MCPArgs:    []string{"--flag", "value"},
	}

	result := GenerateMCPConfig(config)

	// Marshal to JSON
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	jsonStr := string(data)

	// Verify args contains the values
	if !strings.Contains(jsonStr, `"args":["--flag","value"]`) {
		t.Errorf("Expected args to contain provided values, JSON: %s", jsonStr)
	}
}

func TestGenerateAgentConfig_WithArgs(t *testing.T) {
	config := Config{
		AgentName:   "test-agent",
		AgentPrompt: "test prompt",
		MCPCommand:  "test-mcp",
		MCPArgs:     []string{"mcp"},
		WorkDir:     "/tmp",
	}

	result := GenerateAgentConfig(config)

	// Marshal to JSON
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	jsonStr := string(data)

	// Verify args contains the values
	if !strings.Contains(jsonStr, `"args":["mcp"]`) {
		t.Errorf("Expected args to contain provided values, JSON: %s", jsonStr)
	}
}

func TestGenerateMCPConfig_EmptyArgs(t *testing.T) {
	config := Config{
		MCPCommand: "test-mcp",
		MCPArgs:    []string{}, // Explicitly empty
	}

	result := GenerateMCPConfig(config)

	// Marshal to JSON
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	jsonStr := string(data)

	// Verify args is []
	if !strings.Contains(jsonStr, `"args":[]`) {
		t.Errorf("Expected args to be [], JSON: %s", jsonStr)
	}
}
