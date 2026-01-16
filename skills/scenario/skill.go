// Package scenario provides a Claude Code skill for executing multi-domain scenarios.
package scenario

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lex00/wetwire-core-go/agent/results"
	"github.com/lex00/wetwire-core-go/mcp"
	"github.com/lex00/wetwire-core-go/providers"
	scenarioPkg "github.com/lex00/wetwire-core-go/scenario"
)

// Skill implements the /scenario skill for Claude Code.
type Skill struct {
	output    io.Writer
	provider  providers.Provider
	mcpServer *mcp.Server
	outputDir string // Directory for results output
	persona   string // Persona name for results tracking
}

// New creates a new scenario skill.
// Provider and MCPServer can be nil for backward compatibility (instruction generation only).
func New(provider providers.Provider, mcpServer *mcp.Server) *Skill {
	return &Skill{
		output:    os.Stdout,
		provider:  provider,
		mcpServer: mcpServer,
		outputDir: "./output",
	}
}

// SetOutputDir sets the output directory for results.
func (s *Skill) SetOutputDir(dir string) {
	s.outputDir = dir
}

// SetPersona sets the persona name for results tracking.
func (s *Skill) SetPersona(persona string) {
	s.persona = persona
}

// SetOutput sets the output writer for the skill.
func (s *Skill) SetOutput(w io.Writer) {
	s.output = w
}

// Name returns the skill name.
func (s *Skill) Name() string {
	return "scenario"
}

// Description returns the skill description.
func (s *Skill) Description() string {
	return "Execute multi-domain scenario from scenario.yaml"
}

// Run executes the scenario skill.
// If args is empty, it looks for scenario.yaml in the current directory.
// If args is a directory, it looks for scenario.yaml in that directory.
// If args is a file path, it loads that file directly.
//
// If Provider and MCPServer are set, the scenario will be executed autonomously.
// Otherwise, instructions will be generated and output.
func (s *Skill) Run(ctx context.Context, args string) error {
	// Find scenario path
	scenarioPath, err := s.findScenario(args)
	if err != nil {
		return err
	}

	// Load scenario
	config, err := scenarioPkg.LoadFile(scenarioPath)
	if err != nil {
		return fmt.Errorf("failed to load scenario: %w", err)
	}

	// Validate scenario
	result := scenarioPkg.Validate(config)
	if !result.IsValid() {
		return fmt.Errorf("invalid scenario: %s", result.Error())
	}

	// Load prompt if specified
	var prompt string
	if config.Prompts != nil && config.Prompts.Default != "" {
		prompt, err = s.loadPrompt(scenarioPath, config.Prompts.Default)
		if err != nil {
			// Non-fatal: continue without prompt
			prompt = ""
		}
	}

	// If provider and MCP server are available, execute the scenario
	if s.provider != nil && s.mcpServer != nil {
		return s.executeScenario(ctx, config, prompt)
	}

	// Otherwise, just generate and output instructions
	return GenerateInstructions(s.output, config, prompt, nil)
}

// findScenario locates the scenario.yaml file based on args.
func (s *Skill) findScenario(args string) (string, error) {
	if args == "" {
		// Use current directory
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		args = cwd
	}

	// Check if args is a file
	info, err := os.Stat(args)
	if err != nil {
		return "", fmt.Errorf("scenario not found: %w", err)
	}

	if !info.IsDir() {
		// It's a file, use it directly
		return args, nil
	}

	// It's a directory, look for scenario.yaml
	scenarioPath := filepath.Join(args, "scenario.yaml")
	if _, err := os.Stat(scenarioPath); err != nil {
		return "", fmt.Errorf("scenario.yaml not found in %s: %w", args, err)
	}

	return scenarioPath, nil
}

// loadPrompt loads a prompt file relative to the scenario file.
func (s *Skill) loadPrompt(scenarioPath, promptFile string) (string, error) {
	scenarioDir := filepath.Dir(scenarioPath)
	promptPath := filepath.Join(scenarioDir, promptFile)

	data, err := os.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file: %w", err)
	}

	return string(data), nil
}

// executeScenario runs the scenario using an autonomous agent with MCP tools.
func (s *Skill) executeScenario(ctx context.Context, config *scenarioPkg.ScenarioConfig, prompt string) error {
	// Build the full prompt with scenario instructions
	fullPrompt, err := s.buildExecutionPrompt(config, prompt)
	if err != nil {
		return fmt.Errorf("failed to build execution prompt: %w", err)
	}

	// Create session for results tracking
	persona := s.persona
	if persona == "" {
		persona = "default"
	}
	session := results.NewSession(persona, config.Name)

	// Create an agent that uses MCP tools
	agent := NewScenarioAgent(ScenarioAgentConfig{
		Provider:  s.provider,
		MCPServer: s.mcpServer,
		Output:    s.output,
		Session:   session,
	})

	// Run the agent autonomously (no developer interaction)
	if err := agent.Run(ctx, fullPrompt); err != nil {
		return fmt.Errorf("scenario execution failed: %w", err)
	}

	// Mark session complete
	session.Complete()

	// Validate results against scenario requirements
	if err := s.validateResults(config); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Write results
	writer := results.NewWriter(s.outputDir)
	if err := writer.Write(session); err != nil {
		fmt.Fprintf(s.output, "Warning: failed to write results: %v\n", err)
	} else {
		fmt.Fprintf(s.output, "\nResults written to %s/%s/RESULTS.md\n", s.outputDir, persona)
	}

	fmt.Fprintln(s.output, "Scenario completed successfully!")
	return nil
}

// buildExecutionPrompt builds a complete prompt for scenario execution.
func (s *Skill) buildExecutionPrompt(config *scenarioPkg.ScenarioConfig, userPrompt string) (string, error) {
	var promptBuilder strings.Builder

	// Add scenario header
	promptBuilder.WriteString(fmt.Sprintf("# Multi-Domain Scenario: %s\n\n", config.Name))
	if config.Description != "" {
		promptBuilder.WriteString(fmt.Sprintf("%s\n\n", config.Description))
	}

	// Add user's prompt if provided
	if userPrompt != "" {
		promptBuilder.WriteString("## Requirements\n\n")
		promptBuilder.WriteString(userPrompt)
		promptBuilder.WriteString("\n\n")
	}

	// Add domain execution instructions
	order, err := scenarioPkg.GetDomainOrder(config)
	if err != nil {
		return "", fmt.Errorf("failed to determine domain order: %w", err)
	}

	promptBuilder.WriteString("## Execution Plan\n\n")
	promptBuilder.WriteString("Execute the following domains in order:\n\n")

	for i, domainName := range order {
		domain := config.GetDomain(domainName)
		if domain == nil {
			continue
		}

		promptBuilder.WriteString(fmt.Sprintf("%d. **%s** (using %s)\n", i+1, domain.Name, domain.CLI))

		if len(domain.DependsOn) > 0 {
			promptBuilder.WriteString(fmt.Sprintf("   - Depends on: %s\n", strings.Join(domain.DependsOn, ", ")))
		}

		if len(domain.MCPTools) > 0 {
			promptBuilder.WriteString("   - Available MCP tools:\n")
			for purpose, toolName := range domain.MCPTools {
				promptBuilder.WriteString(fmt.Sprintf("     - `%s` for %s\n", toolName, purpose))
			}
		}

		if len(domain.Outputs) > 0 {
			promptBuilder.WriteString(fmt.Sprintf("   - Expected outputs: %s\n", strings.Join(domain.Outputs, ", ")))
		}

		promptBuilder.WriteString("\n")
	}

	// Add cross-domain requirements
	if len(config.CrossDomain) > 0 {
		promptBuilder.WriteString("## Cross-Domain Integration\n\n")
		for _, cd := range config.CrossDomain {
			promptBuilder.WriteString(fmt.Sprintf("- %s → %s (%s)\n", cd.From, cd.To, cd.Type))
			if len(cd.Validation.RequiredRefs) > 0 {
				promptBuilder.WriteString(fmt.Sprintf("  - Required refs: %s\n", strings.Join(cd.Validation.RequiredRefs, ", ")))
			}
		}
		promptBuilder.WriteString("\n")
	}

	// Add validation criteria
	if len(config.Validation) > 0 {
		promptBuilder.WriteString("## Validation Criteria\n\n")
		for domainName, rules := range config.Validation {
			promptBuilder.WriteString(fmt.Sprintf("**%s:**\n", domainName))
			if rules.Stacks != nil {
				promptBuilder.WriteString(fmt.Sprintf("- Stacks: min %d", rules.Stacks.Min))
				if rules.Stacks.Max > 0 {
					promptBuilder.WriteString(fmt.Sprintf(", max %d", rules.Stacks.Max))
				}
				promptBuilder.WriteString("\n")
			}
			if rules.Pipelines != nil {
				promptBuilder.WriteString(fmt.Sprintf("- Pipelines: min %d", rules.Pipelines.Min))
				if rules.Pipelines.Max > 0 {
					promptBuilder.WriteString(fmt.Sprintf(", max %d", rules.Pipelines.Max))
				}
				promptBuilder.WriteString("\n")
			}
			if rules.Workflows != nil {
				promptBuilder.WriteString(fmt.Sprintf("- Workflows: min %d", rules.Workflows.Min))
				if rules.Workflows.Max > 0 {
					promptBuilder.WriteString(fmt.Sprintf(", max %d", rules.Workflows.Max))
				}
				promptBuilder.WriteString("\n")
			}
			if rules.Manifests != nil {
				promptBuilder.WriteString(fmt.Sprintf("- Manifests: min %d", rules.Manifests.Min))
				if rules.Manifests.Max > 0 {
					promptBuilder.WriteString(fmt.Sprintf(", max %d", rules.Manifests.Max))
				}
				promptBuilder.WriteString("\n")
			}
			if rules.Resources != nil {
				promptBuilder.WriteString(fmt.Sprintf("- Resources: min %d", rules.Resources.Min))
				if rules.Resources.Max > 0 {
					promptBuilder.WriteString(fmt.Sprintf(", max %d", rules.Resources.Max))
				}
				promptBuilder.WriteString("\n")
			}
		}
		promptBuilder.WriteString("\n")
	}

	promptBuilder.WriteString("Please execute this scenario autonomously, using the MCP tools available to you.\n")

	return promptBuilder.String(), nil
}

// validateResults validates the scenario execution results against requirements.
func (s *Skill) validateResults(config *scenarioPkg.ScenarioConfig) error {
	// TODO: Implement validation logic
	// This would check:
	// 1. All expected output files exist
	// 2. Validation criteria are met (min/max counts)
	// 3. Cross-domain references are valid

	// For now, just return nil (validation passed)
	// Full implementation would require domain-specific validation logic
	return nil
}
