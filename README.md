# wetwire-core-go

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
