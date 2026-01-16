# Kiro Provider Example

This example demonstrates using the Kiro provider, which delegates AI operations to Claude Code instead of making direct API calls.

## Overview

The Kiro provider allows wetwire to use Claude Code as the AI backend. This is useful when:

- Users are already paying for Claude Code
- You want to leverage Claude Code's MCP integration
- You need Claude Code's extended context and capabilities

## Prerequisites

- `kiro-cli` must be installed and available in PATH
- An MCP server command must be configured

## Running the Example

```bash
# The example checks for kiro-cli availability
go run ./examples/kiro_provider/

# To actually run with Kiro:
# 1. Install kiro-cli
# 2. Configure your MCP server
# 3. Run the example
```

## Key Concepts

### Provider Interface

Both Anthropic and Kiro providers implement the same interface:

```go
type Provider interface {
    CreateMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error)
    StreamMessage(ctx context.Context, req MessageRequest, handler StreamHandler) (*MessageResponse, error)
    Name() string
}
```

### Switching Providers

```go
var provider providers.Provider

if useKiro {
    provider, _ = kiro.New(kiro.Config{
        AgentName:   "wetwire-agent",
        AgentPrompt: "You are a helpful assistant.",
        MCPCommand:  "wetwire-mcp",
    })
} else {
    provider, _ = anthropic.New(anthropic.Config{})
}

// Same code works with either provider
resp, _ := provider.CreateMessage(ctx, req)
```

## Configuration

```go
kiro.Config{
    AgentName:   string  // Name for the kiro agent
    AgentPrompt: string  // System prompt for the agent
    MCPCommand:  string  // MCP server command to run
    WorkDir:     string  // Working directory for the agent
}
```

## See Also

- [docs/KIRO_PROVIDER.md](../../docs/KIRO_PROVIDER.md) - Full Kiro provider documentation
- [examples/unified_agent](../unified_agent/) - Using providers with unified Agent
