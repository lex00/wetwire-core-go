# wetwire-core-go

[![Go Reference](https://pkg.go.dev/badge/github.com/lex00/wetwire-core-go.svg)](https://pkg.go.dev/github.com/lex00/wetwire-core-go)
[![CI](https://github.com/lex00/wetwire-core-go/actions/workflows/ci.yml/badge.svg)](https://github.com/lex00/wetwire-core-go/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/lex00/wetwire-core-go)](https://goreportcard.com/report/github.com/lex00/wetwire-core-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Shared agent infrastructure for wetwire domain packages.

## Overview

wetwire-core-go provides the AI agent framework used by wetwire domain packages (like wetwire-aws-go). It includes:

- **personas** - Developer persona definitions (Beginner, Intermediate, Expert, Terse, Verbose)
- **scoring** - 5-dimension evaluation rubric (0-15 scale)
- **results** - Session tracking and RESULTS.md generation
- **orchestrator** - Developer/Runner agent coordination
- **agents** - Anthropic SDK integration and RunnerAgent
- **version** - Version info exposure via runtime/debug for dependent packages
- **cmd** - CLI command framework with cobra for consistent CLIs across domain packages
- **serialize** - Struct-to-map conversion and JSON/YAML serialization with naming conventions
- **lsp** - Language Server Protocol infrastructure for IDE integration

## Installation

```bash
go get github.com/lex00/wetwire-core-go
```

## Usage

wetwire-core-go is typically used as a dependency of domain packages like wetwire-aws-go.

### Personas

```go
import "github.com/lex00/wetwire-core-go/agent/personas"

// Get a built-in persona
persona, err := personas.Get("beginner")
if err != nil {
    log.Fatal(err)
}
fmt.Println(persona.Name, persona.Description)
```

### RunnerAgent

```go
import "github.com/lex00/wetwire-core-go/agent/agents"

config := agents.RunnerConfig{
    WorkDir:       "./output",
    MaxLintCycles: 3,
    Session:       session,        // Optional: for result tracking
    Developer:     developer,      // Required: responds to questions
    StreamHandler: func(text string) { fmt.Print(text) }, // Optional
}

runner, err := agents.NewRunnerAgent(config)
if err != nil {
    log.Fatal(err)
}

err = runner.Run(ctx, "Create an S3 bucket with versioning")
```

### Session Results

```go
import "github.com/lex00/wetwire-core-go/agent/results"

session := results.NewSession("aws", "my_bucket", "Create a bucket")
// ... run agent workflow ...
session.Complete()

writer := results.NewResultsWriter()
writer.Write(session, "./output/RESULTS.md")
```

### Version

```go
import "github.com/lex00/wetwire-core-go/version"

// Get the module version (returns "dev" for local builds)
v := version.Version()
fmt.Println("wetwire-core-go version:", v)
```

### CLI Commands

```go
import "github.com/lex00/wetwire-core-go/cmd"

func main() {
    root := cmd.NewRootCommand("wetwire-aws", "AWS infrastructure synthesis")
    root.AddCommand(cmd.NewBuildCommand(myBuilder))
    root.AddCommand(cmd.NewLintCommand(myLinter))
    root.AddCommand(cmd.NewInitCommand(myInitializer))
    root.AddCommand(cmd.NewValidateCommand(myValidator))
    root.Execute()
}
```

### Serialization

```go
import "github.com/lex00/wetwire-core-go/serialize"

m := serialize.ToMap(resource, serialize.SnakeCase, serialize.OmitEmpty)
yaml, _ := serialize.ToYAML(resource, serialize.SnakeCase)
json, _ := serialize.ToJSON(resource, serialize.CamelCase)
```

### LSP Server

```go
import "github.com/lex00/wetwire-core-go/lsp"

server := lsp.NewServer(lsp.Config{
    Name:      "wetwire-aws-lsp",
    Linter:    myDiagnosticProvider,
    Completer: myCompletionProvider,
    HoverDocs: myHoverProvider,
})
```

For complete examples, see [wetwire-aws-go](https://github.com/lex00/wetwire-aws-go) which integrates this package.

## License

MIT
