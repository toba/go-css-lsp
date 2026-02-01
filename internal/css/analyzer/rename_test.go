package analyzer

import (
	"testing"

	"github.com/toba/go-css-lsp/internal/css/parser"
)

func TestPrepareRename_OnDeclaration(t *testing.T) {
	src := []byte(`:root { --color: red; }`)
	ss, _ := parser.Parse(src)

	offset := indexOf(src, "--color:")
	loc, found := PrepareRename(ss, src, offset)

	if !found {
		t.Fatal("expected rename to be valid")
	}

	text := string(src[loc.StartPos:loc.EndPos])
	if text != "--color" {
		t.Errorf("expected --color, got %q", text)
	}
}

func TestPrepareRename_NotRenameable(t *testing.T) {
	src := []byte(`.foo { color: red; }`)
	ss, _ := parser.Parse(src)

	offset := indexOf(src, "red")
	_, found := PrepareRename(ss, src, offset)

	if found {
		t.Error("expected rename to not be valid for 'red'")
	}
}

func TestRename(t *testing.T) {
	src := []byte(`:root { --color: red; }
.foo { color: var(--color); }`)
	ss, _ := parser.Parse(src)

	offset := indexOf(src, "--color:")
	edits := Rename(ss, src, offset, "--primary")

	if len(edits) != 2 {
		t.Fatalf("expected 2 edits, got %d", len(edits))
	}

	for _, e := range edits {
		if e.NewText != "--primary" {
			t.Errorf(
				"expected --primary, got %s", e.NewText,
			)
		}
	}
}

func TestRename_AutoPrefix(t *testing.T) {
	src := []byte(`:root { --color: red; }`)
	ss, _ := parser.Parse(src)

	offset := indexOf(src, "--color:")
	edits := Rename(ss, src, offset, "primary")

	if len(edits) != 1 {
		t.Fatalf("expected 1 edit, got %d", len(edits))
	}

	if edits[0].NewText != "--primary" {
		t.Errorf(
			"expected --primary, got %s", edits[0].NewText,
		)
	}
}
