// Package claude provides a Claude Code CLI implementation of the Provider interface.
//
// This provider uses the `claude` CLI (Claude Code) as the AI backend, allowing
// scenarios to run without an Anthropic API key. Claude Code handles its own
// agentic loop internally, so the caller's loop typically runs once.
//
// Usage:
//
//	provider, err := claude.New(claude.Config{
//		WorkDir: "/path/to/project",
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	resp, err := provider.CreateMessage(ctx, providers.MessageRequest{
//		System:   "You are a helpful assistant.",
//		Messages: []providers.Message{providers.NewUserMessage("Hello")},
//	})
package claude

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lex00/wetwire-core-go/providers"
)

// Provider implements the providers.Provider interface using Claude Code CLI.
type Provider struct {
	config Config
}

// Config contains configuration for the Claude provider.
type Config struct {
	// WorkDir is the working directory for claude CLI (default: current directory)
	WorkDir string

	// Model overrides the default model (optional)
	Model string

	// MCPConfigPath is a path to an existing MCP config file (optional)
	// If provided, tools from MessageRequest are ignored
	MCPConfigPath string

	// SystemPrompt is prepended to the request's system prompt (optional)
	SystemPrompt string

	// Verbose enables verbose output for debugging
	Verbose bool

	// AllowedTools restricts which tools claude can use (optional)
	// Example: []string{"Bash", "Read", "Edit"}
	AllowedTools []string

	// PermissionMode sets the permission mode (optional)
	// Options: "default", "acceptEdits", "plan", etc.
	PermissionMode string
}

// New creates a new Claude Code provider.
func New(config Config) (*Provider, error) {
	if config.WorkDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		config.WorkDir = cwd
	}

	return &Provider{
		config: config,
	}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "claude"
}

// Available checks if the claude CLI is installed and available.
func Available() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

// CreateMessage sends a message request and returns the complete response.
// Claude Code runs its own agentic loop, so this executes the full session
// and returns the final result.
func (p *Provider) CreateMessage(ctx context.Context, req providers.MessageRequest) (*providers.MessageResponse, error) {
	if !Available() {
		return nil, fmt.Errorf("claude CLI not found in PATH")
	}

	// Build the prompt from messages
	prompt := p.buildPrompt(req)

	// Build command arguments
	args := p.buildArgs(req, prompt)

	// Execute claude CLI
	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = p.config.WorkDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("claude execution failed: %w\nOutput: %s", err, string(output))
	}

	// Parse the JSON output
	return p.parseJSONOutput(output)
}

// StreamMessage sends a message request and streams the response via the handler.
// This uses --output-format stream-json to get realtime updates.
func (p *Provider) StreamMessage(ctx context.Context, req providers.MessageRequest, handler providers.StreamHandler) (*providers.MessageResponse, error) {
	if !Available() {
		return nil, fmt.Errorf("claude CLI not found in PATH")
	}

	prompt := p.buildPrompt(req)
	args := p.buildStreamArgs(req, prompt)

	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = p.config.WorkDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Capture stderr for error messages
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start claude: %w", err)
	}

	var finalResponse *providers.MessageResponse
	scanner := bufio.NewScanner(stdout)
	// Increase buffer size for large outputs
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		event, err := parseStreamEvent(line)
		if err != nil {
			continue // Skip unparseable lines
		}

		switch event.Type {
		case "assistant":
			// Stream text content to handler
			if event.Message != nil {
				for _, block := range event.Message.Content {
					if block.Type == "text" && handler != nil {
						handler(block.Text)
					}
				}
			}
		case "result":
			// Build final response from result
			finalResponse = &providers.MessageResponse{
				StopReason: providers.StopReasonEndTurn,
				Content: []providers.ContentBlock{
					{Type: "text", Text: event.Result},
				},
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		// Check if context was cancelled
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		// Include stderr output in error message
		stderrOutput := stderrBuf.String()
		if stderrOutput != "" {
			return nil, fmt.Errorf("claude execution failed: %w\nStderr: %s", err, stderrOutput)
		}
		return nil, fmt.Errorf("claude execution failed: %w", err)
	}

	if finalResponse == nil {
		return &providers.MessageResponse{
			StopReason: providers.StopReasonEndTurn,
		}, nil
	}

	return finalResponse, nil
}

// buildPrompt constructs a prompt string from the message request.
func (p *Provider) buildPrompt(req providers.MessageRequest) string {
	var parts []string

	// Extract user messages
	for _, msg := range req.Messages {
		if msg.Role == "user" {
			for _, block := range msg.Content {
				if block.Type == "text" {
					parts = append(parts, block.Text)
				}
			}
		}
	}

	return strings.Join(parts, "\n\n")
}

// buildArgs constructs command line arguments for non-streaming mode.
func (p *Provider) buildArgs(req providers.MessageRequest, prompt string) []string {
	args := []string{"--print", "--output-format", "json"}

	// Add system prompt
	systemPrompt := p.config.SystemPrompt
	if req.System != "" {
		if systemPrompt != "" {
			systemPrompt = systemPrompt + "\n\n" + req.System
		} else {
			systemPrompt = req.System
		}
	}
	if systemPrompt != "" {
		args = append(args, "--system-prompt", systemPrompt)
	}

	// Add model if specified
	if p.config.Model != "" {
		args = append(args, "--model", p.config.Model)
	}

	// Add MCP config if specified
	if p.config.MCPConfigPath != "" {
		args = append(args, "--mcp-config", p.config.MCPConfigPath)
	}

	// Add allowed tools if specified
	if len(p.config.AllowedTools) > 0 {
		args = append(args, "--allowedTools", strings.Join(p.config.AllowedTools, ","))
	}

	// Add permission mode if specified
	if p.config.PermissionMode != "" {
		args = append(args, "--permission-mode", p.config.PermissionMode)
	}

	// Add -- separator and prompt as positional argument
	args = append(args, "--", prompt)

	return args
}

// buildStreamArgs constructs command line arguments for streaming mode.
func (p *Provider) buildStreamArgs(req providers.MessageRequest, prompt string) []string {
	args := []string{"--print", "--output-format", "stream-json", "--verbose"}

	// Add system prompt
	systemPrompt := p.config.SystemPrompt
	if req.System != "" {
		if systemPrompt != "" {
			systemPrompt = systemPrompt + "\n\n" + req.System
		} else {
			systemPrompt = req.System
		}
	}
	if systemPrompt != "" {
		args = append(args, "--system-prompt", systemPrompt)
	}

	// Add model if specified
	if p.config.Model != "" {
		args = append(args, "--model", p.config.Model)
	}

	// Add MCP config if specified
	if p.config.MCPConfigPath != "" {
		args = append(args, "--mcp-config", p.config.MCPConfigPath)
	}

	// Add allowed tools if specified
	if len(p.config.AllowedTools) > 0 {
		args = append(args, "--allowedTools", strings.Join(p.config.AllowedTools, ","))
	}

	// Add permission mode if specified
	if p.config.PermissionMode != "" {
		args = append(args, "--permission-mode", p.config.PermissionMode)
	}

	// Add -- separator and prompt as positional argument
	args = append(args, "--", prompt)

	return args
}

// parseJSONOutput parses the JSON output from claude --print --output-format json
func (p *Provider) parseJSONOutput(output []byte) (*providers.MessageResponse, error) {
	var result jsonResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse claude output: %w", err)
	}

	resp := &providers.MessageResponse{
		StopReason: providers.StopReasonEndTurn,
	}

	if result.IsError {
		resp.Content = []providers.ContentBlock{
			{Type: "text", Text: fmt.Sprintf("Error: %s", result.Result)},
		}
	} else if result.Result != "" {
		resp.Content = []providers.ContentBlock{
			{Type: "text", Text: result.Result},
		}
	}

	return resp, nil
}

// jsonResult represents the JSON output from claude --output-format json
type jsonResult struct {
	Type      string `json:"type"`
	Subtype   string `json:"subtype"`
	IsError   bool   `json:"is_error"`
	Result    string `json:"result"`
	SessionID string `json:"session_id"`
	NumTurns  int    `json:"num_turns"`
}

// streamEvent represents a single event from stream-json output
type streamEvent struct {
	Type    string         `json:"type"`
	Subtype string         `json:"subtype,omitempty"`
	Message *streamMessage `json:"message,omitempty"`
	Result  string         `json:"result,omitempty"`
	IsError bool           `json:"is_error,omitempty"`
}

// streamMessage represents the message field in an assistant event
type streamMessage struct {
	Role    string               `json:"role"`
	Content []streamContentBlock `json:"content"`
}

// streamContentBlock represents a content block in the stream
type streamContentBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text,omitempty"`
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

// parseStreamEvent parses a single line of stream-json output
func parseStreamEvent(line string) (*streamEvent, error) {
	var event streamEvent
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// WriteMCPConfig writes an MCP configuration file for use with --mcp-config.
// This is useful when you want to provide custom MCP tools to Claude Code.
func WriteMCPConfig(path string, servers map[string]MCPServerConfig) error {
	config := MCPConfig{
		MCPServers: servers,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal MCP config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write MCP config: %w", err)
	}

	return nil
}

// MCPConfig represents the MCP configuration file format.
type MCPConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

// MCPServerConfig represents a single MCP server configuration.
type MCPServerConfig struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
	Cwd     string   `json:"cwd,omitempty"`
}
