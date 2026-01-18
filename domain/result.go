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

	// CrossDomain contains outputs from dependency domains, allowing
	// dependent domains to reference values from previously executed domains.
	CrossDomain *CrossDomainContext
}

// CrossDomainContext holds outputs from dependency domains.
// It allows domains to access outputs from their dependencies.
type CrossDomainContext struct {
	// Dependencies maps domain names to their outputs.
	// Each domain's outputs are stored as a DomainOutputs structure.
	Dependencies map[string]*DomainOutputs
}

// DomainOutputs contains outputs for a single domain.
// It maps resource names to their individual outputs.
type DomainOutputs struct {
	// Resources maps resource names to their outputs.
	Resources map[string]*ResourceOutputs
}

// ResourceOutputs contains the output data for a single resource.
type ResourceOutputs struct {
	// Type is the resource type (e.g., "aws_s3_bucket", "gitlab_pipeline")
	Type string

	// Outputs is a map of output names to values
	Outputs map[string]interface{}
}

// NewCrossDomainContext creates a new empty CrossDomainContext.
func NewCrossDomainContext() *CrossDomainContext {
	return &CrossDomainContext{
		Dependencies: make(map[string]*DomainOutputs),
	}
}

// AddDomainOutputs adds outputs for a domain to the cross-domain context.
func (c *CrossDomainContext) AddDomainOutputs(domainName string, outputs *DomainOutputs) {
	if c.Dependencies == nil {
		c.Dependencies = make(map[string]*DomainOutputs)
	}
	c.Dependencies[domainName] = outputs
}

// GetDomainOutputs retrieves outputs for a domain from the cross-domain context.
// Returns nil if the domain is not found.
func (c *CrossDomainContext) GetDomainOutputs(domainName string) *DomainOutputs {
	if c.Dependencies == nil {
		return nil
	}
	return c.Dependencies[domainName]
}

// GetResourceOutput retrieves a specific output value from a domain's resource.
// Returns nil if the domain, resource, or output key is not found.
func (c *CrossDomainContext) GetResourceOutput(domainName, resourceName, outputKey string) interface{} {
	domainOutputs := c.GetDomainOutputs(domainName)
	if domainOutputs == nil {
		return nil
	}

	if domainOutputs.Resources == nil {
		return nil
	}

	resourceOutputs, ok := domainOutputs.Resources[resourceName]
	if !ok || resourceOutputs == nil {
		return nil
	}

	if resourceOutputs.Outputs == nil {
		return nil
	}

	return resourceOutputs.Outputs[outputKey]
}

// HasDependency checks if outputs exist for the specified domain.
func (c *CrossDomainContext) HasDependency(domainName string) bool {
	if c.Dependencies == nil {
		return false
	}
	_, ok := c.Dependencies[domainName]
	return ok
}

// DomainNames returns a list of all domain names in the cross-domain context.
func (c *CrossDomainContext) DomainNames() []string {
	if c.Dependencies == nil {
		return nil
	}
	names := make([]string, 0, len(c.Dependencies))
	for name := range c.Dependencies {
		names = append(names, name)
	}
	return names
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
// The CrossDomain field is left nil; use NewContextWithCrossDomain to set it.
func NewContext(ctx context.Context, workDir string) *Context {
	return &Context{
		Context: ctx,
		WorkDir: workDir,
		Verbose: false,
	}
}

// NewContextWithVerbose creates a new Context with the given background context,
// working directory, and verbosity setting.
// The CrossDomain field is left nil; use NewContextWithCrossDomain to set it.
func NewContextWithVerbose(ctx context.Context, workDir string, verbose bool) *Context {
	return &Context{
		Context: ctx,
		WorkDir: workDir,
		Verbose: verbose,
	}
}

// NewContextWithCrossDomain creates a new Context with cross-domain output support.
// This is used when executing domains that depend on outputs from other domains.
func NewContextWithCrossDomain(ctx context.Context, workDir string, verbose bool, crossDomain *CrossDomainContext) *Context {
	return &Context{
		Context:     ctx,
		WorkDir:     workDir,
		Verbose:     verbose,
		CrossDomain: crossDomain,
	}
}

// WithCrossDomain returns a copy of the context with the given cross-domain context set.
// This is useful for adding cross-domain outputs to an existing context.
func (c *Context) WithCrossDomain(crossDomain *CrossDomainContext) *Context {
	return &Context{
		Context:     c.Context,
		WorkDir:     c.WorkDir,
		Verbose:     c.Verbose,
		CrossDomain: crossDomain,
	}
}
