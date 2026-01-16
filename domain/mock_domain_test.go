package domain_test

import (
	"testing"

	"github.com/lex00/wetwire-core-go/domain"
	"github.com/stretchr/testify/assert"
)

// Complete mock domain implementation demonstrating proper implementation pattern.
// This serves as both a test fixture and documentation for domain implementers.

// completeMockDomain implements the minimal Domain interface
type completeMockDomain struct{}

func (m *completeMockDomain) Name() string    { return "complete-mock" }
func (m *completeMockDomain) Version() string { return "2.0.0" }
func (m *completeMockDomain) Builder() domain.Builder {
	return &completeMockBuilder{}
}
func (m *completeMockDomain) Linter() domain.Linter {
	return &completeMockLinter{}
}
func (m *completeMockDomain) Initializer() domain.Initializer {
	return &completeMockInitializer{}
}
func (m *completeMockDomain) Validator() domain.Validator {
	return &completeMockValidator{}
}

// completeMockFullDomain implements all optional interfaces
type completeMockFullDomain struct {
	completeMockDomain
}

func (m *completeMockFullDomain) Importer() domain.Importer {
	return &completeMockImporter{}
}
func (m *completeMockFullDomain) Lister() domain.Lister {
	return &completeMockLister{}
}
func (m *completeMockFullDomain) Grapher() domain.Grapher {
	return &completeMockGrapher{}
}

// Complete mock operation implementations

type completeMockBuilder struct{}

func (b *completeMockBuilder) Build(ctx *domain.Context, path string, opts domain.BuildOpts) (*domain.Result, error) {
	data := map[string]interface{}{
		"built_at": path,
		"format":   opts.Format,
	}
	return domain.NewResultWithData("Built "+path, data), nil
}

type completeMockLinter struct{}

func (l *completeMockLinter) Lint(ctx *domain.Context, path string, opts domain.LintOpts) (*domain.Result, error) {
	if path == "/test/error" {
		return domain.NewErrorResult("Lint failed", domain.Error{
			Path:     path + "/file.go",
			Line:     42,
			Column:   10,
			Severity: "error",
			Message:  "mock lint error",
			Code:     "MOCK001",
		}), nil
	}
	return domain.NewResult("Lint passed for " + path), nil
}

type completeMockInitializer struct{}

func (i *completeMockInitializer) Init(ctx *domain.Context, path string, opts domain.InitOpts) (*domain.Result, error) {
	data := map[string]interface{}{
		"name": opts.Name,
		"path": path,
	}
	return domain.NewResultWithData("Initialized "+opts.Name, data), nil
}

type completeMockValidator struct{}

func (v *completeMockValidator) Validate(ctx *domain.Context, path string, opts domain.ValidateOpts) (*domain.Result, error) {
	return domain.NewResult("Validation passed for " + path), nil
}

type completeMockImporter struct{}

func (i *completeMockImporter) Import(ctx *domain.Context, source string, opts domain.ImportOpts) (*domain.Result, error) {
	data := map[string]interface{}{
		"source": source,
		"target": opts.Target,
	}
	return domain.NewResultWithData("Imported from "+source, data), nil
}

type completeMockLister struct{}

func (l *completeMockLister) List(ctx *domain.Context, path string, opts domain.ListOpts) (*domain.Result, error) {
	items := []string{"item1", "item2", "item3"}
	data := map[string]interface{}{
		"items": items,
		"count": len(items),
		"type":  opts.Type,
	}
	return domain.NewResultWithData("Listed 3 items", data), nil
}

type completeMockGrapher struct{}

func (g *completeMockGrapher) Graph(ctx *domain.Context, path string, opts domain.GraphOpts) (*domain.Result, error) {
	data := map[string]interface{}{
		"nodes":  10,
		"edges":  15,
		"format": opts.Format,
	}
	return domain.NewResultWithData("Graph generated", data), nil
}

// Compile-time interface checks demonstrate that types satisfy all interfaces
var (
	_ domain.Domain         = (*completeMockDomain)(nil)
	_ domain.ImporterDomain = (*completeMockFullDomain)(nil)
	_ domain.ListerDomain   = (*completeMockFullDomain)(nil)
	_ domain.GrapherDomain  = (*completeMockFullDomain)(nil)
	_ domain.Builder        = (*completeMockBuilder)(nil)
	_ domain.Linter         = (*completeMockLinter)(nil)
	_ domain.Initializer    = (*completeMockInitializer)(nil)
	_ domain.Validator      = (*completeMockValidator)(nil)
	_ domain.Importer       = (*completeMockImporter)(nil)
	_ domain.Lister         = (*completeMockLister)(nil)
	_ domain.Grapher        = (*completeMockGrapher)(nil)
)

// Tests demonstrating complete mock domain usage

func TestCompleteMockDomain_CompileTimeEnforcement(t *testing.T) {
	// This test documents the compile-time interface checking pattern.
	// If interfaces change, these assignments will fail to compile.
	var _ domain.Domain = (*completeMockDomain)(nil)
	var _ domain.ImporterDomain = (*completeMockFullDomain)(nil)
	assert.True(t, true, "Compile-time checks passed")
}

func TestCompleteMockDomain_BasicOperations(t *testing.T) {
	d := &completeMockDomain{}
	ctx := domain.NewContext(nil, "/test")

	t.Run("Build", func(t *testing.T) {
		result, err := d.Builder().Build(ctx, "/test/path", domain.BuildOpts{Format: "json"})
		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Message, "Built")
	})

	t.Run("Lint", func(t *testing.T) {
		result, err := d.Linter().Lint(ctx, "/test/path", domain.LintOpts{})
		assert.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("Init", func(t *testing.T) {
		result, err := d.Initializer().Init(ctx, "/test/path", domain.InitOpts{Name: "test-project"})
		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Message, "Initialized")
	})

	t.Run("Validate", func(t *testing.T) {
		result, err := d.Validator().Validate(ctx, "/test/path", domain.ValidateOpts{})
		assert.NoError(t, err)
		assert.True(t, result.Success)
	})
}

func TestCompleteMockDomain_OptionalOperations(t *testing.T) {
	d := &completeMockFullDomain{}
	ctx := domain.NewContext(nil, "/test")

	t.Run("Import", func(t *testing.T) {
		result, err := d.Importer().Import(ctx, "/source", domain.ImportOpts{Target: "/target"})
		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Message, "Imported")
	})

	t.Run("List", func(t *testing.T) {
		result, err := d.Lister().List(ctx, "/test/path", domain.ListOpts{Type: "all"})
		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Message, "Listed")
	})

	t.Run("Graph", func(t *testing.T) {
		result, err := d.Grapher().Graph(ctx, "/test/path", domain.GraphOpts{Format: "dot"})
		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Message, "Graph")
	})
}

func TestCompleteMockDomain_ErrorHandling(t *testing.T) {
	d := &completeMockDomain{}
	ctx := domain.NewContext(nil, "/test")

	t.Run("Lint error", func(t *testing.T) {
		result, err := d.Linter().Lint(ctx, "/test/error", domain.LintOpts{})
		assert.NoError(t, err)
		assert.False(t, result.Success)
		assert.Len(t, result.Errors, 1)
		assert.Equal(t, "mock lint error", result.Errors[0].Message)
		assert.Equal(t, "MOCK001", result.Errors[0].Code)
	})
}

func TestCompleteMockDomain_CLIGeneration(t *testing.T) {
	t.Run("Basic domain generates 4 commands", func(t *testing.T) {
		d := &completeMockDomain{}
		cmd := domain.Run(d)

		assert.Equal(t, "wetwire-complete-mock", cmd.Use)
		assert.Equal(t, "2.0.0", cmd.Version)

		// Should have build, lint, init, validate
		assert.Len(t, cmd.Commands(), 4)
	})

	t.Run("Full domain generates all commands", func(t *testing.T) {
		d := &completeMockFullDomain{}
		cmd := domain.Run(d)

		// Should have build, lint, init, validate, import, list, graph
		assert.Len(t, cmd.Commands(), 7)
	})
}

func TestCompleteMockDomain_MCPGeneration(t *testing.T) {
	t.Run("Basic domain registers 4 tools", func(t *testing.T) {
		d := &completeMockDomain{}
		server := domain.BuildMCPServer(d)

		tools := server.GetTools()
		assert.Len(t, tools, 4)
	})

	t.Run("Full domain registers all tools", func(t *testing.T) {
		d := &completeMockFullDomain{}
		server := domain.BuildMCPServer(d)

		tools := server.GetTools()
		assert.Len(t, tools, 7)
	})
}

func TestCompleteMockDomain_ContextUsage(t *testing.T) {
	t.Run("Context with verbose", func(t *testing.T) {
		ctx := domain.NewContextWithVerbose(nil, "/test", true)
		assert.True(t, ctx.Verbose)
		assert.Equal(t, "/test", ctx.WorkDir)
	})

	t.Run("Context without verbose", func(t *testing.T) {
		ctx := domain.NewContext(nil, "/test")
		assert.False(t, ctx.Verbose)
		assert.Equal(t, "/test", ctx.WorkDir)
	})
}
