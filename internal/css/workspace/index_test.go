package workspace

import "testing"

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
