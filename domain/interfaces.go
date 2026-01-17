package domain

// Domain is the core interface that all wetwire domain implementations must satisfy.
// It provides access to the domain's metadata and required operations.
type Domain interface {
	// Name returns the domain identifier (e.g., "aws", "honeycomb")
	Name() string

	// Version returns the domain implementation version
	Version() string

	// Builder returns the domain's Builder implementation
	Builder() Builder

	// Linter returns the domain's Linter implementation
	Linter() Linter

	// Initializer returns the domain's Initializer implementation
	Initializer() Initializer

	// Validator returns the domain's Validator implementation
	Validator() Validator
}

// ImporterDomain is an optional interface for domains that support importing
// external resources or configurations.
type ImporterDomain interface {
	Domain
	Importer() Importer
}

// ListerDomain is an optional interface for domains that support listing
// discovered resources.
type ListerDomain interface {
	Domain
	Lister() Lister
}

// GrapherDomain is an optional interface for domains that support visualizing
// resource relationships.
type GrapherDomain interface {
	Domain
	Grapher() Grapher
}

// Builder builds domain resources from source code.
type Builder interface {
	Build(ctx *Context, path string, opts BuildOpts) (*Result, error)
}

// BuildOpts contains options for the Build operation.
type BuildOpts struct {
	// Format specifies the output format (e.g., "json", "pretty")
	Format string

	// Type optionally filters build to specific resource types
	Type string

	// Output specifies the output path for generated files
	Output string

	// DryRun returns content without writing files
	DryRun bool
}

// Linter validates domain resources according to domain-specific rules.
type Linter interface {
	Lint(ctx *Context, path string, opts LintOpts) (*Result, error)
}

// LintOpts contains options for the Lint operation.
type LintOpts struct {
	// Format specifies the output format (e.g., "text", "json")
	Format string

	// Fix automatically fixes fixable issues
	Fix bool

	// Disable specifies rules to disable
	Disable []string
}

// Initializer creates new domain projects with example code.
type Initializer interface {
	Init(ctx *Context, path string, opts InitOpts) (*Result, error)
}

// InitOpts contains options for the Init operation.
type InitOpts struct {
	// Name is the project name
	Name string

	// Path is the output directory (defaults to current directory)
	Path string
}

// Validator validates that generated output conforms to domain specifications.
type Validator interface {
	Validate(ctx *Context, path string, opts ValidateOpts) (*Result, error)
}

// ValidateOpts contains options for the Validate operation.
type ValidateOpts struct {
	// Additional validation-specific options can be added here
}

// Importer imports external resources or configurations into the domain.
type Importer interface {
	Import(ctx *Context, source string, opts ImportOpts) (*Result, error)
}

// ImportOpts contains options for the Import operation.
type ImportOpts struct {
	// Target specifies where to import to
	Target string

	// Additional import-specific options can be added here
}

// Lister discovers and lists domain resources.
type Lister interface {
	List(ctx *Context, path string, opts ListOpts) (*Result, error)
}

// ListOpts contains options for the List operation.
type ListOpts struct {
	// Format specifies the output format (e.g., "text", "json")
	Format string

	// Type optionally filters list to specific resource types
	Type string
}

// Grapher visualizes relationships between domain resources.
type Grapher interface {
	Graph(ctx *Context, path string, opts GraphOpts) (*Result, error)
}

// GraphOpts contains options for the Graph operation.
type GraphOpts struct {
	// Format specifies the output format (e.g., "dot", "mermaid")
	Format string
}
