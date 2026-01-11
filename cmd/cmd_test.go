package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRootCommand(t *testing.T) {
	root := NewRootCommand("test-cli", "Test CLI for wetwire")
	assert.Equal(t, "test-cli", root.Use)
	assert.Contains(t, root.Short, "Test CLI")
}

func TestRootCommandExecute(t *testing.T) {
	root := NewRootCommand("test-cli", "Test CLI")
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"--help"})
	err := root.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Test CLI")
}

type mockBuilder struct {
	called bool
	path   string
	opts   BuildOptions
}

func (m *mockBuilder) Build(ctx context.Context, path string, opts BuildOptions) error {
	m.called = true
	m.path = path
	m.opts = opts
	return nil
}

func TestNewBuildCommand(t *testing.T) {
	mb := &mockBuilder{}
	cmd := NewBuildCommand(mb)
	assert.Equal(t, "build", cmd.Use)
}

func TestBuildCommandExecution(t *testing.T) {
	mb := &mockBuilder{}
	root := NewRootCommand("test-cli", "Test")
	root.AddCommand(NewBuildCommand(mb))
	root.SetArgs([]string{"build", "--path", "."})
	err := root.Execute()
	require.NoError(t, err)
	assert.True(t, mb.called)
	assert.Equal(t, ".", mb.path)
}

type mockLinter struct {
	called bool
	path   string
	issues []Issue
}

func (m *mockLinter) Lint(ctx context.Context, path string, opts LintOptions) ([]Issue, error) {
	m.called = true
	m.path = path
	return m.issues, nil
}

func TestNewLintCommand(t *testing.T) {
	ml := &mockLinter{}
	cmd := NewLintCommand(ml)
	assert.Equal(t, "lint", cmd.Use)
}

func TestLintCommandExecution(t *testing.T) {
	ml := &mockLinter{}
	root := NewRootCommand("test-cli", "Test")
	root.AddCommand(NewLintCommand(ml))
	root.SetArgs([]string{"lint", "--path", "."})
	err := root.Execute()
	require.NoError(t, err)
	assert.True(t, ml.called)
}

type mockInitializer struct {
	called bool
	name   string
}

func (m *mockInitializer) Init(ctx context.Context, name string, opts InitOptions) error {
	m.called = true
	m.name = name
	return nil
}

func TestNewInitCommand(t *testing.T) {
	mi := &mockInitializer{}
	cmd := NewInitCommand(mi)
	assert.Equal(t, "init [name]", cmd.Use)
}

func TestInitCommandExecution(t *testing.T) {
	mi := &mockInitializer{}
	root := NewRootCommand("test-cli", "Test")
	root.AddCommand(NewInitCommand(mi))
	root.SetArgs([]string{"init", "my-project"})
	err := root.Execute()
	require.NoError(t, err)
	assert.True(t, mi.called)
	assert.Equal(t, "my-project", mi.name)
}

type mockValidator struct {
	called bool
	path   string
}

func (m *mockValidator) Validate(ctx context.Context, path string, opts ValidateOptions) ([]ValidationError, error) {
	m.called = true
	m.path = path
	return nil, nil
}

func TestNewValidateCommand(t *testing.T) {
	mv := &mockValidator{}
	cmd := NewValidateCommand(mv)
	assert.Equal(t, "validate", cmd.Use)
}

func TestValidateCommandExecution(t *testing.T) {
	mv := &mockValidator{}
	root := NewRootCommand("test-cli", "Test")
	root.AddCommand(NewValidateCommand(mv))
	root.SetArgs([]string{"validate", "--path", "."})
	err := root.Execute()
	require.NoError(t, err)
	assert.True(t, mv.called)
}
