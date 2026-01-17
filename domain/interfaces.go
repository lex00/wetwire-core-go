package domain

// Domain is the core interface that all wetwire domain implementations must satisfy.
// It provides access to the domain's metadata and required operations.
//
// VALIDATOR-FILE: domain/*_domain.go exists - domain implementation file required
// VALIDATOR-AST: "var _ Domain = (*" present - compile-time interface check required
// VALIDATOR-FILE: type re-exports present - domain package re-exports Context, Result, Error, etc.
// VALIDATOR-FILE: standard commands present - cmd/ contains design.go, test.go, diff.go, watch.go, mcp.go
type Domain interface {
	// Name returns the domain identifier (e.g., "aws", "honeycomb")
	// VALIDATOR: returns non-empty string
	// VALIDATOR: matches module suffix (wetwire-{name}-go)
	Name() string

	// Version returns the domain implementation version
	// VALIDATOR: returns non-empty string
	Version() string

	// Builder returns the domain's Builder implementation
	// VALIDATOR: returns non-nil
	// VALIDATOR-FILE: no newBuildCmd() in cmd/ - CLI auto-generated
	Builder() Builder

	// Linter returns the domain's Linter implementation
	// VALIDATOR: returns non-nil
	// VALIDATOR-FILE: no newLintCmd() in cmd/ - CLI auto-generated
	Linter() Linter

	// Initializer returns the domain's Initializer implementation
	// VALIDATOR: returns non-nil
	// VALIDATOR-FILE: no newInitCmd() in cmd/ - CLI auto-generated
	Initializer() Initializer

	// Validator returns the domain's Validator implementation
	// VALIDATOR: returns non-nil
	// VALIDATOR-FILE: no newValidateCmd() in cmd/ - CLI auto-generated
	Validator() Validator
}

// ImporterDomain is an optional interface for domains that support importing
// external resources or configurations.
//
// VALIDATOR-AST: if implemented, "var _ ImporterDomain = (*" present
type ImporterDomain interface {
	Domain
	// VALIDATOR: returns non-nil if interface claimed
	// VALIDATOR-FILE: no newImportCmd() in cmd/ if interface claimed
	Importer() Importer
}

// ListerDomain is an optional interface for domains that support listing
// discovered resources.
//
// VALIDATOR-AST: if implemented, "var _ ListerDomain = (*" present
type ListerDomain interface {
	Domain
	// VALIDATOR: returns non-nil if interface claimed
	// VALIDATOR-FILE: no newListCmd() in cmd/ if interface claimed
	Lister() Lister
}

// GrapherDomain is an optional interface for domains that support visualizing
// resource relationships.
//
// VALIDATOR-AST: if implemented, "var _ GrapherDomain = (*" present
type GrapherDomain interface {
	Domain
	// VALIDATOR: returns non-nil if interface claimed
	// VALIDATOR-FILE: no newGraphCmd() in cmd/ if interface claimed
	Grapher() Grapher
}

// Builder builds domain resources from source code.
//
// VALIDATOR-AST: signature matches Build(ctx *Context, path string, opts BuildOpts) (*Result, error)
type Builder interface {
	// VALIDATOR: returns *Result not custom type
	Build(ctx *Context, path string, opts BuildOpts) (*Result, error)
}

// BuildOpts contains options for the Build operation.
// VALIDATOR: domain uses these fields, not custom opts
type BuildOpts struct {
	// Format specifies the output format (e.g., "json", "pretty")
	// VALIDATOR: respected in Build output
	Format string

	// Type optionally filters build to specific resource types
	// VALIDATOR: filters resources if non-empty
	Type string

	// Output specifies the output path for generated files
	// VALIDATOR: writes to path if non-empty
	Output string

	// DryRun returns content without writing files
	// VALIDATOR: returns without writing if true
	DryRun bool
}

// Linter validates domain resources according to domain-specific rules.
//
// VALIDATOR-AST: signature matches Lint(ctx *Context, path string, opts LintOpts) (*Result, error)
type Linter interface {
	// VALIDATOR: returns *Result not custom type
	Lint(ctx *Context, path string, opts LintOpts) (*Result, error)
}

// LintOpts contains options for the Lint operation.
// VALIDATOR: domain uses these fields, not custom opts
type LintOpts struct {
	// Format specifies the output format (e.g., "text", "json")
	// VALIDATOR: respected in Lint output
	Format string

	// Fix automatically fixes fixable issues
	// VALIDATOR: applies fixes if true
	Fix bool

	// Disable specifies rules to disable
	// VALIDATOR: skips specified rules
	Disable []string
}

// Initializer creates new domain projects with example code.
//
// VALIDATOR-AST: signature matches Init(ctx *Context, path string, opts InitOpts) (*Result, error)
type Initializer interface {
	// VALIDATOR: returns *Result not custom type
	Init(ctx *Context, path string, opts InitOpts) (*Result, error)
}

// InitOpts contains options for the Init operation.
// VALIDATOR: domain uses these fields, not custom opts
type InitOpts struct {
	// Name is the project name
	// VALIDATOR: used for project directory name
	Name string

	// Path is the output directory (defaults to current directory)
	// VALIDATOR: creates project at path if non-empty
	Path string

	// Scenario indicates whether to create a full scenario structure
	// with prompts/, expected/, scenario.yaml, etc.
	// VALIDATOR: creates scenario structure if true
	// VALIDATOR-AST: if Scenario supported, "ScaffoldScenario" or "WriteScenario" present in Init
	Scenario bool

	// Description is a brief description of the scenario (used when Scenario is true)
	// VALIDATOR: used in scenario.yaml and prompt templates
	Description string
}

// Validator validates that generated output conforms to domain specifications.
//
// VALIDATOR-AST: signature matches Validate(ctx *Context, path string, opts ValidateOpts) (*Result, error)
type Validator interface {
	// VALIDATOR: returns *Result not custom type
	Validate(ctx *Context, path string, opts ValidateOpts) (*Result, error)
}

// ValidateOpts contains options for the Validate operation.
// VALIDATOR: domain uses these fields, not custom opts
type ValidateOpts struct {
	// Additional validation-specific options can be added here
}

// Importer imports external resources or configurations into the domain.
//
// VALIDATOR-AST: signature matches Import(ctx *Context, source string, opts ImportOpts) (*Result, error)
type Importer interface {
	// VALIDATOR: returns *Result not custom type
	Import(ctx *Context, source string, opts ImportOpts) (*Result, error)
}

// ImportOpts contains options for the Import operation.
// VALIDATOR: domain uses these fields, not custom opts
type ImportOpts struct {
	// Target specifies where to import to
	// VALIDATOR: writes imported code to target path
	Target string

	// Additional import-specific options can be added here
}

// Lister discovers and lists domain resources.
//
// VALIDATOR-AST: signature matches List(ctx *Context, path string, opts ListOpts) (*Result, error)
type Lister interface {
	// VALIDATOR: returns *Result not custom type
	List(ctx *Context, path string, opts ListOpts) (*Result, error)
}

// ListOpts contains options for the List operation.
// VALIDATOR: domain uses these fields, not custom opts
type ListOpts struct {
	// Format specifies the output format (e.g., "text", "json")
	// VALIDATOR: respected in List output
	Format string

	// Type optionally filters list to specific resource types
	// VALIDATOR: filters resources if non-empty
	Type string
}

// Grapher visualizes relationships between domain resources.
//
// VALIDATOR-AST: signature matches Graph(ctx *Context, path string, opts GraphOpts) (*Result, error)
type Grapher interface {
	// VALIDATOR: returns *Result not custom type
	Graph(ctx *Context, path string, opts GraphOpts) (*Result, error)
}

// GraphOpts contains options for the Graph operation.
// VALIDATOR: domain uses these fields, not custom opts
type GraphOpts struct {
	// Format specifies the output format (e.g., "dot", "mermaid")
	// VALIDATOR: respected in Graph output
	Format string
}
