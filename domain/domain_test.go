package domain_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lex00/wetwire-core-go/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Result creation
func TestNewResult(t *testing.T) {
	result := domain.NewResult("Operation successful")

	assert.True(t, result.Success)
	assert.Equal(t, "Operation successful", result.Message)
	assert.Nil(t, result.Data)
	assert.Empty(t, result.Errors)
}

func TestNewResultWithData(t *testing.T) {
	data := map[string]string{"key": "value"}
	result := domain.NewResultWithData("Operation successful", data)

	assert.True(t, result.Success)
	assert.Equal(t, "Operation successful", result.Message)
	assert.Equal(t, data, result.Data)
	assert.Empty(t, result.Errors)
}

func TestNewErrorResult(t *testing.T) {
	err := domain.Error{
		Path:     "test.go",
		Line:     10,
		Column:   5,
		Severity: "error",
		Message:  "syntax error",
		Code:     "E001",
	}

	result := domain.NewErrorResult("Operation failed", err)

	assert.False(t, result.Success)
	assert.Equal(t, "Operation failed", result.Message)
	assert.Nil(t, result.Data)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, err, result.Errors[0])
}

func TestNewErrorResultMultiple(t *testing.T) {
	errs := []domain.Error{
		{Message: "error 1"},
		{Message: "error 2"},
	}

	result := domain.NewErrorResultMultiple("Multiple errors", errs)

	assert.False(t, result.Success)
	assert.Equal(t, "Multiple errors", result.Message)
	assert.Len(t, result.Errors, 2)
}

// Test Result.ToJSON()
func TestResultToJSON(t *testing.T) {
	result := domain.NewResult("success")

	jsonBytes, err := result.ToJSON()
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, true, unmarshaled["success"])
	assert.Equal(t, "success", unmarshaled["message"])
}

func TestResultToJSONWithData(t *testing.T) {
	data := map[string]interface{}{
		"count": 5,
		"items": []string{"a", "b", "c"},
	}
	result := domain.NewResultWithData("found items", data)

	jsonBytes, err := result.ToJSON()
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, true, unmarshaled["success"])
	assert.Equal(t, "found items", unmarshaled["message"])
	assert.NotNil(t, unmarshaled["data"])
}

func TestResultToJSONWithErrors(t *testing.T) {
	domainErr := domain.Error{
		Path:     "test.go",
		Line:     10,
		Severity: "error",
		Message:  "test error",
	}
	result := domain.NewErrorResult("failed", domainErr)

	jsonBytes, err := result.ToJSON()
	require.NoError(t, err)

	var unmarshaled map[string]interface{}
	unmarshalErr := json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, unmarshalErr)

	assert.Equal(t, false, unmarshaled["success"])
	assert.Equal(t, "failed", unmarshaled["message"])
	assert.NotNil(t, unmarshaled["errors"])
}

// Test Error formatting
func TestErrorString(t *testing.T) {
	err := domain.Error{
		Path:     "test.go",
		Line:     10,
		Column:   5,
		Severity: "error",
		Message:  "syntax error",
		Code:     "E001",
	}

	str := err.String()

	assert.Contains(t, str, "test.go")
	assert.Contains(t, str, "10")
	assert.Contains(t, str, "5")
	assert.Contains(t, str, "error")
	assert.Contains(t, str, "syntax error")
}

func TestErrorStringMinimal(t *testing.T) {
	err := domain.Error{
		Message: "simple error",
	}

	str := err.String()

	assert.Contains(t, str, "simple error")
}

// Test Context
func TestNewContext(t *testing.T) {
	ctx := domain.NewContext(context.Background(), "/test/path")

	assert.NotNil(t, ctx)
	assert.Equal(t, "/test/path", ctx.WorkDir)
	assert.False(t, ctx.Verbose)
}

func TestNewContextWithVerbose(t *testing.T) {
	ctx := domain.NewContextWithVerbose(context.Background(), "/test/path", true)

	assert.NotNil(t, ctx)
	assert.Equal(t, "/test/path", ctx.WorkDir)
	assert.True(t, ctx.Verbose)
}

// Compile-time interface checks
// These ensure that example implementations satisfy the interfaces

type mockDomain struct{}

func (m *mockDomain) Name() string                    { return "mock" }
func (m *mockDomain) Version() string                 { return "1.0.0" }
func (m *mockDomain) Builder() domain.Builder         { return nil }
func (m *mockDomain) Linter() domain.Linter           { return nil }
func (m *mockDomain) Initializer() domain.Initializer { return nil }
func (m *mockDomain) Validator() domain.Validator     { return nil }

type mockImporterDomain struct {
	mockDomain
}

func (m *mockImporterDomain) Importer() domain.Importer { return nil }

type mockListerDomain struct {
	mockDomain
}

func (m *mockListerDomain) Lister() domain.Lister { return nil }

type mockGrapherDomain struct {
	mockDomain
}

func (m *mockGrapherDomain) Grapher() domain.Grapher { return nil }

// Test interface implementations compile
func TestInterfaceImplementations(t *testing.T) {
	var _ domain.Domain = (*mockDomain)(nil)
	var _ domain.ImporterDomain = (*mockImporterDomain)(nil)
	var _ domain.ListerDomain = (*mockListerDomain)(nil)
	var _ domain.GrapherDomain = (*mockGrapherDomain)(nil)
}

// Test operation interfaces with mock implementations
type mockBuilder struct{}

func (b *mockBuilder) Build(ctx *domain.Context, path string, opts domain.BuildOpts) (*domain.Result, error) {
	return domain.NewResult("build successful"), nil
}

type mockLinter struct{}

func (l *mockLinter) Lint(ctx *domain.Context, path string, opts domain.LintOpts) (*domain.Result, error) {
	return domain.NewResult("lint successful"), nil
}

type mockInitializer struct{}

func (i *mockInitializer) Init(ctx *domain.Context, path string, opts domain.InitOpts) (*domain.Result, error) {
	return domain.NewResult("init successful"), nil
}

type mockValidator struct{}

func (v *mockValidator) Validate(ctx *domain.Context, path string, opts domain.ValidateOpts) (*domain.Result, error) {
	return domain.NewResult("validate successful"), nil
}

type mockImporter struct{}

func (i *mockImporter) Import(ctx *domain.Context, source string, opts domain.ImportOpts) (*domain.Result, error) {
	return domain.NewResult("import successful"), nil
}

type mockLister struct{}

func (l *mockLister) List(ctx *domain.Context, path string, opts domain.ListOpts) (*domain.Result, error) {
	return domain.NewResult("list successful"), nil
}

type mockGrapher struct{}

func (g *mockGrapher) Graph(ctx *domain.Context, path string, opts domain.GraphOpts) (*domain.Result, error) {
	return domain.NewResult("graph successful"), nil
}

func TestOperationInterfaces(t *testing.T) {
	var _ domain.Builder = (*mockBuilder)(nil)
	var _ domain.Linter = (*mockLinter)(nil)
	var _ domain.Initializer = (*mockInitializer)(nil)
	var _ domain.Validator = (*mockValidator)(nil)
	var _ domain.Importer = (*mockImporter)(nil)
	var _ domain.Lister = (*mockLister)(nil)
	var _ domain.Grapher = (*mockGrapher)(nil)
}

func TestOperationExecution(t *testing.T) {
	ctx := domain.NewContext(context.Background(), "/test")

	t.Run("Builder", func(t *testing.T) {
		builder := &mockBuilder{}
		result, err := builder.Build(ctx, "/test/path", domain.BuildOpts{Format: "json"})
		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("Linter", func(t *testing.T) {
		linter := &mockLinter{}
		result, err := linter.Lint(ctx, "/test/path", domain.LintOpts{Format: "text"})
		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("Initializer", func(t *testing.T) {
		initializer := &mockInitializer{}
		result, err := initializer.Init(ctx, "/test/path", domain.InitOpts{Name: "test"})
		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("Validator", func(t *testing.T) {
		validator := &mockValidator{}
		result, err := validator.Validate(ctx, "/test/path", domain.ValidateOpts{})
		require.NoError(t, err)
		assert.True(t, result.Success)
	})
}
