# Kiro Provider

## Overview

The Kiro provider (`providers/kiro`) implements the `providers.Provider` interface for integration with Kiro CLI (Amazon Q Developer CLI). This enables wetwire agents to run using Kiro as the AI backend instead of making direct Anthropic API calls.

This is particularly valuable in enterprise environments where Kiro CLI may be the only approved agentic CLI tool.

## Architecture

### Provider Abstraction

The Kiro provider is part of a broader provider abstraction that allows the same agent code to work with different AI backends:

```
providers.Provider interface
    ├── anthropic.Provider - Direct Anthropic API calls
    └── kiro.Provider - Delegates to Claude Code via kiro-cli
```

### Communication Mechanism

The Kiro provider uses a **subprocess execution model**:

1. It spawns `kiro-cli` as a subprocess
2. Passes configuration via JSON files (`.kiro/mcp.json` and `~/.kiro/agents/{agent_name}.json`)
3. Communicates via command-line arguments (prompt passed as positional argument)
4. Captures stdout/stderr for response parsing
5. Returns parsed output as `MessageResponse`

This is implemented in the `kiro` package (`/Users/alex/Documents/checkouts/wetwire-core-go/kiro/kiro.go`):

```go
// RunTest executes a non-interactive test scenario using Kiro.
func RunTest(ctx context.Context, config Config, prompt string) (*TestResult, error)
```

### File Structure

```
wetwire-core-go/
├── providers/
│   ├── provider.go              # Provider interface definition
│   ├── anthropic/
│   │   ├── anthropic.go         # Anthropic API implementation
│   │   └── anthropic_test.go
│   └── kiro/
│       ├── kiro.go              # Kiro CLI implementation
│       └── kiro_test.go
└── kiro/
    ├── kiro.go                   # Kiro CLI integration utilities
    └── kiro_test.go
```

## Provider Interface

All providers must implement:

```go
type Provider interface {
    // CreateMessage sends a message request and returns the complete response
    CreateMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error)

    // StreamMessage sends a message request and streams the response
    StreamMessage(ctx context.Context, req MessageRequest, handler StreamHandler) (*MessageResponse, error)

    // Name returns the provider name (e.g., "anthropic", "kiro")
    Name() string
}
```

## Kiro Provider Implementation

### Configuration

```go
type Config struct {
    // AgentName is the identifier for this agent (e.g., "wetwire-gitlab-runner")
    AgentName string

    // AgentPrompt is the system prompt for the agent (domain-specific instructions)
    AgentPrompt string

    // MCPCommand is the MCP server command to run (e.g., "wetwire-gitlab-mcp")
    MCPCommand string

    // MCPArgs are optional arguments for the MCP server
    MCPArgs []string

    // WorkDir is the working directory for the agent
    WorkDir string
}
```

### Creating a Provider

```go
import "github.com/lex00/wetwire-core-go/providers/kiro"

provider, err := kiro.New(kiro.Config{
    AgentName:   "wetwire-aws-agent",
    AgentPrompt: "You are an AWS infrastructure code generator...",
    MCPCommand:  "wetwire-aws-mcp",
    WorkDir:     "/path/to/project",
})
if err != nil {
    log.Fatal(err)
}
```

### Using the Provider

```go
import "github.com/lex00/wetwire-core-go/providers"

req := providers.MessageRequest{
    Model:      "claude-sonnet-4-20250514",
    MaxTokens:  4096,
    Messages: []providers.Message{
        providers.NewUserMessage("Create an S3 bucket with versioning"),
    },
}

resp, err := provider.CreateMessage(ctx, req)
if err != nil {
    log.Fatal(err)
}

for _, block := range resp.Content {
    if block.Type == "text" {
        fmt.Println(block.Text)
    }
}
```

## Key Features

### 1. Non-Interactive Execution

The Kiro provider runs in non-interactive mode (`--no-interactive` flag), making it suitable for automated workflows:

```bash
kiro-cli chat --agent wetwire-aws-agent --no-interactive "Create an S3 bucket"
```

### 2. Configuration File Management

The provider automatically installs configuration files:

- **Project config**: `.kiro/mcp.json` in the working directory
- **Agent config**: `~/.kiro/agents/{agent_name}.json` in user home directory

These are generated from the `Config` struct and written before each execution.

### 3. MCP Tool Integration

The Kiro provider integrates MCP servers for tool access:

```json
{
  "name": "wetwire-aws-agent",
  "prompt": "You are an AWS infrastructure code generator...",
  "mcpServers": {
    "wetwire-aws-mcp": {
      "command": "wetwire-aws-mcp",
      "args": ["wetwire-aws-mcp"],
      "cwd": "/path/to/project"
    }
  },
  "tools": ["@wetwire-aws-mcp"]
}
```

### 4. Streaming Support

While kiro-cli doesn't support true streaming, the provider simulates it by:

1. Running `CreateMessage` to get the full response
2. Passing the complete response through the `StreamHandler`

```go
func (p *Provider) StreamMessage(ctx context.Context, req MessageRequest, handler StreamHandler) (*MessageResponse, error) {
    resp, err := p.CreateMessage(ctx, req)
    if err != nil {
        return nil, err
    }

    // Deliver the full response through the handler
    for _, block := range resp.Content {
        if block.Type == "text" {
            handler(block.Text)
        }
    }

    return resp, nil
}
```

## Implementation Details

### Prompt Building

The provider extracts user messages from the conversation history:

```go
func (p *Provider) buildPrompt(req MessageRequest) string {
    var userMessages []string

    for _, msg := range req.Messages {
        if msg.Role == "user" {
            for _, block := range msg.Content {
                if block.Type == "text" {
                    userMessages = append(userMessages, block.Text)
                }
            }
        }
    }

    return strings.Join(userMessages, "\n\n")
}
```

### Output Parsing

The provider converts kiro-cli output into structured responses:

```go
func (p *Provider) parseOutput(output string, exitCode int) *MessageResponse {
    resp := &MessageResponse{
        StopReason: StopReasonEndTurn,
    }

    if output != "" {
        resp.Content = []ContentBlock{
            {
                Type: "text",
                Text: output,
            },
        }
    }

    return resp
}
```

### Error Handling

The provider checks for kiro-cli availability before execution:

```go
func (p *Provider) CreateMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error) {
    if !kiro.KiroAvailable() {
        return nil, fmt.Errorf("kiro-cli not found in PATH")
    }
    // ...
}
```

## Testing

The provider includes comprehensive tests:

```bash
cd /Users/alex/Documents/checkouts/wetwire-core-go
go test ./providers/kiro/... -v
```

Test coverage includes:

- Interface compliance verification
- Provider name verification
- Configuration handling
- Error handling when kiro-cli is unavailable
- Prompt building from various message structures
- Output parsing with different scenarios

## Comparison: Anthropic vs Kiro

| Feature | Anthropic Provider | Kiro Provider |
|---------|-------------------|---------------|
| **Communication** | HTTP API calls | Subprocess execution |
| **Authentication** | ANTHROPIC_API_KEY | Uses Kiro CLI session |
| **Streaming** | True streaming | Simulated (full response) |
| **Tool Execution** | Direct via API | Via MCP server |
| **Use Case** | Production deployments | Enterprise, local dev |
| **Cost** | API usage charges | Uses existing Kiro license |

## When to Use Kiro Provider

Use the Kiro provider when:

1. **Enterprise environments** - Kiro CLI is the approved/only agentic tool
2. **Local development** - Testing agents without API charges
3. **CI/CD pipelines** - If Kiro CLI is already available
4. **MCP integration testing** - Testing MCP server implementations

Use the Anthropic provider when:

1. **Production deployments** - Direct API access, no CLI dependency
2. **True streaming required** - Real-time token streaming
3. **High performance** - Lower latency than subprocess execution
4. **Standalone services** - No external dependencies

## Example: Switching Providers

The provider abstraction makes switching backends trivial:

```go
import (
    "github.com/lex00/wetwire-core-go/providers"
    "github.com/lex00/wetwire-core-go/providers/anthropic"
    "github.com/lex00/wetwire-core-go/providers/kiro"
)

func createProvider(useKiro bool) (providers.Provider, error) {
    if useKiro {
        return kiro.New(kiro.Config{
            AgentName:   "wetwire-agent",
            AgentPrompt: systemPrompt,
            MCPCommand:  "wetwire-mcp",
        })
    }
    return anthropic.New(anthropic.Config{})
}

func main() {
    provider, _ := createProvider(os.Getenv("USE_KIRO") == "true")
    // Same code works with either provider
    resp, _ := provider.CreateMessage(ctx, req)
}
```

## Future Enhancements

Potential improvements to the Kiro provider:

1. **True streaming support** - If kiro-cli adds streaming API
2. **Tool use detection** - Parse tool invocations from output
3. **Multi-turn conversations** - Better conversation state management
4. **Performance optimization** - Reduce subprocess overhead
5. **Extended output formats** - Support structured responses beyond text

## References

- Provider interface: `/Users/alex/Documents/checkouts/wetwire-core-go/providers/provider.go`
- Kiro provider: `/Users/alex/Documents/checkouts/wetwire-core-go/providers/kiro/kiro.go`
- Kiro utilities: `/Users/alex/Documents/checkouts/wetwire-core-go/kiro/kiro.go`
- Anthropic provider: `/Users/alex/Documents/checkouts/wetwire-core-go/providers/anthropic/anthropic.go`
