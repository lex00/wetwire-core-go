# wetwire-core-go

[![Go Reference](https://pkg.go.dev/badge/github.com/lex00/wetwire-core-go.svg)](https://pkg.go.dev/github.com/lex00/wetwire-core-go)
[![CI](https://github.com/lex00/wetwire-core-go/actions/workflows/ci.yml/badge.svg)](https://github.com/lex00/wetwire-core-go/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/lex00/wetwire-core-go)](https://goreportcard.com/report/github.com/lex00/wetwire-core-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Shared agent infrastructure for wetwire domain packages.

## Overview

wetwire-core-go provides the AI agent framework used by wetwire domain packages (like wetwire-aws-go). It includes:

- **personas** - Developer persona definitions (Beginner, Intermediate, Expert, Terse, Verbose)
- **scoring** - 5-dimension evaluation rubric (0-15 scale)
- **results** - Session tracking and RESULTS.md generation
- **orchestrator** - Developer/Runner agent coordination
- **agents** - Anthropic SDK integration and RunnerAgent

## Installation

```bash
go get github.com/lex00/wetwire-core-go
```

## Usage

wetwire-core-go is typically used as a dependency of domain packages like wetwire-aws-go.

```go
import (
    "github.com/lex00/wetwire-core-go/agent/orchestrator"
    "github.com/lex00/wetwire-core-go/agent/personas"
    "github.com/lex00/wetwire-core-go/agent/scoring"
)
```

## License

Apache License 2.0
