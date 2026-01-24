---
title: "FAQ"
---

## What is wetwire-core-go?

wetwire-core-go is the shared agent infrastructure used by all wetwire domain packages. It provides:

- AI provider abstraction (Anthropic, Claude Code, Kiro)
- MCP server for Claude Code integration
- Persona-based testing framework
- Scenario execution and validation
- Session tracking and results generation

## How do I add wetwire-core-go to my domain package?

```bash
go get github.com/lex00/wetwire-core-go
```

Then implement the `domain.Domain` interface:

```go
import "github.com/lex00/wetwire-core-go/domain"

type MyDomain struct{}
var _ domain.Domain = (*MyDomain)(nil)

func (d *MyDomain) Name() string { return "mydomain" }
// ... implement other required methods
```

## What providers are supported?

| Provider | Use Case | API Key Required |
|----------|----------|------------------|
| Anthropic | Direct API access | Yes |
| Claude | Claude Code CLI integration | No |
| Kiro | Enterprise environments | No (uses kiro-cli) |

## How does scoring work?

Scenarios are scored on a 0-12 scale across 4 dimensions:

| Dimension | Points | Description |
|-----------|--------|-------------|
| Completeness | 0-3 | Were all required resources generated? |
| Lint Quality | 0-3 | How many lint cycles needed? |
| Output Validity | 0-3 | Does generated output validate? |
| Question Efficiency | 0-3 | Appropriate number of clarifying questions? |

## What are personas?

Personas simulate different developer skill levels:

- **Beginner** — Uncertain, asks many questions
- **Intermediate** — Some knowledge, may miss details
- **Expert** — Precise requirements, minimal hand-holding

Custom personas can be registered for domain-specific testing.

## How do I run a scenario?

```bash
go run ./cmd/run_scenario ./examples/my_scenario beginner ./results
```

Or run all personas:

```bash
go run ./cmd/run_scenario ./examples/my_scenario --all ./results
```
