package analyzer

import (
	"github.com/toba/go-css-lsp/internal/css/parser"
)

// FoldingRangeKind constants.
const (
	FoldingRangeComment = "comment"
	FoldingRangeRegion  = "region"
)

// FoldingRange represents a foldable range in the document.
type FoldingRange struct {
	StartPos int
	EndPos   int
	Kind     string
}

// FindFoldingRanges returns foldable ranges in the CSS
// document.
func FindFoldingRanges(
	ss *parser.Stylesheet,
	src []byte,
) []FoldingRange {
	if ss == nil {
		return nil
	}

	var ranges []FoldingRange

	parser.Walk(ss, func(n parser.Node) bool {
		switch node := n.(type) {
		case *parser.Ruleset:
			ranges = append(ranges, FoldingRange{
				StartPos: node.StartPos,
				EndPos:   node.EndPos,
				Kind:     FoldingRangeRegion,
			})
		case *parser.AtRule:
			if node.Block != nil {
				ranges = append(ranges, FoldingRange{
					StartPos: node.StartPos,
					EndPos:   node.EndPos,
					Kind:     FoldingRangeRegion,
				})
			}
		case *parser.Comment:
			// Only fold multi-line comments
			if hasNewline(src, node.StartPos, node.EndPos) {
				ranges = append(ranges, FoldingRange{
					StartPos: node.StartPos,
					EndPos:   node.EndPos,
					Kind:     FoldingRangeComment,
				})
			}
		}
		return true
	})

	return ranges
}

func hasNewline(src []byte, start, end int) bool {
	if end > len(src) {
		end = len(src)
	}
	for i := start; i < end; i++ {
		if src[i] == '\n' {
			return true
		}
	}
	return false
}
