package analyzer

import (
	"strings"

	"github.com/toba/go-css-lsp/internal/css/data"
	"github.com/toba/go-css-lsp/internal/css/parser"
	"github.com/toba/go-css-lsp/internal/css/scanner"
)

// HoverResult holds hover content and an optional byte-offset
// range for the hovered span. When RangeStart < RangeEnd the
// editor should highlight that range instead of using its own
// word detection.
type HoverResult struct {
	Content    string
	RangeStart int
	RangeEnd   int
	Found      bool
}

// Hover returns markdown hover content for the given byte
// offset. An optional VariableResolver enables cross-file
// custom property value lookup.
func Hover(
	ss *parser.Stylesheet,
	src []byte,
	offset int,
	resolvers ...VariableResolver,
) HoverResult {
	if ss == nil {
		return HoverResult{}
	}

	var resolver VariableResolver
	if len(resolvers) > 0 {
		resolver = resolvers[0]
	}

	tok := tokenAtOffset(ss, offset)
	if tok == nil {
		// Check selectors for pseudo-classes/elements
		content, found := hoverSelector(ss, offset)
		return HoverResult{Content: content, Found: found}
	}

	switch tok.Kind {
	case scanner.Ident:
		return hoverIdent(ss, src, tok, offset, resolver)
	case scanner.AtKeyword:
		content, found := hoverAtKeyword(tok)
		return HoverResult{Content: content, Found: found}
	case scanner.Function:
		// Check for var() function — return custom prop hover
		if strings.ToLower(tok.Value) == VarFunctionName {
			return hoverVarFunction(
				ss, src, tok, offset, resolver,
			)
		}
		content, found := hoverFunction(tok)
		return HoverResult{Content: content, Found: found}
	}

	return HoverResult{}
}

func hoverIdent(
	ss *parser.Stylesheet,
	src []byte,
	tok *scanner.Token,
	offset int,
	resolver VariableResolver,
) HoverResult {
	decl := declarationAtOffset(ss, offset)

	// Check if it's a property name
	if decl != nil && decl.Property.Offset == tok.Offset {
		if IsCustomProperty(tok.Value) {
			return hoverCustomProperty(
				ss, src, tok.Value,
				tok.Offset, tok.End,
				resolver,
			)
		}
		content, found := hoverProperty(tok.Value)
		return HoverResult{
			Content: content, Found: found,
		}
	}

	// Check for var() reference: custom property ident
	// inside a var() call
	if decl != nil && IsCustomProperty(tok.Value) &&
		decl.Value != nil {
		return hoverVarReference(
			ss, src, decl, tok, resolver,
		)
	}

	// Check if it's a known value keyword
	prop := data.LookupProperty(tok.Value)
	if prop != nil {
		content, found := hoverProperty(tok.Value)
		return HoverResult{Content: content, Found: found}
	}

	return HoverResult{}
}

// declarationAtOffset finds the Declaration node enclosing
// the given offset, or nil.
func declarationAtOffset(
	ss *parser.Stylesheet,
	offset int,
) *parser.Declaration {
	var result *parser.Declaration
	parser.Walk(ss, func(n parser.Node) bool {
		if n.Offset() > offset || n.End() < offset {
			return false
		}
		if d, ok := n.(*parser.Declaration); ok {
			result = d
		}
		return true
	})
	return result
}

// hoverCustomProperty returns hover content for a custom
// property declaration, with a range covering the property
// name token. When the variable is not found in the current
// file, the resolver (if non-nil) is consulted.
func hoverCustomProperty(
	ss *parser.Stylesheet,
	src []byte,
	name string,
	start, end int,
	resolver VariableResolver,
) HoverResult {
	var b strings.Builder
	b.WriteString("`")
	b.WriteString(name)
	b.WriteString("`")

	// Find the declared value in the current file
	var found bool
	parser.Walk(ss, func(n parser.Node) bool {
		decl, ok := n.(*parser.Declaration)
		if !ok {
			return true
		}
		if decl.Property.Value == name &&
			decl.Value != nil {
			valText := strings.TrimSpace(
				string(src[decl.Value.Offset():decl.Value.End()]),
			)
			if valText != "" {
				b.WriteString("\n\n")
				b.WriteString(valText)
				found = true
			}
			return false
		}
		return true
	})

	// Fall back to workspace index
	if !found && resolver != nil {
		if val, ok := resolver.ResolveVariable(name); ok {
			b.WriteString("\n\n")
			b.WriteString(val)
		}
	}

	return HoverResult{
		Content:    b.String(),
		RangeStart: start,
		RangeEnd:   end,
		Found:      true,
	}
}

// hoverVarReference returns hover for a --custom-property ident
// inside a var() call, with a range covering the full var(...)
// expression.
func hoverVarReference(
	ss *parser.Stylesheet,
	src []byte,
	decl *parser.Declaration,
	tok *scanner.Token,
	resolver VariableResolver,
) HoverResult {
	tokens := decl.Value.Tokens
	for _, ref := range findVarRefs(tokens) {
		if tokens[ref.identIdx].Offset == tok.Offset {
			return hoverCustomProperty(
				ss, src, tok.Value,
				ref.varStart, ref.varEnd,
				resolver,
			)
		}
	}
	return HoverResult{}
}

// hoverVarFunction handles hover when the cursor is on the
// "var" function token itself. It finds the custom property
// ident inside and returns hover with range covering the whole
// var(...) expression.
func hoverVarFunction(
	ss *parser.Stylesheet,
	src []byte,
	funcTok *scanner.Token,
	offset int,
	resolver VariableResolver,
) HoverResult {
	// Find the declaration containing this token
	decl := declarationAtOffset(ss, offset)
	if decl == nil || decl.Value == nil {
		return HoverResult{}
	}

	tokens := decl.Value.Tokens
	for _, ref := range findVarRefs(tokens) {
		if tokens[ref.funcIdx].Offset == funcTok.Offset {
			return hoverCustomProperty(
				ss, src,
				tokens[ref.identIdx].Value,
				ref.varStart, ref.varEnd,
				resolver,
			)
		}
	}
	return HoverResult{}
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

	if prop.IsExperimental() {
		b.WriteString("\n\n*Experimental — limited browser support*")
	}

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
	name := strings.ToLower(tok.Value)
	fn := data.LookupFunction(name)
	if fn == nil {
		return "", false
	}

	var b strings.Builder
	b.WriteString("```\n")
	for _, sig := range fn.Signatures {
		b.WriteString(sig)
		b.WriteString("\n")
	}
	b.WriteString("```\n\n")
	b.WriteString(fn.Description)

	if fn.MDN != "" {
		b.WriteString("\n\n[MDN Reference](")
		b.WriteString(fn.MDN)
		b.WriteString(")")
	}

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
