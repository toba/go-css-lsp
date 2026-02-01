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
	expected := ".foo {\n  color: red;\n}\n"
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
