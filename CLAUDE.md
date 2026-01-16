# wetwire-core-go

Shared agent infrastructure for wetwire domain packages (Go).

## Package Structure

```
wetwire-core-go/
└── agent/
    ├── personas/      # 5 built-in developer personas
    ├── scoring/       # 5-dimension evaluation rubric
    ├── orchestrator/  # Developer/Runner coordination
    ├── results/       # Session tracking, RESULTS.md generation
    └── agents/        # DeveloperAgent, RunnerAgent
```

## Core Components

### Personas

Five built-in personas for testing AI-human collaboration:

- **Beginner** — Uncertain, asks many questions, needs guidance
- **Intermediate** — Some knowledge, specifies requirements but may miss details
- **Expert** — Deep knowledge, precise requirements, minimal hand-holding
- **Terse** — Minimal information, expects system to infer defaults
- **Verbose** — Over-explains, buries requirements in prose

```go
import "github.com/lex00/wetwire-core-go/agent/personas"

persona, err := personas.Get("beginner")
// persona.Name, persona.Description, persona.SystemPrompt
```

### Scoring

4-dimension evaluation rubric (0-12 scale):

| Dimension | Points | Description |
|-----------|--------|-------------|
| Completeness | 0-3 | Were all required resources generated? |
| Lint Quality | 0-3 | How many lint cycles needed? |
| Output Validity | 0-3 | Does generated output validate? |
| Question Efficiency | 0-3 | Appropriate number of clarifying questions? |

```go
import "github.com/lex00/wetwire-core-go/agent/scoring"

score := scoring.NewScore("persona", "scenario")
score.Completeness.Rating = 3
score.LintQuality.Rating = 2
score.OutputValidity.Rating = 3
score.QuestionEfficiency.Rating = 2
// score.Total(), score.Threshold()
```

### Orchestrator

Coordinates DeveloperAgent and RunnerAgent conversation:

```go
import "github.com/lex00/wetwire-core-go/agent/orchestrator"

orch := orchestrator.New(orchestrator.Config{
    Domain:    "aws",
    Developer: developerAgent,
    Runner:    runnerAgent,
})
result, err := orch.Run(ctx, initialPrompt)
```

### Results

Session tracking and RESULTS.md generation:

```go
import "github.com/lex00/wetwire-core-go/agent/results"

session := results.NewSession("aws", "my_stack", "Create S3 bucket")
// ... run agent workflow ...
session.Complete()

writer := results.NewWriter()
writer.Write(session, "./output/RESULTS.md")
```

## Integration Pattern

Domain packages (wetwire-aws-go, etc.) integrate wetwire-core-go via:

1. Import agents and orchestrator packages
2. Define domain-specific tools (init_package, write_file, run_lint, run_build)
3. Configure RunnerAgent with domain tools
4. Use orchestrator for design/test commands

## Key Principles

1. **Two-agent model** — Developer asks, Runner generates
2. **Lint enforcement** — RunnerAgent must lint after every write
3. **Pass before done** — Code must pass linting before completion
4. **Persona-based testing** — Test across all 5 developer styles

## Running Tests

```bash
go test -v ./...
```
