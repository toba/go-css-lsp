package analyzer

import (
	"testing"

	"github.com/toba/go-css-lsp/internal/css/parser"
)

func TestFindReferences_FromDeclaration(t *testing.T) {
	src := []byte(`:root { --color: red; }
.foo { color: var(--color); }
.bar { background: var(--color); }`)
	ss, _ := parser.Parse(src)

	// Cursor on the declaration of --color
	offset := indexOf(src, "--color: red")
	refs := FindReferences(ss, src, offset)

	// Should find: 1 declaration + 2 usages = 3
	if len(refs) != 3 {
		t.Fatalf("expected 3 references, got %d", len(refs))
	}
}

func TestFindReferences_FromUsage(t *testing.T) {
	src := []byte(`:root { --primary: blue; }
.foo { color: var(--primary); }`)
	ss, _ := parser.Parse(src)

	// Cursor on var(--primary)
	offset := indexOf(src, "var(--primary)")
	refs := FindReferences(ss, src, offset)

	// 1 declaration + 1 usage = 2
	if len(refs) != 2 {
		t.Fatalf("expected 2 references, got %d", len(refs))
	}
}

func TestFindReferences_NotCustomProperty(t *testing.T) {
	src := []byte(`.foo { color: red; }`)
	ss, _ := parser.Parse(src)

	offset := indexOf(src, "red")
	refs := FindReferences(ss, src, offset)

	if len(refs) != 0 {
		t.Fatalf("expected 0 references, got %d", len(refs))
	}
}

func TestFindReferences_NoUsages(t *testing.T) {
	src := []byte(`:root { --unused: red; }`)
	ss, _ := parser.Parse(src)

	offset := indexOf(src, "--unused")
	refs := FindReferences(ss, src, offset)

	// Just the declaration
	if len(refs) != 1 {
		t.Fatalf("expected 1 reference, got %d", len(refs))
	}
}
