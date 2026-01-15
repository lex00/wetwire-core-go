# AWS + GitLab Example Scenario

This example demonstrates a multi-domain scenario that generates AWS infrastructure alongside a GitLab CI/CD pipeline.

## Demo Recording

![Demo](recordings/aws_gitlab_demo.svg)

The animated recording shows a developer/agent conversation creating AWS infrastructure and a GitLab pipeline.

## Overview

The scenario creates:

**AWS (CloudFormation)**
- VPC with public/private subnets across multiple AZs
- EKS managed Kubernetes cluster
- RDS PostgreSQL database

**GitLab (CI/CD)**
- Pipeline that deploys AWS stacks in dependency order
- Integration tests and manual approval stages
- References to AWS stack outputs

## Files

```
aws_gitlab/
├── README.md           # This file
├── scenario.yaml       # Scenario definition
├── prompt.md           # Default user prompt
├── prompts/
│   └── minimal.md      # Minimal variant
└── recordings/
    └── aws_gitlab_demo.svg  # Animated demo
```

## Running the Scenario

### Load and Validate

```go
import "github.com/lex00/wetwire-core-go/scenario"

config, err := scenario.Load("./examples/aws_gitlab")
if err != nil {
    log.Fatal(err)
}

result := scenario.Validate(config)
if !result.IsValid() {
    log.Fatal(result.Error())
}
```

### Execute with Orchestrator

```go
import "github.com/lex00/wetwire-core-go/agent/orchestrator"

orch := orchestrator.New(config, developer, runner)
session, err := orch.Run(ctx)
```

The orchestrator handles:
1. Domain ordering (AWS before GitLab due to dependency)
2. Developer/Runner conversation coordination
3. Cross-domain validation

## Recording the Conversation

After running the scenario, you can record the conversation as an animated SVG:

```go
import "github.com/lex00/wetwire-core-go/scenario"

// Adapt your session to the SessionMessages interface
adapter := &SessionAdapter{session: session}

err := scenario.RecordSession(adapter, scenario.SessionRecordOptions{
    OutputDir:  "./recordings",
    TermWidth:  80,
    TermHeight: 30,
})
// Creates: ./recordings/<session-name>.svg
```

The recording shows:
- Developer messages with typing simulation (green text)
- Runner responses appearing line-by-line (white text)
- Black terminal background

## Prompt Variants

Use the `minimal` variant for simpler output:

```go
config, _ := scenario.Load("./examples/aws_gitlab")
prompt, _ := config.GetPrompt("minimal")
```

## Cross-Domain Validation

The scenario validates that GitLab pipelines correctly reference AWS outputs:

```yaml
cross_domain:
  - from: aws
    to: gitlab
    type: artifact_reference
    validation:
      required_refs:
        - "${aws.vpc.outputs.vpc_id}"
        - "${aws.eks.outputs.cluster_name}"
        - "${aws.rds.outputs.endpoint}"
```

This ensures the GitLab pipeline can't be generated without proper AWS stack references.

## Expected Output

When complete, the scenario generates:

```
output/
├── cfn-templates/
│   ├── vpc.yaml
│   ├── eks.yaml
│   └── rds.yaml
├── .gitlab-ci.yml
└── deploy/
    └── stages.yaml
```
