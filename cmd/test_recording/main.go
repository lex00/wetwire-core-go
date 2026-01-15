package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lex00/wetwire-core-go/scenario"
	skillscenario "github.com/lex00/wetwire-core-go/skills/scenario"
)

func main() {
	scenarioDir := "/tmp/wetwire-core-go/examples/aws_gitlab"
	outputDir := "/tmp/recordings"

	os.MkdirAll(outputDir, 0755)

	// Load the actual prompt from the scenario
	promptPath := filepath.Join(scenarioDir, "prompt.md")
	promptContent, err := os.ReadFile(promptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading prompt: %v\n", err)
		os.Exit(1)
	}

	err = scenario.RunWithRecording("aws_gitlab_demo", scenario.RecordOptions{
		Enabled:       true,
		OutputDir:     outputDir,
		TermWidth:     80,
		TermHeight:    40, // Taller to fit more content
		LineDelay:     300 * time.Millisecond,
		TypingSpeed:   30 * time.Millisecond, // Faster typing
		ResponseDelay: 500 * time.Millisecond,
		UserPrompt:    string(promptContent), // The actual scenario prompt
	}, func() error {
		skill := skillscenario.New()
		skill.SetOutput(os.Stdout)
		return skill.Run(nil, scenarioDir)
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n\n=== SVG saved to: %s/aws_gitlab_demo.svg ===\n", outputDir)
}
