package analyzer

import (
	"strings"

	"github.com/toba/go-css-lsp/internal/css/data"
	"github.com/toba/go-css-lsp/internal/css/parser"
	"github.com/toba/go-css-lsp/internal/css/scanner"
)

// Hover returns markdown hover content for the given byte
// offset.
func Hover(
	ss *parser.Stylesheet,
	src []byte,
	offset int,
) (string, bool) {
	if ss == nil {
		return "", false
	}

	tok := tokenAtOffset(ss, offset)
	if tok == nil {
		// Check selectors for pseudo-classes/elements
		return hoverSelector(ss, offset)
	}

	switch tok.Kind {
	case scanner.Ident:
		return hoverIdent(ss, tok, offset)
	case scanner.AtKeyword:
		return hoverAtKeyword(tok)
	case scanner.Function:
		return hoverFunction(tok)
	}

	return "", false
}

func hoverIdent(
	ss *parser.Stylesheet,
	tok *scanner.Token,
	offset int,
) (string, bool) {
	// Check if it's a property name
	node, _ := nodeAtOffset(ss, offset)
	if decl, ok := node.(*parser.Declaration); ok {
		if decl.Property.Offset == tok.Offset {
			return hoverProperty(tok.Value)
		}
	}

	// Check if it's a known value keyword
	prop := data.LookupProperty(tok.Value)
	if prop != nil {
		return hoverProperty(tok.Value)
	}

	return "", false
}

func hoverProperty(name string) (string, bool) {
	if strings.HasPrefix(name, "--") {
		return "", false
	}

	prop := data.LookupProperty(name)
	if prop == nil {
		return "", false
	}

	var b strings.Builder
	b.WriteString("**")
	b.WriteString(prop.Name)
	b.WriteString("**\n\n")
	b.WriteString(prop.Description)

	if len(prop.Values) > 0 {
		b.WriteString("\n\nValues: `")
		b.WriteString(strings.Join(prop.Values, "` | `"))
		b.WriteString("`")
	}

	if prop.MDN != "" {
		b.WriteString("\n\n[MDN Reference](")
		b.WriteString(prop.MDN)
		b.WriteString(")")
	}

	return b.String(), true
}

func hoverAtKeyword(
	tok *scanner.Token,
) (string, bool) {
	rule := data.LookupAtRule(tok.Value)
	if rule == nil {
		return "", false
	}

	var b strings.Builder
	b.WriteString("**@")
	b.WriteString(rule.Name)
	b.WriteString("**\n\n")
	b.WriteString(rule.Description)
	return b.String(), true
}

func hoverFunction(
	tok *scanner.Token,
) (string, bool) {
	// Check if it's a known function
	name := tok.Value
	desc := functionDescription(name)
	if desc == "" {
		return "", false
	}

	var b strings.Builder
	b.WriteString("**")
	b.WriteString(name)
	b.WriteString("()**\n\n")
	b.WriteString(desc)
	return b.String(), true
}

func hoverSelector(
	ss *parser.Stylesheet,
	offset int,
) (string, bool) {
	// Walk AST looking for selector parts
	var result string
	found := false

	parser.Walk(ss, func(n parser.Node) bool {
		sel, ok := n.(*parser.Selector)
		if !ok {
			return true
		}

		for _, part := range sel.Parts {
			if part.Token.Kind == scanner.Colon &&
				part.Token.Offset <= offset &&
				offset < part.Token.End {
				// Look at next part for pseudo name
				found = true
			}
		}

		// Check pseudo-class/element in selector tokens
		for i, part := range sel.Parts {
			if part.Token.Kind != scanner.Ident {
				continue
			}
			if part.Token.Offset > offset ||
				offset >= part.Token.End {
				continue
			}
			// Check if preceded by : or ::
			if i > 0 {
				prev := sel.Parts[i-1]
				if prev.Token.Kind == scanner.Colon {
					// Could be pseudo-class or
					// pseudo-element
					pc := data.LookupPseudoClass(
						part.Token.Value,
					)
					if pc != nil {
						result = "**:" + pc.Name +
							"**\n\n" + pc.Description
						found = true
						return false
					}
					pe := data.LookupPseudoElement(
						part.Token.Value,
					)
					if pe != nil {
						result = "**::" + pe.Name +
							"**\n\n" + pe.Description
						found = true
						return false
					}
				}
			}
		}

		return !found
	})

	return result, found
}

func functionDescription(name string) string {
	switch strings.ToLower(name) {
	case "calc":
		return "Performs calculations to determine CSS property values."
	case "var":
		return "Inserts the value of a CSS custom property."
	case "rgb", "rgba":
		return "Specifies a color using red, green, blue (and alpha) values."
	case "hsl", "hsla":
		return "Specifies a color using hue, saturation, lightness (and alpha) values."
	case "min":
		return "Returns the smallest of the given values."
	case "max":
		return "Returns the largest of the given values."
	case "clamp":
		return "Clamps a value between a minimum and maximum."
	case "url":
		return "References a resource by URL."
	case "linear-gradient":
		return "Creates a linear gradient image."
	case "radial-gradient":
		return "Creates a radial gradient image."
	case "conic-gradient":
		return "Creates a conic gradient image."
	case "env":
		return "Inserts the value of a user-agent defined environment variable."
	case "hwb":
		return "Specifies a color using hue, whiteness, blackness values."
	case "oklch":
		return "Specifies a color in the OKLCH color space."
	case "oklab":
		return "Specifies a color in the OKLAB color space."
	case "color-mix":
		return "Mixes two colors in a given color space."
	case "light-dark":
		return "Returns one of two colors depending on the user's color scheme preference."
	default:
		return ""
	}
}
