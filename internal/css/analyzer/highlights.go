package analyzer

import (
	"github.com/toba/go-css-lsp/internal/css/parser"
	"github.com/toba/go-css-lsp/internal/css/scanner"
)

// DocumentHighlight represents a highlighted range.
type DocumentHighlight struct {
	StartPos int
	EndPos   int
	Kind     int
}

// FindDocumentHighlights returns highlights for the symbol at
// the given offset. Supports CSS custom properties.
func FindDocumentHighlights(
	ss *parser.Stylesheet,
	src []byte,
	offset int,
) []DocumentHighlight {
	name := FindCustomPropertyAt(ss, src, offset)
	if name == "" {
		return nil
	}

	var highlights []DocumentHighlight

	// Highlight declarations
	parser.Walk(ss, func(n parser.Node) bool {
		decl, ok := n.(*parser.Declaration)
		if !ok {
			return true
		}

		if decl.Property.Value == name {
			highlights = append(highlights, DocumentHighlight{
				StartPos: decl.Property.Offset,
				EndPos:   decl.Property.End,
				Kind:     HighlightWrite,
			})
		}

		return true
	})

	// Highlight var() usages
	ForEachVarUsage(ss, name, func(tok scanner.Token) {
		highlights = append(highlights, DocumentHighlight{
			StartPos: tok.Offset,
			EndPos:   tok.End,
			Kind:     HighlightRead,
		})
	})

	return highlights
}
