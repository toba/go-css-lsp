package analyzer

import (
	"testing"

	"github.com/toba/go-css-lsp/internal/css/parser"
	"github.com/toba/go-css-lsp/internal/css/scanner"
)

func TestIsCustomProperty(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"--color", true},
		{"--", true},
		{"color", false},
		{"-color", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsCustomProperty(tt.name); got != tt.want {
			t.Errorf(
				"IsCustomProperty(%q) = %v, want %v",
				tt.name, got, tt.want,
			)
		}
	}
}

func TestFindCustomPropertyAt_Declaration(t *testing.T) {
	src := []byte(`:root { --primary: red; }`)
	ss, _ := parser.Parse(src)

	offset := indexOf(src, "--primary")
	name := FindCustomPropertyAt(ss, src, offset)
	if name != "--primary" {
		t.Errorf("expected --primary, got %q", name)
	}
}

func TestFindCustomPropertyAt_VarUsage(t *testing.T) {
	src := []byte(`.foo { color: var(--primary); }`)
	ss, _ := parser.Parse(src)

	offset := indexOf(src, "var(--primary)")
	name := FindCustomPropertyAt(ss, src, offset)
	if name != "--primary" {
		t.Errorf("expected --primary, got %q", name)
	}
}

func TestFindCustomPropertyAt_NotFound(t *testing.T) {
	src := []byte(`.foo { color: red; }`)
	ss, _ := parser.Parse(src)

	offset := indexOf(src, "red")
	name := FindCustomPropertyAt(ss, src, offset)
	if name != "" {
		t.Errorf("expected empty, got %q", name)
	}
}

func TestFindVarReferenceAt(t *testing.T) {
	src := []byte(`:root { --x: 1; }
.foo { color: var(--x); }`)
	ss, _ := parser.Parse(src)

	offset := indexOf(src, "var(--x)")
	name := FindVarReferenceAt(ss, src, offset)
	if name != "--x" {
		t.Errorf("expected --x, got %q", name)
	}
}

func TestFindVarReferenceAt_DirectIdent(t *testing.T) {
	src := []byte(`.foo { color: var(--abc); }`)
	ss, _ := parser.Parse(src)

	offset := indexOf(src, "--abc);")
	name := FindVarReferenceAt(ss, src, offset)
	if name != "--abc" {
		t.Errorf("expected --abc, got %q", name)
	}
}

func TestForEachVarUsage(t *testing.T) {
	src := []byte(`:root { --c: red; }
.a { color: var(--c); }
.b { background: var(--c); }`)
	ss, _ := parser.Parse(src)

	var found []scanner.Token
	ForEachVarUsage(ss, "--c", func(tok scanner.Token) {
		found = append(found, tok)
	})

	if len(found) != 2 {
		t.Fatalf("expected 2 usages, got %d", len(found))
	}
	for _, tok := range found {
		if tok.Value != "--c" {
			t.Errorf("expected --c token, got %q", tok.Value)
		}
	}
}

func TestForEachVarUsage_NoMatch(t *testing.T) {
	src := []byte(`.foo { color: var(--other); }`)
	ss, _ := parser.Parse(src)

	var count int
	ForEachVarUsage(ss, "--missing", func(_ scanner.Token) {
		count++
	})

	if count != 0 {
		t.Errorf("expected 0 usages, got %d", count)
	}
}
