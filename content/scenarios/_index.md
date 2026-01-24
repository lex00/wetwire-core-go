---
title: "Scenarios"
---

Scenarios are structured test cases for AI agent interactions.

## Scenario Structure

```
my_scenario/
├── scenario.yaml       # Configuration and validation rules
├── system_prompt.md    # Domain-specific system prompt
├── prompts/
│   ├── beginner.md     # Beginner persona prompt
│   ├── intermediate.md # Intermediate persona prompt
│   └── expert.md       # Expert persona prompt
└── expected/           # Expected output structure (optional)
```

## scenario.yaml

```yaml
name: my_scenario
description: "Create infrastructure with specific requirements"
domains:
  - aws
  - k8s

validation:
  required_files:
    - "network.go"
    - "compute.go"
  must_contain:
    - pattern: "var.*VPC"
      file: "*.go"
  must_not_contain:
    - pattern: "TODO"
      file: "*.go"
```

## Multi-Domain Scenarios

Scenarios can span multiple domains:

```yaml
domains:
  - aws       # CloudFormation infrastructure
  - k8s       # Kubernetes workloads
  - gitlab    # CI/CD pipeline
```

Each domain contributes its resources to the final output.

## Running Scenarios

```bash
# Single persona
go run ./cmd/run_scenario ./examples/my_scenario beginner

# All personas
go run ./cmd/run_scenario ./examples/my_scenario --all ./results
```

## Output Structure

```
results/
├── SUMMARY.md           # Results table with all personas
├── beginner/
│   ├── RESULTS.md       # Human-readable summary
│   ├── session.json     # Complete session data
│   └── generated/       # Generated files
├── intermediate/
└── expert/
```

## Validation

```bash
go run ./cmd/validate_scenario ./examples/my_scenario ./results
```

Checks:
- Required files exist
- Pattern matching rules
- Lint passes on generated code
- Output validates with domain validators
