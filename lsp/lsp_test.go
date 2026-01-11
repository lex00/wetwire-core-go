package lsp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDiagnosticProvider struct {
	diagnostics []Diagnostic
	called      bool
}

func (m *mockDiagnosticProvider) Diagnose(ctx context.Context, uri string) ([]Diagnostic, error) {
	m.called = true
	return m.diagnostics, nil
}

type mockCompletionProvider struct {
	items  []CompletionItem
	called bool
}

func (m *mockCompletionProvider) Complete(ctx context.Context, uri string, pos Position) ([]CompletionItem, error) {
	m.called = true
	return m.items, nil
}

type mockHoverProvider struct {
	hover  *Hover
	called bool
}

func (m *mockHoverProvider) Hover(ctx context.Context, uri string, pos Position) (*Hover, error) {
	m.called = true
	return m.hover, nil
}

func TestNewServer(t *testing.T) {
	config := Config{
		Name: "test-lsp",
	}
	server := NewServer(config)
	assert.NotNil(t, server)
	assert.Equal(t, "test-lsp", server.Name())
}

func TestServerWithDiagnostics(t *testing.T) {
	provider := &mockDiagnosticProvider{
		diagnostics: []Diagnostic{
			{
				Range: Range{
					Start: Position{Line: 0, Character: 0},
					End:   Position{Line: 0, Character: 10},
				},
				Severity: SeverityError,
				Message:  "test error",
			},
		},
	}

	config := Config{
		Name:   "test-lsp",
		Linter: provider,
	}
	server := NewServer(config)

	ctx := context.Background()
	diags, err := server.Diagnose(ctx, "file:///test.yaml")
	require.NoError(t, err)
	assert.True(t, provider.called)
	assert.Len(t, diags, 1)
	assert.Equal(t, "test error", diags[0].Message)
}

func TestServerWithCompletion(t *testing.T) {
	provider := &mockCompletionProvider{
		items: []CompletionItem{
			{
				Label:  "bucket",
				Kind:   CompletionKindProperty,
				Detail: "S3 bucket resource",
			},
		},
	}

	config := Config{
		Name:      "test-lsp",
		Completer: provider,
	}
	server := NewServer(config)

	ctx := context.Background()
	pos := Position{Line: 5, Character: 10}
	items, err := server.Complete(ctx, "file:///test.yaml", pos)
	require.NoError(t, err)
	assert.True(t, provider.called)
	assert.Len(t, items, 1)
	assert.Equal(t, "bucket", items[0].Label)
}

func TestServerWithHover(t *testing.T) {
	provider := &mockHoverProvider{
		hover: &Hover{
			Contents: "Bucket documentation",
		},
	}

	config := Config{
		Name:      "test-lsp",
		HoverDocs: provider,
	}
	server := NewServer(config)

	ctx := context.Background()
	pos := Position{Line: 5, Character: 10}
	hover, err := server.Hover(ctx, "file:///test.yaml", pos)
	require.NoError(t, err)
	assert.True(t, provider.called)
	assert.NotNil(t, hover)
	assert.Equal(t, "Bucket documentation", hover.Contents)
}

func TestDiagnosticSeverity(t *testing.T) {
	assert.Equal(t, DiagnosticSeverity(1), SeverityError)
	assert.Equal(t, DiagnosticSeverity(2), SeverityWarning)
	assert.Equal(t, DiagnosticSeverity(3), SeverityInformation)
	assert.Equal(t, DiagnosticSeverity(4), SeverityHint)
}

func TestPositionAndRange(t *testing.T) {
	pos := Position{Line: 10, Character: 5}
	assert.Equal(t, 10, pos.Line)
	assert.Equal(t, 5, pos.Character)

	r := Range{Start: pos, End: Position{Line: 10, Character: 15}}
	assert.Equal(t, 10, r.Start.Line)
	assert.Equal(t, 15, r.End.Character)
}

func TestCompletionItemKinds(t *testing.T) {
	assert.Equal(t, CompletionItemKind(6), CompletionKindProperty)
	assert.Equal(t, CompletionItemKind(7), CompletionKindClass)
	assert.Equal(t, CompletionItemKind(10), CompletionKindValue)
}

func TestServerNilProviders(t *testing.T) {
	config := Config{
		Name: "test-lsp",
		// No providers configured
	}
	server := NewServer(config)

	ctx := context.Background()

	// Should return empty results without error when providers are nil
	diags, err := server.Diagnose(ctx, "file:///test.yaml")
	require.NoError(t, err)
	assert.Empty(t, diags)

	items, err := server.Complete(ctx, "file:///test.yaml", Position{})
	require.NoError(t, err)
	assert.Empty(t, items)

	hover, err := server.Hover(ctx, "file:///test.yaml", Position{})
	require.NoError(t, err)
	assert.Nil(t, hover)
}
