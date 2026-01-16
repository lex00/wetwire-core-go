// run_scenario executes a wetwire scenario using Claude Code as the AI backend.
//
// Usage:
//
//	go run ./cmd/run_scenario [scenario_path] [persona] [output_dir] [flags]
//
// Flags:
//
//	--all      Run all personas
//	--record   Generate SVG recordings (requires termsvg)
//
// Examples:
//
//	go run ./cmd/run_scenario ./examples/aws_gitlab
//	go run ./cmd/run_scenario ./examples/aws_gitlab beginner
//	go run ./cmd/run_scenario ./examples/aws_gitlab expert ./results
//	go run ./cmd/run_scenario ./examples/aws_gitlab --all ./results
//	go run ./cmd/run_scenario ./examples/aws_gitlab --all --record ./results
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lex00/wetwire-core-go/scenario/runner"
)

func main() {
	ctx := context.Background()

	// Parse arguments
	scenarioPath := "./examples/aws_gitlab"
	personaName := ""
	outputDir := ""
	runAll := false
	generateRecordings := false

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--all" {
			runAll = true
		} else if arg == "--record" {
			generateRecordings = true
		} else if arg == "--help" || arg == "-h" {
			printUsage()
			return
		} else if !strings.HasPrefix(arg, "-") {
			if scenarioPath == "./examples/aws_gitlab" && i == 0 {
				scenarioPath = arg
			} else if personaName == "" && !strings.HasPrefix(arg, ".") && !strings.HasPrefix(arg, "/") {
				personaName = arg
			} else {
				outputDir = arg
			}
		}
	}

	// Default output dir
	if outputDir == "" {
		outputDir = scenarioPath + "/results"
	}

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║           Wetwire Scenario Runner (Claude Code)            ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	cfg := runner.Config{
		ScenarioPath:       scenarioPath,
		OutputDir:          outputDir,
		GenerateRecordings: generateRecordings,
	}

	if runAll {
		fmt.Printf("Running scenario with all %d personas...\n", len(runner.DefaultPersonas))
		fmt.Printf("Output directory: %s\n\n", outputDir)
	} else {
		if personaName == "" {
			personaName = "intermediate"
		}
		cfg.SinglePersona = personaName
		fmt.Printf("Persona: %s\n", personaName)
		fmt.Printf("Output:  %s\n\n", outputDir)
	}

	// Run scenarios
	results, err := runner.Run(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Print results
	if runAll {
		printSummary(results)
	} else if len(results) > 0 {
		r := results[0]
		status := "FAILED"
		if r.Success {
			status = "SUCCESS"
		}
		fmt.Printf("Status:   %s (%s)\n", status, r.Duration.Round(time.Millisecond))
		if r.Score != nil {
			fmt.Printf("Score:    %d/15 (%s)\n", r.Score.Total(), r.Score.Threshold())
		}
		if !r.Success {
			os.Exit(1)
		}
	}
}

func printUsage() {
	fmt.Println(`Usage: run_scenario [scenario_path] [persona] [output_dir] [flags]

Flags:
  --all      Run all personas (beginner, intermediate, expert, terse, verbose)
  --record   Generate SVG recordings (requires termsvg)
  --help     Show this help

Examples:
  run_scenario ./examples/aws_gitlab
  run_scenario ./examples/aws_gitlab beginner
  run_scenario ./examples/aws_gitlab expert ./results
  run_scenario ./examples/aws_gitlab --all ./results
  run_scenario ./examples/aws_gitlab --all --record ./results`)
}

func printSummary(results []runner.Result) {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                        SUMMARY                             ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	successCount := 0
	for _, r := range results {
		status := "✗ FAILED"
		if r.Success {
			status = "✓ SUCCESS"
			successCount++
		}
		scoreStr := ""
		if r.Score != nil {
			scoreStr = fmt.Sprintf(" [%d/15]", r.Score.Total())
		}
		fmt.Printf("  %-12s %s%s  (%s)\n", r.Persona, status, scoreStr, r.Duration.Round(time.Millisecond))
	}

	fmt.Println()
	fmt.Printf("Results: %d/%d passed\n", successCount, len(results))
}
