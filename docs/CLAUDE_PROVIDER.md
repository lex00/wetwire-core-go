# Claude Code Provider

## Overview

The Claude provider (`providers/claude`) implements the `providers.Provider` interface for integration with Claude Code CLI. This enables wetwire agents to run using Claude Code as the AI backend **without requiring an Anthropic API key**.

## Key Difference from Kiro Provider

| Provider | Backend | Use Case |
|----------|---------|----------|
| `claude` | `claude` CLI (Claude Code) | **Recommended**, uses Claude Code directly |
| `kiro` | `kiro-cli` (Amazon Q Developer) | Enterprise environments with approved kiro-cli |

The Claude provider uses the `claude` CLI that comes with Claude Code, making it the preferred choice for local development and testing.

## Architecture

### How It Works

```
┌─────────────────────────────────────────────────────┐
│ Claude Code Provider                                 │
├─────────────────────────────────────────────────────┤
│ provider.CreateMessage(req)                         │
│          ↓                                          │
│ claude --print --output-format json \               │
│        --system-prompt "..." \                      │
│        --mcp-config mcp.json \                      │
│        "prompt"                                     │
│          ↓                                          │
│ Claude Code runs full agentic loop                  │
│ (tools executed via --mcp-config)                   │
│          ↓                                          │
│ Returns final result (StopReasonEndTurn)            │
└─────────────────────────────────────────────────────┘
```

### Important: Claude Code Runs Its Own Loop

Unlike the Anthropic provider which returns tool_use blocks for the caller to handle, the Claude provider runs Claude Code which has its own agentic loop. This means:

1. **The provider returns completed sessions** (`StopReasonEndTurn`)
2. **MCP tools must be external** (configured via `--mcp-config`)
3. **The caller's agentic loop runs only once**

## Installation

The Claude provider requires the `claude` CLI to be installed:

```bash
# Verify installation
claude --version
```

## Usage

### Basic Usage

```go
import (
    "github.com/lex00/wetwire-core-go/providers"
    "github.com/lex00/wetwire-core-go/providers/claude"
)

// Create provider
provider, err := claude.New(claude.Config{
    WorkDir:      "/path/to/project",
    SystemPrompt: "You are an infrastructure code generator.",
})
if err != nil {
    log.Fatal(err)
}

// Send message
resp, err := provider.CreateMessage(ctx, providers.MessageRequest{
    Messages: []providers.Message{
        providers.NewUserMessage("Create an S3 bucket with versioning"),
    },
})

// Process response
for _, block := range resp.Content {
    if block.Type == "text" {
        fmt.Println(block.Text)
    }
}
```

### With MCP Tools

```go
// Write MCP config for domain tools
err := claude.WriteMCPConfig("/tmp/mcp.json", map[string]claude.MCPServerConfig{
    "wetwire-aws": {
        Command: "wetwire-aws",
        Args:    []string{"mcp"},
    },
    "wetwire-gitlab": {
        Command: "wetwire-gitlab",
        Args:    []string{"mcp"},
    },
})

// Create provider with MCP config
provider, _ := claude.New(claude.Config{
    MCPConfigPath: "/tmp/mcp.json",
    SystemPrompt:  "You are an infrastructure code generator...",
})
```

### Streaming Mode

```go
resp, err := provider.StreamMessage(ctx, req, func(text string) {
    fmt.Print(text)  // Print as tokens arrive
})
```

## Configuration

```go
type Config struct {
    // WorkDir is the working directory for claude CLI (default: current directory)
    WorkDir string

    // Model overrides the default model (optional)
    // Examples: "sonnet", "opus", "haiku"
    Model string

    // MCPConfigPath is a path to an MCP config file (optional)
    // Tools from this file will be available to Claude Code
    MCPConfigPath string

    // SystemPrompt is prepended to the request's system prompt (optional)
    SystemPrompt string

    // AllowedTools restricts which tools claude can use (optional)
    // Example: []string{"Bash", "Read", "Edit"}
    AllowedTools []string

    // PermissionMode sets the permission mode (optional)
    // Options: "default", "acceptEdits", "plan", "bypassPermissions"
    PermissionMode string
}
```

## CLI Flags Used

The provider uses these `claude` CLI flags:

| Flag | Purpose |
|------|---------|
| `--print` | Non-interactive mode, output to stdout |
| `--output-format json` | Get structured JSON response |
| `--output-format stream-json` | Get streaming JSON events |
| `--verbose` | Required for stream-json |
| `--system-prompt` | Set system prompt |
| `--model` | Override model |
| `--mcp-config` | Load MCP server configuration |
| `--allowedTools` | Restrict available tools |
| `--permission-mode` | Set permission mode |

## Output Formats

### JSON Mode (CreateMessage)

```json
{
  "type": "result",
  "subtype": "success",
  "is_error": false,
  "result": "The generated response text",
  "session_id": "abc123",
  "num_turns": 2
}
```

### Stream-JSON Mode (StreamMessage)

```json
{"type":"system","subtype":"init","session_id":"abc123","tools":["Bash","Read"]}
{"type":"assistant","message":{"role":"assistant","content":[{"type":"text","text":"Hello"}]}}
{"type":"result","subtype":"success","result":"Hello","session_id":"abc123"}
```

## Comparison: Claude vs Anthropic Provider

| Feature | Claude Provider | Anthropic Provider |
|---------|-----------------|-------------------|
| **Requires API Key** | No | Yes |
| **Agentic Loop** | Claude Code handles it | Caller handles it |
| **Tool Execution** | Via `--mcp-config` | Caller executes tools |
| **Streaming** | Simulated from final result | True token streaming |
| **Cost** | Free (uses Claude Code) | API usage charges |
| **Best For** | Local dev, testing, scenarios | Production, fine control |

## Running Scenarios

The Claude provider is ideal for running scenarios without an API key:

```go
// Generate scenario instructions
instructionSkill := scenario.New(nil, nil)
instructionSkill.Run(ctx, "./examples/aws_gitlab")

// Create provider with domain MCP servers
provider, _ := claude.New(claude.Config{
    MCPConfigPath: mcpConfigPath,
    SystemPrompt:  "Execute this wetwire scenario...",
})

// Run scenario
resp, _ := provider.CreateMessage(ctx, providers.MessageRequest{
    Messages: []providers.Message{
        providers.NewUserMessage(scenarioInstructions),
    },
})
```

## Testing

```bash
# Run provider tests
go test ./providers/claude/... -v

# Run scenario test
go run ./cmd/test_scenario_claude
```

## Limitations

1. **No tool interception** - Claude Code executes tools internally; caller cannot intercept
2. **Single response** - Returns completed session, not intermediate tool_use blocks
3. **Requires claude CLI** - Must have Claude Code installed
4. **Streaming is simulated** - Full response delivered through handler, not true streaming

## Examples

See:
- `examples/claude_provider/main.go` - Basic usage
- `cmd/test_scenario_claude/main.go` - Scenario integration

## References

- Provider interface: `providers/provider.go`
- Claude provider: `providers/claude/claude.go`
- Claude provider tests: `providers/claude/claude_test.go`
