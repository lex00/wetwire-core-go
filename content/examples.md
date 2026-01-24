---
title: "Examples"
---

The `examples/` directory contains reference implementations demonstrating core wetwire patterns.

## Available Examples

### [mcp_server](https://github.com/lex00/wetwire-core-go/tree/main/examples/mcp_server)

Demonstrates how to create an MCP (Model Context Protocol) server that exposes wetwire tools to Claude Code and other MCP clients.

**Concepts demonstrated:**
- Creating an MCP server
- Registering standard tools
- Registering custom tools
- Claude Code integration

```go
server := mcp.NewServer(mcp.Config{
    Name:    "wetwire-mydomain",
    Version: "1.0.0",
})

mcp.RegisterStandardToolsWithDefaults(server, "mydomain", handlers)
```

### [claude_provider](https://github.com/lex00/wetwire-core-go/tree/main/examples/claude_provider)

Shows how to use the Claude CLI provider for AI-assisted code generation without an API key.

**Concepts demonstrated:**
- Provider configuration
- Claude CLI integration
- No API key required workflow

### [kiro_provider](https://github.com/lex00/wetwire-core-go/tree/main/examples/kiro_provider)

Demonstrates the Kiro provider for alternative AI model integration.

**Concepts demonstrated:**
- Kiro provider setup
- Model configuration
- Provider switching

### [unified_agent](https://github.com/lex00/wetwire-core-go/tree/main/examples/unified_agent)

A complete example of the unified agent that combines MCP server, providers, and scenarios.

**Concepts demonstrated:**
- Agent orchestration
- Multi-provider support
- Scenario execution

## Running Examples

```bash
# Run any example
cd examples/mcp_server
go run .

# Or from repo root
go run ./examples/mcp_server/
```

## Example Structure

Each example follows this structure:

```
example-name/
├── main.go         # Entry point
├── README.md       # Example documentation
└── go.mod          # (optional) Go module file
```

## See Also

- [Providers](/providers/) - Provider configuration reference
- [Scenarios](/scenarios/) - Scenario testing documentation
- [CLI](/cli/) - Command-line interface
