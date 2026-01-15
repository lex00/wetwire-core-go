// Package scenario provides a Claude Code skill for executing multi-domain scenarios.
package scenario

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	scenarioPkg "github.com/lex00/wetwire-core-go/scenario"
)

// Skill implements the /scenario skill for Claude Code.
type Skill struct {
	output io.Writer
}

// New creates a new scenario skill.
func New() *Skill {
	return &Skill{
		output: os.Stdout,
	}
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

	// Generate and output instructions
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
