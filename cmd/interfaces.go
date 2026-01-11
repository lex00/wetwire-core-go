// Package cmd provides a command framework for wetwire CLIs.
//
// Domain packages use this framework to build consistent CLIs with shared
// flags, help text, and command structure.
package cmd

import "context"

// BuildOptions contains options for the build command.
type BuildOptions struct {
	Output  string
	Verbose bool
	DryRun  bool
}

// LintOptions contains options for the lint command.
type LintOptions struct {
	Fix     bool
	Verbose bool
}

// InitOptions contains options for the init command.
type InitOptions struct {
	Template string
	Force    bool
}

// ValidateOptions contains options for the validate command.
type ValidateOptions struct {
	Strict  bool
	Verbose bool
}

// Issue represents a linting issue found in source files.
type Issue struct {
	File     string
	Line     int
	Column   int
	Severity string // "error", "warning", "info"
	Message  string
	Rule     string
}

// ValidationError represents a validation error in generated output.
type ValidationError struct {
	Path    string
	Message string
	Code    string
}

// Builder builds synthesized infrastructure from source definitions.
type Builder interface {
	Build(ctx context.Context, path string, opts BuildOptions) error
}

// Linter checks source files for issues.
type Linter interface {
	Lint(ctx context.Context, path string, opts LintOptions) ([]Issue, error)
}

// Initializer creates new project scaffolding.
type Initializer interface {
	Init(ctx context.Context, name string, opts InitOptions) error
}

// Validator validates generated output.
type Validator interface {
	Validate(ctx context.Context, path string, opts ValidateOptions) ([]ValidationError, error)
}
