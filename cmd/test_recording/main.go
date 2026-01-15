package main

import (
	"fmt"
	"os"
	"time"

	"github.com/lex00/wetwire-core-go/scenario"
	skillscenario "github.com/lex00/wetwire-core-go/skills/scenario"
)

func main() {
	scenarioDir := "/tmp/wetwire-core-go/examples/aws_gitlab"
	outputDir := "/tmp/recordings"

	os.MkdirAll(outputDir, 0755)

	err := scenario.RunWithRecording("aws_gitlab_demo", scenario.RecordOptions{
		Enabled:       true,
		OutputDir:     outputDir,
		TermWidth:     80,
		TermHeight:    24,
		LineDelay:     300 * time.Millisecond,
		TypingSpeed:   50 * time.Millisecond,
		ResponseDelay: 500 * time.Millisecond,
		// Use defaults for AgentGreeting, UserPrompt, AgentResponse
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
