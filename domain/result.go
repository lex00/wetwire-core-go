package domain

import (
	"context"
	"encoding/json"
	"fmt"
)

// Result represents the outcome of a domain operation.
// It provides a unified structure for returning success or failure information,
// optional data, and detailed error information.
//
// VALIDATOR: all operations return *Result
// VALIDATOR-FILE: no custom result types in domain/
type Result struct {
	Success bool    `json:"success"`
	Message string  `json:"message,omitempty"`
	Data    any     `json:"data,omitempty"`
	Errors  []Error `json:"errors,omitempty"`
}

// Error represents a structured error with location and context information.
// It can represent linting errors, validation errors, or any other structured
// diagnostic information.
//
// VALIDATOR: used for lint issues, build errors, validation errors
type Error struct {
	Path     string `json:"path,omitempty"`
	Line     int    `json:"line,omitempty"`
	Column   int    `json:"column,omitempty"`
	Severity string `json:"severity,omitempty"`
	Message  string `json:"message"`
	Code     string `json:"code,omitempty"`
}

// Context wraps context.Context with additional domain operation context.
// It provides working directory and verbosity information for operations.
type Context struct {
	context.Context
	WorkDir string
	Verbose bool
}

// NewResult creates a successful Result with a message.
//
// VALIDATOR-AST: domain uses NewResult() not Result{Success: true}
func NewResult(message string) *Result {
	return &Result{
		Success: true,
		Message: message,
	}
}

// NewResultWithData creates a successful Result with a message and data.
//
// VALIDATOR-AST: domain uses NewResultWithData() for data returns
func NewResultWithData(message string, data any) *Result {
	return &Result{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResult creates a failed Result with a message and a single error.
//
// VALIDATOR-AST: domain uses NewErrorResult() for single errors
func NewErrorResult(message string, err Error) *Result {
	return &Result{
		Success: false,
		Message: message,
		Errors:  []Error{err},
	}
}

// NewErrorResultMultiple creates a failed Result with a message and multiple errors.
//
// VALIDATOR-AST: domain uses NewErrorResultMultiple() for multiple errors
func NewErrorResultMultiple(message string, errs []Error) *Result {
	return &Result{
		Success: false,
		Message: message,
		Errors:  errs,
	}
}

// ToJSON serializes the Result to JSON bytes.
func (r *Result) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// String formats the Error as a human-readable string.
// It includes all available information (path, line, column, severity, message).
func (e *Error) String() string {
	if e.Path == "" {
		return e.Message
	}

	result := e.Path
	if e.Line > 0 {
		result += fmt.Sprintf(":%d", e.Line)
		if e.Column > 0 {
			result += fmt.Sprintf(":%d", e.Column)
		}
	}

	if e.Severity != "" {
		result += fmt.Sprintf(" [%s]", e.Severity)
	}

	result += ": " + e.Message

	if e.Code != "" {
		result += fmt.Sprintf(" (%s)", e.Code)
	}

	return result
}

// NewContext creates a new Context with the given background context and working directory.
func NewContext(ctx context.Context, workDir string) *Context {
	return &Context{
		Context: ctx,
		WorkDir: workDir,
		Verbose: false,
	}
}

// NewContextWithVerbose creates a new Context with the given background context,
// working directory, and verbosity setting.
func NewContextWithVerbose(ctx context.Context, workDir string, verbose bool) *Context {
	return &Context{
		Context: ctx,
		WorkDir: workDir,
		Verbose: verbose,
	}
}
