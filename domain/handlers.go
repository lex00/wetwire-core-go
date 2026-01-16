package domain

import (
	"context"
	"fmt"
)

// createBuildHandler creates an MCP tool handler for the Build operation.
// It converts MCP arguments to BuildOpts and returns the result as JSON.
func createBuildHandler(builder Builder) func(ctx context.Context, args map[string]any) (string, error) {
	return func(ctx context.Context, args map[string]any) (string, error) {
		// Extract path from args (typically "package" param maps to path)
		path := ""
		if pkg, ok := args["package"].(string); ok {
			path = pkg
		}

		// Build options from args
		opts := BuildOpts{}
		if format, ok := args["format"].(string); ok {
			opts.Format = format
		}
		if typ, ok := args["type"].(string); ok {
			opts.Type = typ
		}

		// Create domain context
		domainCtx := NewContext(ctx, path)

		// Execute build
		result, err := builder.Build(domainCtx, path, opts)
		if err != nil {
			return "", fmt.Errorf("build operation failed: %w", err)
		}

		// Convert to JSON
		jsonBytes, err := result.ToJSON()
		if err != nil {
			return "", fmt.Errorf("failed to serialize result: %w", err)
		}

		return string(jsonBytes), nil
	}
}

// createLintHandler creates an MCP tool handler for the Lint operation.
// It converts MCP arguments to LintOpts and returns the result as JSON.
func createLintHandler(linter Linter) func(ctx context.Context, args map[string]any) (string, error) {
	return func(ctx context.Context, args map[string]any) (string, error) {
		// Extract path from args
		path := ""
		if pkg, ok := args["package"].(string); ok {
			path = pkg
		}

		// Lint options from args
		opts := LintOpts{}
		if format, ok := args["format"].(string); ok {
			opts.Format = format
		}

		// Create domain context
		domainCtx := NewContext(ctx, path)

		// Execute lint
		result, err := linter.Lint(domainCtx, path, opts)
		if err != nil {
			return "", fmt.Errorf("lint operation failed: %w", err)
		}

		// Convert to JSON
		jsonBytes, err := result.ToJSON()
		if err != nil {
			return "", fmt.Errorf("failed to serialize result: %w", err)
		}

		return string(jsonBytes), nil
	}
}

// createInitHandler creates an MCP tool handler for the Init operation.
// It converts MCP arguments to InitOpts and returns the result as JSON.
func createInitHandler(initializer Initializer) func(ctx context.Context, args map[string]any) (string, error) {
	return func(ctx context.Context, args map[string]any) (string, error) {
		// Extract path from args
		path := ""
		if p, ok := args["path"].(string); ok {
			path = p
		}

		// Init options from args
		opts := InitOpts{
			Path: path,
		}
		if name, ok := args["name"].(string); ok {
			opts.Name = name
		}

		// Create domain context
		domainCtx := NewContext(ctx, path)

		// Execute init
		result, err := initializer.Init(domainCtx, path, opts)
		if err != nil {
			return "", fmt.Errorf("init operation failed: %w", err)
		}

		// Convert to JSON
		jsonBytes, err := result.ToJSON()
		if err != nil {
			return "", fmt.Errorf("failed to serialize result: %w", err)
		}

		return string(jsonBytes), nil
	}
}

// createValidateHandler creates an MCP tool handler for the Validate operation.
// It converts MCP arguments to ValidateOpts and returns the result as JSON.
func createValidateHandler(validator Validator) func(ctx context.Context, args map[string]any) (string, error) {
	return func(ctx context.Context, args map[string]any) (string, error) {
		// Extract path from args
		path := ""
		if p, ok := args["path"].(string); ok {
			path = p
		}

		// Validate options from args
		opts := ValidateOpts{}

		// Create domain context
		domainCtx := NewContext(ctx, path)

		// Execute validate
		result, err := validator.Validate(domainCtx, path, opts)
		if err != nil {
			return "", fmt.Errorf("validate operation failed: %w", err)
		}

		// Convert to JSON
		jsonBytes, err := result.ToJSON()
		if err != nil {
			return "", fmt.Errorf("failed to serialize result: %w", err)
		}

		return string(jsonBytes), nil
	}
}

// createImportHandler creates an MCP tool handler for the Import operation.
// It converts MCP arguments to ImportOpts and returns the result as JSON.
func createImportHandler(importer Importer) func(ctx context.Context, args map[string]any) (string, error) {
	return func(ctx context.Context, args map[string]any) (string, error) {
		// Extract source from args
		source := ""
		if s, ok := args["source"].(string); ok {
			source = s
		}

		// Import options from args
		opts := ImportOpts{}
		if target, ok := args["target"].(string); ok {
			opts.Target = target
		}

		// Create domain context
		domainCtx := NewContext(ctx, source)

		// Execute import
		result, err := importer.Import(domainCtx, source, opts)
		if err != nil {
			return "", fmt.Errorf("import operation failed: %w", err)
		}

		// Convert to JSON
		jsonBytes, err := result.ToJSON()
		if err != nil {
			return "", fmt.Errorf("failed to serialize result: %w", err)
		}

		return string(jsonBytes), nil
	}
}

// createListHandler creates an MCP tool handler for the List operation.
// It converts MCP arguments to ListOpts and returns the result as JSON.
func createListHandler(lister Lister) func(ctx context.Context, args map[string]any) (string, error) {
	return func(ctx context.Context, args map[string]any) (string, error) {
		// Extract path from args
		path := ""
		if pkg, ok := args["package"].(string); ok {
			path = pkg
		}

		// List options from args
		opts := ListOpts{}
		if format, ok := args["format"].(string); ok {
			opts.Format = format
		}
		if typ, ok := args["type"].(string); ok {
			opts.Type = typ
		}

		// Create domain context
		domainCtx := NewContext(ctx, path)

		// Execute list
		result, err := lister.List(domainCtx, path, opts)
		if err != nil {
			return "", fmt.Errorf("list operation failed: %w", err)
		}

		// Convert to JSON
		jsonBytes, err := result.ToJSON()
		if err != nil {
			return "", fmt.Errorf("failed to serialize result: %w", err)
		}

		return string(jsonBytes), nil
	}
}

// createGraphHandler creates an MCP tool handler for the Graph operation.
// It converts MCP arguments to GraphOpts and returns the result as JSON.
func createGraphHandler(grapher Grapher) func(ctx context.Context, args map[string]any) (string, error) {
	return func(ctx context.Context, args map[string]any) (string, error) {
		// Extract path from args
		path := ""
		if pkg, ok := args["package"].(string); ok {
			path = pkg
		}

		// Graph options from args
		opts := GraphOpts{}
		if format, ok := args["format"].(string); ok {
			opts.Format = format
		}

		// Create domain context
		domainCtx := NewContext(ctx, path)

		// Execute graph
		result, err := grapher.Graph(domainCtx, path, opts)
		if err != nil {
			return "", fmt.Errorf("graph operation failed: %w", err)
		}

		// Convert to JSON
		jsonBytes, err := result.ToJSON()
		if err != nil {
			return "", fmt.Errorf("failed to serialize result: %w", err)
		}

		return string(jsonBytes), nil
	}
}
