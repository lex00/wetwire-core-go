package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
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

	// Convert WorkDir to absolute path
	absWorkDir, err := filepath.Abs(cfg.WorkDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve working directory: %w", err)
	}

	// Build MCP server config from domains
	mcpServers := make(map[string]claude.MCPServerConfig)
	for _, domain := range cfg.ScenarioConfig.Domains {
		if domain.CLI == "" {
			continue
		}
		// Resolve CLI path to full path
		cliPath, err := resolveCLIPath(domain.CLI)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve CLI path for %s: %w", domain.Name, err)
		}
		mcpServers[domain.Name] = claude.MCPServerConfig{
			Command: cliPath,
			Args:    []string{"mcp"},
			Cwd:     absWorkDir,
		}
	}

	if len(mcpServers) == 0 {
		return nil, fmt.Errorf("no domains with CLI specified in scenario")
	}

	// Write MCP config to temp file (use absolute path)
	mcpConfigPath := filepath.Join(absWorkDir, ".mcp-config.json")
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
		WorkDir:        absWorkDir,
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

	sb.WriteString("You are an infrastructure code generator using wetwire domain MCP tools.\n\n")

	sb.WriteString("## ABSOLUTE REQUIREMENTS\n\n")
	sb.WriteString("1. You MUST use the MCP tools listed below to generate infrastructure.\n")
	sb.WriteString("2. You MUST NOT write raw YAML, JSON, or CloudFormation files directly.\n")
	sb.WriteString("3. You MUST write Go code using wetwire patterns, then call domain build tools.\n")
	sb.WriteString("4. All files MUST be created in the current working directory (use relative paths).\n\n")

	sb.WriteString("## Required Workflow (follow exactly)\n\n")
	sb.WriteString("For each domain:\n")
	sb.WriteString("1. Call `mcp__<domain>__wetwire_init` to scaffold the project\n")
	sb.WriteString("2. Use Write tool to create Go code (e.g., `infra/bucket.go` for AWS)\n")
	sb.WriteString("3. Call `mcp__<domain>__wetwire_lint` to validate the Go code\n")
	sb.WriteString("4. Fix lint errors and re-lint until passing\n")
	sb.WriteString("5. Call `mcp__<domain>__wetwire_build` to generate the final output\n\n")

	sb.WriteString("## Available MCP Tools\n\n")

	for _, domain := range config.Domains {
		if domain.CLI == "" {
			continue
		}

		sb.WriteString(fmt.Sprintf("### %s Domain\n\n", strings.ToUpper(domain.Name[:1])+domain.Name[1:]))

		// List tools with full MCP names
		sb.WriteString("Tools (use these exact names):\n")
		sb.WriteString(fmt.Sprintf("- `mcp__%s__wetwire_init` - Initialize project\n", domain.Name))
		sb.WriteString(fmt.Sprintf("- `mcp__%s__wetwire_lint` - Lint Go code\n", domain.Name))
		sb.WriteString(fmt.Sprintf("- `mcp__%s__wetwire_build` - Generate output\n", domain.Name))
		sb.WriteString("\n")

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

// resolveCLIPath attempts to find the full path to a CLI command.
// It first tries exec.LookPath, then checks common Go binary locations.
func resolveCLIPath(cli string) (string, error) {
	// If it's already an absolute path, use it
	if filepath.IsAbs(cli) {
		if _, err := os.Stat(cli); err == nil {
			return cli, nil
		}
		return "", fmt.Errorf("CLI not found at path: %s", cli)
	}

	// Try to find in PATH
	if path, err := exec.LookPath(cli); err == nil {
		return path, nil
	}

	// Try common Go binary locations
	homeDir, err := os.UserHomeDir()
	if err == nil {
		goPathBin := filepath.Join(homeDir, "go", "bin", cli)
		if _, err := os.Stat(goPathBin); err == nil {
			return goPathBin, nil
		}
	}

	// Try GOPATH/bin
	if goPath := os.Getenv("GOPATH"); goPath != "" {
		goPathBin := filepath.Join(goPath, "bin", cli)
		if _, err := os.Stat(goPathBin); err == nil {
			return goPathBin, nil
		}
	}

	return "", fmt.Errorf("CLI %q not found in PATH or common Go binary locations", cli)
}
