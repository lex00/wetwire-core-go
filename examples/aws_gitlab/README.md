# AWS + GitLab Example Scenario

This example demonstrates a multi-domain scenario that generates an S3 bucket CloudFormation template alongside a GitLab CI/CD pipeline for publishing it.

## Overview

The scenario creates:

**AWS (CloudFormation)**
- S3 bucket template with versioning, encryption, and private access

**GitLab (CI/CD)**
- Pipeline that validates and publishes the template (does NOT execute it)

## Files

```
aws_gitlab/
├── README.md           # This file
├── go.mod              # Go module (standalone)
├── main.go             # CLI entry point
├── scenario.yaml       # Scenario definition
├── prompt.md           # Default user prompt
├── system_prompt.md    # System prompt for the agent
├── prompts/
│   ├── beginner.md     # Detailed explanations for newcomers
│   ├── intermediate.md # Standard instructions
│   ├── expert.md       # Brief, assumes knowledge
│   ├── terse.md        # Minimal words
│   └── verbose.md      # Highly detailed requirements
└── results/            # Generated results (gitignored SVGs)
```

## Personas

The scenario includes prompts for different developer personas:

| Persona | Description |
|---------|-------------|
| `beginner` | New to AWS/GitLab, needs explanations |
| `intermediate` | Knows the basics, wants clear structure |
| `expert` | Experienced, prefers concise instructions |
| `terse` | Minimal words, just the essentials |
| `verbose` | Comprehensive requirements with context |

## Running the Scenario

### Prerequisites

- Go 1.23+
- [Claude Code CLI](https://github.com/anthropics/claude-code) installed and authenticated

### Quick Start

```bash
cd examples/aws_gitlab

# Run single persona (default: intermediate)
go run . --persona intermediate --verbose

# Run all 5 personas in parallel
go run . --all

# Specify output directory
go run . --all --output ./my-results
```

### CLI Options

| Flag | Default | Description |
|------|---------|-------------|
| `--persona` | `intermediate` | Persona to run (beginner, intermediate, expert, terse, verbose) |
| `--all` | false | Run all 5 personas in parallel |
| `--verbose` | false | Show streaming output |
| `--output` | `./results` | Output directory for results |

### Results

Results are saved per persona:

```
results/
├── SUMMARY.md              # Comparison table across all personas
├── beginner/
│   ├── RESULTS.md          # Score and file list
│   ├── conversation.txt    # Full prompt and response
│   ├── cfn-templates/
│   │   └── s3-bucket.yaml  # Generated CloudFormation template
│   └── .gitlab-ci.yml      # Generated GitLab pipeline
├── intermediate/
│   └── ...
└── ...
```

### Scoring

Each persona is scored on 4 dimensions (0-12 scale):

| Dimension | Description |
|-----------|-------------|
| Completeness | Were all required files generated? |
| Lint Quality | Deferred to domain tools |
| Output Validity | Are outputs well-formed? |
| Question Efficiency | Appropriate clarifying questions? |

## Cross-Domain Validation

The scenario validates that the GitLab pipeline correctly references AWS outputs:

```yaml
cross_domain:
  - from: aws
    to: gitlab
    type: artifact_reference
    validation:
      required_refs:
        - "${aws.s3.outputs.bucket_name}"
```

This ensures the GitLab pipeline references the target S3 bucket for publishing templates.

## Expected Output

When complete, the scenario generates:

```
output/
├── cfn-templates/
│   └── s3-bucket.yaml     # CloudFormation template
└── .gitlab-ci.yml          # GitLab pipeline (validate + publish)
```

## Dependencies

This example uses [wetwire-core-go](https://github.com/lex00/wetwire-core-go) for the scenario runner. The `go.mod` includes a replace directive for local development.
