---
title: "FAQ"
---

<details>
<summary>What is wetwire-core-go?</summary>

wetwire-core-go is the shared agent infrastructure used by all wetwire domain packages. It provides:

- AI provider abstraction (Anthropic, Claude Code, Kiro)
- MCP server for Claude Code integration
- Persona-based testing framework
- Scenario execution and validation
- Session tracking and results generation
</details>

<details>
<summary>How do I add wetwire-core-go to my domain package?</summary>

```bash
go get github.com/lex00/wetwire-core-go
```

Then implement the `domain.Domain` interface:

```go
import "github.com/lex00/wetwire-core-go/domain"

type MyDomain struct{}
var _ domain.Domain = (*MyDomain)(nil)

func (d *MyDomain) Name() string { return "mydomain" }
// ... implement other required methods
```
</details>

<details>
<summary>What providers are supported?</summary>

| Provider | Use Case | API Key Required |
|----------|----------|------------------|
| Anthropic | Direct API access | Yes |
| Claude | Claude Code CLI integration | No |
| Kiro | Enterprise environments | No (uses kiro-cli) |
</details>

<details>
<summary>How do I create a custom provider?</summary>

Implement the `providers.Provider` interface:

```go
type Provider interface {
    CreateMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error)
    StreamMessage(ctx context.Context, req MessageRequest, handler StreamHandler) error
}
```

Then register it with your agent configuration:

```go
provider := &MyCustomProvider{}
agent, _ := agents.NewAgent(agents.AgentConfig{
    Provider: provider,
})
```
</details>

<details>
<summary>How do personas affect agent behavior?</summary>

Personas simulate different developer skill levels during scenario testing:

- **Beginner** — Uncertain, asks many questions, needs guidance
- **Intermediate** — Some knowledge, specifies requirements but may miss details
- **Expert** — Deep knowledge, precise requirements, minimal hand-holding

Custom personas can be registered for domain-specific testing:

```go
personas.Register(personas.Persona{
    Name:         "security-auditor",
    Description:  "Security-focused reviewer",
    SystemPrompt: "You are a security auditor...",
})
```
</details>

<details>
<summary>How does scoring work?</summary>

Scenarios are scored on a 0-12 scale across 4 dimensions:

| Dimension | Points | Description |
|-----------|--------|-------------|
| Completeness | 0-3 | Were all required resources generated? |
| Lint Quality | 0-3 | How many lint cycles needed? |
| Output Validity | 0-3 | Does generated output validate? |
| Question Efficiency | 0-3 | Appropriate number of clarifying questions? |
</details>

<details>
<summary>What's the recommended way to integrate with domain packages?</summary>

Domain packages should implement the `domain.Domain` interface:

```go
type MyDomain struct{}
var _ domain.Domain = (*MyDomain)(nil)

func (d *MyDomain) Name() string    { return "mydomain" }
func (d *MyDomain) Version() string { return "1.0.0" }
func (d *MyDomain) Builder() domain.Builder { return &MyBuilder{} }
// ... other required methods
```

Then use `domain.Run()` to generate CLI and `domain.BuildMCPServer()` for MCP integration:

```go
func main() {
    cli := domain.Run(&MyDomain{})
    cli.Execute()
}
```
</details>

<details>
<summary>How do I run scenario tests?</summary>

```bash
# Run with specific persona
go run ./cmd/run_scenario ./examples/my_scenario beginner ./results

# Run all personas
go run ./cmd/run_scenario ./examples/my_scenario --all ./results

# Validate results
go run ./cmd/validate_scenario ./examples/my_scenario ./results
```
</details>
