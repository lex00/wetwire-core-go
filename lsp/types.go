// Package lsp provides Language Server Protocol infrastructure for IDE integration.
//
// Domain packages implement DiagnosticProvider, CompletionProvider, and HoverProvider
// interfaces while this package handles LSP protocol communication.
package lsp

import "context"

// Position represents a position in a text document (0-based line and character).
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range represents a range in a text document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// DiagnosticSeverity represents the severity of a diagnostic.
type DiagnosticSeverity int

const (
	SeverityError       DiagnosticSeverity = 1
	SeverityWarning     DiagnosticSeverity = 2
	SeverityInformation DiagnosticSeverity = 3
	SeverityHint        DiagnosticSeverity = 4
)

// Diagnostic represents a diagnostic (error, warning, info, hint).
type Diagnostic struct {
	Range    Range              `json:"range"`
	Severity DiagnosticSeverity `json:"severity"`
	Code     string             `json:"code,omitempty"`
	Source   string             `json:"source,omitempty"`
	Message  string             `json:"message"`
}

// CompletionItemKind represents the kind of completion item.
type CompletionItemKind int

const (
	CompletionKindText          CompletionItemKind = 1
	CompletionKindMethod        CompletionItemKind = 2
	CompletionKindFunction      CompletionItemKind = 3
	CompletionKindConstructor   CompletionItemKind = 4
	CompletionKindField         CompletionItemKind = 5
	CompletionKindProperty      CompletionItemKind = 6
	CompletionKindClass         CompletionItemKind = 7
	CompletionKindInterface     CompletionItemKind = 8
	CompletionKindModule        CompletionItemKind = 9
	CompletionKindValue         CompletionItemKind = 10
	CompletionKindKeyword       CompletionItemKind = 14
	CompletionKindSnippet       CompletionItemKind = 15
	CompletionKindColor         CompletionItemKind = 16
	CompletionKindFile          CompletionItemKind = 17
	CompletionKindReference     CompletionItemKind = 18
	CompletionKindFolder        CompletionItemKind = 19
	CompletionKindEnumMember    CompletionItemKind = 20
	CompletionKindConstant      CompletionItemKind = 21
	CompletionKindStruct        CompletionItemKind = 22
	CompletionKindEvent         CompletionItemKind = 23
	CompletionKindOperator      CompletionItemKind = 24
	CompletionKindTypeParameter CompletionItemKind = 25
)

// CompletionItem represents a completion suggestion.
type CompletionItem struct {
	Label         string             `json:"label"`
	Kind          CompletionItemKind `json:"kind,omitempty"`
	Detail        string             `json:"detail,omitempty"`
	Documentation string             `json:"documentation,omitempty"`
	InsertText    string             `json:"insertText,omitempty"`
	FilterText    string             `json:"filterText,omitempty"`
}

// Hover represents hover information.
type Hover struct {
	Contents string `json:"contents"`
	Range    *Range `json:"range,omitempty"`
}

// Location represents a location in a document.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// DiagnosticProvider provides diagnostics for a document.
type DiagnosticProvider interface {
	Diagnose(ctx context.Context, uri string) ([]Diagnostic, error)
}

// CompletionProvider provides completion items at a position.
type CompletionProvider interface {
	Complete(ctx context.Context, uri string, pos Position) ([]CompletionItem, error)
}

// HoverProvider provides hover information at a position.
type HoverProvider interface {
	Hover(ctx context.Context, uri string, pos Position) (*Hover, error)
}

// DefinitionProvider provides go-to-definition support.
type DefinitionProvider interface {
	Definition(ctx context.Context, uri string, pos Position) ([]Location, error)
}
