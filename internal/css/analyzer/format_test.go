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
	// Source has only 1 newline between rules → no blank line.
	src := []byte(".foo { color: red; }\n.bar {\n  background: blue;\n}")
	ss, _ := parser.Parse(src)
	result := Format(ss, src, preserveOpts())
	expected := ".foo { color: red; }\n.bar {\n  background: blue;\n}\n"
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
	// compact mode: no blank line before nested rule.
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

func TestFormat_ShortFunctionArgsStaySingleLine(t *testing.T) {
	// Value has commas only inside function parens, but the line
	// fits within printWidth so it stays single-line.
	src := []byte(`.foo { color: rgb(100, 200, 300); }`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
		PrintWidth:   80,
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

func TestFormat_LongValueWrapsAtShallowestCommas(t *testing.T) {
	// All commas are inside light-dark() at depth 1, but the line
	// exceeds printWidth so it should still wrap.
	src := []byte(`:root {
  --color-bg-top: light-dark(hsl(from var(--brand-deep-pockets) h calc(s - 40) calc(l + 40)), hsl(from var(--brand-deep-pockets) h calc(s - 40) l));
}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
		PrintWidth:   80,
	})

	expected := `:root {
  --color-bg-top: light-dark(
    hsl(from var(--brand-deep-pockets) h calc(s - 40) calc(l + 40)),
    hsl(from var(--brand-deep-pockets) h calc(s - 40) l));
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
	// source is single-line: no blank line in gap.
	expected := `.parent {
  color: red;
  .child { font-size: 14px; }
}
`
	if result != expected {
		t.Errorf("got:\n%s\nwant:\n%s", result, expected)
	}
}

// --- Detect mode tests ---

func detectOpts(printWidth int) FormatOptions {
	return FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
		Mode:         FormatDetect,
		PrintWidth:   printWidth,
	}
}

func TestFormatDetect_InlineFirstPropFits(t *testing.T) {
	// First prop is inline with { and fits → single line
	src := []byte(`.foo { color: red;
  background: blue;
}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, detectOpts(80))
	expected := ".foo { color: red; background: blue; }\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatDetect_FirstPropOnNewLine(t *testing.T) {
	// First prop on new line → stays multi-line
	src := []byte(`.foo {
  color: red;
  background: blue;
}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, detectOpts(80))
	expected := ".foo {\n  color: red;\n  background: blue;\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatDetect_InlineExceedsPrintWidth(t *testing.T) {
	// First prop inline but result exceeds print-width → multi-line
	src := []byte(`.foo { color: red; background: blue; }`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, detectOpts(20))
	expected := ".foo {\n  color: red;\n  background: blue;\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatDetect_MultiSelectorInlineFits(t *testing.T) {
	// Multi-selector with first prop inline → single line
	src := []byte(`ul, ol { margin: 0; }`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, detectOpts(80))
	expected := "ul, ol { margin: 0; }\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatDetect_MultiSelectorNewLine(t *testing.T) {
	// Multi-selector on same line, first prop on new line →
	// selectors stay inline, body expanded
	src := []byte("ul, ol {\n  margin: 0;\n}")
	ss, _ := parser.Parse(src)
	result := Format(ss, src, detectOpts(80))
	expected := "ul, ol {\n  margin: 0;\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatDetect_MultiSelectorInlineExceedsPrintWidth(t *testing.T) {
	// Multi-selector inline, exceeds single-line print-width →
	// selectors stay inline (they fit), body expands
	src := []byte(`ul, ol { margin: 0; }`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, detectOpts(15))
	expected := "ul, ol {\n  margin: 0;\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatDetect_SelectorListDetectsInline(t *testing.T) {
	// Second selector on same line → all selectors inline
	src := []byte("button, input,\nselect,\ntextarea {\n  font: inherit;\n}")
	ss, _ := parser.Parse(src)
	result := Format(ss, src, detectOpts(80))
	expected := "button, input, select, textarea {\n  font: inherit;\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatDetect_SelectorListDetectsNewLine(t *testing.T) {
	// Second selector on new line → all selectors on new lines
	src := []byte("button,\ninput,\nselect,\ntextarea {\n  font: inherit;\n}")
	ss, _ := parser.Parse(src)
	result := Format(ss, src, detectOpts(80))
	expected := "button,\ninput,\nselect,\ntextarea {\n  font: inherit;\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatDetect_SelectorListInlineExceedsWidth(t *testing.T) {
	// Selectors inline in source but don't fit → multi-line
	src := []byte("button, input, select, textarea {\n  font: inherit;\n}")
	ss, _ := parser.Parse(src)
	result := Format(ss, src, detectOpts(20))
	expected := "button,\ninput,\nselect,\ntextarea {\n  font: inherit;\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatDetect_LeadingCombinatorInlineSelectorList(t *testing.T) {
	// Selectors starting with > should not produce extra space
	src := []byte(`.parent {
  > header h1, > h1 {
    margin: 0;
  }
}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, detectOpts(80))
	expected := ".parent {\n  > header h1, > h1 {\n    margin: 0;\n  }\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatCompact_LeadingCombinatorSingleLine(t *testing.T) {
	src := []byte(`.parent {
  > h1 { margin: 0; }
}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, compactOpts(80))
	expected := ".parent {\n  > h1 { margin: 0; }\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

// --- At-rule single-line tests ---

func TestFormatDetect_AtRuleInlineFirstPropFits(t *testing.T) {
	// Nested @media with first prop inline → single line
	src := []byte(`.foo {
  @media (width <= 767px) { margin-left: 0; }
}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, detectOpts(80))
	expected := ".foo {\n  @media (width <= 767px) { margin-left: 0; }\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatDetect_AtRuleFirstPropOnNewLine(t *testing.T) {
	// Nested @media with first prop on new line → stays multi-line
	src := []byte(`.foo {
  @media (width <= 767px) {
    margin-left: 0;
  }
}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, detectOpts(80))
	expected := ".foo {\n  @media (width <= 767px) {\n    margin-left: 0;\n  }\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatDetect_AtRuleInlineExceedsPrintWidth(t *testing.T) {
	// Nested @media inline but exceeds print-width → multi-line
	src := []byte(`.foo {
  @media (width <= 767px) { margin-left: 0; }
}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, detectOpts(30))
	expected := ".foo {\n  @media (width <= 767px) {\n    margin-left: 0;\n  }\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatCompact_AtRuleSingleLine(t *testing.T) {
	// Compact mode: nested at-rule with only declarations → single line
	src := []byte(`.foo {
  @media (width <= 767px) {
    margin-left: 0;
  }
}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, compactOpts(80))
	// Parent has nested rules so it stays expanded; @media collapses
	expected := ".foo {\n  @media (width <= 767px) { margin-left: 0; }\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatPreserve_AtRuleSingleLine(t *testing.T) {
	// Preserve mode: at-rule originally single-line → stays single line
	src := []byte(`.foo {
  @media (width <= 767px) { margin-left: 0; }
}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
		Mode:         FormatPreserve,
		PrintWidth:   80,
	})
	expected := ".foo {\n  @media (width <= 767px) { margin-left: 0; }\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatPreserve_AtRuleMultiLine(t *testing.T) {
	// Preserve mode: at-rule originally multi-line → stays multi-line
	src := []byte(`.foo {
  @media (width <= 767px) {
    margin-left: 0;
  }
}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
		Mode:         FormatPreserve,
		PrintWidth:   80,
	})
	expected := ".foo {\n  @media (width <= 767px) {\n    margin-left: 0;\n  }\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

// --- Blank line handling tests ---

func TestFormatCompact_NoBlankLinesBetweenTopLevelRules(t *testing.T) {
	src := []byte(".foo { color: red; }\n\n.bar { color: blue; }")
	ss, _ := parser.Parse(src)
	result := Format(ss, src, compactOpts(80))
	expected := ".foo { color: red; }\n.bar { color: blue; }\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatPreserve_NoBlankLinesInSource(t *testing.T) {
	// 0 blank lines in source → 0 in output
	src := []byte(".foo { color: red; }\n.bar { color: blue; }")
	ss, _ := parser.Parse(src)
	result := Format(ss, src, preserveOpts())
	expected := ".foo { color: red; }\n.bar { color: blue; }\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatPreserve_OneBlankLineInSource(t *testing.T) {
	// 1 blank line in source → 1 in output
	src := []byte(".foo { color: red; }\n\n.bar { color: blue; }")
	ss, _ := parser.Parse(src)
	result := Format(ss, src, preserveOpts())
	expected := ".foo { color: red; }\n\n.bar { color: blue; }\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatPreserve_ManyBlankLinesCollapsed(t *testing.T) {
	// 3+ blank lines in source → collapsed to 1
	src := []byte(".foo { color: red; }\n\n\n\n.bar { color: blue; }")
	ss, _ := parser.Parse(src)
	result := Format(ss, src, preserveOpts())
	expected := ".foo { color: red; }\n\n.bar { color: blue; }\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatCompact_NoBlankLineBeforeNestedRules(t *testing.T) {
	src := []byte(".parent {\n  color: red;\n\n  &:hover {\n    color: blue;\n  }\n}")
	ss, _ := parser.Parse(src)
	result := Format(ss, src, compactOpts(80))
	expected := ".parent {\n  color: red;\n  &:hover { color: blue; }\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormatPreserve_NestedNoBlankLinesInSource(t *testing.T) {
	src := []byte(".parent {\n  color: red;\n  &:hover {\n    color: blue;\n  }\n}")
	ss, _ := parser.Parse(src)
	result := Format(ss, src, preserveOpts())
	expected := ".parent {\n  color: red;\n  &:hover {\n    color: blue;\n  }\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormat_ExpandedStillHasBlankLines(t *testing.T) {
	// Expanded mode: blank line before nested rule (regression check)
	src := []byte(`.parent{color:red;&:hover{color:blue;}}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
	})
	expected := ".parent {\n  color: red;\n\n  &:hover {\n    color: blue;\n  }\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}

func TestFormat_TopLevelCommasBreakAfterColon(t *testing.T) {
	// Top-level commas (depth 0) should break after colon, not inline prefix
	src := []byte(`.glow {
  background: linear-gradient(var(--a), var(--b)) padding-box, conic-gradient(from var(--angle), transparent 40%, var(--teal), transparent 60%) border-box;
}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
		PrintWidth:   80,
	})

	expected := `.glow {
  background:
    linear-gradient(var(--a), var(--b)) padding-box,
    conic-gradient(from var(--angle), transparent 40%, var(--teal), transparent 60%) border-box;
}
`
	if result != expected {
		t.Errorf("format mismatch:\ngot:\n%s\nwant:\n%s",
			result, expected)
	}
}

func TestFormat_NestedFunctionCommasInlinePrefix(t *testing.T) {
	// Commas at depth 2 (inside rgb() inside light-dark()) — prefix up to
	// the depth-2 function opening should stay inline
	src := []byte(`:root {
  --shadow: light-dark(0 1px 3px rgb(0, 0, 0, 0.12), 0 1px 3px rgb(255, 255, 255, 0.12));
}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
		PrintWidth:   60,
	})

	expected := `:root {
  --shadow: light-dark(
    0 1px 3px rgb(0, 0, 0, 0.12),
    0 1px 3px rgb(255, 255, 255, 0.12));
}
`
	if result != expected {
		t.Errorf("format mismatch:\ngot:\n%s\nwant:\n%s",
			result, expected)
	}
}

func TestFormatPreserve_SingleLineExceedsPrintWidth(t *testing.T) {
	// Single-line source that exceeds print-width → expands
	src := []byte(`.foo{color:red;background:blue;}`)
	ss, _ := parser.Parse(src)
	result := Format(ss, src, FormatOptions{
		TabSize:      2,
		InsertSpaces: true,
		Mode:         FormatPreserve,
		PrintWidth:   20,
	})
	expected := ".foo {\n  color: red;\n  background: blue;\n}\n"
	if result != expected {
		t.Errorf("got:\n%q\nwant:\n%q", result, expected)
	}
}
