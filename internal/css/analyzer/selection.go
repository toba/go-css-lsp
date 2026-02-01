package analyzer

import (
	"github.com/toba/go-css-lsp/internal/css/parser"
)

// SelectionRange represents a nested selection range.
type SelectionRange struct {
	StartPos int
	EndPos   int
	Parent   *SelectionRange
}

// FindSelectionRange returns the selection range at the given
// offset. Expands: value -> declaration -> ruleset ->
// stylesheet.
func FindSelectionRange(
	ss *parser.Stylesheet,
	offset int,
) *SelectionRange {
	if ss == nil {
		return nil
	}

	// Build the range chain by walking the AST
	var chain []*SelectionRange

	parser.Walk(ss, func(n parser.Node) bool {
		if n.Offset() > offset || n.End() < offset {
			return false
		}

		// Skip selector nodes since they're part of rulesets
		switch n.(type) {
		case *parser.SelectorList, *parser.Selector:
			return true
		}

		chain = append(chain, &SelectionRange{
			StartPos: n.Offset(),
			EndPos:   n.End(),
		})
		return true
	})

	// Link the chain (innermost to outermost)
	for i := len(chain) - 1; i > 0; i-- {
		chain[i].Parent = chain[i-1]
	}

	if len(chain) == 0 {
		return nil
	}
	return chain[len(chain)-1]
}
