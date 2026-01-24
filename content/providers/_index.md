---
title: "Providers"
---

wetwire-core-go supports multiple AI providers through a unified interface.

## Available Providers

| Provider | Package | API Key | Use Case |
|----------|---------|---------|----------|
| Anthropic | `providers/anthropic` | Required | Direct API access, production |
| Claude | `providers/claude` | Not required | Claude Code CLI, local dev |
| Kiro | `providers/kiro` | Not required | Enterprise environments |

## Provider Interface

All providers implement `providers.Provider`:

```go
type Provider interface {
    CreateMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error)
    StreamMessage(ctx context.Context, req MessageRequest, handler StreamHandler) error
}
```

## Anthropic Provider

Direct API access to Claude models.

```go
import "github.com/lex00/wetwire-core-go/providers/anthropic"

provider, err := anthropic.New(anthropic.Config{
    APIKey: os.Getenv("ANTHROPIC_API_KEY"),
    Model:  "claude-sonnet-4-20250514",  // optional
})
```

## Claude Provider

Uses Claude Code CLI - no API key needed for local development.

```go
import "github.com/lex00/wetwire-core-go/providers/claude"

provider, err := claude.New(claude.Config{
    SystemPrompt:  "You are an infrastructure generator...",
    MCPConfigPath: "/path/to/mcp.json",  // optional
})
```

## Kiro Provider

Enterprise provider using kiro-cli.

```go
import "github.com/lex00/wetwire-core-go/providers/kiro"

provider, err := kiro.New(kiro.Config{
    AgentName:  "my-agent",
    MCPCommand: "my-mcp-server",
})
```

## Choosing a Provider

| Scenario | Recommended Provider |
|----------|---------------------|
| Local development | Claude (no API key) |
| CI/CD pipelines | Anthropic (API key in secrets) |
| Enterprise with Kiro | Kiro |
