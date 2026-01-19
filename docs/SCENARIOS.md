# Scenarios

Scenarios define multi-domain infrastructure generation workflows with cross-domain validation.

## Overview

A scenario orchestrates multiple wetwire domain packages (e.g., AWS + GitLab) to generate related infrastructure that works together. The scenario system handles:

- **Domain ordering** - Respects dependencies between domains
- **Cross-domain validation** - Ensures outputs from one domain are correctly referenced by another
- **Conversation capture** - Records the full user/agent interaction for replay and analysis

## Scenario Structure

A scenario lives in a directory with:

```
my-scenario/
├── scenario.yaml      # Scenario definition
├── prompt.md          # Default user prompt
└── prompts/           # Optional prompt variants
    └── minimal.md
```

## scenario.yaml

```yaml
name: aws_gitlab_deployment
description: AWS infrastructure with GitLab CI/CD pipeline

prompts:
  default: prompt.md
  variants:
    minimal: prompts/minimal.md

domains:
  - name: aws
    cli: wetwire-aws
    mcp_tools:
      lint: wetwire_lint
      build: wetwire_build
    outputs:
      - cfn-templates/*.yaml

  - name: gitlab
    cli: wetwire-gitlab
    depends_on:
      - aws
    outputs:
      - .gitlab-ci.yml

cross_domain:
  - from: aws
    to: gitlab
    type: artifact_reference
    validation:
      required_refs:
        - "${aws.vpc.outputs.vpc_id}"
        - "${aws.eks.outputs.cluster_name}"

validation:
  aws:
    stacks:
      min: 3
      max: 10
    resources:
      min: 5
  gitlab:
    pipelines:
      min: 1
```

## Loading Scenarios

```go
import "github.com/lex00/wetwire-core-go/scenario"

// Load from directory
config, err := scenario.Load("./my-scenario")

// Or load from file directly
config, err := scenario.LoadFile("./my-scenario/scenario.yaml")
```

## Domain Ordering

The scenario system automatically determines execution order based on `depends_on` relationships:

```go
order, err := scenario.GetDomainOrder(config)
// Returns: ["aws", "gitlab"] (aws first since gitlab depends on it)
```

Circular dependencies are detected and return an error.

## Validation

```go
result := scenario.Validate(config)
if !result.IsValid() {
    fmt.Println(result.Error())
}
```

Validation checks:
- Required fields (name, domains)
- Domain names are unique
- Dependencies reference existing domains
- No circular dependencies
- Cross-domain references are valid

## Conversation Flow

When a scenario runs through the orchestrator, the conversation is captured:

```go
// The orchestrator coordinates Developer (user) and Runner (agent)
orchestrator := orchestrator.New(config, developer, runner)
session, err := orchestrator.Run(ctx)

// Session contains the full conversation
for _, msg := range session.Messages {
    fmt.Printf("[%s]: %s\n", msg.Role, msg.Content)
}
```

The conversation flow:
1. **Developer** provides initial prompt
2. **Runner** (agent) works on the request
3. **Runner** may ask clarifying questions
4. **Developer** answers
5. **Runner** continues until complete

The conversation supports clarifying questions - when requirements are ambiguous, the agent may generate questions, and the developer (human or simulated persona) responds.

## Recording Conversations

Captured sessions can be recorded as animated SVGs. See [RECORDING.md](RECORDING.md) for details.

## Example

See [`examples/aws_gitlab/`](../examples/aws_gitlab/) for a complete multi-domain scenario with:

- AWS VPC, EKS, and RDS CloudFormation templates
- GitLab CI/CD pipeline with deployment stages
- Cross-domain validation ensuring pipeline references AWS outputs
- Animated demo recording of the developer/agent conversation
