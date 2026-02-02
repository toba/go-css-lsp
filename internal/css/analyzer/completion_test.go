package analyzer

import (
	"testing"

	"github.com/toba/go-css-lsp/internal/css/parser"
)

func TestCompletePropertyNames(t *testing.T) {
	src := []byte("body { col }")
	ss, _ := parser.Parse(src)

	// offset 10 is after "col" inside block
	items := Complete(ss, src, 10, LintOptions{})

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
	items := Complete(ss, src, 3, LintOptions{})

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

	items := Complete(ss, src, 0, LintOptions{})
	// Should return top-level completions without panic
	if items == nil {
		t.Error("expected non-nil completions")
	}
}

func TestCompleteNoSuggestionsInBlockComment(t *testing.T) {
	src := []byte("/* :ho */")
	ss, _ := parser.Parse(src)

	// offset 6 is after ":ho" inside the comment
	items := Complete(ss, src, 6, LintOptions{})
	if len(items) > 0 {
		t.Errorf(
			"expected no completions inside comment, got %d",
			len(items),
		)
	}
}

func TestCompleteNoSuggestionsInBlockComment_AtRule(
	t *testing.T,
) {
	src := []byte("/* @me */")
	ss, _ := parser.Parse(src)

	items := Complete(ss, src, 6, LintOptions{})
	if len(items) > 0 {
		t.Errorf(
			"expected no completions inside comment, got %d",
			len(items),
		)
	}
}

func TestCompleteNoSuggestionsInMultilineComment(
	t *testing.T,
) {
	src := []byte("body { }\n/* :hov\n */")
	ss, _ := parser.Parse(src)

	// offset 15 is after ":hov" inside the comment
	items := Complete(ss, src, 15, LintOptions{})
	if len(items) > 0 {
		t.Errorf(
			"expected no completions inside comment, got %d",
			len(items),
		)
	}
}

func TestCompleteAfterClosedComment(t *testing.T) {
	// After a closed comment, completions should work
	src := []byte("/* comment */ @me")
	ss, _ := parser.Parse(src)

	// offset 17 is after "@me"
	items := Complete(ss, src, 17, LintOptions{})
	found := false
	for _, item := range items {
		if item.Label == "@media" {
			found = true
			break
		}
	}
	if !found {
		t.Error(
			"expected '@media' after closed comment",
		)
	}
}

func TestCompleteNotInString(t *testing.T) {
	// Outside a comment, pseudo-class should still work
	src := []byte("a:ho")
	ss, _ := parser.Parse(src)

	items := Complete(ss, src, 4, LintOptions{})
	found := false
	for _, item := range items {
		if item.Label == ":hover" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ':hover' in pseudo-class completions")
	}
}
