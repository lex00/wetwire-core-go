// Example demonstrating the unified Agent architecture
//
// This example shows how to:
// 1. Create an MCP server with tools
// 2. Wrap it with MCPServerAdapter
// 3. Create a unified Agent
// 4. Run the agent in autonomous or interactive mode
//
// This is the recommended pattern for all wetwire domain packages.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/lex00/wetwire-core-go/agent/agents"
	"github.com/lex00/wetwire-core-go/mcp"
	"github.com/lex00/wetwire-core-go/providers/anthropic"
)

func main() {
	// Example 1: Autonomous Agent (no human interaction)
	fmt.Println("=== Example 1: Autonomous Agent ===")
	autonomousExample()

	// Example 2: Interactive Agent (with developer)
	fmt.Println("\n=== Example 2: Interactive Agent ===")
	interactiveExample()
}

// autonomousExample shows how to run an agent without human interaction.
// This is useful for scenarios, automated testing, and batch processing.
func autonomousExample() {
	// Step 1: Create MCP server with tools
	mcpServer := mcp.NewServer(mcp.Config{
		Name:    "example-domain",
		Version: "1.0.0",
	})

	// Step 2: Register tools
	// Option A: Register standard tools with handlers
	mcp.RegisterStandardToolsWithDefaults(mcpServer, "example", mcp.StandardToolHandlers{
		Init: func(ctx context.Context, args map[string]any) (string, error) {
			name, _ := args["name"].(string)
			return fmt.Sprintf("Initialized project: %s", name), nil
		},
		Build: func(ctx context.Context, args map[string]any) (string, error) {
			pkg, _ := args["package"].(string)
			return fmt.Sprintf("Built package: %s", pkg), nil
		},
		Lint: func(ctx context.Context, args map[string]any) (string, error) {
			pkg, _ := args["package"].(string)
			return fmt.Sprintf("Linted %s: no issues", pkg), nil
		},
	})

	// Option B: Register custom tools individually
	mcpServer.RegisterToolWithSchema(
		"custom_tool",
		"A custom domain-specific tool",
		func(ctx context.Context, args map[string]any) (string, error) {
			return "Custom tool executed", nil
		},
		map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	)

	// Step 3: Create provider
	provider, err := anthropic.New(anthropic.Config{
		// APIKey defaults to ANTHROPIC_API_KEY env var
	})
	if err != nil {
		log.Printf("Failed to create provider: %v", err)
		log.Println("(Set ANTHROPIC_API_KEY to run this example)")
		return
	}

	// Step 4: Create unified Agent
	agent, err := agents.NewAgent(agents.AgentConfig{
		Provider:  provider,
		MCPServer: agents.NewMCPServerAdapter(mcpServer),
		SystemPrompt: `You are an infrastructure code generator.
When asked to create something, use the available tools.
Work autonomously without asking questions.`,
		// Developer: nil means autonomous mode
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Step 5: Run the agent
	ctx := context.Background()
	err = agent.Run(ctx, "Initialize a project called 'my-app' and build it")
	if err != nil {
		log.Printf("Agent error: %v", err)
	}

	fmt.Println("Autonomous agent completed")
}

// interactiveExample shows how to run an agent with human interaction.
// This is useful for design mode where the agent can ask clarifying questions.
func interactiveExample() {
	// Create MCP server (same as autonomous)
	mcpServer := mcp.NewServer(mcp.Config{
		Name: "example-domain",
	})

	mcp.RegisterStandardToolsWithDefaults(mcpServer, "example", mcp.StandardToolHandlers{
		Init: func(ctx context.Context, args map[string]any) (string, error) {
			return "Project initialized", nil
		},
	})

	// Create provider
	provider, err := anthropic.New(anthropic.Config{})
	if err != nil {
		log.Printf("Failed to create provider: %v", err)
		return
	}

	// Create a developer that provides answers
	// In real usage, this would read from stdin or another source
	developer := &ExampleDeveloper{
		answers: []string{
			"production",
			"us-west-2",
			"yes",
		},
	}

	// Create agent with developer for interactive mode
	agent, err := agents.NewAgent(agents.AgentConfig{
		Provider:  provider,
		MCPServer: agents.NewMCPServerAdapter(mcpServer),
		Developer: developer, // Enables interactive mode
		SystemPrompt: `You are an infrastructure designer.
Ask clarifying questions using ask_developer when requirements are unclear.
Available questions: What environment? What region? Confirm creation?`,
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	ctx := context.Background()
	err = agent.Run(ctx, "Set up infrastructure for my application")
	if err != nil {
		log.Printf("Agent error: %v", err)
	}

	fmt.Println("Interactive agent completed")
	fmt.Printf("Questions asked: %d\n", developer.questionsAsked)
}

// ExampleDeveloper implements agents.Developer for demonstration.
type ExampleDeveloper struct {
	answers        []string
	questionsAsked int
}

// Respond implements agents.Developer interface.
func (d *ExampleDeveloper) Respond(ctx context.Context, message string) (string, error) {
	fmt.Printf("[Agent asks]: %s\n", message)

	if d.questionsAsked >= len(d.answers) {
		return "I don't know", nil
	}

	answer := d.answers[d.questionsAsked]
	d.questionsAsked++

	fmt.Printf("[Developer answers]: %s\n", answer)
	return answer, nil
}
