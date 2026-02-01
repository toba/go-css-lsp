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

func TestHoverNoPanic(t *testing.T) {
	src := []byte(`body { color: red; }`)
	ss, _ := parser.Parse(src)

	// Just ensure no panic at various offsets
	for i := range src {
		Hover(ss, src, i)
	}
}
