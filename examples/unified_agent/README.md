<picture>
  <source media="(prefers-color-scheme: dark)" srcset="../../docs/wetwire-dark.svg">
  <img src="../../docs/wetwire-light.svg" width="100" height="67">
</picture>

This example demonstrates the recommended pattern for using the unified Agent architecture in wetwire domain packages.

## Overview

The unified Agent replaces the deprecated `RunnerAgent` with a cleaner architecture:

- **MCP Server** provides tools (no hardcoded tools in the agent)
- **MCPServerAdapter** connects the MCP server to the Agent
- **Agent** runs the agentic loop, fetching tools dynamically

## Running the Example

```bash
# Set your API key
export ANTHROPIC_API_KEY=your-key-here

# Run the example
go run ./examples/unified_agent/
```

## Key Concepts

### 1. Autonomous Mode (Developer = nil)

```go
agent, _ := agents.NewAgent(agents.AgentConfig{
    Provider:     provider,
    MCPServer:    agents.NewMCPServerAdapter(mcpServer),
    SystemPrompt: "...",
    // Developer: nil - runs without human interaction
})
```

Use for: scenarios, automated testing, batch processing.

### 2. Interactive Mode (Developer != nil)

```go
agent, _ := agents.NewAgent(agents.AgentConfig{
    Provider:     provider,
    MCPServer:    agents.NewMCPServerAdapter(mcpServer),
    Developer:    myDeveloper, // implements agents.Developer
    SystemPrompt: "...",
})
```

Use for: design mode, interactive sessions.

### 3. Tool Registration

```go
// Standard tools with default file handlers
mcp.RegisterStandardToolsWithDefaults(server, "domain", mcp.StandardToolHandlers{
    Init:  myInitHandler,
    Build: myBuildHandler,
    Lint:  myLintHandler,
})

// Custom tools
server.RegisterToolWithSchema("custom_tool", "Description", handler, schema)
```

## Migration from RunnerAgent

```go
// OLD (deprecated):
runner, _ := agents.NewRunnerAgent(agents.RunnerConfig{
    Domain:    myDomain,
    WorkDir:   "./output",
    Developer: developer,
})
runner.Run(ctx, prompt)

// NEW (recommended):
mcpServer := mcp.NewServer(mcp.Config{Name: "my-domain"})
mcp.RegisterStandardToolsWithDefaults(mcpServer, "domain", handlers)

agent, _ := agents.NewAgent(agents.AgentConfig{
    Provider:     provider,
    MCPServer:    agents.NewMCPServerAdapter(mcpServer),
    Developer:    developer,
    SystemPrompt: mySystemPrompt,
})
agent.Run(ctx, prompt)
```

## Benefits

1. **Single tool system** - MCP tools work for both Claude Code integration and direct agent usage
2. **Provider agnostic** - Same code works with Anthropic API or Kiro (Claude Code)
3. **Extensible** - Easy to add domain-specific tools
4. **Testable** - Mock the MCP server for unit tests
