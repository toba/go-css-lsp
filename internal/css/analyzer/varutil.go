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
		for i, tok := range tokens {
			if tok.Kind == scanner.Function &&
				strings.ToLower(tok.Value) == VarFunctionName {
				for j := i + 1; j < len(tokens); j++ {
					if tokens[j].Kind == scanner.Whitespace {
						continue
					}
					if tokens[j].Kind == scanner.Ident &&
						IsCustomProperty(tokens[j].Value) {
						// Check if cursor is in var(...) range
						varEnd := tokens[j].End
						for k := j + 1; k < len(tokens); k++ {
							if tokens[k].Kind == scanner.ParenClose {
								varEnd = tokens[k].End
								break
							}
						}
						if offset >= tok.Offset &&
							offset <= varEnd {
							result = tokens[j].Value
							return false
						}
					}
					break
				}
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
		for i, tok := range tokens {
			// Check if cursor is on a var() function call
			if tok.Kind == scanner.Function &&
				strings.ToLower(tok.Value) == VarFunctionName {
				for j := i + 1; j < len(tokens); j++ {
					if tokens[j].Kind == scanner.Whitespace {
						continue
					}
					if tokens[j].Kind == scanner.Ident &&
						IsCustomProperty(tokens[j].Value) {
						varStart := tok.Offset
						varEnd := tokens[j].End
						for k := j + 1; k < len(tokens); k++ {
							if tokens[k].Kind == scanner.ParenClose {
								varEnd = tokens[k].End
								break
							}
						}
						if offset >= varStart && offset <= varEnd {
							result = tokens[j].Value
							return false
						}
					}
					break
				}
			}

			// Check if cursor is directly on a --variable ident
			// inside a var()
			if tok.Kind == scanner.Ident &&
				IsCustomProperty(tok.Value) &&
				offset >= tok.Offset && offset <= tok.End {
				if i > 0 {
					for k := i - 1; k >= 0; k-- {
						if tokens[k].Kind == scanner.Whitespace {
							continue
						}
						if tokens[k].Kind == scanner.Function &&
							strings.ToLower(tokens[k].Value) == VarFunctionName {
							result = tok.Value
							return false
						}
						break
					}
				}
			}
		}
		return true
	})

	return result
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

		for i, tok := range decl.Value.Tokens {
			if tok.Kind == scanner.Function &&
				strings.ToLower(tok.Value) == VarFunctionName {
				for j := i + 1; j < len(decl.Value.Tokens); j++ {
					t := decl.Value.Tokens[j]
					if t.Kind == scanner.Whitespace {
						continue
					}
					if t.Kind == scanner.Ident &&
						t.Value == name {
						fn(t)
					}
					break
				}
			}
		}

		return true
	})
}
