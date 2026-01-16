# MCP Package

The MCP (Model Context Protocol) package provides server infrastructure for exposing wetwire tools to Claude Code and other MCP clients.

## Standard Tools

All domain packages inherit these standard tools:

### File Operations (Default Handlers Provided)

- **wetwire_write**: Write content to a file
  - Parameters: `path` (string), `content` (string)
  - Default implementation creates parent directories and writes the file

- **wetwire_read**: Read content from a file
  - Parameters: `path` (string)
  - Default implementation reads and returns file content

### Domain-Specific Tools (Require Implementation)

- **wetwire_init**: Initialize a new project
  - Parameters: `name` (string), `path` (string, optional)

- **wetwire_build**: Generate domain output from code
  - Parameters: `package` (string), `output` (string), `format` (string), `dry_run` (boolean)

- **wetwire_lint**: Run domain linter on code
  - Parameters: `package` (string), `fix` (boolean), `format` (string)

- **wetwire_validate**: Validate generated output
  - Parameters: `path` (string), `format` (string)

- **wetwire_import**: Convert existing configs to wetwire code
  - Parameters: `files` (array), `output` (string), `single_file` (boolean)

- **wetwire_list**: List discovered resources
  - Parameters: `package` (string), `format` (string)

- **wetwire_graph**: Visualize resource dependencies
  - Parameters: `package` (string), `format` (string), `output` (string)

- **wetwire_scenario**: Load and execute a scenario
  - Parameters: `path` (string), `prompt` (string, optional)

## Usage

### Basic Setup

```go
import "github.com/lex00/wetwire-core-go/mcp"

// Create server
server := mcp.NewServer(mcp.Config{
    Name:    "wetwire-mydomain",
    Version: "1.0.0",
    Debug:   true,
})

// Define handlers
handlers := mcp.StandardToolHandlers{
    Init:  myInitHandler,
    Build: myBuildHandler,
    Lint:  myLintHandler,
    // ... other handlers
}

// Register tools with defaults for file operations
mcp.RegisterStandardToolsWithDefaults(server, "mydomain", handlers)

// Start server
server.Start(context.Background())
```

### Custom File Handlers

Override default file operations by providing custom handlers:

```go
handlers := mcp.StandardToolHandlers{
    Write: func(ctx context.Context, args map[string]any) (string, error) {
        // Custom write logic
        return "Custom write result", nil
    },
    Read: func(ctx context.Context, args map[string]any) (string, error) {
        // Custom read logic
        return "file contents", nil
    },
    // ... domain-specific handlers
}
```

### Tool Schemas

All tool schemas are exported and can be used directly:

- `InitSchema`
- `WriteSchema`
- `ReadSchema`
- `BuildSchema`
- `LintSchema`
- `ValidateSchema`
- `ImportSchema`
- `ListSchema`
- `GraphSchema`
- `ScenarioSchema`

## Testing

See `tools_test.go` for comprehensive examples of:
- Testing default file handlers
- Testing custom handlers
- Testing schema validation
- Testing tool registration

## Installation Instructions

Generate MCP client configuration:

```go
instructions := mcp.GetInstallInstructions("wetwire-mydomain", "/path/to/binary")
fmt.Println(instructions)
```

This outputs configuration for Claude Code's settings file.
