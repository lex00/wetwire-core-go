// Package kiro provides integration with Kiro CLI for AI agent execution.
//
// This package provides infrastructure for launching and managing Kiro agents,
// which allow for interactive AI sessions with configurable MCP servers.
package kiro

import (
	"context"
	"fmt"
)

// Config contains configuration for launching a Kiro agent.
type Config struct {
	// AgentName is the identifier for this agent (e.g., "wetwire-gitlab-runner")
	AgentName string

	// AgentPrompt is the system prompt for the agent (domain-specific instructions)
	AgentPrompt string

	// MCPCommand is the MCP server command to run (e.g., "wetwire-gitlab-mcp")
	MCPCommand string

	// WorkDir is the working directory for the agent
	WorkDir string
}

// Launch starts a Kiro agent session with the given configuration and prompt.
// This is a placeholder for the full Kiro integration.
func Launch(ctx context.Context, config Config, prompt string) error {
	// TODO: Implement Kiro CLI launching
	return fmt.Errorf("kiro integration not yet implemented")
}

// Install installs Kiro configuration files to the user's config directory.
// This is a placeholder for the full Kiro installation.
func Install(config Config) error {
	// TODO: Implement Kiro config installation
	return fmt.Errorf("kiro installation not yet implemented")
}

// RunTest executes a non-interactive test scenario using Kiro.
// This is a placeholder for the full Kiro test runner.
func RunTest(ctx context.Context, config Config, prompt string) (string, error) {
	// TODO: Implement Kiro test runner
	return "", fmt.Errorf("kiro test runner not yet implemented")
}
