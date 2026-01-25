---
title: "Wetwire Core"
---

[![Go Reference](https://pkg.go.dev/badge/github.com/lex00/wetwire-core-go.svg)](https://pkg.go.dev/github.com/lex00/wetwire-core-go)
[![CI](https://github.com/lex00/wetwire-core-go/actions/workflows/ci.yml/badge.svg)](https://github.com/lex00/wetwire-core-go/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/lex00/wetwire-core-go/graph/badge.svg)](https://codecov.io/gh/lex00/wetwire-core-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/lex00/wetwire-core-go)](https://goreportcard.com/report/github.com/lex00/wetwire-core-go)
[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Agent infrastructure for wetwire.

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
