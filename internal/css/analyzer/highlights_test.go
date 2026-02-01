package analyzer

import (
	"testing"

	"github.com/toba/go-css-lsp/internal/css/parser"
)

func TestFindDocumentHighlights(t *testing.T) {
	src := []byte(`:root { --color: red; }
.foo { color: var(--color); }
.bar { background: var(--color); }`)
	ss, _ := parser.Parse(src)

	offset := indexOf(src, "--color: red")
	highlights := FindDocumentHighlights(ss, src, offset)

	// 1 write (declaration) + 2 reads (var() usages) = 3
	if len(highlights) != 3 {
		t.Fatalf("expected 3 highlights, got %d", len(highlights))
	}

	if highlights[0].Kind != HighlightWrite {
		t.Errorf("expected write highlight for declaration")
	}
	if highlights[1].Kind != HighlightRead {
		t.Errorf("expected read highlight for var() usage")
	}
}

func TestFindDocumentHighlights_NotCustomProp(t *testing.T) {
	src := []byte(`.foo { color: red; }`)
	ss, _ := parser.Parse(src)

	offset := indexOf(src, "red")
	highlights := FindDocumentHighlights(ss, src, offset)

	if len(highlights) != 0 {
		t.Fatalf("expected 0 highlights, got %d", len(highlights))
	}
}
