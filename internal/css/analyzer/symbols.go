package analyzer

import (
	"strings"

	"github.com/toba/go-css-lsp/internal/css/parser"
)

// DocumentSymbol represents a symbol in a document with
// optional children for hierarchy.
type DocumentSymbol struct {
	Name           string
	Kind           int
	StartPos       int
	EndPos         int
	SelectionStart int
	SelectionEnd   int
	Children       []DocumentSymbol
}

// FindDocumentSymbols returns a hierarchical list of symbols
// in the CSS document.
func FindDocumentSymbols(
	ss *parser.Stylesheet,
	src []byte,
) []DocumentSymbol {
	if ss == nil {
		return nil
	}

	var symbols []DocumentSymbol

	for _, child := range ss.Children {
		switch n := child.(type) {
		case *parser.Ruleset:
			sym := rulesetSymbol(n, src)
			symbols = append(symbols, sym)

		case *parser.AtRule:
			sym := atRuleSymbol(n, src)
			symbols = append(symbols, sym)
		}
	}

	return symbols
}

func rulesetSymbol(
	r *parser.Ruleset,
	src []byte,
) DocumentSymbol {
	name := selectorText(r.Selectors, src)
	selStart := r.StartPos
	selEnd := selStart + len(name)
	if r.Selectors != nil {
		selStart = r.Selectors.Offset()
		selEnd = r.Selectors.End()
	}

	sym := DocumentSymbol{
		Name:           name,
		Kind:           SymbolKindClass,
		StartPos:       r.StartPos,
		EndPos:         r.EndPos,
		SelectionStart: selStart,
		SelectionEnd:   selEnd,
	}

	// Add children: custom properties, nested rulesets, at-rules
	for _, child := range r.Children {
		switch n := child.(type) {
		case *parser.Declaration:
			propName := n.Property.Value
			if IsCustomProperty(propName) {
				sym.Children = append(sym.Children,
					DocumentSymbol{
						Name:           propName,
						Kind:           SymbolKindVariable,
						StartPos:       n.StartPos,
						EndPos:         n.EndPos,
						SelectionStart: n.Property.Offset,
						SelectionEnd:   n.Property.End,
					},
				)
			}
		case *parser.Ruleset:
			sym.Children = append(
				sym.Children, rulesetSymbol(n, src),
			)
		case *parser.AtRule:
			sym.Children = append(
				sym.Children, atRuleSymbol(n, src),
			)
		}
	}

	return sym
}

func atRuleSymbol(
	ar *parser.AtRule,
	src []byte,
) DocumentSymbol {
	name := "@" + ar.Name
	if len(ar.Prelude) > 0 {
		start := ar.Prelude[0].Offset
		end := ar.Prelude[len(ar.Prelude)-1].End
		if start < len(src) && end <= len(src) {
			prelude := strings.TrimSpace(
				string(src[start:end]),
			)
			name += " " + prelude
		}
	}

	sym := DocumentSymbol{
		Name:           name,
		Kind:           SymbolKindString,
		StartPos:       ar.StartPos,
		EndPos:         ar.EndPos,
		SelectionStart: ar.StartPos,
		SelectionEnd:   ar.StartPos + len("@"+ar.Name),
	}

	// Add children from block
	if ar.Block != nil {
		for _, child := range ar.Block.Children {
			switch n := child.(type) {
			case *parser.Ruleset:
				sym.Children = append(
					sym.Children, rulesetSymbol(n, src),
				)
			case *parser.AtRule:
				sym.Children = append(
					sym.Children, atRuleSymbol(n, src),
				)
			}
		}
	}

	return sym
}

func selectorText(
	sl *parser.SelectorList,
	src []byte,
) string {
	if sl == nil {
		return "<unknown>"
	}
	start := sl.Offset()
	end := sl.End()
	if start >= len(src) || end > len(src) {
		return "<unknown>"
	}
	text := string(src[start:end])
	return strings.TrimSpace(text)
}
