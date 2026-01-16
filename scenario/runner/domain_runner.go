package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lex00/wetwire-core-go/providers"
	"github.com/lex00/wetwire-core-go/providers/anthropic"
	"github.com/lex00/wetwire-core-go/scenario"
)

// DomainRunner executes scenarios using domain-specific MCP tools.
type DomainRunner struct {
	mcpManager *MCPManager
	provider   *anthropic.Provider
	config     *scenario.ScenarioConfig
	output     io.Writer
	verbose    bool
	model      string
}

// DomainRunnerConfig configures a DomainRunner.
type DomainRunnerConfig struct {
	// ScenarioConfig is the loaded scenario configuration
	ScenarioConfig *scenario.ScenarioConfig

	// WorkDir is the working directory for MCP servers
	WorkDir string

	// Output is where to write progress messages
	Output io.Writer

	// Verbose enables detailed output
	Verbose bool

	// Model overrides the scenario's model setting
	Model string

	// Debug enables MCP debug logging
	Debug bool
}

// NewDomainRunner creates a new domain runner.
func NewDomainRunner(ctx context.Context, cfg DomainRunnerConfig) (*DomainRunner, error) {
	if cfg.ScenarioConfig == nil {
		return nil, fmt.Errorf("scenario config is required")
	}

	output := cfg.Output
	if output == nil {
		output = os.Stdout
	}

	// Create Anthropic provider (requires API key)
	provider, err := anthropic.New(anthropic.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create anthropic provider: %w", err)
	}

	// Create MCP manager
	mcpManager := NewMCPManager(cfg.WorkDir, cfg.Debug)

	// Start MCP servers for all domains
	if err := mcpManager.Start(ctx, cfg.ScenarioConfig.Domains); err != nil {
		return nil, fmt.Errorf("failed to start MCP servers: %w", err)
	}

	model := cfg.Model
	if model == "" {
		model = cfg.ScenarioConfig.Model
	}

	return &DomainRunner{
		mcpManager: mcpManager,
		provider:   provider,
		config:     cfg.ScenarioConfig,
		output:     output,
		verbose:    cfg.Verbose,
		model:      model,
	}, nil
}

// Close releases resources.
func (r *DomainRunner) Close() error {
	var errs []error
	if r.mcpManager != nil {
		if err := r.mcpManager.Stop(); err != nil {
			errs = append(errs, err)
		}
	}
	if r.provider != nil {
		if err := r.provider.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// Run executes the scenario with the given user prompt.
func (r *DomainRunner) Run(ctx context.Context, userPrompt string) (*DomainRunResult, error) {
	result := &DomainRunResult{
		ToolCalls: make([]ToolCallRecord, 0),
	}

	// Build system prompt with domain tool instructions
	systemPrompt := r.buildSystemPrompt()

	// Get all tools from all domains
	tools := r.buildToolList()

	// Build initial messages
	messages := []providers.Message{
		providers.NewUserMessage(userPrompt),
	}

	// Agentic loop
	maxTurns := 50
	for turn := 0; turn < maxTurns; turn++ {
		if r.verbose {
			fmt.Fprintf(r.output, "\n[Turn %d]\n", turn+1)
		}

		// Send message
		resp, err := r.provider.StreamMessage(ctx, providers.MessageRequest{
			System:   systemPrompt,
			Messages: messages,
			Tools:    tools,
			Model:    r.model,
		}, func(text string) {
			if r.verbose {
				fmt.Fprint(r.output, text)
			}
			result.Response += text
		})
		if err != nil {
			return result, fmt.Errorf("API call failed: %w", err)
		}

		// Check stop reason
		if resp.StopReason == providers.StopReasonEndTurn {
			result.Success = true
			break
		}

		if resp.StopReason != providers.StopReasonToolUse {
			// Unexpected stop reason
			break
		}

		// Process tool calls
		toolResults := r.processToolCalls(ctx, resp.Content, result)

		// Add assistant message and tool results to conversation
		messages = append(messages, providers.Message{
			Role:    "assistant",
			Content: resp.Content,
		})
		messages = append(messages, providers.Message{
			Role:    "user",
			Content: toolResults,
		})
	}

	return result, nil
}

// DomainRunResult contains the result of a domain scenario run.
type DomainRunResult struct {
	Success   bool
	Response  string
	ToolCalls []ToolCallRecord
}

// ToolCallRecord records a tool call made during execution.
type ToolCallRecord struct {
	Domain    string
	Tool      string
	Arguments map[string]any
	Result    string
	IsError   bool
}

// buildSystemPrompt creates the system prompt with domain tool instructions.
func (r *DomainRunner) buildSystemPrompt() string {
	var sb strings.Builder

	sb.WriteString("You are an infrastructure code generator using wetwire domain tools.\n\n")

	sb.WriteString("## CRITICAL INSTRUCTIONS\n\n")
	sb.WriteString("You MUST use the domain-specific wetwire tools to generate infrastructure.\n")
	sb.WriteString("Do NOT write raw YAML or JSON files directly.\n")
	sb.WriteString("Do NOT use generic file writing tools.\n\n")

	sb.WriteString("## Workflow\n\n")
	sb.WriteString("For each domain, follow this workflow:\n")
	sb.WriteString("1. Call `{domain}.wetwire_init` to initialize the project\n")
	sb.WriteString("2. Write Go code using wetwire patterns (typed structs, direct references)\n")
	sb.WriteString("3. Call `{domain}.wetwire_lint` to validate the code\n")
	sb.WriteString("4. Fix any lint issues and re-lint until passing\n")
	sb.WriteString("5. Call `{domain}.wetwire_build` to generate the final output\n\n")

	sb.WriteString("## Available Domain Tools\n\n")

	for _, domain := range r.config.Domains {
		tools := r.mcpManager.GetTools(domain.Name)
		if len(tools) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("### %s Domain (%s)\n\n", strings.ToUpper(domain.Name[:1])+domain.Name[1:], domain.CLI))

		for _, tool := range tools {
			sb.WriteString(fmt.Sprintf("- `%s.%s`: %s\n", domain.Name, tool.Name, tool.Description))
		}
		sb.WriteString("\n")
	}

	// Add domain execution order
	order, err := scenario.GetDomainOrder(r.config)
	if err == nil && len(order) > 1 {
		sb.WriteString("## Domain Execution Order\n\n")
		sb.WriteString("Execute domains in this order:\n")
		for i, name := range order {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, name))
		}
		sb.WriteString("\n")
	}

	// Add cross-domain requirements
	if len(r.config.CrossDomain) > 0 {
		sb.WriteString("## Cross-Domain Integration\n\n")
		for _, cd := range r.config.CrossDomain {
			sb.WriteString(fmt.Sprintf("- %s → %s (%s)\n", cd.From, cd.To, cd.Type))
			if len(cd.Validation.RequiredRefs) > 0 {
				sb.WriteString(fmt.Sprintf("  Required references: %s\n", strings.Join(cd.Validation.RequiredRefs, ", ")))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// buildToolList creates the tool list from all domain MCP servers.
func (r *DomainRunner) buildToolList() []providers.Tool {
	mcpTools := r.mcpManager.GetAllTools()
	tools := make([]providers.Tool, 0, len(mcpTools))

	for _, t := range mcpTools {
		tool := providers.Tool{
			Name:        t.Name,
			Description: t.Description,
		}

		// Convert MCP input schema to provider schema
		if t.InputSchema != nil {
			if props, ok := t.InputSchema["properties"].(map[string]any); ok {
				tool.InputSchema.Properties = props
			}
			if req, ok := t.InputSchema["required"].([]any); ok {
				for _, r := range req {
					if s, ok := r.(string); ok {
						tool.InputSchema.Required = append(tool.InputSchema.Required, s)
					}
				}
			}
		}

		tools = append(tools, tool)
	}

	return tools
}

// processToolCalls handles tool use blocks and returns tool result blocks.
func (r *DomainRunner) processToolCalls(ctx context.Context, content []providers.ContentBlock, result *DomainRunResult) []providers.ContentBlock {
	var results []providers.ContentBlock

	for _, block := range content {
		if block.Type != "tool_use" {
			continue
		}

		// Parse prefixed tool name
		domain, toolName, err := parsePrefixedTool(block.Name)
		if err != nil {
			results = append(results, providers.ContentBlock{
				Type:      "tool_result",
				ToolUseID: block.ID,
				Content:   fmt.Sprintf("Error: %v", err),
				IsError:   true,
			})
			continue
		}

		// Parse arguments
		var arguments map[string]any
		if block.Input != nil {
			if err := json.Unmarshal(block.Input, &arguments); err != nil {
				results = append(results, providers.ContentBlock{
					Type:      "tool_result",
					ToolUseID: block.ID,
					Content:   fmt.Sprintf("Error parsing arguments: %v", err),
					IsError:   true,
				})
				continue
			}
		}

		if r.verbose {
			fmt.Fprintf(r.output, "\n[Tool Call] %s.%s\n", domain, toolName)
		}

		// Call the tool via MCP
		mcpResult, err := r.mcpManager.CallTool(ctx, domain, toolName, arguments)

		var resultText string
		var isError bool

		if err != nil {
			resultText = fmt.Sprintf("Error: %v", err)
			isError = true
		} else {
			// Concatenate text content blocks
			var sb strings.Builder
			for _, c := range mcpResult.Content {
				if c.Type == "text" {
					sb.WriteString(c.Text)
				}
			}
			resultText = sb.String()
			isError = mcpResult.IsError
		}

		// Record the tool call
		result.ToolCalls = append(result.ToolCalls, ToolCallRecord{
			Domain:    domain,
			Tool:      toolName,
			Arguments: arguments,
			Result:    resultText,
			IsError:   isError,
		})

		if r.verbose {
			if isError {
				fmt.Fprintf(r.output, "[Tool Error] %s\n", resultText)
			} else {
				// Truncate long results
				display := resultText
				if len(display) > 500 {
					display = display[:500] + "..."
				}
				fmt.Fprintf(r.output, "[Tool Result] %s\n", display)
			}
		}

		results = append(results, providers.ContentBlock{
			Type:      "tool_result",
			ToolUseID: block.ID,
			Content:   resultText,
			IsError:   isError,
		})
	}

	return results
}
