# Scenario Specification

This document defines the required structure and conventions for wetwire scenarios.
All domain projects should follow this specification to ensure consistency.

## Directory Structure

```
examples/<scenario_name>/
├── scenario.yaml          # Required: Scenario metadata and configuration
├── system_prompt.md       # Required: Agent system prompt
├── prompt.md              # Required: Default user prompt
├── prompts/               # Required: Persona-specific prompts
│   ├── beginner.md
│   ├── intermediate.md
│   └── expert.md
├── expected/              # Optional: Golden files for scoring comparison
│   └── <expected_outputs>
├── results/               # Generated: Scenario run outputs (gitignored)
└── .gitignore             # Required: Ignore results/ and *.svg
```

## Required Files

### scenario.yaml

Defines scenario metadata and expected outputs.

```yaml
name: scenario_name           # Unique identifier (snake_case)
description: Brief description of what this scenario creates

prompts:
  default: prompt.md
  variants:
    beginner: prompts/beginner.md
    intermediate: prompts/intermediate.md
    expert: prompts/expert.md

domains:
  - name: aws                 # Domain identifier
    outputs:                  # Expected output files
      - path/to/output.yaml

validation:                   # Optional: Validation rules
  aws:
    resources:
      min: 1
```

### system_prompt.md

Instructions for the agent. Should be domain-agnostic where possible.

```markdown
You are a helpful infrastructure engineer assistant.

Your task is to help users create infrastructure files based on their requirements.
Use the Write tool to create files. Use mkdir via Bash if needed.

Guidelines:
- Always generate complete, production-quality infrastructure
- If the user asks questions, answer them
- Include best practices even if not explicitly requested
```

### prompt.md (Default Prompt)

The default user prompt used when no persona is specified.

```markdown
# Scenario Title

Brief description of what the user wants.

## Requirements

- Requirement 1
- Requirement 2

## Expected Outputs

- output1.yaml
- output2.yaml
```

### Persona Prompts

Each persona represents a different user communication style. The prompts should
request the **same outcome** but with different levels of detail and expertise.

| Persona | Description | Prompt Style |
|---------|-------------|--------------|
| `beginner` | New to the domain, needs guidance | Asks questions, requests explanations |
| `intermediate` | Knows basics, wants structure | Clear requirements, some context |
| `expert` | Deep knowledge, concise | Brief, technical, assumes knowledge |

Custom personas can be registered for domain-specific testing using `personas.Register()`.

#### beginner.md Example

```markdown
I'm new to [domain]. Please help me create:

1. [Output 1] - I think I need this for [reason]?
2. [Output 2]

I want:
- [Feature 1] (not sure how this works)
- [Feature 2]

Please explain what each part does.

## Questions I have

- How does X work?
- What's the difference between Y and Z?
```

### .gitignore

```
# Scenario run outputs
results/

# SVG recordings
*.svg
```

## Optional Files

### expected/

Golden/reference files for scoring comparison. Place expected outputs here
to enable more accurate quality scoring.

```
expected/
├── output1.yaml
└── output2.yaml
```

## Running Scenarios

### Using the Runner Package

Domain projects should use the shared runner:

```go
import "github.com/lex00/wetwire-core-go/scenario/runner"

results, err := runner.Run(ctx, runner.Config{
    ScenarioPath: "./examples/my_scenario",
    OutputDir:    "./examples/my_scenario/results",
})
```

### Using the CLI

```bash
# Run single persona
go run ./cmd/run_scenario ./examples/my_scenario

# Run all personas
go run ./cmd/run_scenario ./examples/my_scenario --all

# With recordings
go run ./cmd/run_scenario ./examples/my_scenario --all --record
```

## Output Structure

After running, the results directory contains:

```
results/
├── SUMMARY.md              # Overall results table
├── beginner/
│   ├── RESULTS.md          # Score and file links
│   ├── conversation.txt    # Full prompt/response
│   ├── <generated_files>   # Files created by agent
│   └── *_scenario.svg      # Recording (if --record)
├── intermediate/
└── expert/
```

## Scoring

Scenarios are scored on 4 dimensions (0-3 each, max 12):

| Dimension | Description |
|-----------|-------------|
| Completeness | Were all expected files created? |
| Lint Quality | Does the output pass linting? |
| Output Validity | Is the output valid/deployable? |
| Question Efficiency | Appropriate clarification? |

Thresholds:
- 11-12: Excellent
- 8-10: Success
- 5-7: Partial
- 0-4: Failure

**Note:** LLM outputs are non-deterministic. Scores may vary between runs even with identical inputs. For reliable baselines, run scenarios multiple times and track score distributions rather than single values.

## Validation

Use `scenario.ValidateStructure()` in tests:

```go
func TestScenarioStructure(t *testing.T) {
    err := scenario.ValidateStructure("./examples/my_scenario")
    if err != nil {
        t.Errorf("Invalid scenario structure: %v", err)
    }
}
```

## Checklist for New Scenarios

- [ ] Created `scenario.yaml` with name, description, outputs
- [ ] Created `system_prompt.md` with agent instructions
- [ ] Created `prompt.md` with default requirements
- [ ] Created all 3 persona prompts in `prompts/`
- [ ] Created `.gitignore` for results/ and *.svg
- [ ] All prompts request the same outcome (different styles)
- [ ] Ran with `--all` to verify all personas produce valid output
- [ ] Ran multiple times to establish score variance per persona
