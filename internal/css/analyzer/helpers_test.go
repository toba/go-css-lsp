package analyzer

import (
	"testing"

	"github.com/toba/go-css-lsp/internal/css/parser"
)

// indexOf returns the byte offset of substr in src, or -1.
func indexOf(src []byte, substr string) int {
	s := string(src)
	for i := range len(s) - len(substr) + 1 {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// parseCSS parses CSS source and fails the test on nil.
func parseCSS(
	t *testing.T,
	src []byte,
) *parser.Stylesheet {
	t.Helper()
	ss, _ := parser.Parse(src)
	if ss == nil {
		t.Fatal("failed to parse CSS")
	}
	return ss
}

// findDiagnostic searches for a diagnostic with the given
// message.
func findDiagnostic(
	diags []Diagnostic,
	msg string,
) (Diagnostic, bool) { //nolint:unparam
	for _, d := range diags {
		if d.Message == msg {
			return d, true
		}
	}
	return Diagnostic{}, false
}
