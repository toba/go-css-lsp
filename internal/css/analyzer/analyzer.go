// Package analyzer provides semantic analysis for CSS:
// diagnostics, hover, and completion.
package analyzer

// Severity levels for diagnostics.
const (
	SeverityError   = 1
	SeverityWarning = 2
	SeverityInfo    = 3
	SeverityHint    = 4
)

// Completion item kinds matching LSP spec.
const (
	KindProperty = 10
	KindKeyword  = 14
	KindUnit     = 11
	KindValue    = 12
	KindFunction = 3
	KindColor    = 16
)

// SymbolKind mirrors LSP SymbolKind values.
const (
	SymbolKindClass    = 5
	SymbolKindProperty = 7
	SymbolKindVariable = 13
	SymbolKindString   = 15
)

// DocumentHighlightKind constants.
const (
	HighlightText  = 1
	HighlightRead  = 2
	HighlightWrite = 3
)

// Diagnostic represents a diagnostic message.
type Diagnostic struct {
	Message   string
	StartLine int
	StartChar int
	EndLine   int
	EndChar   int
	Severity  int
}

// CompletionItem represents a completion suggestion.
type CompletionItem struct {
	Label         string
	Kind          int
	Detail        string
	Documentation string
	InsertText    string
}
