// Example demonstrating the Claude Code provider for wetwire-core-go
//
// This example shows how to:
// 1. Create a Claude Code provider instance
// 2. Configure MCP servers for tool access
// 3. Send messages and receive responses
//
// Prerequisites:
// - claude CLI must be installed and available in PATH
// - For MCP tools, an MCP server binary must be available

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/lex00/wetwire-core-go/providers"
	"github.com/lex00/wetwire-core-go/providers/claude"
)

func main() {
	ctx := context.Background()

	// Example 1: Basic Claude provider
	fmt.Println("=== Example 1: Basic Claude Provider ===")
	provider1, err := claude.New(claude.Config{
		SystemPrompt: "You are a helpful assistant. Be concise.",
	})
	if err != nil {
		log.Fatalf("Failed to create Claude provider: %v", err)
	}
	fmt.Printf("Created provider: %s\n", provider1.Name())

	// Check if claude is available
	if !claude.Available() {
		fmt.Println("claude CLI not found in PATH - skipping actual execution")
		fmt.Println("This example demonstrates the API structure")
		demonstrateMCPConfig()
		return
	}

	// Example 2: Send a simple message
	fmt.Println("\n=== Example 2: Send Message ===")
	req := providers.MessageRequest{
		Messages: []providers.Message{
			providers.NewUserMessage("What is 2+2? Reply with just the number."),
		},
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
		System: "Be very brief.",
		Messages: []providers.Message{
			providers.NewUserMessage("Say hello in one word."),
		},
	}

	fmt.Print("Streaming: ")
	streamResp, err := provider1.StreamMessage(ctx, streamReq, func(text string) {
		fmt.Print(text)
	})
	if err != nil {
		log.Fatalf("Failed to stream message: %v", err)
	}
	fmt.Printf("\nStop reason: %s\n", streamResp.StopReason)

	// Example 4: With MCP tools
	fmt.Println("\n=== Example 4: With MCP Configuration ===")
	demonstrateMCPConfig()

	// Example 5: Provider switching pattern
	fmt.Println("\n=== Example 5: Provider Switching ===")
	demonstrateProviderSwitching()
}

// demonstrateMCPConfig shows how to write an MCP config file
func demonstrateMCPConfig() {
	tmpDir := os.TempDir()
	mcpConfigPath := filepath.Join(tmpDir, "wetwire-mcp.json")

	// Write MCP config for a domain MCP server
	err := claude.WriteMCPConfig(mcpConfigPath, map[string]claude.MCPServerConfig{
		"wetwire-aws": {
			Command: "wetwire-aws-mcp",
			Args:    []string{},
			Cwd:     ".",
		},
	})
	if err != nil {
		fmt.Printf("Failed to write MCP config: %v\n", err)
		return
	}

	fmt.Printf("MCP config written to: %s\n", mcpConfigPath)

	// Create provider with MCP config
	_, err = claude.New(claude.Config{
		MCPConfigPath: mcpConfigPath,
		SystemPrompt: `You are an infrastructure code generator.
Use the wetwire-aws MCP tools to create AWS resources.`,
		PermissionMode: "acceptEdits",
	})
	if err != nil {
		fmt.Printf("Failed to create provider: %v\n", err)
		return
	}

	fmt.Println("Provider created with MCP tools")
}

// demonstrateProviderSwitching shows how to switch between providers
func demonstrateProviderSwitching() {
	useClaudeCode := os.Getenv("USE_CLAUDE_CODE") == "true"

	var provider providers.Provider
	var err error

	if useClaudeCode {
		fmt.Println("Using Claude Code provider (no API key needed)")
		provider, err = claude.New(claude.Config{
			SystemPrompt: "You are a helpful assistant.",
		})
	} else {
		fmt.Println("Using Claude Code provider by default")
		provider, err = claude.New(claude.Config{
			SystemPrompt: "You are a helpful assistant.",
		})
	}

	if err != nil {
		log.Printf("Failed to create provider: %v", err)
		return
	}

	fmt.Printf("Created provider: %s\n", provider.Name())
	fmt.Println("Same interface works with Anthropic API or Claude Code!")
}

// Output example (when claude CLI is available):
//
// === Example 1: Basic Claude Provider ===
// Created provider: claude
//
// === Example 2: Send Message ===
// Response:
// 4
//
// === Example 3: Streaming Response ===
// Streaming: Hello
// Stop reason: end_turn
//
// === Example 4: With MCP Configuration ===
// MCP config written to: /tmp/wetwire-mcp.json
// Provider created with MCP tools
//
// === Example 5: Provider Switching ===
// Using Claude Code provider by default
// Created provider: claude
// Same interface works with Anthropic API or Claude Code!
