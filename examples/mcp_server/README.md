# MCP Server Example

This example demonstrates how to create an MCP (Model Context Protocol) server that exposes wetwire tools to Claude Code and other MCP clients.

## Overview

MCP servers allow AI assistants like Claude Code to use wetwire tools directly. When you run `wetwire-honeycomb mcp`, it starts an MCP server that Claude Code can connect to.

## Running the Example

```bash
# Build and run the MCP server
go run ./examples/mcp_server/

# The server runs on stdio - it expects JSON-RPC input
# In practice, Claude Code launches this as a subprocess
```

## Key Concepts

### Creating an MCP Server

```go
server := mcp.NewServer(mcp.Config{
    Name:    "wetwire-mydomain",
    Version: "1.0.0",
    Debug:   true,  // Logs to stderr
})
```

### Registering Standard Tools

```go
handlers := mcp.StandardToolHandlers{
    Init:  myInitHandler,
    Build: myBuildHandler,
    Lint:  myLintHandler,
    // Write and Read have default implementations
}

mcp.RegisterStandardToolsWithDefaults(server, "mydomain", handlers)
```

### Registering Custom Tools

```go
server.RegisterToolWithSchema(
    "custom_tool",
    "Description of what this tool does",
    func(ctx context.Context, args map[string]any) (string, error) {
        // Tool implementation
        return "result", nil
    },
    map[string]any{
        "type": "object",
        "properties": map[string]any{
            "param1": map[string]any{"type": "string"},
        },
    },
)
```

### Starting the Server

```go
// Blocks until stdin is closed
err := server.Start(context.Background())
```

## Claude Code Integration

To use your MCP server with Claude Code, add to your settings:

```json
{
  "mcpServers": {
    "wetwire-mydomain": {
      "command": "/path/to/your/binary",
      "args": ["mcp"]
    }
  }
}
```

## Standard Tools

All wetwire MCP servers should provide these tools:

| Tool | Description |
|------|-------------|
| `wetwire_init` | Initialize a new project |
| `wetwire_write` | Write content to a file |
| `wetwire_read` | Read content from a file |
| `wetwire_build` | Generate output from code |
| `wetwire_lint` | Run domain linter |
| `wetwire_list` | List discovered resources |
| `wetwire_graph` | Visualize dependencies |

## See Also

- [mcp/README.md](../../mcp/README.md) - Full MCP package documentation
- [examples/unified_agent](../unified_agent/) - Using MCP server with unified Agent
