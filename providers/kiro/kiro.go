// Package kiro provides a Kiro CLI implementation of the Provider interface.
package kiro

import (
	"context"
	"fmt"
	"strings"

	"github.com/lex00/wetwire-core-go/kiro"
	"github.com/lex00/wetwire-core-go/providers"
)

// Provider implements the providers.Provider interface using Kiro CLI.
type Provider struct {
	config Config
}

// Config contains configuration for the Kiro provider.
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

// New creates a new Kiro provider.
func New(config Config) (*Provider, error) {
	return &Provider{
		config: config,
	}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "kiro"
}

// CreateMessage sends a message request and returns the complete response.
// It runs kiro-cli in non-interactive mode and parses the output.
func (p *Provider) CreateMessage(ctx context.Context, req providers.MessageRequest) (*providers.MessageResponse, error) {
	if !kiro.KiroAvailable() {
		return nil, fmt.Errorf("kiro-cli not found in PATH")
	}

	prompt := p.buildPrompt(req)

	kiroConfig := kiro.Config{
		AgentName:   p.config.AgentName,
		AgentPrompt: p.config.AgentPrompt,
		MCPCommand:  p.config.MCPCommand,
		MCPArgs:     p.config.MCPArgs,
		WorkDir:     p.config.WorkDir,
	}

	result, err := kiro.RunTest(ctx, kiroConfig, prompt)
	if err != nil {
		return nil, fmt.Errorf("kiro execution failed: %w", err)
	}

	return p.parseOutput(result.Output, result.ExitCode), nil
}

// StreamMessage sends a message request and streams the response via the handler.
// Note: Kiro CLI doesn't support true streaming, so this calls CreateMessage
// and delivers the full response through the handler.
func (p *Provider) StreamMessage(ctx context.Context, req providers.MessageRequest, handler providers.StreamHandler) (*providers.MessageResponse, error) {
	resp, err := p.CreateMessage(ctx, req)
	if err != nil {
		return nil, err
	}

	// Deliver the full response through the handler
	for _, block := range resp.Content {
		if block.Type == "text" {
			handler(block.Text)
		}
	}

	return resp, nil
}

// buildPrompt constructs a prompt string from the message request.
// It extracts user messages and concatenates them.
func (p *Provider) buildPrompt(req providers.MessageRequest) string {
	var userMessages []string

	for _, msg := range req.Messages {
		if msg.Role == "user" {
			for _, block := range msg.Content {
				if block.Type == "text" {
					userMessages = append(userMessages, block.Text)
				}
			}
		}
	}

	return strings.Join(userMessages, "\n\n")
}

// parseOutput converts Kiro CLI output into a MessageResponse.
func (p *Provider) parseOutput(output string, exitCode int) *providers.MessageResponse {
	resp := &providers.MessageResponse{
		StopReason: providers.StopReasonEndTurn,
	}

	if output != "" {
		resp.Content = []providers.ContentBlock{
			{
				Type: "text",
				Text: output,
			},
		}
	}

	return resp
}
