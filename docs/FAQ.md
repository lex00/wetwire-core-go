# wetwire-core-go FAQ

This FAQ covers questions specific to the Go agent infrastructure. For general wetwire questions, see the [central FAQ](https://github.com/lex00/wetwire/blob/main/docs/FAQ.md).

---

## General

### What is wetwire-core-go?

wetwire-core-go provides the agent infrastructure for AI-assisted infrastructure generation:
- Developer and Runner agent implementations
- Personas for testing
- Scoring and evaluation
- Session tracking and results

### How does it relate to domain packages?

Domain packages (like wetwire-aws-go) use wetwire-core-go for `design` and `test` commands. The core package provides the agent orchestration; domain packages provide the tools and resources.

---

## Agents

### What is the two-agent model?

1. **DeveloperAgent** — Represents the human (or simulates via persona)
2. **RunnerAgent** — Generates infrastructure code using domain CLI tools

### What tools does the RunnerAgent have?

| Tool | Purpose |
|------|---------|
| `init_package` | Create new project directory |
| `write_file` | Write source file |
| `read_file` | Read file contents |
| `run_lint` | Execute linter |
| `run_build` | Build output template |
| `ask_developer` | Query developer for clarification |

### What are the enforcement rules?

1. **Lint After Write** — After `write_file`, agent must call `run_lint`
2. **Pass Before Done** — Agent cannot complete until lint passes

---

## Personas

### What personas are available?

| Name | Behavior |
|------|----------|
| **Beginner** | Uncertain, asks many questions |
| **Intermediate** | Some knowledge, may miss details |
| **Expert** | Precise requirements, minimal hand-holding |
| **Terse** | Minimal information, expects inference |
| **Verbose** | Over-explains, buries requirements |

### How do I use personas?

```go
import "github.com/lex00/wetwire-core-go/agent/personas"

persona, err := personas.Get("beginner")
if err != nil {
    log.Fatal(err)
}
```

### Can I create custom personas?

Yes. Create a `Persona` struct with name, description, system prompt, and traits.

---

## Scoring

### How is scoring calculated?

4 dimensions, 0-3 points each (12 total):

| Dimension | Measures |
|-----------|----------|
| Completeness | Were all resources generated? |
| Lint Quality | How many lint cycles needed? |
| Output Validity | Does output validate? |
| Question Efficiency | Appropriate question count? |

### What scores are passing?

| Score | Grade |
|-------|-------|
| 0-4 | Failure |
| 5-7 | Partial |
| 8-10 | Success |
| 11-12 | Excellent |

Minimum passing score is 5.

**Note:** LLM outputs are non-deterministic. Scores may vary between runs.

---

## Orchestrator

### What does the Orchestrator do?

Coordinates the Developer and Runner agents:
1. Initializes session
2. Runs conversation loop
3. Enforces rules (lint after write, pass before done)
4. Calculates score
5. Generates results

### How do I use the Orchestrator?

```go
import "github.com/lex00/wetwire-core-go/agent/orchestrator"

config := orchestrator.DefaultConfig()
orch := orchestrator.New(config, developer, runner)
session, err := orch.Run(ctx)
```

---

## Results

### What output files are generated?

| File | Content |
|------|---------|
| `RESULTS.md` | Human-readable summary |
| `session.json` | Complete session data |
| `score.json` | Score breakdown |

### Where are results saved?

By default in the output directory specified in the config. Usually `./output/` or a timestamped directory.

---

## Integration

### How do domain packages integrate?

Domain packages use the RunnerAgent with their CLI on PATH. The agent internally calls the domain CLI (e.g., `wetwire-aws lint`, `wetwire-aws build`):

```go
// In domain package (e.g., wetwire-aws-go cmd/wetwire-aws/design.go)
import "github.com/lex00/wetwire-core-go/agent/agents"

config := agents.RunnerConfig{
    WorkDir:       outputDir,
    MaxLintCycles: maxLintCycles,
    Session:       session,
    Developer:     developer,
    StreamHandler: streamFunc, // Optional
}
runner, err := agents.NewRunnerAgent(config)
```

The RunnerAgent executes `wetwire-aws lint` and `wetwire-aws build` commands internally, so the domain CLI must be installed and on PATH.

---

## Resources

- [Wetwire Specification](https://github.com/lex00/wetwire/blob/main/docs/WETWIRE_SPEC.md)
- [wetwire-aws-go](https://github.com/lex00/wetwire-aws-go) (example integration)
