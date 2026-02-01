package analyzer

import (
	"testing"

	"github.com/toba/go-css-lsp/internal/css/parser"
)

func TestFindDefinition_VarReference(t *testing.T) {
	src := []byte(`:root { --primary: red; }
.foo { color: var(--primary); }`)
	ss, _ := parser.Parse(src)

	// Cursor on "var(--primary)" — offset at "var"
	// .foo { color: var(--primary); }
	// Position within var(--primary)
	varOffset := indexOf(src, "var(--primary)")
	loc, found := FindDefinition(ss, src, varOffset)

	if !found {
		t.Fatal("expected to find definition")
	}

	propText := string(src[loc.StartPos:loc.EndPos])
	if propText != "--primary" {
		t.Errorf(
			"expected --primary, got %q", propText,
		)
	}
}

func TestFindDefinition_CursorOnVarName(t *testing.T) {
	src := []byte(`:root { --color: blue; }
.foo { background: var(--color); }`)
	ss, _ := parser.Parse(src)

	// Cursor directly on "--color" inside var()
	varOffset := indexOf(src, "--color);")
	loc, found := FindDefinition(ss, src, varOffset)

	if !found {
		t.Fatal("expected to find definition")
	}

	propText := string(src[loc.StartPos:loc.EndPos])
	if propText != "--color" {
		t.Errorf("expected --color, got %q", propText)
	}
}

func TestFindDefinition_NotOnVar(t *testing.T) {
	src := []byte(`.foo { color: red; }`)
	ss, _ := parser.Parse(src)

	// Cursor on "red"
	offset := indexOf(src, "red")
	_, found := FindDefinition(ss, src, offset)

	if found {
		t.Error("expected not to find definition for 'red'")
	}
}

func TestFindDefinition_UndefinedVar(t *testing.T) {
	src := []byte(`.foo { color: var(--undefined); }`)
	ss, _ := parser.Parse(src)

	offset := indexOf(src, "var(--undefined)")
	_, found := FindDefinition(ss, src, offset)

	if found {
		t.Error(
			"expected not to find definition for undefined var",
		)
	}
}
