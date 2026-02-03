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

func TestCompleteValueContext(t *testing.T) {
	src := []byte("body { justify-content: ; }")
	ss, _ := parser.Parse(src)

	// offset 23 is after "justify-content: " (right before ;)
	items := Complete(ss, src, 23, LintOptions{})

	// Should include value keywords like center, space-between
	hasCenter := false
	hasPropertyKind := false
	for _, item := range items {
		if item.Label == "center" {
			hasCenter = true
		}
		if item.Kind == KindProperty {
			hasPropertyKind = true
		}
	}
	if !hasCenter {
		t.Error(
			"expected 'center' in value completions for " +
				"justify-content",
		)
	}
	if hasPropertyKind {
		t.Error(
			"property completions should not appear " +
				"in value context",
		)
	}
}

func TestCompleteValueContext_WhiteSpace(t *testing.T) {
	src := []byte("body { white-space: ; }")
	ss, _ := parser.Parse(src)

	// offset 20 is right before the ; (after "white-space: ")
	items := Complete(ss, src, 19, LintOptions{})

	expected := map[string]bool{
		"normal":       false,
		"nowrap":       false,
		"pre":          false,
		"pre-wrap":     false,
		"pre-line":     false,
		"break-spaces": false,
	}
	for _, item := range items {
		if _, ok := expected[item.Label]; ok {
			expected[item.Label] = true
		}
	}
	for name, found := range expected {
		if !found {
			t.Errorf("expected %q in value completions for white-space", name)
		}
	}
}

func TestCompleteDeprecatedPropertyTagged(t *testing.T) {
	src := []byte("body { cli }")
	ss, _ := parser.Parse(src)

	items := Complete(ss, src, 11, LintOptions{
		Deprecated: DeprecatedWarn,
	})

	for _, item := range items {
		if item.Label == "clip" {
			if !item.Deprecated {
				t.Error("expected Deprecated=true for 'clip'")
			}
			if len(item.Detail) < 12 ||
				item.Detail[:13] != "(deprecated) " {
				t.Errorf(
					"expected '(deprecated) ' prefix, got %q",
					item.Detail,
				)
			}
			return
		}
	}
	t.Error("expected 'clip' in property completions")
}

func TestCompleteDeprecatedPropertyNotTaggedWhenIgnored(
	t *testing.T,
) {
	src := []byte("body { cli }")
	ss, _ := parser.Parse(src)

	items := Complete(ss, src, 11, LintOptions{
		Deprecated: DeprecatedIgnore,
	})

	for _, item := range items {
		if item.Label == "clip" {
			if item.Deprecated {
				t.Error("Deprecated should be false when ignored")
			}
			return
		}
	}
	t.Error("expected 'clip' in property completions")
}
