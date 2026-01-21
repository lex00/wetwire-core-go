# CLI Reference

wetwire-core-go provides command-line tools for scenario testing and validation. These tools are used during development and CI to test AI agent interactions across different personas.

## Quick Reference

| Command | Description |
|---------|-------------|
| `init_scenario` | Scaffold a new scenario with required files |
| `run_scenario` | Execute a scenario using Claude Code |
| `validate_scenario` | Validate scenario results against rules |
| `record_example` | Generate SVG recordings of sessions |

---

## init_scenario

Scaffold a new wetwire scenario with all required files.

```bash
go run ./cmd/init_scenario <path> <description>
```

### Examples

```bash
go run ./cmd/init_scenario ./examples/eks_cluster "EKS cluster with ArgoCD"
go run ./cmd/init_scenario ./examples/lambda_api "Lambda API Gateway"
```

### Generated Files

| File | Purpose |
|------|---------|
| `scenario.yaml` | Scenario configuration |
| `system_prompt.md` | Domain-specific system prompt |
| `prompts/beginner.md` | Beginner persona prompt |
| `prompts/intermediate.md` | Intermediate persona prompt |
| `prompts/expert.md` | Expert persona prompt |

---

## run_scenario

Execute a wetwire scenario using Claude Code as the AI backend.

```bash
go run ./cmd/run_scenario [scenario_path] [persona] [output_dir] [flags]
```

### Options

| Option | Description |
|--------|-------------|
| `scenario_path` | Path to scenario directory |
| `persona` | Persona to run (beginner, intermediate, expert) |
| `output_dir` | Directory for results (default: ./results) |
| `--all` | Run all personas |
| `--verbose` | Show streaming output from Claude |
| `--record` | Generate SVG recordings (requires termsvg) |

### Examples

```bash
# Run with default persona
go run ./cmd/run_scenario ./examples/aws_gitlab

# Run specific persona
go run ./cmd/run_scenario ./examples/aws_gitlab beginner

# Run with output directory
go run ./cmd/run_scenario ./examples/aws_gitlab expert ./results

# Run all personas
go run ./cmd/run_scenario ./examples/aws_gitlab --all ./results

# Run with verbose output
go run ./cmd/run_scenario ./examples/aws_gitlab --verbose
```

### Output Files

| File | Content |
|------|---------|
| `RESULTS.md` | Human-readable summary |
| `session.json` | Complete session data |
| `score.json` | Score breakdown |

---

## validate_scenario

Validate scenario results against defined validation rules and expected files.

```bash
go run ./cmd/validate_scenario [scenario_path] [results_dir] [persona] [flags]
```

### Options

| Option | Description |
|--------|-------------|
| `scenario_path` | Path to scenario directory |
| `results_dir` | Directory containing results |
| `persona` | Specific persona results to validate |
| `--markdown` | Output report in markdown format |
| `--json` | Output report in JSON format |
| `--quiet` | Only output pass/fail status |

### Examples

```bash
# Validate default results
go run ./cmd/validate_scenario ./examples/honeycomb_k8s

# Validate specific results directory
go run ./cmd/validate_scenario ./examples/honeycomb_k8s ./results/intermediate

# Validate specific persona
go run ./cmd/validate_scenario ./examples/honeycomb_k8s intermediate

# Generate markdown report
go run ./cmd/validate_scenario ./examples/honeycomb_k8s --markdown > VALIDATION.md
```

### Validation Rules

Scenarios define validation rules in `scenario.yaml`:

```yaml
validation:
  required_files:
    - "network.go"
    - "compute.go"
  must_contain:
    - pattern: "var.*Deployment"
      file: "*.go"
  must_not_contain:
    - pattern: "TODO"
      file: "*.go"
```

---

## record_example

Generate animated SVG recordings of agent sessions.

```bash
go run ./cmd/record_example
```

Recordings are saved to the scenario's `recordings/` directory.

### Requirements

- termsvg must be installed for recording

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `ANTHROPIC_API_KEY` | API key for Anthropic provider |
| `WETWIRE_VERBOSE` | Enable verbose logging |
| `WETWIRE_OUTPUT_DIR` | Default output directory |

---

## Resources

- [Scenario Specification](SCENARIO_SPEC.md)
- [Personas](../CLAUDE.md#personas)
- [Scoring](FAQ.md#scoring)
