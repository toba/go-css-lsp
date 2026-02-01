package analyzer

import (
	"strings"

	"github.com/toba/go-css-lsp/internal/css/data"
	"github.com/toba/go-css-lsp/internal/css/parser"
	"github.com/toba/go-css-lsp/internal/css/scanner"
)

// Analyze returns diagnostics for the parsed stylesheet.
func Analyze(
	ss *parser.Stylesheet,
	src []byte,
) []Diagnostic {
	if ss == nil {
		return nil
	}
	a := &diagAnalyzer{src: src}
	a.analyzeStylesheet(ss)
	return a.diags
}

type diagAnalyzer struct {
	src   []byte
	diags []Diagnostic
}

func (a *diagAnalyzer) analyzeStylesheet(ss *parser.Stylesheet) {
	for _, child := range ss.Children {
		switch n := child.(type) {
		case *parser.Ruleset:
			a.analyzeRuleset(n)
		case *parser.AtRule:
			a.analyzeAtRule(n)
		}
	}
}

func (a *diagAnalyzer) analyzeRuleset(rs *parser.Ruleset) {
	// Check for empty rulesets
	if len(rs.Declarations) == 0 {
		line, char := offsetToLineChar(
			a.src, rs.Offset(),
		)
		endLine, endChar := offsetToLineChar(
			a.src, rs.End(),
		)
		a.diags = append(a.diags, Diagnostic{
			Message:   "empty ruleset",
			StartLine: line,
			StartChar: char,
			EndLine:   endLine,
			EndChar:   endChar,
			Severity:  SeverityHint,
		})
	}

	seen := make(map[string]bool)

	for _, decl := range rs.Declarations {
		a.analyzeDeclaration(decl, seen)
	}
}

func (a *diagAnalyzer) analyzeDeclaration(
	decl *parser.Declaration,
	seen map[string]bool,
) {
	propName := decl.Property.Value

	// Skip custom properties
	if strings.HasPrefix(propName, "--") {
		return
	}

	// Check for unknown properties
	if !data.IsKnownProperty(propName) {
		line, char := offsetToLineChar(
			a.src, decl.Property.Offset,
		)
		endLine, endChar := offsetToLineChar(
			a.src, decl.Property.End,
		)
		a.diags = append(a.diags, Diagnostic{
			Message:   "unknown property '" + propName + "'",
			StartLine: line,
			StartChar: char,
			EndLine:   endLine,
			EndChar:   endChar,
			Severity:  SeverityWarning,
		})
	}

	// Check for duplicate properties
	if seen[propName] {
		line, char := offsetToLineChar(
			a.src, decl.Property.Offset,
		)
		endLine, endChar := offsetToLineChar(
			a.src, decl.Property.End,
		)
		a.diags = append(a.diags, Diagnostic{
			Message:   "duplicate property '" + propName + "'",
			StartLine: line,
			StartChar: char,
			EndLine:   endLine,
			EndChar:   endChar,
			Severity:  SeverityWarning,
		})
	}
	seen[propName] = true
}

func (a *diagAnalyzer) analyzeAtRule(rule *parser.AtRule) {
	if !data.IsKnownAtRule(rule.Name) {
		line, char := offsetToLineChar(
			a.src, rule.Offset(),
		)
		endLine, endChar := offsetToLineChar(
			a.src, rule.Offset()+len(rule.Name)+1,
		)
		a.diags = append(a.diags, Diagnostic{
			Message:   "unknown at-rule '@" + rule.Name + "'",
			StartLine: line,
			StartChar: char,
			EndLine:   endLine,
			EndChar:   endChar,
			Severity:  SeverityWarning,
		})
	}

	if rule.Block != nil {
		a.analyzeStylesheet(rule.Block)
	}
}

// offsetToLineChar converts a byte offset to line/character.
func offsetToLineChar(src []byte, offset int) (int, int) {
	line := 0
	char := 0
	for i := range offset {
		if i >= len(src) {
			break
		}
		if src[i] == '\n' {
			line++
			char = 0
		} else {
			char++
		}
	}
	return line, char
}

// nodeAtOffset finds what kind of context the offset is in.
func nodeAtOffset(
	ss *parser.Stylesheet,
	offset int,
) (node parser.Node, inDeclarationBlock bool) {
	var found parser.Node
	inBlock := false

	parser.Walk(ss, func(n parser.Node) bool {
		if n.Offset() > offset || n.End() < offset {
			return false
		}
		found = n

		if _, ok := n.(*parser.Ruleset); ok {
			inBlock = true
		}

		return true
	})

	return found, inBlock
}

// tokenAtOffset finds the token at the given byte offset in
// the value tokens or declaration.
func tokenAtOffset(
	ss *parser.Stylesheet,
	offset int,
) *scanner.Token {
	var result *scanner.Token

	parser.Walk(ss, func(n parser.Node) bool {
		if n.Offset() > offset || n.End() < offset {
			return false
		}

		switch node := n.(type) {
		case *parser.Declaration:
			if node.Property.Offset <= offset &&
				offset < node.Property.End {
				result = &node.Property
				return false
			}
		case *parser.Value:
			for i := range node.Tokens {
				t := &node.Tokens[i]
				if t.Offset <= offset && offset < t.End {
					result = t
					return false
				}
			}
		}

		return true
	})

	return result
}
