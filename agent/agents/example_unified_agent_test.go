package agents_test

import (
	"context"
	"fmt"

	"github.com/lex00/wetwire-core-go/agent/agents"
	"github.com/lex00/wetwire-core-go/agent/results"
	"github.com/lex00/wetwire-core-go/mcp"
	"github.com/lex00/wetwire-core-go/providers/anthropic"
)

// Example_unifiedAgent_autonomous shows how to use the unified Agent in autonomous mode.
// The agent runs without human interaction, using only the tools from the MCP server.
func Example_unifiedAgent_autonomous() {
	// 1. Create an MCP server with domain-specific tools
	mcpServer := mcp.NewServer(mcp.Config{
		Name:    "example-domain",
		Version: "1.0.0",
	})

	// Register tools
	mcpServer.RegisterToolWithSchema(
		"init_project",
		"Initialize a new project",
		func(ctx context.Context, args map[string]any) (string, error) {
			name := args["name"].(string)
			return fmt.Sprintf("Created project: %s", name), nil
		},
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
			},
			"required": []string{"name"},
		},
	)

	// 2. Create provider
	provider, _ := anthropic.New(anthropic.Config{
		// APIKey from environment
	})

	// 3. Create unified agent
	agent, _ := agents.NewAgent(agents.AgentConfig{
		Provider:  provider,
		MCPServer: agents.NewMCPServerAdapter(mcpServer),
		SystemPrompt: `You are a project generator.
When asked to create a project, use init_project tool.`,
	})

	// 4. Run autonomously
	ctx := context.Background()
	_ = agent.Run(ctx, "Create a project called 'my-app'")
}

// Example_unifiedAgent_withDeveloper shows how to use the unified Agent with a Developer
// for interactive mode where the agent can ask clarifying questions.
func Example_unifiedAgent_withDeveloper() {
	// 1. Create MCP server
	mcpServer := mcp.NewServer(mcp.Config{
		Name: "example-domain",
	})

	// Register domain tools
	mcpServer.RegisterTool("build", "Build the project",
		func(ctx context.Context, args map[string]any) (string, error) {
			return "Build successful", nil
		},
	)

	// 2. Create a developer (could be human or AI)
	// In real code, you would implement the Developer interface:
	//
	// type MyDeveloper struct{}
	// func (d *MyDeveloper) Respond(ctx context.Context, message string) (string, error) {
	//     // Get user input or AI response
	//     return "user answer", nil
	// }
	//
	// For this example, we'll use nil (autonomous mode)
	var developer agents.Developer // nil = autonomous mode

	// 3. Create session for tracking
	session := results.NewSession("developer", "example-scenario")

	// 4. Create provider
	provider, _ := anthropic.New(anthropic.Config{})

	// 5. Create agent with developer
	agent, _ := agents.NewAgent(agents.AgentConfig{
		Provider:  provider,
		MCPServer: agents.NewMCPServerAdapter(mcpServer),
		Developer: developer, // Enable interactive mode
		Session:   session,   // Track questions/answers
		SystemPrompt: `You are a helpful assistant.
You can use the 'ask_developer' tool to ask questions.
You can use the 'build' tool to build the project.`,
	})

	// 6. Run with developer interaction
	ctx := context.Background()
	_ = agent.Run(ctx, "Build my project")

	// Session will contain all questions asked and answers received
	fmt.Printf("Questions asked: %d\n", len(session.Questions))
}

// Example_unifiedAgent_migration shows how to migrate from RunnerAgent to Agent.
func Example_unifiedAgent_migration() {
	// OLD WAY (deprecated):
	// runner, _ := agents.NewRunnerAgent(agents.RunnerConfig{
	//     Domain:    myDomainConfig,  // domain is required
	//     Provider:  provider,
	//     WorkDir:   "/tmp/work",
	//     Session:   session,
	//     Developer: developer,
	// })
	// runner.Run(ctx, prompt)

	// NEW WAY:

	// 1. Create MCP server and register tools
	mcpServer := mcp.NewServer(mcp.Config{
		Name:    "wetwire-aws",
		Version: "1.0.0",
	})

	// Register standard wetwire tools
	mcp.RegisterStandardTools(mcpServer, "aws", mcp.StandardToolHandlers{
		Init: func(ctx context.Context, args map[string]any) (string, error) {
			// Implementation
			return "Initialized", nil
		},
		Build: func(ctx context.Context, args map[string]any) (string, error) {
			// Implementation
			return "Built", nil
		},
		Lint: func(ctx context.Context, args map[string]any) (string, error) {
			// Implementation
			return "Linted", nil
		},
		// ... other handlers
	})

	// 2. Create provider and session
	provider, _ := anthropic.New(anthropic.Config{})
	session := results.NewSession("dev", "scenario")

	// 3. Create unified agent
	agent, _ := agents.NewAgent(agents.AgentConfig{
		Provider:  provider,
		MCPServer: agents.NewMCPServerAdapter(mcpServer),
		Session:   session,
		SystemPrompt: `You are an infrastructure code generator using wetwire-aws.
Your job is to generate Go code that defines AWS CloudFormation resources.`,
	})

	// 4. Run
	ctx := context.Background()
	_ = agent.Run(ctx, "Create an S3 bucket")

	fmt.Println("Migration complete")
}
