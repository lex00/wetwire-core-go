# wetwire-core-go

[![Go Reference](https://pkg.go.dev/badge/github.com/lex00/wetwire-core-go.svg)](https://pkg.go.dev/github.com/lex00/wetwire-core-go)
[![CI](https://github.com/lex00/wetwire-core-go/actions/workflows/ci.yml/badge.svg)](https://github.com/lex00/wetwire-core-go/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/lex00/wetwire-core-go/branch/main/graph/badge.svg)](https://codecov.io/gh/lex00/wetwire-core-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/lex00/wetwire-core-go)](https://goreportcard.com/report/github.com/lex00/wetwire-core-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Shared agent infrastructure for wetwire domain packages.

## Overview

wetwire-core-go provides the AI agent framework used by wetwire domain packages (like wetwire-aws-go).

### Core Hypothesis

Wetwire validates that **typed constraints reduce required model capability**:

> Typed input + smaller model ≈ Semantic input + larger model

The type system and lint rules act as a force multiplier — cheaper models can produce quality output when guided by schema-generated types and iterative lint feedback. Scenarios test this by comparing output quality across model/constraint combinations.

### Package Summary

- **agents** - Unified Agent architecture with MCP tool integration
- **mcp** - MCP server for Claude Code integration with standard tool definitions
- **providers** - AI provider abstraction (Anthropic API, Claude Code, Kiro)
- **personas** - Developer persona definitions (Beginner, Intermediate, Expert) with custom persona support
- **scoring** - 4-dimension evaluation rubric (0-12 scale)
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
    "github.com/lex00/wetwire-core-go/providers/claude"
)

var provider providers.Provider

if useClaudeCode {
    // No API key needed - uses Claude Code CLI
    provider, _ = claude.New(claude.Config{
        SystemPrompt: "You are an infrastructure generator...",
    })
} else {
    // Direct API access
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
| [claude_provider](examples/claude_provider/) | Using Claude Code as AI backend (no API key) |
| [kiro_provider](examples/kiro_provider/) | Using Kiro provider (enterprise) |

### Running Scenarios

Results are organized by persona:
```
output/
├── SUMMARY.md           # Results table with all personas
├── default/
│   ├── RESULTS.md       # Response and generated files
│   ├── cfn-templates/s3-bucket.yaml
│   └── .gitlab-ci.yml
├── beginner/
├── intermediate/
└── expert/
```

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
// Claude Code (no API key needed - recommended for local dev)
provider, _ := claude.New(claude.Config{
    SystemPrompt:  "You are an infrastructure generator...",
    MCPConfigPath: "/path/to/mcp.json",  // Optional: MCP tools
})

// Anthropic (direct API - recommended for production)
provider, _ := anthropic.New(anthropic.Config{APIKey: "..."})

// Kiro (enterprise - uses kiro-cli)
provider, _ := kiro.New(kiro.Config{AgentName: "...", MCPCommand: "..."})

// All implement providers.Provider interface
resp, _ := provider.CreateMessage(ctx, req)
resp, _ := provider.StreamMessage(ctx, req, handler)
```

### personas

```go
persona, _ := personas.Get("beginner")  // beginner, intermediate, expert (+ custom)
fmt.Println(persona.Name, persona.Description)
```

### results

```go
session := results.NewSession("domain", "scenario")
session.Complete()
writer := results.NewResultsWriter()
writer.Write(session, "./RESULTS.md")
```

## Implementing a Domain

Domain packages implement the `domain.Domain` interface to get automatic CLI and MCP generation.

### Required Interface

```go
import "github.com/lex00/wetwire-core-go/domain"

type MyDomain struct{}

// Compile-time check - fails if any method is missing
var _ domain.Domain = (*MyDomain)(nil)

func (d *MyDomain) Name() string    { return "mydomain" }
func (d *MyDomain) Version() string { return "1.0.0" }
func (d *MyDomain) Builder() domain.Builder       { return &MyBuilder{} }
func (d *MyDomain) Linter() domain.Linter         { return &MyLinter{} }
func (d *MyDomain) Initializer() domain.Initializer { return &MyInitializer{} }
func (d *MyDomain) Validator() domain.Validator   { return &MyValidator{} }
```

### Usage

```go
func main() {
    cli := domain.Run(&MyDomain{})
    cli.Execute()
}
```

This generates:
- CLI with `build`, `lint`, `init`, `validate` commands
- Persistent `--format` and `--verbose` flags

### MCP Server

```go
server := domain.BuildMCPServer(&MyDomain{})
server.Start()
```

This generates MCP tools: `wetwire_build`, `wetwire_lint`, `wetwire_init`, `wetwire_validate`

### Optional Interfaces

Domains may implement additional capabilities:

```go
// Import external configs
func (d *MyDomain) Importer() domain.Importer { return &MyImporter{} }
var _ domain.ImporterDomain = (*MyDomain)(nil)

// List discovered resources
func (d *MyDomain) Lister() domain.Lister { return &MyLister{} }
var _ domain.ListerDomain = (*MyDomain)(nil)

// Visualize dependencies
func (d *MyDomain) Grapher() domain.Grapher { return &MyGrapher{} }
var _ domain.GrapherDomain = (*MyDomain)(nil)

// Compare outputs semantically
func (d *MyDomain) Differ() domain.Differ { return &MyDiffer{} }
var _ domain.DifferDomain = (*MyDomain)(nil)
```

### Adding Custom Commands

Domain-specific commands are added after `Run()`:

```go
func main() {
    cli := domain.Run(&MyDomain{})
    cli.AddCommand(newDesignCmd())  // AI-assisted design
    cli.AddCommand(newTestCmd())    // Persona testing
    cli.AddCommand(newMCPCmd())     // MCP server
    cli.Execute()
}
```

## Documentation

- [mcp/README.md](mcp/README.md) - MCP server and standard tools
- [docs/CLAUDE_PROVIDER.md](docs/CLAUDE_PROVIDER.md) - Claude Code provider (no API key)
- [docs/KIRO_PROVIDER.md](docs/KIRO_PROVIDER.md) - Kiro provider (enterprise)
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
