package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// StandardTools contains the standard tool schemas that all wetwire domain packages
// should implement. Domain packages can use these definitions and provide their own handlers.

// InitSchema is the JSON schema for wetwire_init tool.
var InitSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"name": map[string]any{
			"type":        "string",
			"description": "Project name",
		},
		"path": map[string]any{
			"type":        "string",
			"description": "Output directory (default: current directory)",
		},
	},
}

// BuildSchema is the JSON schema for wetwire_build tool.
var BuildSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"package": map[string]any{
			"type":        "string",
			"description": "Package path to discover resources from",
		},
		"output": map[string]any{
			"type":        "string",
			"description": "Output directory for generated files",
		},
		"format": map[string]any{
			"type":        "string",
			"enum":        []string{"yaml", "json"},
			"description": "Output format (default: yaml)",
		},
		"dry_run": map[string]any{
			"type":        "boolean",
			"description": "Return content without writing files",
		},
	},
}

// LintSchema is the JSON schema for wetwire_lint tool.
var LintSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"package": map[string]any{
			"type":        "string",
			"description": "Package path to lint",
		},
		"fix": map[string]any{
			"type":        "boolean",
			"description": "Automatically fix fixable issues",
		},
		"format": map[string]any{
			"type":        "string",
			"enum":        []string{"text", "json"},
			"description": "Output format (default: text)",
		},
	},
}

// ValidateSchema is the JSON schema for wetwire_validate tool.
var ValidateSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"path": map[string]any{
			"type":        "string",
			"description": "Path to file or directory to validate",
		},
		"format": map[string]any{
			"type":        "string",
			"enum":        []string{"text", "json"},
			"description": "Output format (default: text)",
		},
	},
}

// ImportSchema is the JSON schema for wetwire_import tool.
var ImportSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"files": map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "string"},
			"description": "Files to import",
		},
		"output": map[string]any{
			"type":        "string",
			"description": "Output directory for generated code",
		},
		"single_file": map[string]any{
			"type":        "boolean",
			"description": "Generate all code in a single file",
		},
	},
}

// ListSchema is the JSON schema for wetwire_list tool.
var ListSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"package": map[string]any{
			"type":        "string",
			"description": "Package path to discover from",
		},
		"format": map[string]any{
			"type":        "string",
			"enum":        []string{"table", "json"},
			"description": "Output format (default: table)",
		},
	},
}

// GraphSchema is the JSON schema for wetwire_graph tool.
var GraphSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"package": map[string]any{
			"type":        "string",
			"description": "Package path to analyze",
		},
		"format": map[string]any{
			"type":        "string",
			"enum":        []string{"dot", "mermaid"},
			"description": "Output format (default: mermaid)",
		},
		"output": map[string]any{
			"type":        "string",
			"description": "Output file path",
		},
	},
}

// WriteSchema is the JSON schema for wetwire_write tool.
var WriteSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"path": map[string]any{
			"type":        "string",
			"description": "File path to write to",
		},
		"content": map[string]any{
			"type":        "string",
			"description": "Content to write to the file",
		},
	},
	"required": []string{"path", "content"},
}

// ReadSchema is the JSON schema for wetwire_read tool.
var ReadSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"path": map[string]any{
			"type":        "string",
			"description": "File path to read from",
		},
	},
	"required": []string{"path"},
}

// ScenarioSchema is the JSON schema for wetwire_scenario tool.
var ScenarioSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"path": map[string]any{
			"type":        "string",
			"description": "Path to scenario directory",
		},
		"prompt": map[string]any{
			"type":        "string",
			"description": "Prompt variant (optional)",
		},
	},
}

// StandardToolDefinitions returns tool definitions for standard wetwire tools.
// Domain packages can use these to ensure consistent tool interfaces.
func StandardToolDefinitions(domain string) []ToolInfo {
	return []ToolInfo{
		{
			Name:        "wetwire_init",
			Description: "Initialize a new wetwire-" + domain + " project with example code",
			InputSchema: InitSchema,
		},
		{
			Name:        "wetwire_write",
			Description: "Write content to a file",
			InputSchema: WriteSchema,
		},
		{
			Name:        "wetwire_read",
			Description: "Read content from a file",
			InputSchema: ReadSchema,
		},
		{
			Name:        "wetwire_build",
			Description: "Generate " + domain + " output from wetwire declarations",
			InputSchema: BuildSchema,
		},
		{
			Name:        "wetwire_lint",
			Description: "Check code quality and style (domain lint rules)",
			InputSchema: LintSchema,
		},
		{
			Name:        "wetwire_validate",
			Description: "Validate generated output using external validator",
			InputSchema: ValidateSchema,
		},
		{
			Name:        "wetwire_import",
			Description: "Convert existing " + domain + " configs to wetwire code",
			InputSchema: ImportSchema,
		},
		{
			Name:        "wetwire_list",
			Description: "List discovered resources",
			InputSchema: ListSchema,
		},
		{
			Name:        "wetwire_graph",
			Description: "Visualize resource dependencies (DOT/Mermaid)",
			InputSchema: GraphSchema,
		},
		{
			Name:        "wetwire_scenario",
			Description: "Load and execute a scenario",
			InputSchema: ScenarioSchema,
		},
	}
}

// StandardToolHandlers is a map of tool names to handler functions.
// Domain packages provide these handlers to implement the standard tools.
type StandardToolHandlers struct {
	Init     ToolHandler
	Write    ToolHandler
	Read     ToolHandler
	Build    ToolHandler
	Lint     ToolHandler
	Validate ToolHandler
	Import   ToolHandler
	List     ToolHandler
	Graph    ToolHandler
	Scenario ToolHandler
}

// RegisterStandardTools registers all standard tools with a server.
// If a handler is nil, that tool is not registered.
func RegisterStandardTools(server *Server, domain string, handlers StandardToolHandlers) {
	defs := StandardToolDefinitions(domain)

	handlerMap := map[string]ToolHandler{
		"wetwire_init":     handlers.Init,
		"wetwire_write":    handlers.Write,
		"wetwire_read":     handlers.Read,
		"wetwire_build":    handlers.Build,
		"wetwire_lint":     handlers.Lint,
		"wetwire_validate": handlers.Validate,
		"wetwire_import":   handlers.Import,
		"wetwire_list":     handlers.List,
		"wetwire_graph":    handlers.Graph,
		"wetwire_scenario": handlers.Scenario,
	}

	for _, def := range defs {
		handler := handlerMap[def.Name]
		if handler == nil {
			continue
		}

		server.RegisterToolWithSchema(
			def.Name,
			def.Description,
			handler,
			def.InputSchema,
		)
	}
}

// WrapHandler wraps a simple string-returning function as a ToolHandler.
// This is a convenience for handlers that don't need the context.
func WrapHandler(fn func(args map[string]any) (string, error)) ToolHandler {
	return func(_ context.Context, args map[string]any) (string, error) {
		return fn(args)
	}
}

// DefaultFileWriteHandler provides a default implementation for wetwire_write tool.
func DefaultFileWriteHandler(_ context.Context, args map[string]any) (string, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return "", fmt.Errorf("path is required")
	}

	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("content is required")
	}

	// Create parent directories if they don't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write the file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path), nil
}

// DefaultFileReadHandler provides a default implementation for wetwire_read tool.
func DefaultFileReadHandler(_ context.Context, args map[string]any) (string, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return "", fmt.Errorf("path is required")
	}

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}

	return string(data), nil
}

// PlaceholderHandler returns an error indicating the tool requires domain-specific implementation.
func PlaceholderHandler(toolName string) ToolHandler {
	return func(_ context.Context, _ map[string]any) (string, error) {
		return "", fmt.Errorf("%s requires a domain-specific handler - not implemented", toolName)
	}
}

// RegisterStandardToolsWithDefaults registers standard tools with default handlers
// where available. File operations (write, read) get default implementations.
// Domain-specific tools (build, lint, validate, import, list, graph, scenario, init)
// require explicit handlers from the domain package.
func RegisterStandardToolsWithDefaults(server *Server, domain string, handlers StandardToolHandlers) {
	// Apply default handlers for file operations if not provided
	if handlers.Write == nil {
		handlers.Write = DefaultFileWriteHandler
	}
	if handlers.Read == nil {
		handlers.Read = DefaultFileReadHandler
	}

	// Register all tools with their handlers
	RegisterStandardTools(server, domain, handlers)
}
