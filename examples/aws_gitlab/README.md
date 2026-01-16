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
├── scenario.yaml       # Scenario definition
├── prompt.md           # Default user prompt
├── prompts/
│   ├── beginner.md     # Detailed explanations for newcomers
│   ├── intermediate.md # Standard instructions
│   ├── expert.md       # Brief, assumes knowledge
│   ├── terse.md        # Minimal words
│   └── verbose.md      # Highly detailed requirements
└── recordings/
    └── aws_gitlab_demo.svg  # Animated demo
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

### Load a Specific Persona

```go
config, _ := scenario.Load("./examples/aws_gitlab")

// Load beginner prompt
prompt, _ := config.GetPrompt("beginner")

// Or load expert prompt
prompt, _ := config.GetPrompt("expert")
```

### Execute with Unified Agent

```go
import (
    "github.com/lex00/wetwire-core-go/agent/agents"
    "github.com/lex00/wetwire-core-go/mcp"
    "github.com/lex00/wetwire-core-go/providers/anthropic"
)

// Create MCP server with AWS/GitLab tools
mcpServer := mcp.NewServer(mcp.Config{Name: "aws-gitlab"})
mcp.RegisterStandardToolsWithDefaults(mcpServer, "aws", handlers)

// Create provider
provider, _ := anthropic.New(anthropic.Config{})

// Create agent
agent, _ := agents.NewAgent(agents.AgentConfig{
    Provider:  provider,
    MCPServer: agents.NewMCPServerAdapter(mcpServer),
    SystemPrompt: systemPrompt,
})

// Run with chosen persona
prompt, _ := config.GetPrompt("intermediate")
agent.Run(ctx, prompt)
```

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

## See Also

- [unified_agent example](../unified_agent/) - Agent architecture pattern
- [mcp_server example](../mcp_server/) - MCP server creation
