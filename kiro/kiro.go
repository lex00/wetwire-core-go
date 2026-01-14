// Package kiro provides integration with Kiro CLI for AI agent execution.
//
// This package provides infrastructure for launching and managing Kiro agents,
// which allow for interactive AI sessions with configurable MCP servers.
package kiro

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Config contains configuration for launching a Kiro agent.
type Config struct {
	// AgentName is the identifier for this agent (e.g., "wetwire-gitlab-runner")
	AgentName string

	// AgentPrompt is the system prompt for the agent (domain-specific instructions)
	AgentPrompt string

	// MCPCommand is the MCP server command to run (e.g., "wetwire-gitlab-mcp")
	MCPCommand string

	// MCPArgs are optional arguments for the MCP server
	MCPArgs []string

	// WorkDir is the working directory for the agent
	WorkDir string
}

// TestResult contains the result of running a test scenario.
type TestResult struct {
	// Output is the captured output from the Kiro session
	Output string

	// ExitCode is the exit code from kiro-cli
	ExitCode int

	// Error contains any error message if the test failed
	Error string
}

// KiroAvailable checks if Kiro CLI is installed and available.
func KiroAvailable() bool {
	_, err := exec.LookPath("kiro-cli")
	if err == nil {
		return true
	}
	_, err = exec.LookPath("kiro")
	return err == nil
}

// getKiroCommand returns the kiro command name that's available.
func getKiroCommand() (string, error) {
	if _, err := exec.LookPath("kiro-cli"); err == nil {
		return "kiro-cli", nil
	}
	if _, err := exec.LookPath("kiro"); err == nil {
		return "kiro", nil
	}
	return "", errors.New("kiro-cli not found in PATH")
}

// MCPConfig represents the MCP server configuration structure.
type MCPConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

// MCPServerConfig represents a single MCP server configuration.
type MCPServerConfig struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Cwd     string   `json:"cwd,omitempty"`
}

// AgentConfig represents the custom agent configuration structure.
type AgentConfig struct {
	Name       string                       `json:"name"`
	Prompt     string                       `json:"prompt"`
	MCPServers map[string]MCPServerConfig   `json:"mcpServers"`
	Tools      []string                     `json:"tools,omitempty"`
}

// GenerateMCPConfig generates the MCP server configuration.
func GenerateMCPConfig(config Config) MCPConfig {
	args := []string{config.MCPCommand}
	args = append(args, config.MCPArgs...)

	return MCPConfig{
		MCPServers: map[string]MCPServerConfig{
			config.MCPCommand: {
				Command: "uvx",
				Args:    args,
			},
		},
	}
}

// GenerateAgentConfig generates the custom agent configuration.
func GenerateAgentConfig(config Config) AgentConfig {
	args := []string{config.MCPCommand}
	args = append(args, config.MCPArgs...)

	// Ensure WorkDir is set - default to current directory
	workDir := config.WorkDir
	if workDir == "" {
		workDir, _ = os.Getwd()
	}

	return AgentConfig{
		Name:   config.AgentName,
		Prompt: config.AgentPrompt,
		MCPServers: map[string]MCPServerConfig{
			config.MCPCommand: {
				Command: config.MCPCommand,
				Args:    args,
				Cwd:     workDir,
			},
		},
		// Tools array uses @server_name format to include all tools from that MCP server
		// See: https://github.com/aws/amazon-q-developer-cli/issues/2640
		Tools: []string{"@" + config.MCPCommand},
	}
}

// Install installs Kiro configuration files to the user's config directory.
//
// It creates two configuration files:
// - .kiro/mcp.json in the project directory (or WorkDir from config)
// - ~/.kiro/agents/{agent_name}.json in the user's home directory
func Install(config Config) error {
	projectDir := config.WorkDir
	if projectDir == "" {
		var err error
		projectDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create MCP config in project
	mcpDir := filepath.Join(projectDir, ".kiro")
	if err := os.MkdirAll(mcpDir, 0755); err != nil {
		return fmt.Errorf("failed to create .kiro directory: %w", err)
	}

	mcpConfigPath := filepath.Join(mcpDir, "mcp.json")
	mcpConfig := GenerateMCPConfig(config)
	mcpJSON, err := json.MarshalIndent(mcpConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal MCP config: %w", err)
	}
	if err := os.WriteFile(mcpConfigPath, mcpJSON, 0644); err != nil {
		return fmt.Errorf("failed to write MCP config: %w", err)
	}

	// Create agent config in home directory
	agentsDir := filepath.Join(homeDir, ".kiro", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	agentConfigPath := filepath.Join(agentsDir, config.AgentName+".json")
	agentConfig := GenerateAgentConfig(config)
	agentJSON, err := json.MarshalIndent(agentConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal agent config: %w", err)
	}
	if err := os.WriteFile(agentConfigPath, agentJSON, 0644); err != nil {
		return fmt.Errorf("failed to write agent config: %w", err)
	}

	return nil
}

// BuildCommand builds the kiro-cli command arguments.
func BuildCommand(agentName, prompt string, nonInteractive bool) ([]string, error) {
	kiroCmd, err := getKiroCommand()
	if err != nil {
		return nil, err
	}

	args := []string{kiroCmd, "chat", "--agent", agentName}

	if nonInteractive {
		args = append(args, "--no-interactive")
	}

	// Prompt is passed as a positional argument
	args = append(args, prompt)

	return args, nil
}

// Launch starts a Kiro agent session with the given configuration and prompt.
//
// This function installs the configuration files and then launches the Kiro CLI
// in interactive mode. It replaces the current process with kiro-cli.
func Launch(ctx context.Context, config Config, prompt string) error {
	// Install configs first
	if err := Install(config); err != nil {
		return fmt.Errorf("failed to install configs: %w", err)
	}

	args, err := BuildCommand(config.AgentName, prompt, false)
	if err != nil {
		return err
	}

	// Get the full path to kiro-cli
	kiroPath, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("failed to find kiro-cli: %w", err)
	}

	// Set working directory
	workDir := config.WorkDir
	if workDir == "" {
		workDir, _ = os.Getwd()
	}

	// Use exec.CommandContext for cancellation support
	cmd := exec.CommandContext(ctx, kiroPath, args[1:]...)
	cmd.Dir = workDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// RunTest executes a non-interactive test scenario using Kiro.
//
// This runs kiro-cli in non-interactive mode and captures the output.
// It's useful for automated testing of domain packages.
func RunTest(ctx context.Context, config Config, prompt string) (*TestResult, error) {
	// Install configs first
	if err := Install(config); err != nil {
		return nil, fmt.Errorf("failed to install configs: %w", err)
	}

	args, err := BuildCommand(config.AgentName, prompt, true)
	if err != nil {
		return nil, err
	}

	// Set working directory
	workDir := config.WorkDir
	if workDir == "" {
		workDir, _ = os.Getwd()
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	result := &TestResult{
		Output:   string(output),
		ExitCode: 0,
	}

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
			result.Error = err.Error()
		} else {
			return nil, fmt.Errorf("failed to run kiro: %w", err)
		}
	}

	return result, nil
}
