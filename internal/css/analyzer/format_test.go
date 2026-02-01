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
