package analyzer

import (
	"strings"

	"github.com/toba/go-css-lsp/internal/css/parser"
	"github.com/toba/go-css-lsp/internal/css/scanner"
)

// RenameEdit represents a text replacement for a rename.
type RenameEdit struct {
	StartPos int
	EndPos   int
	NewText  string
}

// PrepareRename checks if rename is valid at the offset and
// returns the range to rename.
func PrepareRename(
	ss *parser.Stylesheet,
	src []byte,
	offset int,
) (Location, bool) {
	name := FindCustomPropertyAt(ss, src, offset)
	if name == "" {
		return Location{}, false
	}

	// Find the token at the cursor
	var loc Location
	found := false

	parser.Walk(ss, func(n parser.Node) bool {
		if found {
			return false
		}

		decl, ok := n.(*parser.Declaration)
		if !ok {
			return true
		}

		// Check declaration property
		if decl.Property.Value == name &&
			offset >= decl.Property.Offset &&
			offset <= decl.Property.End {
			loc = Location{
				StartPos: decl.Property.Offset,
				EndPos:   decl.Property.End,
			}
			found = true
			return false
		}

		// Check var() usages
		if decl.Value == nil {
			return true
		}

		tokens := decl.Value.Tokens
		for i, tok := range tokens {
			if tok.Kind == scanner.Function &&
				strings.ToLower(tok.Value) == VarFunctionName {
				for j := i + 1; j < len(tokens); j++ {
					if tokens[j].Kind == scanner.Whitespace {
						continue
					}
					if tokens[j].Kind == scanner.Ident &&
						tokens[j].Value == name &&
						offset >= tok.Offset {
						// Check if within var() range
						varEnd := tokens[j].End
						for k := j + 1; k < len(tokens); k++ {
							if tokens[k].Kind == scanner.ParenClose {
								varEnd = tokens[k].End
								break
							}
						}
						if offset <= varEnd {
							loc = Location{
								StartPos: tokens[j].Offset,
								EndPos:   tokens[j].End,
							}
							found = true
							return false
						}
					}
					break
				}
			}
		}

		return true
	})

	return loc, found
}

// Rename renames a CSS custom property at the given offset.
func Rename(
	ss *parser.Stylesheet,
	src []byte,
	offset int,
	newName string,
) []RenameEdit {
	name := FindCustomPropertyAt(ss, src, offset)
	if name == "" {
		return nil
	}

	// Ensure new name starts with --
	if !IsCustomProperty(newName) {
		newName = CustomPropertyPrefix + newName
	}

	refs := FindReferences(ss, src, offset)
	edits := make([]RenameEdit, len(refs))
	for i, ref := range refs {
		edits[i] = RenameEdit{
			StartPos: ref.StartPos,
			EndPos:   ref.EndPos,
			NewText:  newName,
		}
	}
	return edits
}
