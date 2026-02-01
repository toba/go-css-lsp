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
