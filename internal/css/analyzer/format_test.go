package analyzer

import (
	"testing"

	"github.com/toba/go-css-lsp/internal/css/parser"
)

func TestFormat_BasicRuleset(t *testing.T) {
	src := []byte(`.foo{color:red;background:blue;}`)
	ss, _ := parser.Parse(src)

	result := Format(ss, src, FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
	})

	expected := `.foo {
  color: red;
  background: blue;
}
`
	if result != expected {
		t.Errorf("format mismatch:\ngot:\n%s\nwant:\n%s",
			result, expected)
	}
}

func TestFormat_MediaQuery(t *testing.T) {
	src := []byte(`@media (max-width: 768px){.foo{color:red;}}`)
	ss, _ := parser.Parse(src)

	result := Format(ss, src, FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
	})

	expected := `@media (max-width: 768px) {
  .foo {
    color: red;
  }
}
`
	if result != expected {
		t.Errorf("format mismatch:\ngot:\n%s\nwant:\n%s",
			result, expected)
	}
}

func TestFormat_TabIndent(t *testing.T) {
	src := []byte(`.foo{color:red;}`)
	ss, _ := parser.Parse(src)

	result := Format(ss, src, FormatOptions{
		TabSize:      1,
		InsertSpaces: false,
	})

	expected := ".foo {\n\tcolor: red;\n}\n"
	if result != expected {
		t.Errorf("format mismatch:\ngot:\n%q\nwant:\n%q",
			result, expected)
	}
}

func TestFormat_Important(t *testing.T) {
	src := []byte(`.foo{color:red !important;}`)
	ss, _ := parser.Parse(src)

	result := Format(ss, src, FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
	})

	expected := `.foo {
  color: red !important;
}
`
	if result != expected {
		t.Errorf("format mismatch:\ngot:\n%s\nwant:\n%s",
			result, expected)
	}
}

func TestFormat_Nil(t *testing.T) {
	src := []byte(`some text`)
	result := Format(nil, src, FormatOptions{})
	if result != "some text" {
		t.Errorf("expected original text for nil AST")
	}
}

// --- Compact mode tests ---

func compactOpts(printWidth int) FormatOptions {
	return FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
		Mode:         FormatCompact,
		PrintWidth:   printWidth,
	}
}

func TestFormatCompact_SingleDeclFits(t *testing.T) {
	src := []byte(`.foo { color: red; }`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, compactOpts(80))
	expected := ".foo { color: red; }\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatCompact_MultipleDeclsFit(t *testing.T) {
	src := []byte(`.foo{color:red;background:blue;}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, compactOpts(80))
	expected := ".foo { color: red; background: blue; }\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatCompact_ExceedsPrintWidth(t *testing.T) {
	src := []byte(`.foo{color:red;background:blue;}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, compactOpts(20))
	expected := ".foo {\n  color: red;\n  background: blue;\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatCompact_MultipleSelectorsFit(t *testing.T) {
	src := []byte(`ul, ol { margin: 0; }`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, compactOpts(80))
	expected := "ul, ol { margin: 0; }\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatCompact_MultipleSelectorsDontFit(t *testing.T) {
	src := []byte(`ul, ol { margin: 0; }`)
	ss, _ := parser.Parse(src)
	// Width too narrow for single line — falls back to expanded
	result := Format(ss, src, compactOpts(15))
	expected := "ul,\nol {\n  margin: 0;\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatCompact_EmptyRuleset(t *testing.T) {
	src := []byte(`.foo{}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, compactOpts(80))
	expected := ".foo {}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatCompact_CustomPrintWidth(t *testing.T) {
	src := []byte(`.foo{color:red;}`)
	ss, _ := parser.Parse(src)
	// ".foo { color: red; }" is 20 chars
	result := Format(ss, src, compactOpts(20))
	expected := ".foo { color: red; }\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
	// At 19, should expand
	result = Format(ss, src, compactOpts(19))
	expected = ".foo {\n  color: red;\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatCompact_NestedInMedia(t *testing.T) {
	src := []byte(`@media screen{.foo{color:red;}}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, compactOpts(80))
	expected := "@media screen {\n  .foo { color: red; }\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatCompact_NestedExceedsWidthBudget(t *testing.T) {
	src := []byte(`@media screen{.foo{color:red;}}`)
	ss, _ := parser.Parse(src)
	// "  .foo { color: red; }" = 22 chars with indent
	result := Format(ss, src, compactOpts(21))
	expected := "@media screen {\n  .foo {\n    color: red;\n  }\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatCompact_RulesetWithComment(t *testing.T) {
	src := []byte(`.foo{/* comment */color:red;}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, compactOpts(80))
	// Should always fall back to expanded when comment present
	expected := ".foo {\n  /* comment */\n  color: red;\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

// --- Preserve mode tests ---

func preserveOpts() FormatOptions {
	return FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
		Mode:         FormatPreserve,
	}
}

func TestFormatPreserve_OriginalSingleLine(t *testing.T) {
	src := []byte(`.foo{color:red;background:blue;}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, preserveOpts())
	expected := ".foo { color: red; background: blue; }\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatPreserve_OriginalMultiLine(t *testing.T) {
	src := []byte(".foo {\n  color: red;\n  background: blue;\n}")
	ss, _ := parser.Parse(src)
	result := Format(ss, src, preserveOpts())
	expected := ".foo {\n  color: red;\n  background: blue;\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatPreserve_NormalizeBadSpacing(t *testing.T) {
	src := []byte(`.foo   {   color:   red   ;   }`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, preserveOpts())
	// Single line input stays single line with normalized spacing
	expected := ".foo { color: red; }\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatPreserve_NormalizeBadSpacingMultiLine(t *testing.T) {
	src := []byte(".foo   {\n   color:   red   ;\n   background:   blue   ;\n}")
	ss, _ := parser.Parse(src)
	result := Format(ss, src, preserveOpts())
	expected := ".foo {\n  color: red;\n  background: blue;\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatPreserve_MixedRules(t *testing.T) {
	src := []byte(".foo { color: red; }\n.bar {\n  background: blue;\n}")
	ss, _ := parser.Parse(src)
	result := Format(ss, src, preserveOpts())
	expected := ".foo { color: red; }\n\n.bar {\n  background: blue;\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

// --- Nesting tests ---

func TestFormat_NestedRuleset(t *testing.T) {
	src := []byte(`.parent{color:red;&:hover{color:blue;}}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
	})

	expected := `.parent {
  color: red;

  &:hover {
    color: blue;
  }
}
`
	if result != expected {
		t.Errorf("format mismatch:\ngot:\n%s\nwant:\n%s",
			result, expected)
	}
}

func TestFormat_NestedClassSelector(t *testing.T) {
	src := []byte(`.parent{color:red;.child{font-size:14px;}}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
	})

	expected := `.parent {
  color: red;

  .child {
    font-size: 14px;
  }
}
`
	if result != expected {
		t.Errorf("format mismatch:\ngot:\n%s\nwant:\n%s",
			result, expected)
	}
}

func TestFormat_NestedAtRule(t *testing.T) {
	src := []byte(`.parent{color:red;@media (hover){color:blue;}}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
	})

	expected := `.parent {
  color: red;

  @media (hover) {
    color: blue;
  }
}
`
	if result != expected {
		t.Errorf("format mismatch:\ngot:\n%s\nwant:\n%s",
			result, expected)
	}
}

func TestFormat_MultiLevelNesting(t *testing.T) {
	src := []byte(`.a{.b{.c{color:red;}}}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
	})

	expected := `.a {
  .b {
    .c {
      color: red;
    }
  }
}
`
	if result != expected {
		t.Errorf("format mismatch:\ngot:\n%s\nwant:\n%s",
			result, expected)
	}
}

func TestFormatCompact_NestedFallsBackToExpanded(t *testing.T) {
	src := []byte(`.parent{color:red;&:hover{color:blue;}}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, compactOpts(80))
	// Parent falls back to expanded due to nested rules;
	// nested leaf rulesets may still be compacted.
	expected := `.parent {
  color: red;

  &:hover { color: blue; }
}
`
	if result != expected {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}

// --- Value wrapping tests ---

func TestFormat_LongValueWrapsAtTopLevelCommas(t *testing.T) {
	src := []byte(`.glow {
  background:
    linear-gradient(var(--color-bg), var(--color-bg)) padding-box,
    conic-gradient(from var(--glow-angle), transparent 40%, var(--brand-teal), transparent 60%) border-box;
}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
		PrintWidth:   120,
	})

	expected := `.glow {
  background:
    linear-gradient(var(--color-bg), var(--color-bg)) padding-box,
    conic-gradient(from var(--glow-angle), transparent 40%, var(--brand-teal), transparent 60%) border-box;
}
`
	if result != expected {
		t.Errorf("format mismatch:\ngot:\n%s\nwant:\n%s",
			result, expected)
	}
}

func TestFormat_ShortValueStaysSingleLine(t *testing.T) {
	src := []byte(`.foo { background: red, blue; }`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
		PrintWidth:   80,
	})

	expected := `.foo {
  background: red, blue;
}
`
	if result != expected {
		t.Errorf("format mismatch:\ngot:\n%s\nwant:\n%s",
			result, expected)
	}
}

func TestFormat_LongValueNoTopLevelCommasStaysSingleLine(t *testing.T) {
	// Value is long but all commas are inside function parens
	src := []byte(`.foo { color: rgb(100, 200, 300); }`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
		PrintWidth:   20,
	})

	expected := `.foo {
  color: rgb(100, 200, 300);
}
`
	if result != expected {
		t.Errorf("format mismatch:\ngot:\n%s\nwant:\n%s",
			result, expected)
	}
}

func TestFormat_LongValueWrapsWithImportant(t *testing.T) {
	src := []byte(
		`.foo { background: linear-gradient(red, blue) padding-box, conic-gradient(green, yellow) border-box !important; }`,
	)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
		PrintWidth:   40,
	})

	expected := `.foo {
  background:
    linear-gradient(red, blue) padding-box,
    conic-gradient(green, yellow) border-box !important;
}
`
	if result != expected {
		t.Errorf("format mismatch:\ngot:\n%s\nwant:\n%s",
			result, expected)
	}
}

func TestFormatCompact_LongValueWraps(t *testing.T) {
	src := []byte(
		`.foo { background: linear-gradient(red, blue) padding-box, conic-gradient(green, yellow) border-box; }`,
	)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, compactOpts(40))

	expected := `.foo {
  background:
    linear-gradient(red, blue) padding-box,
    conic-gradient(green, yellow) border-box;
}
`
	if result != expected {
		t.Errorf("format mismatch:\ngot:\n%s\nwant:\n%s",
			result, expected)
	}
}

func TestFormat_NoSpaceAfterFunctionParen(t *testing.T) {
	// Source has whitespace after ( — formatter must not preserve it
	src := []byte(`.foo { background: conic-gradient( from var( --angle), red, blue); }`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
	})

	expected := `.foo {
  background: conic-gradient(from var(--angle), red, blue);
}
`
	if result != expected {
		t.Errorf("format mismatch:\ngot:\n%s\nwant:\n%s",
			result, expected)
	}
}

func TestFormatCompact_NoSpaceAfterFunctionParen(t *testing.T) {
	// Source has whitespace after ( — formatter must not preserve it
	src := []byte(`.foo { color: rgb( 255, 0, 0); }`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, compactOpts(80))
	expected := ".foo { color: rgb(255, 0, 0); }\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatPreserve_NestedFallsBackToExpanded(t *testing.T) {
	src := []byte(`.parent { color: red; .child { font-size: 14px; } }`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, preserveOpts())
	// Parent falls back to expanded due to nested rules;
	// nested leaf rulesets keep original single-line layout.
	expected := `.parent {
  color: red;

  .child { font-size: 14px; }
}
`
	if result != expected {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}
