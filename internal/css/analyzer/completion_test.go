package analyzer

import (
	"testing"

	"github.com/toba/go-css-lsp/internal/css/parser"
)

func TestCompletePropertyNames(t *testing.T) {
	src := []byte("body { col }")
	ss, _ := parser.Parse(src)

	// offset 10 is after "col" inside block
	items := Complete(ss, src, 10)

	found := false
	for _, item := range items {
		if item.Label == "color" {
			found = true
			break
		}
	}
	if !found {
		t.Error(
			"expected 'color' in property completions",
		)
	}
}

func TestCompleteAtRules(t *testing.T) {
	src := []byte("@me")
	ss, _ := parser.Parse(src)

	// offset 3 is after "@me"
	items := Complete(ss, src, 3)

	found := false
	for _, item := range items {
		if item.Label == "@media" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected '@media' in at-rule completions")
	}
}

func TestCompleteEmpty(t *testing.T) {
	src := []byte("")
	ss, _ := parser.Parse(src)

	items := Complete(ss, src, 0)
	// Should return top-level completions without panic
	if items == nil {
		t.Error("expected non-nil completions")
	}
}
