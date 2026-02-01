package analyzer

import (
	"strings"
	"testing"

	"github.com/toba/go-css-lsp/internal/css/parser"
)

func TestHoverProperty(t *testing.T) {
	src := []byte(`body { color: red; }`)
	ss, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	// "color" is at bytes 7-12
	content, found := Hover(ss, src, 8)
	if !found {
		t.Fatal("expected hover to find content")
	}
	if !strings.Contains(content, "**color**") {
		t.Errorf(
			"expected property name, got %q", content,
		)
	}
}

func TestHoverUnknownProperty(t *testing.T) {
	src := []byte(`body { foobar: red; }`)
	ss, _ := parser.Parse(src)

	_, found := Hover(ss, src, 8)
	if found {
		t.Error(
			"expected hover not to find unknown property",
		)
	}
}

func TestHoverFunctionSignatures(t *testing.T) {
	src := []byte(`body { color: rgb(255, 0, 0); }`)
	ss, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	// "rgb" function token starts at byte 14
	content, found := Hover(ss, src, 14)
	if !found {
		t.Fatal("expected hover to find rgb function")
	}
	if !strings.Contains(content, "rgb(") {
		t.Errorf(
			"expected signature lines, got %q", content,
		)
	}
	if !strings.Contains(content, "```") {
		t.Errorf(
			"expected code block, got %q", content,
		)
	}
	if !strings.Contains(content, "MDN Reference") {
		t.Errorf(
			"expected MDN link, got %q", content,
		)
	}
	if !strings.Contains(content, "red, green, blue") {
		t.Errorf(
			"expected description, got %q", content,
		)
	}
}

func TestHoverFunctionCalc(t *testing.T) {
	src := []byte(`body { width: calc(100% - 20px); }`)
	ss, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	// "calc" function token starts at byte 14
	content, found := Hover(ss, src, 14)
	if !found {
		t.Fatal("expected hover to find calc function")
	}
	if !strings.Contains(content, "calc(<expression>)") {
		t.Errorf(
			"expected calc signature, got %q", content,
		)
	}
	if !strings.Contains(content, "MDN Reference") {
		t.Errorf(
			"expected MDN link, got %q", content,
		)
	}
}

func TestHoverFunctionUnknown(t *testing.T) {
	src := []byte(
		`body { color: notafunction(1, 2); }`,
	)
	ss, _ := parser.Parse(src)

	_, found := Hover(ss, src, 14)
	if found {
		t.Error(
			"expected hover not to find unknown function",
		)
	}
}

func TestHoverNoPanic(t *testing.T) {
	src := []byte(`body { color: red; }`)
	ss, _ := parser.Parse(src)

	// Just ensure no panic at various offsets
	for i := range src {
		Hover(ss, src, i)
	}
}
