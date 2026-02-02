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

// ExperimentalMode controls how experimental CSS features are
// reported.
type ExperimentalMode int

const (
	// ExperimentalWarn emits a warning diagnostic (default).
	ExperimentalWarn ExperimentalMode = iota
	// ExperimentalIgnore suppresses experimental diagnostics
	// and completion tags.
	ExperimentalIgnore
	// ExperimentalError treats experimental features as errors.
	ExperimentalError
)

// DeprecatedMode controls how deprecated (obsolete) CSS
// features are reported.
type DeprecatedMode int

const (
	// DeprecatedWarn emits a warning diagnostic (default).
	DeprecatedWarn DeprecatedMode = iota
	// DeprecatedIgnore suppresses deprecated diagnostics
	// and completion tags.
	DeprecatedIgnore
	// DeprecatedError treats deprecated features as errors.
	DeprecatedError
)

// UnknownValueMode controls how unrecognized value keywords are
// reported.
type UnknownValueMode int

const (
	// UnknownValueWarn emits a warning diagnostic (default).
	UnknownValueWarn UnknownValueMode = iota
	// UnknownValueIgnore suppresses unknown value diagnostics.
	UnknownValueIgnore
	// UnknownValueError treats unknown values as errors.
	UnknownValueError
)

// LintOptions configures analyzer behavior.
type LintOptions struct {
	Experimental     ExperimentalMode
	Deprecated       DeprecatedMode
	UnknownValues    UnknownValueMode
	StrictColorNames bool
}

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
	Tags          []int
	Deprecated    bool
}
