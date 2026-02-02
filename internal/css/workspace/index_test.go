package workspace

import (
	"testing"

	"github.com/toba/go-css-lsp/internal/css/parser"
)

func TestIndex_IndexFile(t *testing.T) {
	idx := NewIndex()

	src := []byte(`:root {
  --primary: #ff0000;
  --secondary: blue;
}`)
	idx.IndexFile("file:///test.css", src)

	defs := idx.LookupDefinitions("--primary")
	if len(defs) != 1 {
		t.Fatalf(
			"expected 1 definition for --primary, got %d",
			len(defs),
		)
	}
	if defs[0].URI != "file:///test.css" {
		t.Errorf("unexpected URI: %s", defs[0].URI)
	}

	defs = idx.LookupDefinitions("--secondary")
	if len(defs) != 1 {
		t.Fatalf(
			"expected 1 definition for --secondary, got %d",
			len(defs),
		)
	}
}

func TestIndex_MultipleFiles(t *testing.T) {
	idx := NewIndex()

	idx.IndexFile("file:///a.css", []byte(
		`:root { --color: red; }`,
	))
	idx.IndexFile("file:///b.css", []byte(
		`.dark { --color: blue; }`,
	))

	defs := idx.LookupDefinitions("--color")
	if len(defs) != 2 {
		t.Fatalf(
			"expected 2 definitions for --color, got %d",
			len(defs),
		)
	}
}

func TestIndex_RemoveFile(t *testing.T) {
	idx := NewIndex()

	idx.IndexFile("file:///a.css", []byte(
		`:root { --color: red; }`,
	))
	idx.IndexFile("file:///b.css", []byte(
		`.dark { --color: blue; }`,
	))

	idx.RemoveFile("file:///a.css")

	defs := idx.LookupDefinitions("--color")
	if len(defs) != 1 {
		t.Fatalf(
			"expected 1 definition after removal, got %d",
			len(defs),
		)
	}
	if defs[0].URI != "file:///b.css" {
		t.Errorf("expected b.css, got %s", defs[0].URI)
	}
}

func TestIndex_ReindexFile(t *testing.T) {
	idx := NewIndex()

	idx.IndexFile("file:///a.css", []byte(
		`:root { --old: red; }`,
	))

	defs := idx.LookupDefinitions("--old")
	if len(defs) != 1 {
		t.Fatal("expected --old definition")
	}

	// Re-index with different content
	idx.IndexFile("file:///a.css", []byte(
		`:root { --new: blue; }`,
	))

	defs = idx.LookupDefinitions("--old")
	if len(defs) != 0 {
		t.Error("--old should be gone after re-index")
	}

	defs = idx.LookupDefinitions("--new")
	if len(defs) != 1 {
		t.Error("expected --new definition after re-index")
	}
}

func TestIndex_CrossFileDefinitionLookup(t *testing.T) {
	idx := NewIndex()

	// File A defines the variable
	srcA := []byte(`:root { --color-surface: #fff; }`)
	idx.IndexFile("file:///tokens.css", srcA)

	// File B uses the variable but does not define it
	srcB := []byte(`.card { background: var(--color-surface); }`)
	idx.IndexFile("file:///card.css", srcB)

	// Same-file lookup in card.css should find nothing
	defs := idx.LookupDefinitions("--color-surface")
	found := false
	for _, d := range defs {
		if d.URI == "file:///card.css" {
			found = true
		}
	}
	if found {
		t.Error("card.css should not contain a definition")
	}

	// Workspace lookup should find the definition in tokens.css
	if len(defs) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(defs))
	}
	if defs[0].URI != "file:///tokens.css" {
		t.Errorf(
			"expected tokens.css, got %s", defs[0].URI,
		)
	}

	// Verify offsets point to the correct text
	propText := string(
		srcA[defs[0].StartPos:defs[0].EndPos],
	)
	if propText != "--color-surface" {
		t.Errorf(
			"expected --color-surface, got %q", propText,
		)
	}
}

func TestVariableDefinitionRawValue(t *testing.T) {
	idx := NewIndex()

	src := []byte(`:root {
  --color: #ff0000;
  --bg: rgb(0, 0, 0);
  --size: 16px;
}`)
	idx.IndexFile("file:///test.css", src)

	defs := idx.LookupDefinitions("--color")
	if len(defs) != 1 {
		t.Fatalf("expected 1 def, got %d", len(defs))
	}
	if defs[0].RawValue != "#ff0000" {
		t.Errorf(
			"expected RawValue #ff0000, got %q",
			defs[0].RawValue,
		)
	}

	defs = idx.LookupDefinitions("--bg")
	if len(defs) != 1 {
		t.Fatalf("expected 1 def, got %d", len(defs))
	}
	if defs[0].RawValue != "rgb(0, 0, 0)" {
		t.Errorf(
			"expected RawValue rgb(0, 0, 0), got %q",
			defs[0].RawValue,
		)
	}

	defs = idx.LookupDefinitions("--size")
	if len(defs) != 1 {
		t.Fatalf("expected 1 def, got %d", len(defs))
	}
	if defs[0].RawValue != "16px" {
		t.Errorf(
			"expected RawValue 16px, got %q",
			defs[0].RawValue,
		)
	}
}

func TestResolveVariable(t *testing.T) {
	idx := NewIndex()

	idx.IndexFile("file:///test.css", []byte(
		`:root { --primary: #ff0000; }`,
	))

	val, ok := idx.ResolveVariable("--primary")
	if !ok {
		t.Fatal("expected ResolveVariable to return ok")
	}
	if val != "#ff0000" {
		t.Errorf("expected #ff0000, got %q", val)
	}

	_, ok = idx.ResolveVariable("--unknown")
	if ok {
		t.Error("expected ResolveVariable to return !ok for unknown")
	}
}

func TestIndex_AllVariableNames(t *testing.T) {
	idx := NewIndex()

	idx.IndexFile("file:///a.css", []byte(
		`:root { --color: red; --size: 16px; }`,
	))

	names := idx.AllVariableNames()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
}

func TestFindReferences(t *testing.T) {
	idx := NewIndex()

	srcA := []byte(`:root { --color: red; }`)
	srcB := []byte(`.card { background: var(--color); }`)
	srcC := []byte(`.btn { color: var(--color); }`)

	idx.IndexFile("file:///a.css", srcA)
	idx.IndexFile("file:///b.css", srcB)
	idx.IndexFile("file:///c.css", srcC)

	files := map[string][]byte{
		"file:///a.css": srcA,
		"file:///b.css": srcB,
		"file:///c.css": srcC,
	}

	refs := idx.FindReferences(
		"--color", files,
		map[string]*parser.Stylesheet{},
	)

	// 1 definition + 2 usages = 3
	if len(refs) != 3 {
		t.Fatalf("expected 3 references, got %d", len(refs))
	}
}

func TestIndexFileWithStylesheet_NilStylesheet(t *testing.T) {
	idx := NewIndex()

	// Should not panic with nil stylesheet
	idx.IndexFileWithStylesheet("file:///x.css", nil, nil)

	names := idx.AllVariableNames()
	if len(names) != 0 {
		t.Errorf("expected 0 names, got %d", len(names))
	}
}

func TestIndexFile_EmptySource(t *testing.T) {
	idx := NewIndex()

	idx.IndexFile("file:///empty.css", []byte(""))

	names := idx.AllVariableNames()
	if len(names) != 0 {
		t.Errorf("expected 0 names, got %d", len(names))
	}
}

func TestIndexFile_VarsInMediaRule(t *testing.T) {
	idx := NewIndex()

	src := []byte(`@media (prefers-color-scheme: dark) {
  :root { --bg: #000; }
}`)
	idx.IndexFile("file:///theme.css", src)

	defs := idx.LookupDefinitions("--bg")
	if len(defs) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(defs))
	}
	if defs[0].RawValue != "#000" {
		t.Errorf(
			"expected RawValue '#000', got %q",
			defs[0].RawValue,
		)
	}
}
