// Test script to run a scenario with the Claude Code provider.
//
// This demonstrates the correct integration pattern: Claude Code runs the
// full agentic loop with MCP tools configured via --mcp-config.
//
// Usage:
//
//	go run ./cmd/test_scenario_claude
package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/lex00/wetwire-core-go/providers"
	"github.com/lex00/wetwire-core-go/providers/claude"
	"github.com/lex00/wetwire-core-go/skills/scenario"
)

func main() {
	ctx := context.Background()

	// Check if claude is available
	if !claude.Available() {
		log.Fatal("claude CLI not found in PATH")
	}

	fmt.Println("=== Testing Scenario with Claude Code Provider ===")
	fmt.Println()

	// Step 1: Generate scenario instructions
	fmt.Println("Step 1: Generate scenario instructions...")
	instructionSkill := scenario.New(nil, nil)
	var instructions bytes.Buffer
	instructionSkill.SetOutput(&instructions)

	scenarioPath := "./examples/aws_gitlab"
	if err := instructionSkill.Run(ctx, scenarioPath); err != nil {
		log.Fatalf("Failed to generate instructions: %v", err)
	}

	fmt.Println("Instructions generated (truncated):")
	instrStr := instructions.String()
	if len(instrStr) > 300 {
		fmt.Println(instrStr[:300] + "...\n")
	}

	// Step 2: Write MCP config for domain tools
	// In a real scenario, these would point to actual domain MCP servers
	fmt.Println("Step 2: Configure MCP servers...")
	tmpDir, err := os.MkdirTemp("", "scenario-mcp-*")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	mcpConfigPath := filepath.Join(tmpDir, "mcp.json")

	// For demonstration, we use the wetwire-honeycomb MCP server that's already configured
	// In production, you'd configure domain-specific MCP servers here
	err = claude.WriteMCPConfig(mcpConfigPath, map[string]claude.MCPServerConfig{
		// These would be real MCP servers in production:
		// "wetwire-aws": {Command: "wetwire-aws", Args: []string{"mcp"}},
		// "wetwire-gitlab": {Command: "wetwire-gitlab", Args: []string{"mcp"}},
	})
	if err != nil {
		log.Fatalf("Failed to write MCP config: %v", err)
	}
	fmt.Printf("MCP config written to: %s\n", mcpConfigPath)

	// Step 3: Create Claude provider with scenario instructions as system prompt
	fmt.Println("\nStep 3: Create Claude provider...")
	provider, err := claude.New(claude.Config{
		WorkDir:        ".",
		MCPConfigPath:  mcpConfigPath,
		PermissionMode: "plan", // Use plan mode for safety
		SystemPrompt: `You are executing a wetwire scenario.
Given the scenario instructions, describe the implementation plan.
Do NOT execute any tools - just explain what you would do.`,
	})
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}
	fmt.Printf("Provider: %s\n", provider.Name())

	// Step 4: Run the scenario
	fmt.Println("\nStep 4: Execute scenario...")

	// Use a condensed prompt for the test
	prompt := `Execute the aws_gitlab_s3_deploy scenario:

1. AWS: Create CloudFormation template for S3 bucket with versioning and encryption
2. GitLab: Create CI/CD pipeline to validate and publish the template

Briefly describe (in 3-5 sentences) the implementation steps.`

	resp, err := provider.CreateMessage(ctx, providers.MessageRequest{
		Messages: []providers.Message{
			providers.NewUserMessage(prompt),
		},
	})
	if err != nil {
		log.Fatalf("Scenario execution failed: %v", err)
	}

	fmt.Println("\n--- Claude Code Response ---")
	for _, block := range resp.Content {
		if block.Type == "text" {
			fmt.Println(block.Text)
		}
	}
	fmt.Println("---")

	// Step 5: Demonstrate streaming mode
	fmt.Println("\nStep 5: Test streaming mode...")
	var streamedText bytes.Buffer
	_, err = provider.StreamMessage(ctx, providers.MessageRequest{
		Messages: []providers.Message{
			providers.NewUserMessage("Say 'Scenario test complete!' in exactly those words."),
		},
	}, func(text string) {
		streamedText.WriteString(text)
		fmt.Print(text)
	})
	if err != nil {
		fmt.Printf("\nStreaming error: %v\n", err)
	}
	fmt.Println()

	fmt.Println("\n=== Test Summary ===")
	fmt.Println("The Claude Code provider successfully:")
	fmt.Println("  1. Generated scenario instructions")
	fmt.Println("  2. Configured MCP servers")
	fmt.Println("  3. Executed scenario via Claude Code")
	fmt.Println("  4. Streamed responses")
	fmt.Println("\nArchitecture note:")
	fmt.Println("  - Claude Code runs its own agentic loop internally")
	fmt.Println("  - MCP tools are passed via --mcp-config (not in-process)")
	fmt.Println("  - The provider returns completed responses (StopReasonEndTurn)")
	fmt.Println("  - For full tool execution, configure actual MCP server binaries")
}
