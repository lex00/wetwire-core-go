// Package agents provides AI agents backed by configurable AI providers.
//
// The package provides two main agent types:
// - RunnerAgent: Generates infrastructure code with tool access
// - DeveloperAgent: Simulates a developer using a persona
package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lex00/wetwire-core-go/agent/orchestrator"
	"github.com/lex00/wetwire-core-go/agent/results"
	"github.com/lex00/wetwire-core-go/providers"
	anthropicprovider "github.com/lex00/wetwire-core-go/providers/anthropic"
)

// DomainConfig provides domain-specific configuration for the RunnerAgent.
// This enables the same agent infrastructure to work across different wetwire domains
// (AWS, Kubernetes, Honeycomb, etc.) by configuring the CLI command and prompts.
type DomainConfig struct {
	// Name is the domain identifier (e.g., "aws", "honeycomb", "k8s")
	Name string

	// CLICommand is the domain CLI binary name (e.g., "wetwire-aws", "wetwire-honeycomb")
	CLICommand string

	// SystemPrompt provides domain-specific agent instructions
	SystemPrompt string

	// OutputFormat describes what the build command produces (e.g., "CloudFormation JSON", "Query JSON")
	OutputFormat string
}

// RunnerAgent generates infrastructure code using a configurable AI provider.
//
// Deprecated: Use Agent with MCPServerAdapter instead. RunnerAgent hardcodes
// its tools, while the new Agent architecture gets tools from an MCP server.
// This provides better extensibility and consistency across wetwire domains.
//
// Migration example:
//
//	// Old:
//	runner, _ := NewRunnerAgent(RunnerConfig{...})
//	runner.Run(ctx, prompt)
//
//	// New:
//	mcpServer := mcp.NewServer(mcp.Config{Name: "domain"})
//	// Register tools with mcpServer...
//	agent, _ := NewAgent(AgentConfig{
//		Provider:     provider,
//		MCPServer:    NewMCPServerAdapter(mcpServer),
//		SystemPrompt: "...",
//	})
//	agent.Run(ctx, prompt)
type RunnerAgent struct {
	provider       providers.Provider
	model          string
	domain         DomainConfig
	session        *results.Session
	developer      orchestrator.Developer
	workDir        string
	generatedFiles []string
	templateJSON   string
	maxLintCycles  int
	streamHandler  providers.StreamHandler

	// Lint enforcement state
	lintCalled  bool // Has lint been run at least once?
	lintPassed  bool // Did lint pass on the most recent run?
	pendingLint bool // Does code need linting (written since last lint)?
	lintCycles  int  // Number of lint attempts
}

// StreamHandler is called for each text chunk during streaming.
// The handler receives text chunks as they are generated.
// Deprecated: Use providers.StreamHandler instead.
type StreamHandler = providers.StreamHandler

// RunnerConfig configures the RunnerAgent.
type RunnerConfig struct {
	// Domain provides domain-specific configuration (required).
	Domain DomainConfig

	// Provider is the AI provider to use. If nil, defaults to Anthropic.
	Provider providers.Provider

	// APIKey for Anthropic (defaults to ANTHROPIC_API_KEY env var)
	// Only used when Provider is nil.
	APIKey string

	// Model to use (defaults to claude-sonnet-4-20250514)
	Model string

	// WorkDir is the directory to write generated files
	WorkDir string

	// MaxLintCycles is the maximum number of lint/fix attempts
	MaxLintCycles int

	// Session for tracking results
	Session *results.Session

	// Developer to ask clarifying questions
	Developer orchestrator.Developer

	// StreamHandler is called for each text chunk during streaming.
	// If nil, responses are not streamed.
	StreamHandler providers.StreamHandler
}

// NewRunnerAgent creates a new RunnerAgent.
func NewRunnerAgent(config RunnerConfig) (*RunnerAgent, error) {
	provider := config.Provider

	// Default to Anthropic provider if none specified
	if provider == nil {
		var err error
		provider, err = anthropicprovider.New(anthropicprovider.Config{
			APIKey: config.APIKey,
		})
		if err != nil {
			return nil, err
		}
	}

	// Domain is required
	domain := config.Domain
	if domain.CLICommand == "" {
		return nil, fmt.Errorf("domain.CLICommand is required")
	}

	if config.WorkDir == "" {
		config.WorkDir = "."
	}
	if config.MaxLintCycles == 0 {
		config.MaxLintCycles = 3
	}

	model := config.Model
	if model == "" {
		model = anthropicprovider.DefaultModel
	}

	return &RunnerAgent{
		provider:      provider,
		model:         model,
		domain:        domain,
		session:       config.Session,
		developer:     config.Developer,
		workDir:       config.WorkDir,
		maxLintCycles: config.MaxLintCycles,
		streamHandler: config.StreamHandler,
	}, nil
}

// Run executes the runner workflow.
func (r *RunnerAgent) Run(ctx context.Context, prompt string) error {
	// Use domain-specific system prompt
	systemPrompt := r.domain.SystemPrompt

	tools := r.getTools()

	messages := []providers.Message{
		providers.NewUserMessage(prompt),
	}

	// Agentic loop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req := providers.MessageRequest{
			Model:     r.model,
			MaxTokens: 4096,
			System:    systemPrompt,
			Messages:  messages,
			Tools:     tools,
		}

		var resp *providers.MessageResponse
		var err error

		if r.streamHandler != nil {
			// Use streaming API
			resp, err = r.provider.StreamMessage(ctx, req, r.streamHandler)
		} else {
			// Use non-streaming API
			resp, err = r.provider.CreateMessage(ctx, req)
		}
		if err != nil {
			return fmt.Errorf("API call failed: %w", err)
		}

		// Add assistant response to messages
		messages = append(messages, providers.NewAssistantMessage(resp.Content))

		// Check for stop reason
		if resp.StopReason == providers.StopReasonEndTurn {
			// Completion gate: check if lint requirements are met
			if enforcement := r.checkCompletionGate(resp); enforcement != "" {
				// Force agent to continue
				messages = append(messages, providers.NewUserMessage(enforcement))
				continue
			}
			// Agent is done
			break
		}

		// Process tool calls
		if resp.StopReason == providers.StopReasonToolUse {
			var toolResults []providers.ContentBlock
			var toolsCalled []string

			for _, block := range resp.Content {
				if block.Type == "tool_use" {
					result := r.executeTool(ctx, block.Name, block.Input)
					toolResults = append(toolResults, providers.NewToolResult(
						block.ID,
						result,
						false,
					))
					toolsCalled = append(toolsCalled, block.Name)
				}
			}

			messages = append(messages, providers.NewToolResultMessage(toolResults))

			// Check for lint enforcement violations after this turn
			if enforcement := r.checkLintEnforcement(toolsCalled); enforcement != "" {
				messages = append(messages, providers.NewUserMessage(enforcement))
			}
		}
	}

	return nil
}

// checkLintEnforcement checks if the agent violated lint enforcement rules.
// Returns an enforcement message if a violation occurred, empty string otherwise.
func (r *RunnerAgent) checkLintEnforcement(toolsCalled []string) string {
	wroteFile := false
	ranLint := false

	for _, tool := range toolsCalled {
		if tool == "write_file" {
			wroteFile = true
		}
		if tool == "run_lint" {
			ranLint = true
		}
	}

	// Enforcement: If write_file was called but run_lint wasn't in the same turn
	if wroteFile && !ranLint {
		return `ENFORCEMENT: You wrote a file but did not call run_lint in the same turn.
You MUST call run_lint immediately after writing code to check for issues.
Call run_lint now before proceeding.`
	}

	return ""
}

// checkCompletionGate checks if the agent can complete.
// Returns an enforcement message if completion is not allowed.
func (r *RunnerAgent) checkCompletionGate(resp *providers.MessageResponse) string {
	// Extract text from response to check for completion indicators
	var responseText string
	for _, block := range resp.Content {
		if block.Type == "text" {
			responseText += block.Text
		}
	}

	// Check for completion indicators
	lowerText := strings.ToLower(responseText)
	isCompletionAttempt := strings.Contains(lowerText, "done") ||
		strings.Contains(lowerText, "complete") ||
		strings.Contains(lowerText, "finished") ||
		strings.Contains(lowerText, "that's it") ||
		strings.Contains(lowerText, "all set")

	if !isCompletionAttempt && len(r.generatedFiles) == 0 {
		// Agent hasn't written any files yet, let it continue thinking
		return ""
	}

	// Gate 1: Must have called lint at least once
	if !r.lintCalled {
		return `ENFORCEMENT: You cannot complete without running the linter.
You MUST call run_lint to validate your code before finishing.
Call run_lint now.`
	}

	// Gate 2: Code must not be pending lint (written since last lint)
	if r.pendingLint {
		return `ENFORCEMENT: You have written code since the last lint run.
You MUST call run_lint to validate your latest changes before finishing.
Call run_lint now.`
	}

	// Gate 3: Lint must have passed
	if !r.lintPassed {
		return `ENFORCEMENT: The linter found issues that have not been resolved.
You MUST fix the lint errors and run_lint again until it passes.
Review the lint output and fix the issues.`
	}

	// All gates passed
	return ""
}

// AskDeveloper sends a question to the Developer.
func (r *RunnerAgent) AskDeveloper(ctx context.Context, question string) (string, error) {
	if r.developer == nil {
		return "", fmt.Errorf("no developer configured")
	}

	answer, err := r.developer.Respond(ctx, question)
	if err != nil {
		return "", err
	}

	if r.session != nil {
		r.session.AddQuestion(question, answer)
	}

	return answer, nil
}

// GetGeneratedFiles returns the list of generated file paths.
func (r *RunnerAgent) GetGeneratedFiles() []string {
	return r.generatedFiles
}

// GetTemplate returns the generated CloudFormation template JSON.
func (r *RunnerAgent) GetTemplate() string {
	return r.templateJSON
}

// GetLintCycles returns the number of lint attempts.
func (r *RunnerAgent) GetLintCycles() int {
	return r.lintCycles
}

// LintPassed returns whether the last lint run passed.
func (r *RunnerAgent) LintPassed() bool {
	return r.lintPassed
}

// getTools returns the tool definitions for the agent.
// Tool descriptions are domain-agnostic where possible.
//
// Deprecated: This method hardcodes tools. Use the unified Agent with
// an MCP server instead, which allows dynamic tool registration.
func (r *RunnerAgent) getTools() []providers.Tool {
	return []providers.Tool{
		{
			Name:        "init_package",
			Description: fmt.Sprintf("Initialize a new %s package directory", r.domain.CLICommand),
			InputSchema: providers.ToolInputSchema{
				Properties: map[string]any{
					"name": map[string]any{
						"type":        "string",
						"description": "Package name (directory name)",
					},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "write_file",
			Description: "Write content to a Go file",
			InputSchema: providers.ToolInputSchema{
				Properties: map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "File path relative to work directory",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "File content",
					},
				},
				Required: []string{"path", "content"},
			},
		},
		{
			Name:        "read_file",
			Description: "Read a file's contents",
			InputSchema: providers.ToolInputSchema{
				Properties: map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "File path relative to work directory",
					},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "run_lint",
			Description: fmt.Sprintf("Run the %s linter on the package", r.domain.CLICommand),
			InputSchema: providers.ToolInputSchema{
				Properties: map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Package path to lint",
					},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "run_build",
			Description: fmt.Sprintf("Build the %s from the package", r.domain.OutputFormat),
			InputSchema: providers.ToolInputSchema{
				Properties: map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Package path to build",
					},
				},
				Required: []string{"path"},
			},
		},
		{
			Name:        "ask_developer",
			Description: "Ask the developer a clarifying question",
			InputSchema: providers.ToolInputSchema{
				Properties: map[string]any{
					"question": map[string]any{
						"type":        "string",
						"description": "The question to ask",
					},
				},
				Required: []string{"question"},
			},
		},
	}
}

// executeTool executes a tool and returns the result.
//
// Deprecated: This method hardcodes tool execution. Use the unified Agent
// which executes tools via MCP server for better extensibility.
func (r *RunnerAgent) executeTool(ctx context.Context, name string, input json.RawMessage) string {
	var params map[string]string
	if err := json.Unmarshal(input, &params); err != nil {
		return fmt.Sprintf("Error parsing input: %v", err)
	}

	switch name {
	case "init_package":
		return r.toolInitPackage(params["name"])
	case "write_file":
		return r.toolWriteFile(params["path"], params["content"])
	case "read_file":
		return r.toolReadFile(params["path"])
	case "run_lint":
		return r.toolRunLint(params["path"])
	case "run_build":
		return r.toolRunBuild(params["path"])
	case "ask_developer":
		answer, err := r.AskDeveloper(ctx, params["question"])
		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}
		return answer
	default:
		return fmt.Sprintf("Unknown tool: %s", name)
	}
}

func (r *RunnerAgent) toolInitPackage(name string) string {
	dir := filepath.Join(r.workDir, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Sprintf("Error creating directory: %v", err)
	}
	return fmt.Sprintf("Created package directory: %s", dir)
}

func (r *RunnerAgent) toolWriteFile(path, content string) string {
	fullPath := filepath.Join(r.workDir, path)

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Sprintf("Error creating directory: %v", err)
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Sprintf("Error writing file: %v", err)
	}

	r.generatedFiles = append(r.generatedFiles, path)

	// Update lint enforcement state: code needs linting
	r.pendingLint = true
	r.lintPassed = false

	return fmt.Sprintf("Wrote %d bytes to %s", len(content), path)
}

func (r *RunnerAgent) toolReadFile(path string) string {
	fullPath := filepath.Join(r.workDir, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Sprintf("Error reading file: %v", err)
	}
	return string(content)
}

func (r *RunnerAgent) toolRunLint(path string) string {
	fullPath := filepath.Join(r.workDir, path)
	cmd := exec.Command(r.domain.CLICommand, "lint", fullPath, "--format", "json")
	output, err := cmd.CombinedOutput()

	result := string(output)

	// Update lint enforcement state
	r.lintCalled = true
	r.pendingLint = false
	r.lintCycles++

	if err != nil {
		// Lint found issues but didn't crash
		r.lintPassed = false
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 2 {
			// Parse issues for session tracking
			var lintResult struct {
				Success bool `json:"success"`
				Issues  []struct {
					Message string `json:"message"`
				} `json:"issues"`
			}
			if json.Unmarshal(output, &lintResult) == nil && r.session != nil {
				issues := make([]string, len(lintResult.Issues))
				for i, issue := range lintResult.Issues {
					issues[i] = issue.Message
				}
				r.session.AddLintCycle(issues, r.lintCycles, false)
			}
		}
	} else {
		// Lint passed
		r.lintPassed = true
		if r.session != nil {
			r.session.AddLintCycle(nil, r.lintCycles, true)
		}
	}

	return result
}

func (r *RunnerAgent) toolRunBuild(path string) string {
	fullPath := filepath.Join(r.workDir, path)
	cmd := exec.Command(r.domain.CLICommand, "build", fullPath, "--format", "json")
	output, err := cmd.CombinedOutput()

	result := string(output)
	if err == nil {
		// Extract template JSON
		var buildResult struct {
			Success  bool        `json:"success"`
			Template interface{} `json:"template"`
		}
		if json.Unmarshal(output, &buildResult) == nil && buildResult.Success {
			if templateData, err := json.Marshal(buildResult.Template); err == nil {
				r.templateJSON = string(templateData)
			}
		}
	}

	return result
}

// CreateDeveloperResponder creates a responder function for AIDeveloper.
func CreateDeveloperResponder(apiKey string) func(ctx context.Context, systemPrompt, message string) (string, error) {
	return CreateDeveloperResponderWithProvider(nil, apiKey)
}

// CreateDeveloperResponderWithProvider creates a responder function for AIDeveloper using the specified provider.
func CreateDeveloperResponderWithProvider(provider providers.Provider, apiKey string) func(ctx context.Context, systemPrompt, message string) (string, error) {
	return func(ctx context.Context, systemPrompt, message string) (string, error) {
		p := provider
		if p == nil {
			var err error
			p, err = anthropicprovider.New(anthropicprovider.Config{APIKey: apiKey})
			if err != nil {
				return "", err
			}
		}

		req := providers.MessageRequest{
			Model:     "claude-3-5-haiku-latest",
			MaxTokens: 1024,
			System:    systemPrompt,
			Messages:  []providers.Message{providers.NewUserMessage(message)},
		}

		resp, err := p.CreateMessage(ctx, req)
		if err != nil {
			return "", err
		}

		var response strings.Builder
		for _, block := range resp.Content {
			if block.Type == "text" {
				response.WriteString(block.Text)
			}
		}

		return response.String(), nil
	}
}

// ============================================================================
// Unified Agent Architecture (Issue #56)
// ============================================================================

// Developer is the interface for asking clarifying questions during agent execution.
// This is optional - if nil, the agent runs autonomously.
type Developer interface {
	// Respond generates a response to a question from the agent.
	Respond(ctx context.Context, message string) (string, error)
}

// MCPServer is the interface for MCP servers that provide tools to agents.
type MCPServer interface {
	// ExecuteTool executes a tool directly without stdio.
	ExecuteTool(ctx context.Context, name string, args map[string]any) (string, error)

	// GetTools returns the list of registered tools.
	GetTools() []MCPToolInfo
}

// MCPToolInfo describes a tool available from an MCP server.
type MCPToolInfo struct {
	Name        string
	Description string
	InputSchema map[string]any
}

// Agent represents the unified agent that can operate in multiple modes:
// - Autonomous: developer=nil, runs without human interaction
// - Interactive: developer!=nil, can ask clarifying questions
// - Scenario: runs test scenarios with AI personas
//
// The Agent gets all its tools from an MCP server, not hardcoded methods.
// This makes it extensible and domain-agnostic.
type Agent struct {
	provider      providers.Provider
	model         string
	mcpServer     MCPServer
	session       *results.Session
	developer     Developer
	systemPrompt  string
	streamHandler providers.StreamHandler
}

// AgentConfig configures the unified Agent.
type AgentConfig struct {
	// Provider is the AI provider (required)
	Provider providers.Provider

	// Model to use (defaults to claude-sonnet-4-20250514)
	Model string

	// MCPServer provides tools (required)
	MCPServer MCPServer

	// Session tracks execution results (optional)
	Session *results.Session

	// Developer to ask questions (nil for autonomous mode)
	Developer Developer

	// SystemPrompt for the agent (required)
	SystemPrompt string

	// StreamHandler for streaming responses (optional)
	StreamHandler providers.StreamHandler
}

// NewAgent creates a new unified Agent.
func NewAgent(config AgentConfig) (*Agent, error) {
	if config.Provider == nil {
		return nil, fmt.Errorf("provider is required")
	}
	if config.MCPServer == nil {
		return nil, fmt.Errorf("mcpServer is required")
	}
	if config.SystemPrompt == "" {
		return nil, fmt.Errorf("systemPrompt is required")
	}

	model := config.Model
	if model == "" {
		model = anthropicprovider.DefaultModel
	}

	return &Agent{
		provider:      config.Provider,
		model:         model,
		mcpServer:     config.MCPServer,
		session:       config.Session,
		developer:     config.Developer,
		systemPrompt:  config.SystemPrompt,
		streamHandler: config.StreamHandler,
	}, nil
}

// Run executes the agent's workflow with the given prompt.
func (a *Agent) Run(ctx context.Context, prompt string) error {
	// Get tools from MCP server
	mcpTools := a.mcpServer.GetTools()
	tools := make([]providers.Tool, len(mcpTools))
	for i, t := range mcpTools {
		tools[i] = providers.Tool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: providers.ToolInputSchema{
				Properties: a.extractProperties(t.InputSchema),
				Required:   a.extractRequired(t.InputSchema),
			},
		}
	}

	messages := []providers.Message{
		providers.NewUserMessage(prompt),
	}

	// Agentic loop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req := providers.MessageRequest{
			Model:     a.model,
			MaxTokens: 4096,
			System:    a.systemPrompt,
			Messages:  messages,
			Tools:     tools,
		}

		var resp *providers.MessageResponse
		var err error

		if a.streamHandler != nil {
			resp, err = a.provider.StreamMessage(ctx, req, a.streamHandler)
		} else {
			resp, err = a.provider.CreateMessage(ctx, req)
		}
		if err != nil {
			return fmt.Errorf("API call failed: %w", err)
		}

		// Add assistant response to messages
		messages = append(messages, providers.NewAssistantMessage(resp.Content))

		// Check for stop reason
		if resp.StopReason == providers.StopReasonEndTurn {
			break
		}

		// Process tool calls
		if resp.StopReason == providers.StopReasonToolUse {
			var toolResults []providers.ContentBlock

			for _, block := range resp.Content {
				if block.Type == "tool_use" {
					result, err := a.executeTool(ctx, block.Name, block.Input)
					toolResults = append(toolResults, providers.NewToolResult(
						block.ID,
						result,
						err != nil,
					))
				}
			}

			messages = append(messages, providers.NewToolResultMessage(toolResults))
		}
	}

	return nil
}

// executeTool executes a tool via the MCP server.
func (a *Agent) executeTool(ctx context.Context, name string, input json.RawMessage) (string, error) {
	var args map[string]any
	if err := json.Unmarshal(input, &args); err != nil {
		return "", fmt.Errorf("error parsing input: %w", err)
	}

	// Special handling for ask_developer if available
	if name == "ask_developer" && a.developer != nil {
		question, ok := args["question"].(string)
		if !ok {
			return "", fmt.Errorf("ask_developer requires a 'question' string parameter")
		}

		answer, err := a.developer.Respond(ctx, question)
		if err != nil {
			return "", err
		}

		if a.session != nil {
			a.session.AddQuestion(question, answer)
		}

		return answer, nil
	}

	// Execute via MCP server
	result, err := a.mcpServer.ExecuteTool(ctx, name, args)
	if err != nil {
		return "", err
	}

	return result, nil
}

// extractProperties extracts the properties map from a JSON schema.
func (a *Agent) extractProperties(schema map[string]any) map[string]any {
	if props, ok := schema["properties"].(map[string]any); ok {
		return props
	}
	return map[string]any{}
}

// extractRequired extracts the required fields from a JSON schema.
func (a *Agent) extractRequired(schema map[string]any) []string {
	if req, ok := schema["required"].([]any); ok {
		required := make([]string, len(req))
		for i, r := range req {
			if s, ok := r.(string); ok {
				required[i] = s
			}
		}
		return required
	}
	if req, ok := schema["required"].([]string); ok {
		return req
	}
	return []string{}
}
