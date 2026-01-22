<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./wetwire-dark.svg">
  <img src="./wetwire-light.svg" width="100" height="67">
</picture>

wetwire-core-go is an infrastructure library, not a domain package. It does not define domain-specific lint rules like `WAW` (AWS) or `WK8` (Kubernetes).

Instead, wetwire-core-go provides:

1. **Scoring Dimensions** - Evaluation criteria for agent output
2. **Validation Rules** - Scenario validation patterns
3. **Enforcement Rules** - Agent behavior constraints

---

## Scoring Dimensions

Agent output is evaluated on 4 dimensions (0-3 points each, 12 total):

| Code | Dimension | Description |
|------|-----------|-------------|
| `SC01` | Completeness | Were all required resources generated? |
| `SC02` | Lint Quality | How many lint cycles were needed? |
| `SC03` | Output Validity | Does generated output validate? |
| `SC04` | Question Efficiency | Were questions appropriate and minimal? |

### Scoring Thresholds

| Score | Grade | Meaning |
|-------|-------|---------|
| 0-4 | Failure | Output unusable |
| 5-7 | Partial | Output needs significant fixes |
| 8-10 | Success | Output usable with minor tweaks |
| 11-12 | Excellent | Output production-ready |

---

## Enforcement Rules

The orchestrator enforces these rules during agent execution:

| Code | Rule | Description |
|------|------|-------------|
| `EN01` | Lint After Write | Agent must call `run_lint` after every `write_file` |
| `EN02` | Pass Before Done | Agent cannot complete until lint passes |
| `EN03` | Max Lint Cycles | Agent fails if lint cycles exceed threshold |

---

## Validation Patterns

Scenarios define validation rules in `scenario.yaml`:

### Required Files (`VR01`)

```yaml
validation:
  required_files:
    - "network.go"
    - "compute.go"
```

Fails if any listed file is missing from output.

### Must Contain (`VR02`)

```yaml
validation:
  must_contain:
    - pattern: "var.*Deployment"
      file: "*.go"
```

Fails if pattern is not found in matching files.

### Must Not Contain (`VR03`)

```yaml
validation:
  must_not_contain:
    - pattern: "TODO"
      file: "*.go"
```

Fails if pattern is found in matching files.

### Build Success (`VR04`)

```yaml
validation:
  must_build: true
```

Fails if `go build ./...` fails on output.

### Lint Clean (`VR05`)

```yaml
validation:
  must_lint_clean: true
```

Fails if domain linter reports errors.

---

## Domain Lint Rules

Domain packages define their own lint rules with unique prefixes:

| Prefix | Domain | Example |
|--------|--------|---------|
| `WAW` | wetwire-aws-go | `WAW001`: No pointers in declarations |
| `WK8` | wetwire-k8s-go | `WK8001`: Label selector must match template |
| `WAZ` | wetwire-azure-go | `WAZ001`: Resource naming constraints |
| `WGH` | wetwire-github-go | `WGH001`: Workflow naming conventions |
| `WGL` | wetwire-gitlab-go | `WGL001`: Pipeline structure rules |
| `WHC` | wetwire-honeycomb-go | `WHC001`: Query validation |
| `WN4` | wetwire-neo4j-go | `WN4001`: Node label conventions |
| `WOB` | wetwire-observability-go | `WOB001`: PromQL validation |

See individual domain package documentation for complete lint rule references.

---

## Resources

- [Scenario Specification](SCENARIO_SPEC.md)
- [Wetwire Specification](https://github.com/lex00/wetwire/blob/main/docs/WETWIRE_SPEC.md)
