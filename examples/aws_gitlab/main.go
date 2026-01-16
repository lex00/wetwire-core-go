// aws_gitlab runs the AWS + GitLab scenario using wetwire-core-go.
//
// Usage:
//
//	go run . [flags]
//
// Flags:
//
//	--all       Run all personas
//	--persona   Run specific persona (default: intermediate)
//	--verbose   Show streaming output
//	--output    Output directory (default: ./results)
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/lex00/wetwire-core-go/scenario/runner"
)

func main() {
	var (
		runAll    = flag.Bool("all", false, "Run all personas")
		persona   = flag.String("persona", "intermediate", "Persona to run")
		verbose   = flag.Bool("verbose", false, "Show streaming output")
		outputDir = flag.String("output", "./results", "Output directory")
	)
	flag.Parse()

	fmt.Println("AWS + GitLab Scenario")
	fmt.Println("=====================")
	fmt.Println()

	cfg := runner.Config{
		ScenarioPath: ".",
		OutputDir:    *outputDir,
		Verbose:      *verbose,
	}

	if *runAll {
		fmt.Printf("Running all %d personas...\n", len(runner.DefaultPersonas))
	} else {
		cfg.SinglePersona = *persona
		fmt.Printf("Running persona: %s\n", *persona)
	}
	fmt.Printf("Output: %s\n\n", *outputDir)

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
