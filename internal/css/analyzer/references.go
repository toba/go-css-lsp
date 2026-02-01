package analyzer

import (
	"github.com/toba/go-css-lsp/internal/css/parser"
	"github.com/toba/go-css-lsp/internal/css/scanner"
)

// FindReferences finds all references to the symbol at the
// given offset. Supports CSS custom properties.
func FindReferences(
	ss *parser.Stylesheet,
	src []byte,
	offset int,
) []Location {
	// Determine the custom property name at cursor
	name := FindCustomPropertyAt(ss, src, offset)
	if name == "" {
		return nil
	}

	// Find all occurrences: declarations and var() usages
	var refs []Location

	parser.Walk(ss, func(n parser.Node) bool {
		decl, ok := n.(*parser.Declaration)
		if !ok {
			return true
		}

		// Check property declaration
		if decl.Property.Value == name {
			refs = append(refs, Location{
				StartPos: decl.Property.Offset,
				EndPos:   decl.Property.End,
			})
		}

		return true
	})

	// Check var() usages
	ForEachVarUsage(ss, name, func(tok scanner.Token) {
		refs = append(refs, Location{
			StartPos: tok.Offset,
			EndPos:   tok.End,
		})
	})

	return refs
}
