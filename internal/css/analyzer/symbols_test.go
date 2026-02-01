package analyzer

import (
	"testing"

	"github.com/toba/go-css-lsp/internal/css/parser"
)

func TestFindDocumentSymbols_Rulesets(t *testing.T) {
	src := []byte(`
.foo { color: red; }
#bar { display: block; }
`)
	ss, _ := parser.Parse(src)
	symbols := FindDocumentSymbols(ss, src)

	if len(symbols) != 2 {
		t.Fatalf("expected 2 symbols, got %d", len(symbols))
	}

	if symbols[0].Name != ".foo" {
		t.Errorf("expected .foo, got %s", symbols[0].Name)
	}
	if symbols[0].Kind != SymbolKindClass {
		t.Errorf("expected class kind, got %d", symbols[0].Kind)
	}

	if symbols[1].Name != "#bar" {
		t.Errorf("expected #bar, got %s", symbols[1].Name)
	}
}

func TestFindDocumentSymbols_AtRules(t *testing.T) {
	src := []byte(`
@media (max-width: 768px) {
  .foo { color: red; }
}
`)
	ss, _ := parser.Parse(src)
	symbols := FindDocumentSymbols(ss, src)

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	if symbols[0].Kind != SymbolKindString {
		t.Errorf(
			"expected string kind for at-rule, got %d",
			symbols[0].Kind,
		)
	}

	if len(symbols[0].Children) != 1 {
		t.Fatalf(
			"expected 1 child, got %d",
			len(symbols[0].Children),
		)
	}

	if symbols[0].Children[0].Name != ".foo" {
		t.Errorf(
			"expected .foo child, got %s",
			symbols[0].Children[0].Name,
		)
	}
}

func TestFindDocumentSymbols_CustomProperties(t *testing.T) {
	src := []byte(`
:root {
  --primary: #ff0000;
  --secondary: blue;
}
`)
	ss, _ := parser.Parse(src)
	symbols := FindDocumentSymbols(ss, src)

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	children := symbols[0].Children
	if len(children) != 2 {
		t.Fatalf(
			"expected 2 variable children, got %d",
			len(children),
		)
	}

	if children[0].Name != "--primary" {
		t.Errorf(
			"expected --primary, got %s", children[0].Name,
		)
	}
	if children[0].Kind != SymbolKindVariable {
		t.Errorf(
			"expected variable kind, got %d", children[0].Kind,
		)
	}

	if children[1].Name != "--secondary" {
		t.Errorf(
			"expected --secondary, got %s", children[1].Name,
		)
	}
}

func TestFindDocumentSymbols_NestedRulesets(t *testing.T) {
	src := []byte(`
.parent {
  color: red;
  &:hover { color: blue; }
  .child { font-size: 14px; }
}
`)
	ss, _ := parser.Parse(src)
	symbols := FindDocumentSymbols(ss, src)

	if len(symbols) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(symbols))
	}

	parent := symbols[0]
	if parent.Name != ".parent" {
		t.Errorf("expected .parent, got %s", parent.Name)
	}

	// Should have 2 nested ruleset children
	if len(parent.Children) != 2 {
		t.Fatalf(
			"expected 2 children, got %d",
			len(parent.Children),
		)
	}
}

func TestFindDocumentSymbols_Empty(t *testing.T) {
	src := []byte(`/* empty */`)
	ss, _ := parser.Parse(src)
	symbols := FindDocumentSymbols(ss, src)

	if len(symbols) != 0 {
		t.Fatalf("expected 0 symbols, got %d", len(symbols))
	}
}

func TestFindDocumentSymbols_Nil(t *testing.T) {
	symbols := FindDocumentSymbols(nil, nil)
	if symbols != nil {
		t.Fatalf("expected nil for nil stylesheet")
	}
}
