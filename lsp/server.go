package lsp

import "context"

// Config configures the LSP server.
type Config struct {
	// Name is the server name (e.g., "wetwire-aws-lsp")
	Name string

	// Linter provides diagnostics for documents
	Linter DiagnosticProvider

	// Completer provides completion items
	Completer CompletionProvider

	// HoverDocs provides hover documentation
	HoverDocs HoverProvider

	// Definitions provides go-to-definition support
	Definitions DefinitionProvider
}

// Server implements the LSP protocol.
type Server struct {
	config Config
}

// NewServer creates a new LSP server with the given configuration.
func NewServer(config Config) *Server {
	return &Server{
		config: config,
	}
}

// Name returns the server name.
func (s *Server) Name() string {
	return s.config.Name
}

// Diagnose runs diagnostics on the specified document.
func (s *Server) Diagnose(ctx context.Context, uri string) ([]Diagnostic, error) {
	if s.config.Linter == nil {
		return []Diagnostic{}, nil
	}
	return s.config.Linter.Diagnose(ctx, uri)
}

// Complete returns completion items at the specified position.
func (s *Server) Complete(ctx context.Context, uri string, pos Position) ([]CompletionItem, error) {
	if s.config.Completer == nil {
		return []CompletionItem{}, nil
	}
	return s.config.Completer.Complete(ctx, uri, pos)
}

// Hover returns hover information at the specified position.
func (s *Server) Hover(ctx context.Context, uri string, pos Position) (*Hover, error) {
	if s.config.HoverDocs == nil {
		return nil, nil
	}
	return s.config.HoverDocs.Hover(ctx, uri, pos)
}

// Definition returns definition locations at the specified position.
func (s *Server) Definition(ctx context.Context, uri string, pos Position) ([]Location, error) {
	if s.config.Definitions == nil {
		return []Location{}, nil
	}
	return s.config.Definitions.Definition(ctx, uri, pos)
}
