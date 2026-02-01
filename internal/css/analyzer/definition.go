package analyzer

import (
	"github.com/toba/go-css-lsp/internal/css/parser"
)

// Location represents a position range in a document.
type Location struct {
	StartPos int
	EndPos   int
}

// FindDefinition finds the definition of the symbol at the
// given offset. Currently supports CSS custom properties
// (var(--name) -> --name declaration).
func FindDefinition(
	ss *parser.Stylesheet,
	src []byte,
	offset int,
) (Location, bool) {
	// Find what's at the cursor
	varName := FindVarReferenceAt(ss, src, offset)
	if varName == "" {
		return Location{}, false
	}

	// Search for the custom property declaration
	loc, found := findCustomPropertyDecl(ss, varName)
	return loc, found
}

// findCustomPropertyDecl finds the declaration of a custom
// property by name.
func findCustomPropertyDecl(
	ss *parser.Stylesheet,
	name string,
) (Location, bool) {
	var loc Location
	var found bool

	parser.Walk(ss, func(n parser.Node) bool {
		if found {
			return false
		}

		decl, ok := n.(*parser.Declaration)
		if !ok {
			return true
		}

		if decl.Property.Value == name {
			loc = Location{
				StartPos: decl.Property.Offset,
				EndPos:   decl.Property.End,
			}
			found = true
			return false
		}
		return true
	})

	return loc, found
}
