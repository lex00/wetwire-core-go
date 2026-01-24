---
title: "CLI"
slug: cli
---

wetwire-core-go provides command-line tools for scenario testing and validation.

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

# Run all personas
go run ./cmd/run_scenario ./examples/aws_gitlab --all ./results
```

---

## validate_scenario

Validate scenario results against defined validation rules.

```bash
go run ./cmd/validate_scenario [scenario_path] [results_dir] [persona] [flags]
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
```

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `ANTHROPIC_API_KEY` | API key for Anthropic provider |
| `WETWIRE_VERBOSE` | Enable verbose logging |
| `WETWIRE_OUTPUT_DIR` | Default output directory |
