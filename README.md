# wetwire-core-go

[![Go Reference](https://pkg.go.dev/badge/github.com/lex00/wetwire-core-go.svg)](https://pkg.go.dev/github.com/lex00/wetwire-core-go)
[![CI](https://github.com/lex00/wetwire-core-go/actions/workflows/ci.yml/badge.svg)](https://github.com/lex00/wetwire-core-go/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/lex00/wetwire-core-go/branch/main/graph/badge.svg)](https://codecov.io/gh/lex00/wetwire-core-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/lex00/wetwire-core-go)](https://goreportcard.com/report/github.com/lex00/wetwire-core-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Shared agent infrastructure for wetwire domain packages.

## Overview

wetwire-core-go provides the AI agent framework used by wetwire domain packages (like wetwire-honeycomb-go). It includes:

- **agents** - Unified Agent architecture with MCP tool integration
- **mcp** - MCP server for Claude Code integration with standard tool definitions
- **providers** - AI provider abstraction (Anthropic API, Kiro/Claude Code)
- **personas** - Developer persona definitions (Beginner, Intermediate, Expert, Terse, Verbose)
- **scoring** - 5-dimension evaluation rubric (0-15 scale)
- **results** - Session tracking and RESULTS.md generation
- **orchestrator** - Developer/Runner agent coordination
- **scenario** - Multi-domain scenario definitions with cross-domain validation
- **recording** - Animated SVG recordings of user/agent conversations
- **version** - Version info exposure via runtime/debug
- **cmd** - CLI command framework with cobra
- **serialize** - Struct-to-map conversion and JSON/YAML serialization

## Installation

```bash
go get github.com/lex00/wetwire-core-go
```

## Quick Start

### Unified Agent (Recommended)

The unified Agent is the recommended pattern for all domain packages:

```go
import (
    "github.com/lex00/wetwire-core-go/agent/agents"
    "github.com/lex00/wetwire-core-go/mcp"
    "github.com/lex00/wetwire-core-go/providers/anthropic"
)

// 1. Create MCP server with tools
mcpServer := mcp.NewServer(mcp.Config{
    Name:    "wetwire-mydomain",
    Version: "1.0.0",
})

// 2. Register standard tools
mcp.RegisterStandardToolsWithDefaults(mcpServer, "mydomain", mcp.StandardToolHandlers{
    Init:  myInitHandler,
    Build: myBuildHandler,
    Lint:  myLintHandler,
})

// 3. Create provider
provider, _ := anthropic.New(anthropic.Config{})

// 4. Create unified Agent
agent, _ := agents.NewAgent(agents.AgentConfig{
    Provider:     provider,
    MCPServer:    agents.NewMCPServerAdapter(mcpServer),
    SystemPrompt: "You are an infrastructure code generator...",
    Developer:    developer,  // nil for autonomous mode
})

// 5. Run
agent.Run(ctx, "Create an S3 bucket with versioning")
```

### MCP Server for Claude Code

Expose tools to Claude Code via MCP:

```go
import "github.com/lex00/wetwire-core-go/mcp"

server := mcp.NewServer(mcp.Config{
    Name:    "wetwire-mydomain",
    Version: "1.0.0",
})

mcp.RegisterStandardToolsWithDefaults(server, "mydomain", handlers)
server.Start(context.Background())  // Runs on stdio
```

### Provider Abstraction

Same code works with different AI backends:

```go
import (
    "github.com/lex00/wetwire-core-go/providers"
    "github.com/lex00/wetwire-core-go/providers/anthropic"
    "github.com/lex00/wetwire-core-go/providers/kiro"
)

var provider providers.Provider

if useClaudeCode {
    provider, _ = kiro.New(kiro.Config{...})
} else {
    provider, _ = anthropic.New(anthropic.Config{})
}

// Same API for both
resp, _ := provider.CreateMessage(ctx, req)
```

## Examples

| Example | Description |
|---------|-------------|
| [unified_agent](examples/unified_agent/) | Unified Agent with MCP tools (recommended pattern) |
| [mcp_server](examples/mcp_server/) | MCP server for Claude Code integration |
| [kiro_provider](examples/kiro_provider/) | Using Kiro provider with Claude Code |
| [aws_gitlab](examples/aws_gitlab/) | Multi-domain scenario example |

## Package Reference

### agents

```go
// Create unified Agent
agent, _ := agents.NewAgent(agents.AgentConfig{
    Provider:     provider,
    MCPServer:    mcpServer,
    SystemPrompt: "...",
    Developer:    developer,  // nil for autonomous mode
    Session:      session,    // optional, for result tracking
})

// Wrap MCP server for Agent
adapter := agents.NewMCPServerAdapter(mcpServer)
```

### mcp

```go
// Create server
server := mcp.NewServer(mcp.Config{Name: "domain", Version: "1.0.0"})

// Register tools with default file handlers
mcp.RegisterStandardToolsWithDefaults(server, "domain", handlers)

// In-process tool execution (for Agent)
result, _ := server.ExecuteTool(ctx, "tool_name", args)
tools := server.GetTools()
```

### providers

```go
// Anthropic (direct API)
provider, _ := anthropic.New(anthropic.Config{APIKey: "..."})

// Kiro (Claude Code backend)
provider, _ := kiro.New(kiro.Config{AgentName: "...", MCPCommand: "..."})

// Both implement providers.Provider interface
resp, _ := provider.CreateMessage(ctx, req)
resp, _ := provider.StreamMessage(ctx, req, handler)
```

### personas

```go
persona, _ := personas.Get("beginner")  // beginner, intermediate, expert, terse, verbose
fmt.Println(persona.Name, persona.Description)
```

### results

```go
session := results.NewSession("domain", "scenario")
session.Complete()
writer := results.NewResultsWriter()
writer.Write(session, "./RESULTS.md")
```

## Documentation

- [mcp/README.md](mcp/README.md) - MCP server and standard tools
- [docs/KIRO_PROVIDER.md](docs/KIRO_PROVIDER.md) - Kiro provider for Claude Code
- [docs/SCENARIOS.md](docs/SCENARIOS.md) - Multi-domain scenario definitions
- [docs/RECORDING.md](docs/RECORDING.md) - SVG recording of conversations
- [docs/FAQ.md](docs/FAQ.md) - Frequently asked questions

## Migration from RunnerAgent

The `RunnerAgent` is deprecated. Migrate to the unified Agent:

```go
// OLD (deprecated):
runner, _ := agents.NewRunnerAgent(agents.RunnerConfig{
    Domain:    myDomain,
    WorkDir:   "./output",
    Developer: developer,
})
runner.Run(ctx, prompt)

// NEW (recommended):
mcpServer := mcp.NewServer(mcp.Config{Name: "domain"})
mcp.RegisterStandardToolsWithDefaults(mcpServer, "domain", handlers)

agent, _ := agents.NewAgent(agents.AgentConfig{
    Provider:     provider,
    MCPServer:    agents.NewMCPServerAdapter(mcpServer),
    Developer:    developer,
    SystemPrompt: systemPrompt,
})
agent.Run(ctx, prompt)
```

## License

MIT
