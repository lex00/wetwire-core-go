package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lex00/wetwire-core-go/domain"
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

	// DependencyOutputs contains outputs from previously executed domains.
	// These are included in the system prompt so dependent domains can reference them.
	DependencyOutputs *OutputManifest
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
		SystemPrompt:   buildDomainSystemPrompt(cfg.ScenarioConfig, cfg.DependencyOutputs),
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
// If dependencyOutputs is provided, it includes outputs from previously executed domains.
func buildDomainSystemPrompt(config *scenario.ScenarioConfig, dependencyOutputs *OutputManifest) string {
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

	for _, d := range config.Domains {
		if d.CLI == "" {
			continue
		}

		sb.WriteString(fmt.Sprintf("### %s Domain\n\n", strings.ToUpper(d.Name[:1])+d.Name[1:]))

		// List tools with full MCP names
		sb.WriteString("Tools (use these exact names):\n")
		sb.WriteString(fmt.Sprintf("- `mcp__%s__wetwire_init` - Initialize project\n", d.Name))
		sb.WriteString(fmt.Sprintf("- `mcp__%s__wetwire_lint` - Lint Go code\n", d.Name))
		sb.WriteString(fmt.Sprintf("- `mcp__%s__wetwire_build` - Generate output\n", d.Name))
		sb.WriteString("\n")

		if len(d.Outputs) > 0 {
			sb.WriteString(fmt.Sprintf("Expected outputs: %s\n\n", strings.Join(d.Outputs, ", ")))
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

	// Add dependency outputs if available
	if dependencyOutputs != nil && len(dependencyOutputs.Domains) > 0 {
		sb.WriteString("## Available Dependency Outputs\n\n")
		sb.WriteString("The following outputs from dependency domains are available for reference:\n\n")

		for domainName, domainOutput := range dependencyOutputs.Domains {
			if domainOutput == nil || len(domainOutput.Resources) == 0 {
				continue
			}

			sb.WriteString(fmt.Sprintf("### %s Domain Outputs\n\n", strings.ToUpper(domainName[:1])+domainName[1:]))

			for resourceName, resourceOutput := range domainOutput.Resources {
				sb.WriteString(fmt.Sprintf("**%s** (type: %s)\n", resourceName, resourceOutput.Type))
				if len(resourceOutput.Outputs) > 0 {
					// Format outputs as JSON for readability
					outputJSON, err := json.MarshalIndent(resourceOutput.Outputs, "  ", "  ")
					if err == nil {
						sb.WriteString("```json\n")
						sb.WriteString("  ")
						sb.WriteString(string(outputJSON))
						sb.WriteString("\n```\n")
					}
				}
				sb.WriteString("\n")
			}
		}

		sb.WriteString("Use these outputs when referencing resources from dependency domains.\n")
		sb.WriteString("Reference syntax: `${domain.resource.outputs.field}`\n\n")
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

// OutputManifestToCrossDomainContext converts an OutputManifest to a domain.CrossDomainContext.
// This allows domain implementations to access outputs from dependency domains.
func OutputManifestToCrossDomainContext(manifest *OutputManifest) *domain.CrossDomainContext {
	if manifest == nil || len(manifest.Domains) == 0 {
		return nil
	}

	crossDomain := domain.NewCrossDomainContext()

	for domainName, domainOutput := range manifest.Domains {
		if domainOutput == nil {
			continue
		}

		domainOutputs := &domain.DomainOutputs{
			Resources: make(map[string]*domain.ResourceOutputs),
		}

		for resourceName, resourceOutput := range domainOutput.Resources {
			domainOutputs.Resources[resourceName] = &domain.ResourceOutputs{
				Type:    resourceOutput.Type,
				Outputs: resourceOutput.Outputs,
			}
		}

		crossDomain.AddDomainOutputs(domainName, domainOutputs)
	}

	return crossDomain
}

// CrossDomainContextToOutputManifest converts a domain.CrossDomainContext to an OutputManifest.
// This allows converting domain outputs back to the manifest format for persistence.
func CrossDomainContextToOutputManifest(crossDomain *domain.CrossDomainContext) *OutputManifest {
	if crossDomain == nil || len(crossDomain.Dependencies) == 0 {
		return nil
	}

	manifest := NewOutputManifest()

	for domainName, domainOutputs := range crossDomain.Dependencies {
		if domainOutputs == nil {
			continue
		}

		domainOutput := &DomainOutput{
			Resources: make(map[string]ResourceOutput),
		}

		for resourceName, resourceOutputs := range domainOutputs.Resources {
			if resourceOutputs == nil {
				continue
			}
			domainOutput.Resources[resourceName] = ResourceOutput{
				Type:    resourceOutputs.Type,
				Outputs: resourceOutputs.Outputs,
			}
		}

		manifest.AddDomainOutput(domainName, domainOutput)
	}

	return manifest
}
