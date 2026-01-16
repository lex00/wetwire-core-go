# Claude Provider Example

This example demonstrates using the Claude Code provider, which delegates AI operations to the `claude` CLI instead of making direct API calls.

## Overview

The Claude provider allows wetwire to use Claude Code as the AI backend. This is useful when:

- Users already have Claude Code installed (no API key needed)
- You want to leverage Claude Code's tool permissions and sandboxing
- You need Claude Code's extended context and conversation management

## Prerequisites

- `claude` CLI must be installed and available in PATH
- For MCP tools, an MCP server binary must be available

## Running the Example

```bash
go run ./examples/claude_provider/
```

If `claude` CLI is not available, the example demonstrates the API structure without executing.

## Key Concepts

### Creating a Provider

```go
provider, err := claude.New(claude.Config{
    SystemPrompt: "You are a helpful assistant.",
})
```

### Sending Messages

```go
resp, err := provider.CreateMessage(ctx, providers.MessageRequest{
    Messages: []providers.Message{
        providers.NewUserMessage("What is 2+2?"),
    },
})

for _, block := range resp.Content {
    if block.Type == "text" {
        fmt.Println(block.Text)
    }
}
```

### Streaming Responses

```go
resp, err := provider.StreamMessage(ctx, req, func(text string) {
    fmt.Print(text)  // Print each chunk as it arrives
})
```

### MCP Server Configuration

```go
// Write MCP config for domain tools
claude.WriteMCPConfig(configPath, map[string]claude.MCPServerConfig{
    "wetwire-aws": {
        Command: "wetwire-aws-mcp",
        Args:    []string{},
    },
})

// Create provider with MCP config
provider, _ := claude.New(claude.Config{
    MCPConfigPath:  configPath,
    SystemPrompt:   "Use wetwire-aws tools to create resources.",
    PermissionMode: "acceptEdits",
})
```

## Configuration Options

```go
claude.Config{
    SystemPrompt:   string   // System prompt for the agent
    MCPConfigPath:  string   // Path to MCP server config JSON
    AllowedTools:   []string // Tools to allow (e.g., "Write", "Bash")
    PermissionMode: string   // "acceptEdits" to auto-approve file changes
    WorkDir:        string   // Working directory for file operations
}
```

## Provider Interface

The Claude provider implements the standard `providers.Provider` interface:

```go
type Provider interface {
    CreateMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error)
    StreamMessage(ctx context.Context, req MessageRequest, handler StreamHandler) (*MessageResponse, error)
    Name() string
}
```

This allows seamless switching between Claude Code and direct API providers.

## See Also

- [docs/CLAUDE_PROVIDER.md](../../docs/CLAUDE_PROVIDER.md) - Full provider documentation
- [examples/unified_agent](../unified_agent/) - Using providers with unified Agent
- [examples/mcp_server](../mcp_server/) - Creating MCP servers for Claude Code
