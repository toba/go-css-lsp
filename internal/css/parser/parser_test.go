package parser

import (
	"testing"
)

func TestParseSimpleRule(t *testing.T) {
	src := `body { color: red; }`
	ss, errs := Parse([]byte(src))

	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if len(ss.Children) != 1 {
		t.Fatalf(
			"expected 1 child, got %d", len(ss.Children),
		)
	}

	rs, ok := ss.Children[0].(*Ruleset)
	if !ok {
		t.Fatalf("expected Ruleset, got %T", ss.Children[0])
	}

	if len(rs.Declarations()) != 1 {
		t.Fatalf(
			"expected 1 declaration, got %d",
			len(rs.Declarations()),
		)
	}

	decl := rs.Declarations()[0]
	if decl.Property.Value != "color" {
		t.Errorf(
			"expected property 'color', got %q",
			decl.Property.Value,
		)
	}

	if decl.Value == nil || len(decl.Value.Tokens) == 0 {
		t.Fatal("expected value tokens")
	}

	if decl.Value.Tokens[0].Value != "red" {
		t.Errorf(
			"expected value 'red', got %q",
			decl.Value.Tokens[0].Value,
		)
	}
}

func TestParseMultipleDeclarations(t *testing.T) {
	src := `h1 { color: blue; font-size: 16px; }`
	ss, errs := Parse([]byte(src))

	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	rs := ss.Children[0].(*Ruleset)
	if len(rs.Declarations()) != 2 {
		t.Fatalf(
			"expected 2 declarations, got %d",
			len(rs.Declarations()),
		)
	}

	if rs.Declarations()[0].Property.Value != "color" {
		t.Errorf("expected 'color', got %q",
			rs.Declarations()[0].Property.Value)
	}
	if rs.Declarations()[1].Property.Value != "font-size" {
		t.Errorf("expected 'font-size', got %q",
			rs.Declarations()[1].Property.Value)
	}
}

func TestParseImportant(t *testing.T) {
	src := `p { color: red !important; }`
	ss, errs := Parse([]byte(src))

	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	rs := ss.Children[0].(*Ruleset)
	if !rs.Declarations()[0].Important {
		t.Error("expected !important flag")
	}
}

func TestParseAtRule(t *testing.T) {
	src := `@import url("style.css");`
	ss, errs := Parse([]byte(src))

	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if len(ss.Children) != 1 {
		t.Fatalf(
			"expected 1 child, got %d", len(ss.Children),
		)
	}

	at, ok := ss.Children[0].(*AtRule)
	if !ok {
		t.Fatalf("expected AtRule, got %T", ss.Children[0])
	}

	if at.Name != "import" {
		t.Errorf("expected 'import', got %q", at.Name)
	}
}

func TestParseMediaRule(t *testing.T) {
	src := `@media (max-width: 600px) { body { color: red; } }`
	ss, errs := Parse([]byte(src))

	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	at := ss.Children[0].(*AtRule)
	if at.Name != "media" {
		t.Errorf("expected 'media', got %q", at.Name)
	}
	if at.Block == nil {
		t.Fatal("expected block in @media")
	}
}

func TestParseMultipleSelectors(t *testing.T) {
	src := `h1, h2, h3 { color: blue; }`
	ss, errs := Parse([]byte(src))

	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	rs := ss.Children[0].(*Ruleset)
	if rs.Selectors == nil ||
		len(rs.Selectors.Selectors) != 3 {
		t.Fatalf(
			"expected 3 selectors, got %d",
			len(rs.Selectors.Selectors),
		)
	}
}

func TestParseComment(t *testing.T) {
	src := `/* header */ body { color: red; }`
	ss, errs := Parse([]byte(src))

	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if len(ss.Children) < 2 {
		t.Fatalf(
			"expected at least 2 children, got %d",
			len(ss.Children),
		)
	}

	_, ok := ss.Children[0].(*Comment)
	if !ok {
		t.Errorf(
			"expected Comment, got %T", ss.Children[0],
		)
	}
}

func TestParseErrorRecovery(t *testing.T) {
	src := `body { color: ; font-size: 16px; }`
	ss, _ := Parse([]byte(src))

	// Should still produce a partial AST
	if len(ss.Children) == 0 {
		t.Fatal("expected partial AST after error")
	}
}

func TestParseCustomProperty(t *testing.T) {
	src := `:root { --main-color: #333; }`
	ss, errs := Parse([]byte(src))

	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	rs := ss.Children[0].(*Ruleset)
	if len(rs.Declarations()) != 1 {
		t.Fatalf(
			"expected 1 declaration, got %d",
			len(rs.Declarations()),
		)
	}

	decl := rs.Declarations()[0]
	if decl.Property.Value != "--main-color" {
		t.Errorf("expected '--main-color', got %q",
			decl.Property.Value)
	}
}

func TestParseEmpty(t *testing.T) {
	ss, errs := Parse([]byte(``))

	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if len(ss.Children) != 0 {
		t.Errorf(
			"expected 0 children, got %d",
			len(ss.Children),
		)
	}
}

func TestParseNesting_AmpersandSelector(t *testing.T) {
	src := `.parent { color: red; &:hover { color: blue; } }`
	ss, errs := Parse([]byte(src))

	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	rs := ss.Children[0].(*Ruleset)
	if len(rs.Declarations()) != 1 {
		t.Fatalf(
			"expected 1 declaration, got %d",
			len(rs.Declarations()),
		)
	}

	// Find nested ruleset
	var nested *Ruleset
	for _, child := range rs.Children {
		if r, ok := child.(*Ruleset); ok {
			nested = r
			break
		}
	}
	if nested == nil {
		t.Fatal("expected nested ruleset")
	}
	if len(nested.Declarations()) != 1 {
		t.Fatalf(
			"expected 1 nested declaration, got %d",
			len(nested.Declarations()),
		)
	}
	if nested.Declarations()[0].Property.Value != "color" {
		t.Errorf("expected 'color', got %q",
			nested.Declarations()[0].Property.Value)
	}
}

func TestParseNesting_BareSelector(t *testing.T) {
	src := `.parent { color: red; .child { font-size: 14px; } }`
	ss, errs := Parse([]byte(src))

	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	rs := ss.Children[0].(*Ruleset)
	if len(rs.Declarations()) != 1 {
		t.Fatalf(
			"expected 1 declaration, got %d",
			len(rs.Declarations()),
		)
	}

	var nested *Ruleset
	for _, child := range rs.Children {
		if r, ok := child.(*Ruleset); ok {
			nested = r
			break
		}
	}
	if nested == nil {
		t.Fatal("expected nested ruleset for .child")
	}
	if len(nested.Declarations()) != 1 {
		t.Fatalf(
			"expected 1 nested declaration, got %d",
			len(nested.Declarations()),
		)
	}
}

func TestParseNesting_IdentDisambiguation(t *testing.T) {
	// "a:hover { ... }" starts with ident then colon,
	// but should be parsed as a nested selector, not a
	// declaration.
	src := `.parent { a:hover { color: blue; } }`
	ss, errs := Parse([]byte(src))

	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	rs := ss.Children[0].(*Ruleset)
	// No declarations in parent
	if len(rs.Declarations()) != 0 {
		t.Fatalf(
			"expected 0 declarations, got %d",
			len(rs.Declarations()),
		)
	}

	var nested *Ruleset
	for _, child := range rs.Children {
		if r, ok := child.(*Ruleset); ok {
			nested = r
			break
		}
	}
	if nested == nil {
		t.Fatal("expected nested ruleset for a:hover")
	}
	if len(nested.Declarations()) != 1 {
		t.Errorf("expected 1 declaration in nested rule")
	}
}

func TestParseNesting_NestedAtRule(t *testing.T) {
	src := `.parent { color: red; @media (hover) { color: blue; } }`
	ss, errs := Parse([]byte(src))

	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	rs := ss.Children[0].(*Ruleset)
	if len(rs.Declarations()) != 1 {
		t.Fatalf(
			"expected 1 declaration, got %d",
			len(rs.Declarations()),
		)
	}

	var nested *AtRule
	for _, child := range rs.Children {
		if a, ok := child.(*AtRule); ok {
			nested = a
			break
		}
	}
	if nested == nil {
		t.Fatal("expected nested at-rule")
	}
	if nested.Name != "media" {
		t.Errorf("expected 'media', got %q", nested.Name)
	}
}

func TestParseNesting_MultiLevel(t *testing.T) {
	src := `.a { .b { .c { color: red; } } }`
	ss, errs := Parse([]byte(src))

	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	a := ss.Children[0].(*Ruleset)
	var b *Ruleset
	for _, child := range a.Children {
		if r, ok := child.(*Ruleset); ok {
			b = r
			break
		}
	}
	if b == nil {
		t.Fatal("expected nested .b ruleset")
	}

	var c *Ruleset
	for _, child := range b.Children {
		if r, ok := child.(*Ruleset); ok {
			c = r
			break
		}
	}
	if c == nil {
		t.Fatal("expected nested .c ruleset")
	}
	if len(c.Declarations()) != 1 {
		t.Errorf("expected 1 declaration in .c")
	}
}

func TestParseNesting_MixedDeclsAndNested(t *testing.T) {
	src := `.parent {
	color: red;
	.child { font-size: 14px; }
	background: blue;
}`
	ss, errs := Parse([]byte(src))

	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	rs := ss.Children[0].(*Ruleset)
	decls := rs.Declarations()
	if len(decls) != 2 {
		t.Fatalf("expected 2 declarations, got %d",
			len(decls))
	}
	if decls[0].Property.Value != "color" {
		t.Errorf("expected 'color', got %q",
			decls[0].Property.Value)
	}
	if decls[1].Property.Value != "background" {
		t.Errorf("expected 'background', got %q",
			decls[1].Property.Value)
	}

	var nestedCount int
	for _, child := range rs.Children {
		if _, ok := child.(*Ruleset); ok {
			nestedCount++
		}
	}
	if nestedCount != 1 {
		t.Errorf("expected 1 nested ruleset, got %d",
			nestedCount)
	}
}

func TestParseFunctionValue(t *testing.T) {
	src := `div { background: rgb(255, 0, 0); }`
	ss, errs := Parse([]byte(src))

	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	rs := ss.Children[0].(*Ruleset)
	decl := rs.Declarations()[0]
	if decl.Property.Value != "background" {
		t.Errorf(
			"expected 'background', got %q",
			decl.Property.Value,
		)
	}
}

func TestUnclosedBlock(t *testing.T) {
	// EOF inside block — should recover gracefully
	src := `body { color: red;`
	ss, _ := Parse([]byte(src))

	if len(ss.Children) == 0 {
		t.Fatal("expected partial AST after unclosed block")
	}
}

func TestUnclosedAtRule(t *testing.T) {
	// EOF inside at-rule block
	src := `@media (max-width: 600px) { body { color: red; }`
	ss, _ := Parse([]byte(src))

	if len(ss.Children) == 0 {
		t.Fatal("expected partial AST after unclosed at-rule")
	}
	at, ok := ss.Children[0].(*AtRule)
	if !ok {
		t.Fatalf("expected AtRule, got %T", ss.Children[0])
	}
	if at.Name != "media" {
		t.Errorf("expected 'media', got %q", at.Name)
	}
}

func TestMissingPropertyName(t *testing.T) {
	// Declaration error recovery: missing property name
	src := `body { : red; color: blue; }`
	ss, _ := Parse([]byte(src))

	if len(ss.Children) == 0 {
		t.Fatal("expected partial AST")
	}
	_, ok := ss.Children[0].(*Ruleset)
	if !ok {
		t.Fatalf("expected Ruleset, got %T", ss.Children[0])
	}
}

func TestMissingColon(t *testing.T) {
	// Declaration error recovery: missing colon
	src := `body { color red; font-size: 16px; }`
	ss, _ := Parse([]byte(src))

	if len(ss.Children) == 0 {
		t.Fatal("expected partial AST")
	}
}

func TestNestedFunctionValues(t *testing.T) {
	src := `div { width: calc(100% - (50px + 10px)); }`
	ss, errs := Parse([]byte(src))

	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	rs := ss.Children[0].(*Ruleset)
	decls := rs.Declarations()
	if len(decls) != 1 {
		t.Fatalf("expected 1 declaration, got %d", len(decls))
	}
	if decls[0].Property.Value != "width" {
		t.Errorf(
			"expected 'width', got %q",
			decls[0].Property.Value,
		)
	}
}

func TestParseContainerRule(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{
			name: "name only, no query",
			src:  `@container card { #inner { background-color: skyblue; }}`,
		},
		{
			name: "size query",
			src:  `@container (width <= 150px) { #inner { background-color: skyblue; }}`,
		},
		{
			name: "name with size and style query",
			src:  `@container card (inline-size > 30em) and style(--responsive: true) { }`,
		},
		{
			name: "standalone custom property in style()",
			src:  `@container card style(--responsive) { }`,
		},
		{
			name: "comma-separated queries",
			src:  `@container (inline-size > 30em), style(--responsive: true) { }`,
		},
		{
			name: "comma-separated with name",
			src:  `@container card (inline-size > 30em), style(--responsive: true) { }`,
		},
		{
			name: "comma-separated with different names",
			src:  `@container card (inline-size > 30em), summary style(--responsive: true) { }`,
		},
		{
			name: "nested container",
			src:  `@container card (inline-size > 30em) { @container style(--responsive: true) {} }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss, errs := Parse([]byte(tt.src))
			if len(errs) != 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}
			if len(ss.Children) == 0 {
				t.Fatal("expected at least 1 child")
			}
			at, ok := ss.Children[0].(*AtRule)
			if !ok {
				t.Fatalf("expected AtRule, got %T", ss.Children[0])
			}
			if at.Name != "container" {
				t.Errorf("expected 'container', got %q", at.Name)
			}
			if at.Block == nil {
				t.Error("expected block in @container rule")
			}
		})
	}
}

func TestParseIfFunction(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{
			name: "media condition with else",
			src:  `div { color: if(media(print): black; else: white); }`,
		},
		{
			name: "empty branches",
			src:  `div { color: if(media(print): ; else: ); }`,
		},
		{
			name: "no trailing semicolon",
			src:  `div { color: if(media(print): black; else: white); }`,
		},
		{
			name: "style condition",
			src:  `div { color: if(style(--some-var: true): black); }`,
		},
		{
			name: "else only",
			src:  `div { color: if(else: white); }`,
		},
		{
			name: "nested in value",
			src:  `div { background: if(media(print): white; else: url("bg.png")); }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss, errs := Parse([]byte(tt.src))
			if len(errs) != 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}
			if len(ss.Children) == 0 {
				t.Fatal("expected at least 1 child")
			}
			rs, ok := ss.Children[0].(*Ruleset)
			if !ok {
				t.Fatalf("expected Ruleset, got %T", ss.Children[0])
			}
			decls := rs.Declarations()
			if len(decls) == 0 {
				t.Fatal("expected at least 1 declaration")
			}
			if decls[0].Value == nil || len(decls[0].Value.Tokens) == 0 {
				t.Fatal("expected value tokens in declaration")
			}
		})
	}
}

func TestParseScopeRule(t *testing.T) {
	tests := []struct {
		name        string
		src         string
		wantPrelude bool // true if prelude tokens expected
	}{
		{
			name:        "single scope root",
			src:         `@scope (.card) { .title { color: red; } }`,
			wantPrelude: true,
		},
		{
			name:        "scope root selector list",
			src:         `@scope (.card, .aside) { .title { color: red; } }`,
			wantPrelude: true,
		},
		{
			name:        "scope root and limit",
			src:         `@scope (.card) to (.header) { .title { color: red; } }`,
			wantPrelude: true,
		},
		{
			name:        "scope root and limit both selector lists",
			src:         `@scope (.card, .aside) to (.header, .footer) { .title { color: red; } }`,
			wantPrelude: true,
		},
		{
			name:        "scope without root (implicit)",
			src:         `@scope { .title { color: red; } }`,
			wantPrelude: false,
		},
		{
			name:        "scope limit only",
			src:         `@scope to (.footer) { .title { color: red; } }`,
			wantPrelude: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss, errs := Parse([]byte(tt.src))
			if len(errs) != 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}
			if len(ss.Children) == 0 {
				t.Fatal("expected at least 1 child")
			}
			at, ok := ss.Children[0].(*AtRule)
			if !ok {
				t.Fatalf("expected AtRule, got %T", ss.Children[0])
			}
			if at.Name != "scope" {
				t.Errorf("expected 'scope', got %q", at.Name)
			}
			if at.Block == nil {
				t.Error("expected block in @scope rule")
			}
			if tt.wantPrelude && len(at.Prelude) == 0 {
				t.Error("expected prelude tokens")
			}
			if !tt.wantPrelude && len(at.Prelude) != 0 {
				t.Errorf("expected no prelude tokens, got %d", len(at.Prelude))
			}
		})
	}
}

func TestCommentsBetweenDeclarations(t *testing.T) {
	src := `body {
  color: red;
  /* separator */
  font-size: 16px;
}`
	ss, errs := Parse([]byte(src))

	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	rs := ss.Children[0].(*Ruleset)
	decls := rs.Declarations()
	if len(decls) != 2 {
		t.Fatalf("expected 2 declarations, got %d", len(decls))
	}
}
