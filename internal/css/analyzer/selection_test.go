package analyzer

import (
	"testing"

	"github.com/toba/go-css-lsp/internal/css/parser"
)

func TestFindSelectionRange(t *testing.T) {
	src := []byte(`.foo {
  color: red;
}`)
	ss, _ := parser.Parse(src)

	// Cursor on "red"
	offset := indexOf(src, "red")
	sr := FindSelectionRange(ss, offset)

	if sr == nil {
		t.Fatal("expected a selection range")
	}

	// Should have parent chain
	if sr.Parent == nil {
		t.Fatal("expected parent selection range")
	}
}

func TestFindSelectionRange_Nil(t *testing.T) {
	sr := FindSelectionRange(nil, 0)
	if sr != nil {
		t.Fatal("expected nil for nil stylesheet")
	}
}
