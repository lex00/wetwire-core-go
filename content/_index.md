---
title: "Wetwire Core"
---

[![Go Reference](https://pkg.go.dev/badge/github.com/lex00/wetwire-core-go.svg)](https://pkg.go.dev/github.com/lex00/wetwire-core-go)
[![CI](https://github.com/lex00/wetwire-core-go/actions/workflows/ci.yml/badge.svg)](https://github.com/lex00/wetwire-core-go/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Shared agent infrastructure for wetwire domain packages — providers, MCP servers, personas, and scoring.

## Philosophy

Wetwire uses typed constraints to reduce the model capability required for accurate code generation.

**Core hypothesis:** Typed input + smaller model ≈ Semantic input + larger model

The type system and lint rules act as a force multiplier — cheaper models produce quality output when guided by schema-generated types and iterative lint feedback.

## Documentation

| Document | Description |
|----------|-------------|
| [CLI Reference]({{< relref "/cli" >}}) | Command-line interface |
| [Providers]({{< relref "/providers" >}}) | AI provider abstraction |
| [Scenarios]({{< relref "/scenarios" >}}) | Multi-domain test scenarios |
| [FAQ]({{< relref "/faq" >}}) | Frequently asked questions |

## Installation

```bash
go get github.com/lex00/wetwire-core-go
```

## Quick Example

```go
provider, _ := anthropic.New(anthropic.Config{})

agent, _ := agents.NewAgent(agents.AgentConfig{
    Provider:     provider,
    SystemPrompt: "You are an infrastructure generator...",
})

agent.Run(ctx, "Create an S3 bucket with versioning")
```
