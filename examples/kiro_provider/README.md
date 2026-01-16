# Kiro Provider Example

This example demonstrates using the Kiro provider, which delegates AI operations to Kiro CLI (Amazon Q Developer CLI) instead of making direct API calls.

## Overview

The Kiro provider allows wetwire to use Kiro CLI as the AI backend. This is useful when:

- Enterprise environments where Kiro CLI is the approved agentic tool
- Users want to leverage existing Kiro CLI installation and configuration
- You want to use Kiro's MCP server integration for tool access
- CI/CD pipelines where Kiro CLI is already available

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
        MCPCommand:  "wetwire-aws",
        MCPArgs:     []string{"mcp"},
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
    AgentName:   string   // Name for the kiro agent
    AgentPrompt: string   // System prompt for the agent
    MCPCommand:  string   // MCP server binary name (e.g., "wetwire-aws")
    MCPArgs:     []string // Args for MCP server (e.g., []string{"mcp"})
    WorkDir:     string   // Working directory for the agent
}
```

Example:

```go
provider, _ := kiro.New(kiro.Config{
    AgentName:   "wetwire-aws-agent",
    AgentPrompt: "You are an AWS infrastructure code generator.",
    MCPCommand:  "wetwire-aws",
    MCPArgs:     []string{"mcp"},
    WorkDir:     "/path/to/project",
})
```

## See Also

- [docs/KIRO_PROVIDER.md](../../docs/KIRO_PROVIDER.md) - Full Kiro provider documentation
- [examples/unified_agent](../unified_agent/) - Using providers with unified Agent
