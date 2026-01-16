// Example demonstrating the Kiro provider for wetwire-core-go
//
// This example shows how to:
// 1. Create a Kiro provider instance
// 2. Send a message request
// 3. Handle the response
//
// Prerequisites:
// - kiro-cli must be installed and available in PATH
// - An MCP server command must be configured

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/lex00/wetwire-core-go/providers"
	"github.com/lex00/wetwire-core-go/providers/kiro"
)

func main() {
	ctx := context.Background()

	// Example 1: Create a Kiro provider with configuration
	fmt.Println("=== Example 1: Basic Kiro Provider ===")
	provider1, err := kiro.New(kiro.Config{
		AgentName: "wetwire-example-agent",
		AgentPrompt: `You are a helpful infrastructure code assistant.
Your job is to help users generate infrastructure code.`,
		MCPCommand: "wetwire-aws",   // The MCP server binary name
		MCPArgs:    []string{"mcp"}, // Subcommand to start MCP server
		WorkDir:    ".",
	})
	if err != nil {
		log.Fatalf("Failed to create Kiro provider: %v", err)
	}

	fmt.Printf("Created provider: %s\n", provider1.Name())

	// Example 2: Send a message (non-interactive mode)
	fmt.Println("\n=== Example 2: Send Message ===")

	req := providers.MessageRequest{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 4096,
		Messages: []providers.Message{
			providers.NewUserMessage("What types of AWS resources can you create?"),
		},
	}

	// Note: This will actually call kiro-cli if available
	// For this example to work, you need kiro-cli installed
	if !kiroAvailable() {
		fmt.Println("kiro-cli not found in PATH - skipping actual execution")
		fmt.Println("This example demonstrates the API structure")
		return
	}

	resp, err := provider1.CreateMessage(ctx, req)
	if err != nil {
		log.Fatalf("Failed to create message: %v", err)
	}

	fmt.Println("Response:")
	for _, block := range resp.Content {
		if block.Type == "text" {
			fmt.Println(block.Text)
		}
	}

	// Example 3: Streaming response
	fmt.Println("\n=== Example 3: Streaming Response ===")

	streamReq := providers.MessageRequest{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 1024,
		Messages: []providers.Message{
			providers.NewUserMessage("Hello!"),
		},
	}

	// Define a stream handler to process text chunks
	streamHandler := func(text string) {
		fmt.Print(text)
	}

	streamResp, err := provider1.StreamMessage(ctx, streamReq, streamHandler)
	if err != nil {
		log.Fatalf("Failed to stream message: %v", err)
	}

	fmt.Printf("\nStop reason: %s\n", streamResp.StopReason)

	// Example 4: Provider switching pattern
	fmt.Println("\n=== Example 4: Provider Switching ===")
	demonstrateProviderSwitching()
}

// kiroAvailable checks if kiro-cli is in PATH
func kiroAvailable() bool {
	_, err := exec.LookPath("kiro-cli")
	if err == nil {
		return true
	}
	_, err = exec.LookPath("kiro")
	return err == nil
}

// demonstrateProviderSwitching shows how to switch between providers
func demonstrateProviderSwitching() {
	useKiro := os.Getenv("USE_KIRO") == "true"

	var provider providers.Provider
	var err error

	if useKiro {
		fmt.Println("Using Kiro provider (Amazon Q Developer CLI backend)")
		provider, err = kiro.New(kiro.Config{
			AgentName:   "wetwire-agent",
			AgentPrompt: "You are a helpful assistant.",
			MCPCommand:  "wetwire-aws",
			MCPArgs:     []string{"mcp"},
		})
	} else {
		fmt.Println("Using Anthropic provider (Direct API backend)")
		// Note: This requires anthropic package import
		// provider, err = anthropic.New(anthropic.Config{})
		fmt.Println("(Anthropic provider example omitted - see anthropic package docs)")
		return
	}

	if err != nil {
		log.Printf("Failed to create provider: %v", err)
		return
	}

	fmt.Printf("Created provider: %s\n", provider.Name())
	fmt.Println("Same code works with either provider!")
}

// Output example (when kiro-cli is available):
//
// === Example 1: Basic Kiro Provider ===
// Created provider: kiro
//
// === Example 2: Send Message ===
// Response:
// I can help you create various AWS resources including:
// - S3 buckets
// - Lambda functions
// - DynamoDB tables
// - IAM roles and policies
// - And many more CloudFormation resources
//
// === Example 3: Streaming Response ===
// Hello! How can I assist you today?
// Stop reason: end_turn
//
// === Example 4: Provider Switching ===
// Using Kiro provider (Claude Code backend)
// Created provider: kiro
// Same code works with either provider!
