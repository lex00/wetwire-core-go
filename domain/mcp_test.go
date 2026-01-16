package domain_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lex00/wetwire-core-go/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test buildMCPServer returns configured server
func TestBuildMCPServer(t *testing.T) {
	d := &fullMockDomain{}
	server := domain.BuildMCPServer(d)

	assert.NotNil(t, server)
	assert.Equal(t, "wetwire-mock", server.Name())
}

// Test all 4 required tools are registered
func TestBuildMCPServer_RequiredTools(t *testing.T) {
	d := &mockDomainBasic{}
	server := domain.BuildMCPServer(d)

	tools := server.GetTools()

	// Should have exactly 4 required tools
	assert.Len(t, tools, 4)

	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name] = true
	}

	assert.True(t, toolNames["wetwire_build"], "build tool should be registered")
	assert.True(t, toolNames["wetwire_lint"], "lint tool should be registered")
	assert.True(t, toolNames["wetwire_init"], "init tool should be registered")
	assert.True(t, toolNames["wetwire_validate"], "validate tool should be registered")
}

// Test optional tools only registered if domain implements optional interfaces
func TestBuildMCPServer_OptionalTools(t *testing.T) {
	t.Run("with all optional tools", func(t *testing.T) {
		d := &fullMockDomain{}
		server := domain.BuildMCPServer(d)

		tools := server.GetTools()

		// Should have 4 required + 3 optional = 7 tools
		assert.Len(t, tools, 7)

		toolNames := make(map[string]bool)
		for _, tool := range tools {
			toolNames[tool.Name] = true
		}

		// Required tools
		assert.True(t, toolNames["wetwire_build"])
		assert.True(t, toolNames["wetwire_lint"])
		assert.True(t, toolNames["wetwire_init"])
		assert.True(t, toolNames["wetwire_validate"])

		// Optional tools
		assert.True(t, toolNames["wetwire_import"])
		assert.True(t, toolNames["wetwire_list"])
		assert.True(t, toolNames["wetwire_graph"])
	})

	t.Run("with only import optional tool", func(t *testing.T) {
		d := &mockImporterDomainOnly{}
		server := domain.BuildMCPServer(d)

		tools := server.GetTools()

		// Should have 4 required + 1 optional = 5 tools
		assert.Len(t, tools, 5)

		toolNames := make(map[string]bool)
		for _, tool := range tools {
			toolNames[tool.Name] = true
		}

		assert.True(t, toolNames["wetwire_import"])
		assert.False(t, toolNames["wetwire_list"])
		assert.False(t, toolNames["wetwire_graph"])
	})

	t.Run("with only lister optional tool", func(t *testing.T) {
		d := &mockListerDomainOnly{}
		server := domain.BuildMCPServer(d)

		tools := server.GetTools()

		// Should have 4 required + 1 optional = 5 tools
		assert.Len(t, tools, 5)

		toolNames := make(map[string]bool)
		for _, tool := range tools {
			toolNames[tool.Name] = true
		}

		assert.False(t, toolNames["wetwire_import"])
		assert.True(t, toolNames["wetwire_list"])
		assert.False(t, toolNames["wetwire_graph"])
	})

	t.Run("with only grapher optional tool", func(t *testing.T) {
		d := &mockGrapherDomainOnly{}
		server := domain.BuildMCPServer(d)

		tools := server.GetTools()

		// Should have 4 required + 1 optional = 5 tools
		assert.Len(t, tools, 5)

		toolNames := make(map[string]bool)
		for _, tool := range tools {
			toolNames[tool.Name] = true
		}

		assert.False(t, toolNames["wetwire_import"])
		assert.False(t, toolNames["wetwire_list"])
		assert.True(t, toolNames["wetwire_graph"])
	})
}

// Test handlers return valid JSON
func TestBuildMCPServer_HandlersReturnValidJSON(t *testing.T) {
	d := &fullMockDomain{}
	server := domain.BuildMCPServer(d)
	ctx := context.Background()

	t.Run("build handler", func(t *testing.T) {
		result, err := server.ExecuteTool(ctx, "wetwire_build", map[string]any{
			"package": "/test/path",
			"format":  "json",
		})
		require.NoError(t, err)

		// Verify it's valid JSON
		var parsed map[string]any
		err = json.Unmarshal([]byte(result), &parsed)
		require.NoError(t, err)

		assert.True(t, parsed["success"].(bool))
		assert.Equal(t, "build successful", parsed["message"])
	})

	t.Run("lint handler", func(t *testing.T) {
		result, err := server.ExecuteTool(ctx, "wetwire_lint", map[string]any{
			"package": "/test/path",
			"format":  "text",
		})
		require.NoError(t, err)

		var parsed map[string]any
		err = json.Unmarshal([]byte(result), &parsed)
		require.NoError(t, err)

		assert.True(t, parsed["success"].(bool))
	})

	t.Run("init handler", func(t *testing.T) {
		result, err := server.ExecuteTool(ctx, "wetwire_init", map[string]any{
			"name": "test-project",
			"path": "/test/path",
		})
		require.NoError(t, err)

		var parsed map[string]any
		err = json.Unmarshal([]byte(result), &parsed)
		require.NoError(t, err)

		assert.True(t, parsed["success"].(bool))
	})

	t.Run("validate handler", func(t *testing.T) {
		result, err := server.ExecuteTool(ctx, "wetwire_validate", map[string]any{
			"path": "/test/path",
		})
		require.NoError(t, err)

		var parsed map[string]any
		err = json.Unmarshal([]byte(result), &parsed)
		require.NoError(t, err)

		assert.True(t, parsed["success"].(bool))
	})

	t.Run("import handler", func(t *testing.T) {
		result, err := server.ExecuteTool(ctx, "wetwire_import", map[string]any{
			"source": "/source/path",
			"target": "/target/path",
		})
		require.NoError(t, err)

		var parsed map[string]any
		err = json.Unmarshal([]byte(result), &parsed)
		require.NoError(t, err)

		assert.True(t, parsed["success"].(bool))
	})

	t.Run("list handler", func(t *testing.T) {
		result, err := server.ExecuteTool(ctx, "wetwire_list", map[string]any{
			"package": "/test/path",
			"format":  "json",
		})
		require.NoError(t, err)

		var parsed map[string]any
		err = json.Unmarshal([]byte(result), &parsed)
		require.NoError(t, err)

		assert.True(t, parsed["success"].(bool))
	})

	t.Run("graph handler", func(t *testing.T) {
		result, err := server.ExecuteTool(ctx, "wetwire_graph", map[string]any{
			"package": "/test/path",
			"format":  "dot",
		})
		require.NoError(t, err)

		var parsed map[string]any
		err = json.Unmarshal([]byte(result), &parsed)
		require.NoError(t, err)

		assert.True(t, parsed["success"].(bool))
	})
}

// Test handler errors propagate correctly
func TestBuildMCPServer_HandlerErrorsPropagated(t *testing.T) {
	d := &errorMockDomain{}
	server := domain.BuildMCPServer(d)
	ctx := context.Background()

	t.Run("build error", func(t *testing.T) {
		result, err := server.ExecuteTool(ctx, "wetwire_build", map[string]any{})
		require.NoError(t, err) // ExecuteTool should not return error

		// But the result should contain error information
		var parsed map[string]any
		parseErr := json.Unmarshal([]byte(result), &parsed)
		require.NoError(t, parseErr)

		assert.False(t, parsed["success"].(bool))
		assert.NotEmpty(t, parsed["errors"])
	})

	t.Run("lint error", func(t *testing.T) {
		result, err := server.ExecuteTool(ctx, "wetwire_lint", map[string]any{})
		require.NoError(t, err)

		var parsed map[string]any
		parseErr := json.Unmarshal([]byte(result), &parsed)
		require.NoError(t, parseErr)

		assert.False(t, parsed["success"].(bool))
	})
}

// Mock implementations for testing

type mockDomainBasic struct{}

func (m *mockDomainBasic) Name() string    { return "mock" }
func (m *mockDomainBasic) Version() string { return "1.0.0" }
func (m *mockDomainBasic) Builder() domain.Builder {
	return &mockBuilderSuccess{}
}
func (m *mockDomainBasic) Linter() domain.Linter {
	return &mockLinterSuccess{}
}
func (m *mockDomainBasic) Initializer() domain.Initializer {
	return &mockInitializerSuccess{}
}
func (m *mockDomainBasic) Validator() domain.Validator {
	return &mockValidatorSuccess{}
}

type fullMockDomain struct {
	mockDomainBasic
}

func (m *fullMockDomain) Importer() domain.Importer {
	return &mockImporterSuccess{}
}
func (m *fullMockDomain) Lister() domain.Lister {
	return &mockListerSuccess{}
}
func (m *fullMockDomain) Grapher() domain.Grapher {
	return &mockGrapherSuccess{}
}

type mockImporterDomainOnly struct {
	mockDomainBasic
}

func (m *mockImporterDomainOnly) Importer() domain.Importer {
	return &mockImporterSuccess{}
}

type mockListerDomainOnly struct {
	mockDomainBasic
}

func (m *mockListerDomainOnly) Lister() domain.Lister {
	return &mockListerSuccess{}
}

type mockGrapherDomainOnly struct {
	mockDomainBasic
}

func (m *mockGrapherDomainOnly) Grapher() domain.Grapher {
	return &mockGrapherSuccess{}
}

type errorMockDomain struct {
	mockDomainBasic
}

func (m *errorMockDomain) Builder() domain.Builder {
	return &mockBuilderError{}
}
func (m *errorMockDomain) Linter() domain.Linter {
	return &mockLinterError{}
}

// Success implementations

type mockBuilderSuccess struct{}

func (b *mockBuilderSuccess) Build(ctx *domain.Context, path string, opts domain.BuildOpts) (*domain.Result, error) {
	return domain.NewResult("build successful"), nil
}

type mockLinterSuccess struct{}

func (l *mockLinterSuccess) Lint(ctx *domain.Context, path string, opts domain.LintOpts) (*domain.Result, error) {
	return domain.NewResult("lint successful"), nil
}

type mockInitializerSuccess struct{}

func (i *mockInitializerSuccess) Init(ctx *domain.Context, path string, opts domain.InitOpts) (*domain.Result, error) {
	return domain.NewResult("init successful"), nil
}

type mockValidatorSuccess struct{}

func (v *mockValidatorSuccess) Validate(ctx *domain.Context, path string, opts domain.ValidateOpts) (*domain.Result, error) {
	return domain.NewResult("validate successful"), nil
}

type mockImporterSuccess struct{}

func (i *mockImporterSuccess) Import(ctx *domain.Context, source string, opts domain.ImportOpts) (*domain.Result, error) {
	return domain.NewResult("import successful"), nil
}

type mockListerSuccess struct{}

func (l *mockListerSuccess) List(ctx *domain.Context, path string, opts domain.ListOpts) (*domain.Result, error) {
	return domain.NewResult("list successful"), nil
}

type mockGrapherSuccess struct{}

func (g *mockGrapherSuccess) Graph(ctx *domain.Context, path string, opts domain.GraphOpts) (*domain.Result, error) {
	return domain.NewResult("graph successful"), nil
}

// Error implementations

type mockBuilderError struct{}

func (b *mockBuilderError) Build(ctx *domain.Context, path string, opts domain.BuildOpts) (*domain.Result, error) {
	return domain.NewErrorResult("build failed", domain.Error{Message: "build error"}), nil
}

type mockLinterError struct{}

func (l *mockLinterError) Lint(ctx *domain.Context, path string, opts domain.LintOpts) (*domain.Result, error) {
	return domain.NewErrorResult("lint failed", domain.Error{Message: "lint error"}), nil
}
