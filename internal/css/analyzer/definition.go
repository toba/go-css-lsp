package analyzer

import (
	"github.com/toba/go-css-lsp/internal/css/parser"
)

// Location represents a position range in a document.
type Location struct {
	StartPos int
	EndPos   int
}

// DefinitionResult holds both the origin range (at the usage
// site, e.g. var(--name)) and the target range (at the
// declaration, e.g. --name).
type DefinitionResult struct {
	OriginStart int
	OriginEnd   int
	TargetStart int
	TargetEnd   int
}

// FindDefinition finds the definition of the symbol at the
// given offset. Currently supports CSS custom properties
// (var(--name) -> --name declaration).
func FindDefinition(
	ss *parser.Stylesheet,
	src []byte,
	offset int,
) (DefinitionResult, bool) {
	// Find what's at the cursor and the var() span
	varName, originStart, originEnd := FindVarReferenceWithRange(
		ss, src, offset,
	)
	if varName == "" {
		return DefinitionResult{}, false
	}

	// Search for the custom property declaration
	loc, found := findCustomPropertyDecl(ss, varName)
	if !found {
		return DefinitionResult{}, false
	}

	return DefinitionResult{
		OriginStart: originStart,
		OriginEnd:   originEnd,
		TargetStart: loc.StartPos,
		TargetEnd:   loc.EndPos,
	}, true
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
