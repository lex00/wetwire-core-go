// aws_gitlab runs the AWS + GitLab scenario using wetwire-core-go.
//
// Usage:
//
//	go run . [flags]
//
// Flags:
//
//	--all         Run all personas
//	--persona     Run specific persona (default: intermediate)
//	--verbose     Show streaming output
//	--output      Output directory (default: ./results)
//	--domain-mode Use domain MCP tools (requires ANTHROPIC_API_KEY)
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lex00/wetwire-core-go/scenario"
	"github.com/lex00/wetwire-core-go/scenario/runner"
)

func main() {
	var (
		runAll     = flag.Bool("all", false, "Run all personas")
		persona    = flag.String("persona", "intermediate", "Persona to run")
		verbose    = flag.Bool("verbose", false, "Show streaming output")
		outputDir  = flag.String("output", "./results", "Output directory")
		domainMode = flag.Bool("domain-mode", false, "Use domain MCP tools (requires ANTHROPIC_API_KEY)")
		debug      = flag.Bool("debug", false, "Enable MCP debug logging")
	)
	flag.Parse()

	fmt.Println("AWS + GitLab Scenario")
	fmt.Println("=====================")
	fmt.Println()

	if *domainMode {
		runDomainMode(*persona, *verbose, *outputDir, *debug)
	} else {
		runClaudeMode(*runAll, *persona, *verbose, *outputDir)
	}
}

// runClaudeMode runs using Claude Code CLI (original behavior).
func runClaudeMode(runAll bool, persona string, verbose bool, outputDir string) {
	cfg := runner.Config{
		ScenarioPath: ".",
		OutputDir:    outputDir,
		Verbose:      verbose,
	}

	if runAll {
		fmt.Printf("Running all %d personas...\n", len(runner.DefaultPersonas))
	} else {
		cfg.SinglePersona = persona
		fmt.Printf("Running persona: %s\n", persona)
	}
	fmt.Printf("Output: %s\n\n", outputDir)

	ctx := context.Background()
	results, err := runner.Run(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Print summary
	fmt.Println()
	fmt.Println("Results")
	fmt.Println("-------")
	for _, r := range results {
		status := "FAILED"
		if r.Success {
			status = "SUCCESS"
		}
		scoreStr := ""
		if r.Score != nil {
			scoreStr = fmt.Sprintf(" [%d/12]", r.Score.Total())
		}
		fmt.Printf("  %-12s %s%s (%s)\n", r.Persona, status, scoreStr, r.Duration.Round(time.Millisecond))
	}
}

// runDomainMode runs using domain MCP tools via Anthropic API.
func runDomainMode(persona string, verbose bool, outputDir string, debug bool) {
	fmt.Println("Mode: Domain MCP Tools")
	fmt.Printf("Persona: %s\n", persona)
	fmt.Printf("Output: %s\n\n", outputDir)

	// Check for API key
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		fmt.Fprintln(os.Stderr, "Error: ANTHROPIC_API_KEY environment variable required for domain mode")
		os.Exit(1)
	}

	ctx := context.Background()

	// Load scenario config
	config, err := scenario.Load(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading scenario: %v\n", err)
		os.Exit(1)
	}

	// Create output directory for this persona
	personaDir := filepath.Join(outputDir, persona)
	if err := os.MkdirAll(personaDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Load user prompt for persona
	prompt := loadPrompt(persona)

	fmt.Printf("Starting MCP servers for domains: %s\n", strings.Join(config.DomainNames(), ", "))

	// Create domain runner
	start := time.Now()
	domainRunner, err := runner.NewDomainRunner(ctx, runner.DomainRunnerConfig{
		ScenarioConfig: config,
		WorkDir:        personaDir,
		Output:         os.Stdout,
		Verbose:        verbose,
		Debug:          debug,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating domain runner: %v\n", err)
		os.Exit(1)
	}
	defer domainRunner.Close()

	fmt.Println("MCP servers started successfully")
	fmt.Println()

	// Run the scenario
	result, err := domainRunner.Run(ctx, prompt)
	duration := time.Since(start)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running scenario: %v\n", err)
		os.Exit(1)
	}

	// Print results
	fmt.Println()
	fmt.Println("Results")
	fmt.Println("-------")

	status := "FAILED"
	if result.Success {
		status = "SUCCESS"
	}
	fmt.Printf("  %-12s %s (%s)\n", persona, status, duration.Round(time.Millisecond))

	// Print tool call summary
	if len(result.ToolCalls) > 0 {
		fmt.Println()
		fmt.Println("Tool Calls")
		fmt.Println("----------")
		for _, tc := range result.ToolCalls {
			errStr := ""
			if tc.IsError {
				errStr = " [ERROR]"
			}
			fmt.Printf("  %s.%s%s\n", tc.Domain, tc.Tool, errStr)
		}
	}
}

// loadPrompt loads the prompt for the given persona.
func loadPrompt(persona string) string {
	// Try persona-specific prompt first
	promptPath := filepath.Join("prompts", persona+".md")
	if content, err := os.ReadFile(promptPath); err == nil {
		return stripTitle(string(content))
	}

	// Fall back to default prompt
	if content, err := os.ReadFile("prompt.md"); err == nil {
		return stripTitle(string(content))
	}

	return "Create the required infrastructure files."
}

// stripTitle removes the markdown title line from prompt content.
func stripTitle(content string) string {
	lines := strings.Split(content, "\n")
	startIdx := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		startIdx = i
		break
	}
	return strings.TrimSpace(strings.Join(lines[startIdx:], "\n"))
}
