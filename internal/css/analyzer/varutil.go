package analyzer

import (
	"strings"

	"github.com/toba/go-css-lsp/internal/css/parser"
	"github.com/toba/go-css-lsp/internal/css/scanner"
)

// VarFunctionName is the CSS var() function name.
const VarFunctionName = "var"

// CustomPropertyPrefix is the prefix for CSS custom properties.
const CustomPropertyPrefix = "--"

// IsCustomProperty returns true if the name starts with --.
func IsCustomProperty(name string) bool {
	return strings.HasPrefix(name, CustomPropertyPrefix)
}

// varRef describes a var(--name) reference found in tokens.
type varRef struct {
	// identIdx is the index of the --custom-property ident
	// token in the token slice.
	identIdx int
	// funcIdx is the index of the var( function token.
	funcIdx int
	// varStart is the byte offset of the var( token.
	varStart int
	// varEnd is the byte offset past the closing paren (or
	// the ident end if no paren found).
	varEnd int
}

// findVarRefs finds all var(--name) references in a token
// slice. For each var() call with a custom property ident, it
// records the function token index, ident token index, and the
// full byte range from var( to the closing ).
func findVarRefs(tokens []scanner.Token) []varRef {
	var refs []varRef
	for i, tok := range tokens {
		if tok.Kind != scanner.Function ||
			strings.ToLower(tok.Value) != VarFunctionName {
			continue
		}
		for j := i + 1; j < len(tokens); j++ {
			if tokens[j].Kind == scanner.Whitespace {
				continue
			}
			if tokens[j].Kind == scanner.Ident &&
				IsCustomProperty(tokens[j].Value) {
				varEnd := tokens[j].End
				for k := j + 1; k < len(tokens); k++ {
					if tokens[k].Kind == scanner.ParenClose {
						varEnd = tokens[k].End
						break
					}
				}
				refs = append(refs, varRef{
					identIdx: j,
					funcIdx:  i,
					varStart: tok.Offset,
					varEnd:   varEnd,
				})
			}
			break
		}
	}
	return refs
}

// FindCustomPropertyAt determines the custom property name at
// the cursor position. Works on both declarations and var()
// usages.
func FindCustomPropertyAt(
	ss *parser.Stylesheet,
	_ []byte,
	offset int,
) string {
	var result string

	parser.Walk(ss, func(n parser.Node) bool {
		if result != "" {
			return false
		}

		decl, ok := n.(*parser.Declaration)
		if !ok {
			return true
		}

		// Check if cursor is on a custom property declaration
		if IsCustomProperty(decl.Property.Value) &&
			offset >= decl.Property.Offset &&
			offset <= decl.Property.End {
			result = decl.Property.Value
			return false
		}

		// Check if cursor is inside a var() usage
		if decl.Value == nil {
			return true
		}

		tokens := decl.Value.Tokens
		for _, ref := range findVarRefs(tokens) {
			if offset >= ref.varStart &&
				offset <= ref.varEnd {
				result = tokens[ref.identIdx].Value
				return false
			}
		}

		return true
	})

	return result
}

// FindVarReferenceAt checks if the cursor is on a var()
// reference and returns the variable name. Also handles cursor
// directly on a --variable ident inside var().
func FindVarReferenceAt(
	ss *parser.Stylesheet,
	_ []byte,
	offset int,
) string {
	var result string

	parser.Walk(ss, func(n parser.Node) bool {
		if result != "" {
			return false
		}

		decl, ok := n.(*parser.Declaration)
		if !ok {
			return true
		}
		if decl.Value == nil {
			return true
		}

		tokens := decl.Value.Tokens
		for _, ref := range findVarRefs(tokens) {
			if offset >= ref.varStart &&
				offset <= ref.varEnd {
				result = tokens[ref.identIdx].Value
				return false
			}
		}

		// Check if cursor is directly on a --variable ident
		// inside a var()
		for _, tok := range tokens {
			if tok.Kind == scanner.Ident &&
				IsCustomProperty(tok.Value) &&
				offset >= tok.Offset && offset <= tok.End {
				for _, ref := range findVarRefs(tokens) {
					if ref.identIdx >= 0 &&
						tokens[ref.identIdx].Offset == tok.Offset {
						result = tok.Value
						return false
					}
				}
			}
		}
		return true
	})

	return result
}

// FindVarReferenceWithRange is like FindVarReferenceAt but also
// returns the start and end offsets of the full var(--name)
// expression.
func FindVarReferenceWithRange(
	ss *parser.Stylesheet,
	_ []byte,
	offset int,
) (name string, start, end int) {
	parser.Walk(ss, func(n parser.Node) bool {
		if name != "" {
			return false
		}

		decl, ok := n.(*parser.Declaration)
		if !ok {
			return true
		}
		if decl.Value == nil {
			return true
		}

		tokens := decl.Value.Tokens
		for _, ref := range findVarRefs(tokens) {
			if offset < ref.varStart || offset > ref.varEnd {
				continue
			}
			ident := tokens[ref.identIdx]
			name = ident.Value
			start = ident.Offset
			end = ident.End
			return false
		}

		// Check if cursor is directly on a --variable ident
		// inside a var()
		for _, tok := range tokens {
			if tok.Kind == scanner.Ident &&
				IsCustomProperty(tok.Value) &&
				offset >= tok.Offset && offset <= tok.End {
				for _, ref := range findVarRefs(tokens) {
					if tokens[ref.identIdx].Offset == tok.Offset {
						name = tok.Value
						start = tok.Offset
						end = tok.End
						return false
					}
				}
			}
		}
		return true
	})

	return name, start, end
}

// ForEachVarUsage calls fn for each var(--name) usage token
// matching the given name.
func ForEachVarUsage(
	ss *parser.Stylesheet,
	name string,
	fn func(tok scanner.Token),
) {
	parser.Walk(ss, func(n parser.Node) bool {
		decl, ok := n.(*parser.Declaration)
		if !ok {
			return true
		}
		if decl.Value == nil {
			return true
		}

		for _, ref := range findVarRefs(decl.Value.Tokens) {
			if decl.Value.Tokens[ref.identIdx].Value == name {
				fn(decl.Value.Tokens[ref.identIdx])
			}
		}

		return true
	})
}
