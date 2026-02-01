package parser

import (
	"github.com/toba/go-css-lsp/internal/css/scanner"
)

// Parser performs recursive descent parsing of CSS.
type Parser struct {
	tokens []scanner.Token
	pos    int
	Errors []*Error
}

// Parse parses CSS source into a Stylesheet AST.
func Parse(src []byte) (*Stylesheet, []*Error) {
	tokens := scanner.ScanAll(src)
	p := &Parser{tokens: tokens}
	ss := p.parseStylesheet()
	return ss, p.Errors
}

func (p *Parser) peek() scanner.Token {
	if p.pos >= len(p.tokens) {
		return scanner.Token{Kind: scanner.EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) next() scanner.Token {
	t := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return t
}

func (p *Parser) skipWhitespace() {
	for p.peek().Kind == scanner.Whitespace {
		p.pos++
	}
}

func (p *Parser) skipWhitespaceAndComments() []Node {
	var comments []Node
	for {
		t := p.peek()
		if t.Kind == scanner.Whitespace {
			p.pos++
			continue
		}
		if t.Kind == scanner.Comment {
			p.pos++
			comments = append(comments, &Comment{
				Text:     t.Value,
				StartPos: t.Offset,
				EndPos:   t.End,
			})
			continue
		}
		break
	}
	return comments
}

func (p *Parser) addError(msg string, start, end int) {
	p.Errors = append(p.Errors, &Error{
		Message:  msg,
		StartPos: start,
		EndPos:   end,
	})
}

func (p *Parser) parseStylesheet() *Stylesheet {
	ss := &Stylesheet{}

	for {
		comments := p.skipWhitespaceAndComments()
		ss.Children = append(ss.Children, comments...)

		t := p.peek()
		if t.Kind == scanner.EOF {
			ss.EndPos = t.End
			break
		}

		node := p.parseRule()
		if node != nil {
			ss.Children = append(ss.Children, node)
		}
	}

	return ss
}

func (p *Parser) parseRule() Node {
	t := p.peek()

	if t.Kind == scanner.AtKeyword {
		return p.parseAtRule()
	}

	if t.Kind == scanner.CDO || t.Kind == scanner.CDC {
		p.next()
		return nil
	}

	return p.parseRuleset()
}

func (p *Parser) parseAtRule() *AtRule {
	t := p.next() // consume @keyword
	rule := &AtRule{
		Name:     t.Value,
		StartPos: t.Offset,
	}

	// Collect prelude tokens until { or ;
	for {
		p.skipWhitespace()
		t = p.peek()

		if t.Kind == scanner.EOF {
			rule.EndPos = t.Offset
			p.addError(
				"unexpected end of file in at-rule",
				rule.StartPos, t.Offset,
			)
			return rule
		}

		if t.Kind == scanner.Semicolon {
			p.next()
			rule.EndPos = t.End
			return rule
		}

		if t.Kind == scanner.BraceOpen {
			rule.Block = p.parseBlock()
			rule.EndPos = rule.Block.EndPos
			return rule
		}

		rule.Prelude = append(rule.Prelude, p.next())
	}
}

func (p *Parser) parseBlock() *Stylesheet {
	p.next() // consume {
	ss := &Stylesheet{}

	depth := 1
	for depth > 0 {
		comments := p.skipWhitespaceAndComments()
		ss.Children = append(ss.Children, comments...)

		t := p.peek()
		if t.Kind == scanner.EOF {
			ss.EndPos = t.Offset
			p.addError(
				"unexpected end of file, expected '}'",
				t.Offset, t.Offset,
			)
			return ss
		}

		if t.Kind == scanner.BraceClose {
			p.next()
			depth--
			if depth == 0 {
				ss.EndPos = t.End
				return ss
			}
			continue
		}

		if t.Kind == scanner.AtKeyword {
			node := p.parseAtRule()
			ss.Children = append(ss.Children, node)
			continue
		}

		// Could be a nested ruleset or declarations.
		// Try to determine: if we see ident followed by colon
		// before a brace, it's declarations.
		if p.looksLikeDeclaration() {
			decl := p.parseDeclaration()
			if decl != nil {
				ss.Children = append(ss.Children, decl)
			}
		} else {
			node := p.parseRuleset()
			if node != nil {
				ss.Children = append(ss.Children, node)
			}
		}
	}

	return ss
}

// looksLikeDeclaration peeks ahead to determine whether the
// current position starts a declaration (ident : ...) or a
// selector.
func (p *Parser) looksLikeDeclaration() bool {
	saved := p.pos
	defer func() { p.pos = saved }()

	// Skip whitespace
	for p.pos < len(p.tokens) &&
		p.tokens[p.pos].Kind == scanner.Whitespace {
		p.pos++
	}

	// Must start with ident (including custom properties
	// starting with --)
	if p.pos >= len(p.tokens) {
		return false
	}

	t := p.tokens[p.pos]
	if t.Kind != scanner.Ident {
		return false
	}
	p.pos++

	// Skip whitespace
	for p.pos < len(p.tokens) &&
		p.tokens[p.pos].Kind == scanner.Whitespace {
		p.pos++
	}

	// Must be followed by colon
	if p.pos >= len(p.tokens) {
		return false
	}

	return p.tokens[p.pos].Kind == scanner.Colon
}

func (p *Parser) parseRuleset() *Ruleset {
	rs := &Ruleset{StartPos: p.peek().Offset}

	rs.Selectors = p.parseSelectorList()

	p.skipWhitespace()
	t := p.peek()
	if t.Kind != scanner.BraceOpen {
		p.addError(
			"expected '{' after selector",
			t.Offset, t.End,
		)
		// Error recovery: skip to next { or }
		p.skipToRecovery()
		rs.EndPos = p.peek().Offset
		return rs
	}
	p.next() // consume {

	// Parse declarations
	for {
		p.skipWhitespaceAndComments()
		t = p.peek()

		if t.Kind == scanner.BraceClose {
			p.next()
			rs.EndPos = t.End
			return rs
		}

		if t.Kind == scanner.EOF {
			p.addError(
				"unexpected end of file, expected '}'",
				rs.StartPos, t.Offset,
			)
			rs.EndPos = t.Offset
			return rs
		}

		decl := p.parseDeclaration()
		if decl != nil {
			rs.Declarations = append(
				rs.Declarations, decl,
			)
		}
	}
}

func (p *Parser) parseSelectorList() *SelectorList {
	sl := &SelectorList{StartPos: p.peek().Offset}

	sel := p.parseSelector()
	if sel != nil {
		sl.Selectors = append(sl.Selectors, sel)
	}

	for {
		p.skipWhitespace()
		if p.peek().Kind != scanner.Comma {
			break
		}
		p.next() // consume comma
		p.skipWhitespace()
		sel = p.parseSelector()
		if sel != nil {
			sl.Selectors = append(sl.Selectors, sel)
		}
	}

	if len(sl.Selectors) > 0 {
		last := sl.Selectors[len(sl.Selectors)-1]
		sl.EndPos = last.EndPos
	}

	return sl
}

func (p *Parser) parseSelector() *Selector {
	sel := &Selector{StartPos: p.peek().Offset}

	for {
		t := p.peek()

		if t.Kind == scanner.EOF ||
			t.Kind == scanner.BraceOpen ||
			t.Kind == scanner.Comma {
			break
		}

		// Whitespace could be descendant combinator
		if t.Kind == scanner.Whitespace {
			p.next()
			next := p.peek()
			if next.Kind == scanner.BraceOpen ||
				next.Kind == scanner.Comma ||
				next.Kind == scanner.EOF {
				break
			}
			// Check for explicit combinators
			if next.Kind == scanner.Delim &&
				(next.Value == ">" ||
					next.Value == "+" ||
					next.Value == "~") {
				continue
			}
			sel.Parts = append(sel.Parts, SelectorPart{
				Combinator: " ",
			})
			continue
		}

		// Explicit combinators
		if t.Kind == scanner.Delim &&
			(t.Value == ">" ||
				t.Value == "+" ||
				t.Value == "~") {
			p.next()
			p.skipWhitespace()
			sel.Parts = append(sel.Parts, SelectorPart{
				Combinator: t.Value,
			})
			continue
		}

		sel.Parts = append(sel.Parts, SelectorPart{
			Token: p.next(),
		})
	}

	if len(sel.Parts) > 0 {
		last := sel.Parts[len(sel.Parts)-1]
		if last.Token.Kind != 0 {
			sel.EndPos = last.Token.End
		}
	}

	if len(sel.Parts) == 0 {
		return nil
	}
	return sel
}

func (p *Parser) parseDeclaration() *Declaration {
	p.skipWhitespace()
	t := p.peek()

	if t.Kind != scanner.Ident {
		// Error recovery
		p.addError(
			"expected property name",
			t.Offset, t.End,
		)
		p.skipToSemicolonOrBrace()
		return nil
	}

	decl := &Declaration{
		Property: p.next(),
		StartPos: t.Offset,
	}

	p.skipWhitespace()
	t = p.peek()
	if t.Kind != scanner.Colon {
		p.addError(
			"expected ':' after property name",
			t.Offset, t.End,
		)
		p.skipToSemicolonOrBrace()
		decl.EndPos = p.peek().Offset
		return decl
	}
	p.next() // consume :

	decl.Value = p.parseValue()

	// Check for !important
	if decl.Value != nil && len(decl.Value.Tokens) > 0 {
		tokens := decl.Value.Tokens
		lastIdx := len(tokens) - 1
		if lastIdx >= 1 &&
			tokens[lastIdx].Kind == scanner.Ident &&
			tokens[lastIdx].Value == "important" &&
			tokens[lastIdx-1].Kind == scanner.Delim &&
			tokens[lastIdx-1].Value == "!" {
			decl.Important = true
			decl.Value.Tokens = tokens[:lastIdx-1]
		}
	}

	p.skipWhitespace()
	if p.peek().Kind == scanner.Semicolon {
		p.next()
		decl.Semicolon = true
	}

	if decl.Value != nil {
		decl.EndPos = decl.Value.EndPos
	} else {
		decl.EndPos = p.peek().Offset
	}

	return decl
}

func (p *Parser) parseValue() *Value {
	p.skipWhitespace()
	v := &Value{StartPos: p.peek().Offset}

	depth := 0
	for {
		t := p.peek()

		if t.Kind == scanner.EOF {
			break
		}

		if t.Kind == scanner.Semicolon && depth == 0 {
			break
		}

		if t.Kind == scanner.BraceClose && depth == 0 {
			break
		}

		if t.Kind == scanner.ParenOpen ||
			t.Kind == scanner.Function {
			depth++
		}
		if t.Kind == scanner.ParenClose {
			if depth > 0 {
				depth--
			}
		}

		v.Tokens = append(v.Tokens, p.next())
	}

	// Trim trailing whitespace tokens
	for len(v.Tokens) > 0 &&
		v.Tokens[len(v.Tokens)-1].Kind == scanner.Whitespace {
		v.Tokens = v.Tokens[:len(v.Tokens)-1]
	}

	if len(v.Tokens) > 0 {
		v.EndPos = v.Tokens[len(v.Tokens)-1].End
	} else {
		v.EndPos = v.StartPos
	}

	return v
}

func (p *Parser) skipToRecovery() {
	for {
		t := p.peek()
		if t.Kind == scanner.EOF ||
			t.Kind == scanner.BraceOpen ||
			t.Kind == scanner.BraceClose ||
			t.Kind == scanner.Semicolon {
			return
		}
		p.next()
	}
}

func (p *Parser) skipToSemicolonOrBrace() {
	for {
		t := p.peek()
		if t.Kind == scanner.EOF ||
			t.Kind == scanner.Semicolon ||
			t.Kind == scanner.BraceClose {
			if t.Kind == scanner.Semicolon {
				p.next()
			}
			return
		}
		p.next()
	}
}
