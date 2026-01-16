package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lex00/wetwire-core-go/providers"
	"github.com/lex00/wetwire-core-go/providers/claude"
	"github.com/lex00/wetwire-core-go/scenario"
)

// DomainRunner executes scenarios using domain-specific MCP tools via Claude Code CLI.
type DomainRunner struct {
	provider      *claude.Provider
	config        *scenario.ScenarioConfig
	mcpConfigPath string
	output        io.Writer
	verbose       bool
}

// DomainRunnerConfig configures a DomainRunner.
type DomainRunnerConfig struct {
	// ScenarioConfig is the loaded scenario configuration
	ScenarioConfig *scenario.ScenarioConfig

	// WorkDir is the working directory for Claude Code
	WorkDir string

	// Output is where to write progress messages
	Output io.Writer

	// Verbose enables detailed output
	Verbose bool

	// Model overrides the scenario's model setting
	Model string
}

// NewDomainRunner creates a new domain runner that uses Claude Code CLI with domain MCP tools.
func NewDomainRunner(ctx context.Context, cfg DomainRunnerConfig) (*DomainRunner, error) {
	if cfg.ScenarioConfig == nil {
		return nil, fmt.Errorf("scenario config is required")
	}

	if !claude.Available() {
		return nil, fmt.Errorf("claude CLI not found in PATH")
	}

	output := cfg.Output
	if output == nil {
		output = os.Stdout
	}

	// Build MCP server config from domains
	mcpServers := make(map[string]claude.MCPServerConfig)
	for _, domain := range cfg.ScenarioConfig.Domains {
		if domain.CLI == "" {
			continue
		}
		mcpServers[domain.Name] = claude.MCPServerConfig{
			Command: domain.CLI,
			Args:    []string{"mcp"},
			Cwd:     cfg.WorkDir,
		}
	}

	if len(mcpServers) == 0 {
		return nil, fmt.Errorf("no domains with CLI specified in scenario")
	}

	// Write MCP config to temp file
	mcpConfigPath := filepath.Join(cfg.WorkDir, ".mcp-config.json")
	if err := claude.WriteMCPConfig(mcpConfigPath, mcpServers); err != nil {
		return nil, fmt.Errorf("failed to write MCP config: %w", err)
	}

	// Build allowed tools list from domains
	allowedTools := buildAllowedTools(cfg.ScenarioConfig.Domains)

	// Determine model
	model := cfg.Model
	if model == "" {
		model = cfg.ScenarioConfig.Model
	}

	// Create Claude provider with MCP config
	provider, err := claude.New(claude.Config{
		WorkDir:        cfg.WorkDir,
		Model:          model,
		MCPConfigPath:  mcpConfigPath,
		AllowedTools:   allowedTools,
		PermissionMode: "acceptEdits",
		SystemPrompt:   buildDomainSystemPrompt(cfg.ScenarioConfig),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create claude provider: %w", err)
	}

	return &DomainRunner{
		provider:      provider,
		config:        cfg.ScenarioConfig,
		mcpConfigPath: mcpConfigPath,
		output:        output,
		verbose:       cfg.Verbose,
	}, nil
}

// Close releases resources.
func (r *DomainRunner) Close() error {
	// Remove temp MCP config file
	if r.mcpConfigPath != "" {
		_ = os.Remove(r.mcpConfigPath)
	}
	return nil
}

// Run executes the scenario with the given user prompt.
func (r *DomainRunner) Run(ctx context.Context, userPrompt string) (*DomainRunResult, error) {
	result := &DomainRunResult{}

	// Stream the response
	resp, err := r.provider.StreamMessage(ctx, providers.MessageRequest{
		Messages: []providers.Message{
			providers.NewUserMessage(userPrompt),
		},
	}, func(text string) {
		if r.verbose {
			_, _ = fmt.Fprint(r.output, text)
		}
		result.Response += text
	})

	if err != nil {
		return result, fmt.Errorf("execution failed: %w", err)
	}

	result.Success = resp.StopReason == providers.StopReasonEndTurn
	return result, nil
}

// DomainRunResult contains the result of a domain scenario run.
type DomainRunResult struct {
	Success  bool
	Response string
}

// buildAllowedTools creates the list of allowed tools for Claude Code.
// This includes built-in tools plus MCP tools from domains.
func buildAllowedTools(domains []scenario.DomainSpec) []string {
	// Start with essential built-in tools
	tools := []string{
		"Read",  // For reading files
		"Write", // For writing Go code
		"Bash",  // For mkdir, etc.
		"Glob",  // For finding files
	}

	// Add MCP tools from each domain
	for _, domain := range domains {
		for _, toolName := range domain.MCPTools {
			// MCP tools are prefixed with "mcp__servername__"
			mcpTool := fmt.Sprintf("mcp__%s__%s", domain.Name, toolName)
			tools = append(tools, mcpTool)
		}
	}

	return tools
}

// buildDomainSystemPrompt creates the system prompt with domain tool instructions.
func buildDomainSystemPrompt(config *scenario.ScenarioConfig) string {
	var sb strings.Builder

	sb.WriteString("You are an infrastructure code generator using wetwire domain tools.\n\n")

	sb.WriteString("## CRITICAL INSTRUCTIONS\n\n")
	sb.WriteString("You MUST use the domain-specific wetwire MCP tools to generate infrastructure.\n")
	sb.WriteString("Do NOT write raw YAML or JSON output files directly.\n")
	sb.WriteString("Instead, write Go code using wetwire patterns, then use domain tools to build the output.\n\n")

	sb.WriteString("## Workflow\n\n")
	sb.WriteString("For each domain, follow this workflow:\n")
	sb.WriteString("1. Call the domain's `wetwire_init` tool to initialize the project\n")
	sb.WriteString("2. Use the Write tool to create Go code with wetwire patterns (typed structs, direct references)\n")
	sb.WriteString("3. Call the domain's `wetwire_lint` tool to validate the code\n")
	sb.WriteString("4. Fix any lint issues and re-lint until passing\n")
	sb.WriteString("5. Call the domain's `wetwire_build` tool to generate the final output\n\n")

	sb.WriteString("## Available Domains and Tools\n\n")

	for _, domain := range config.Domains {
		if domain.CLI == "" {
			continue
		}

		sb.WriteString(fmt.Sprintf("### %s Domain\n\n", strings.ToUpper(domain.Name[:1])+domain.Name[1:]))
		sb.WriteString(fmt.Sprintf("CLI: `%s`\n\n", domain.CLI))

		if len(domain.MCPTools) > 0 {
			sb.WriteString("Tools:\n")
			for purpose, toolName := range domain.MCPTools {
				sb.WriteString(fmt.Sprintf("- `%s` - %s\n", toolName, purpose))
			}
			sb.WriteString("\n")
		}

		if len(domain.Outputs) > 0 {
			sb.WriteString(fmt.Sprintf("Expected outputs: %s\n\n", strings.Join(domain.Outputs, ", ")))
		}
	}

	// Add domain execution order
	order, err := scenario.GetDomainOrder(config)
	if err == nil && len(order) > 1 {
		sb.WriteString("## Domain Execution Order\n\n")
		sb.WriteString("Execute domains in this order (respecting dependencies):\n")
		for i, name := range order {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, name))
		}
		sb.WriteString("\n")
	}

	// Add cross-domain requirements
	if len(config.CrossDomain) > 0 {
		sb.WriteString("## Cross-Domain Integration\n\n")
		for _, cd := range config.CrossDomain {
			sb.WriteString(fmt.Sprintf("- %s → %s (%s)\n", cd.From, cd.To, cd.Type))
			if len(cd.Validation.RequiredRefs) > 0 {
				sb.WriteString(fmt.Sprintf("  Required references: %s\n", strings.Join(cd.Validation.RequiredRefs, ", ")))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
