package analyzer

import (
	"strings"

	"github.com/toba/go-css-lsp/internal/css/parser"
	"github.com/toba/go-css-lsp/internal/css/scanner"
)

// DocumentLink represents a link found in a document.
type DocumentLink struct {
	StartPos int
	EndPos   int
	Target   string
}

// FindDocumentLinks returns links found in the CSS document,
// including @import URLs and url() references.
func FindDocumentLinks(
	ss *parser.Stylesheet,
	src []byte,
) []DocumentLink {
	if ss == nil {
		return nil
	}

	var links []DocumentLink

	parser.Walk(ss, func(n parser.Node) bool {
		switch node := n.(type) {
		case *parser.AtRule:
			if node.Name == "import" {
				link := extractImportLink(node)
				if link != nil {
					links = append(links, *link)
				}
			}
		case *parser.Declaration:
			if node.Value != nil {
				links = append(links,
					extractURLLinks(node.Value.Tokens)...,
				)
			}
		}
		return true
	})

	return links
}

func extractImportLink(rule *parser.AtRule) *DocumentLink {
	for _, tok := range rule.Prelude {
		switch tok.Kind {
		case scanner.String:
			target := tok.Value
			return &DocumentLink{
				StartPos: tok.Offset,
				EndPos:   tok.End,
				Target:   target,
			}
		case scanner.URL:
			target := tok.Value
			return &DocumentLink{
				StartPos: tok.Offset,
				EndPos:   tok.End,
				Target:   target,
			}
		case scanner.Function:
			if strings.ToLower(tok.Value) == "url" {
				// Next token should be the URL string
				return nil // handled by URL token type
			}
		}
	}
	return nil
}

func extractURLLinks(
	tokens []scanner.Token,
) []DocumentLink {
	var links []DocumentLink

	for _, tok := range tokens {
		if tok.Kind == scanner.URL && tok.Value != "" {
			links = append(links, DocumentLink{
				StartPos: tok.Offset,
				EndPos:   tok.End,
				Target:   tok.Value,
			})
		}
	}

	return links
}
