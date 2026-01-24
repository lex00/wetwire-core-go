---
title: "Wetwire Core"
---

[![Go Reference](https://pkg.go.dev/badge/github.com/lex00/wetwire-core-go.svg)](https://pkg.go.dev/github.com/lex00/wetwire-core-go)
[![CI](https://github.com/lex00/wetwire-core-go/actions/workflows/ci.yml/badge.svg)](https://github.com/lex00/wetwire-core-go/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Shared agent infrastructure for wetwire domain packages.

## Overview

wetwire-core-go provides the AI agent framework used by wetwire domain packages (like wetwire-aws-go).

### Core Hypothesis

Wetwire validates that **typed constraints reduce required model capability**:

> Typed input + smaller model ≈ Semantic input + larger model

The type system and lint rules act as a force multiplier — cheaper models can produce quality output when guided by schema-generated types and iterative lint feedback.

## Package Summary

| Package | Description |
|---------|-------------|
| [agents](https://pkg.go.dev/github.com/lex00/wetwire-core-go/agent/agents) | Unified Agent architecture with MCP tool integration |
| [mcp](https://pkg.go.dev/github.com/lex00/wetwire-core-go/mcp) | MCP server for Claude Code integration |
| [providers]({{< relref "/providers" >}}) | AI provider abstraction (Anthropic, Claude, Kiro) |
| [personas](https://pkg.go.dev/github.com/lex00/wetwire-core-go/agent/personas) | Developer persona definitions |
| [scoring](https://pkg.go.dev/github.com/lex00/wetwire-core-go/agent/scoring) | 4-dimension evaluation rubric |
| [scenario]({{< relref "/scenarios" >}}) | Multi-domain scenario definitions |
| [domain](https://pkg.go.dev/github.com/lex00/wetwire-core-go/domain) | Domain interface for CLI/MCP generation |

## Installation

```bash
go get github.com/lex00/wetwire-core-go
```

## Quick Start

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
mcp.RegisterStandardToolsWithDefaults(mcpServer, "mydomain", handlers)

// 3. Create provider
provider, _ := anthropic.New(anthropic.Config{})

// 4. Create unified Agent
agent, _ := agents.NewAgent(agents.AgentConfig{
    Provider:     provider,
    MCPServer:    agents.NewMCPServerAdapter(mcpServer),
    SystemPrompt: "You are an infrastructure code generator...",
})

// 5. Run
agent.Run(ctx, "Create an S3 bucket with versioning")
```

## Documentation

| Document | Description |
|----------|-------------|
| [CLI Reference]({{< relref "/cli" >}}) | Command-line interface |
| [Scenarios]({{< relref "/scenarios" >}}) | Multi-domain scenario definitions |
| [Providers]({{< relref "/providers" >}}) | AI provider abstraction |
| [FAQ]({{< relref "/faq" >}}) | Frequently asked questions |
