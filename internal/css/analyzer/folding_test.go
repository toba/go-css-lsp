package analyzer

import (
	"testing"

	"github.com/toba/go-css-lsp/internal/css/parser"
)

func TestFindFoldingRanges_Rulesets(t *testing.T) {
	src := []byte(`.foo {
  color: red;
  background: blue;
}

.bar {
  display: block;
}`)
	ss, _ := parser.Parse(src)
	ranges := FindFoldingRanges(ss, src)

	if len(ranges) != 2 {
		t.Fatalf("expected 2 folding ranges, got %d", len(ranges))
	}

	if ranges[0].Kind != FoldingRangeRegion {
		t.Errorf("expected region kind")
	}
}

func TestFindFoldingRanges_AtRules(t *testing.T) {
	src := []byte(`@media (max-width: 768px) {
  .foo {
    color: red;
  }
}`)
	ss, _ := parser.Parse(src)
	ranges := FindFoldingRanges(ss, src)

	// @media block + .foo ruleset
	if len(ranges) != 2 {
		t.Fatalf("expected 2 folding ranges, got %d", len(ranges))
	}
}

func TestFindFoldingRanges_MultilineComment(t *testing.T) {
	src := []byte(`/*
 * This is a
 * multi-line comment
 */
.foo { color: red; }`)
	ss, _ := parser.Parse(src)
	ranges := FindFoldingRanges(ss, src)

	// comment + ruleset (single-line rulesets still fold)
	found := false
	for _, r := range ranges {
		if r.Kind == FoldingRangeComment {
			found = true
		}
	}
	if !found {
		t.Error("expected a comment folding range")
	}
}

func TestFindFoldingRanges_Empty(t *testing.T) {
	ranges := FindFoldingRanges(nil, nil)
	if ranges != nil {
		t.Fatalf("expected nil for nil stylesheet")
	}
}
