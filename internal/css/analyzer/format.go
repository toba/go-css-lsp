package analyzer

import (
	"strings"

	"github.com/toba/go-css-lsp/internal/css/parser"
	"github.com/toba/go-css-lsp/internal/css/scanner"
)

// FormatOptions controls CSS formatting behavior.
type FormatOptions struct {
	TabSize      int
	InsertSpaces bool
}

// Format formats the CSS document and returns the formatted
// text.
func Format(
	ss *parser.Stylesheet,
	src []byte,
	opts FormatOptions,
) string {
	if ss == nil {
		return string(src)
	}

	if opts.TabSize == 0 {
		opts.TabSize = 2
	}

	f := &formatter{
		src:    src,
		indent: 0,
		opts:   opts,
	}

	f.formatStylesheet(ss)
	return f.buf.String()
}

type formatter struct {
	src    []byte
	buf    strings.Builder
	indent int
	opts   FormatOptions
}

func (f *formatter) writeIndent() {
	if f.opts.InsertSpaces {
		for range f.indent * f.opts.TabSize {
			f.buf.WriteByte(' ')
		}
	} else {
		for range f.indent {
			f.buf.WriteByte('\t')
		}
	}
}

func (f *formatter) formatStylesheet(ss *parser.Stylesheet) {
	for i, child := range ss.Children {
		if i > 0 {
			f.buf.WriteByte('\n')
		}

		switch n := child.(type) {
		case *parser.Ruleset:
			f.formatRuleset(n)
		case *parser.AtRule:
			f.formatAtRule(n)
		case *parser.Comment:
			f.formatComment(n)
		}
	}

	// Ensure trailing newline
	if f.buf.Len() > 0 {
		s := f.buf.String()
		if s[len(s)-1] != '\n' {
			f.buf.WriteByte('\n')
		}
	}
}

func (f *formatter) formatRuleset(rs *parser.Ruleset) {
	f.writeIndent()

	// Write selector
	if rs.Selectors != nil {
		f.writeSelector(rs.Selectors)
	}

	f.buf.WriteString(" {\n")
	f.indent++

	for _, decl := range rs.Declarations {
		f.formatDeclaration(decl)
	}

	f.indent--
	f.writeIndent()
	f.buf.WriteString("}\n")
}

func (f *formatter) writeSelector(sl *parser.SelectorList) {
	for i, sel := range sl.Selectors {
		if i > 0 {
			f.buf.WriteString(",\n")
			f.writeIndent()
		}
		f.writeSingleSelector(sel)
	}
}

func (f *formatter) writeSingleSelector(sel *parser.Selector) {
	for i, part := range sel.Parts {
		if part.Combinator != "" && part.Combinator != " " {
			f.buf.WriteByte(' ')
			f.buf.WriteString(part.Combinator)
			f.buf.WriteByte(' ')
		} else if i > 0 && part.Combinator == " " {
			f.buf.WriteByte(' ')
		}

		if part.Token.Kind != scanner.EOF {
			f.buf.WriteString(
				string(f.src[part.Token.Offset:part.Token.End]),
			)
		}
	}
}

func (f *formatter) formatDeclaration(
	decl *parser.Declaration,
) {
	f.writeIndent()
	f.buf.WriteString(decl.Property.Value)
	f.buf.WriteString(": ")

	if decl.Value != nil {
		f.writeValue(decl.Value)
	}

	if decl.Important {
		f.buf.WriteString(" !important")
	}

	f.buf.WriteString(";\n")
}

func (f *formatter) writeValue(v *parser.Value) {
	prevEnd := -1
	for _, tok := range v.Tokens {
		if tok.Kind == scanner.Whitespace {
			continue
		}
		// Add space between non-adjacent tokens
		if prevEnd >= 0 && tok.Offset > prevEnd {
			// Don't add space after ( or before )
			if tok.Kind != scanner.ParenClose {
				f.buf.WriteByte(' ')
			}
		}
		f.buf.WriteString(
			string(f.src[tok.Offset:tok.End]),
		)
		prevEnd = tok.End
	}
}

func (f *formatter) formatAtRule(ar *parser.AtRule) {
	f.writeIndent()
	f.buf.WriteByte('@')
	f.buf.WriteString(ar.Name)

	if len(ar.Prelude) > 0 {
		f.buf.WriteByte(' ')
		prevEnd := -1
		for _, tok := range ar.Prelude {
			if tok.Kind == scanner.Whitespace {
				continue
			}
			// If there was whitespace between tokens, add a
			// space
			if prevEnd >= 0 && tok.Offset > prevEnd {
				f.buf.WriteByte(' ')
			}
			f.buf.WriteString(
				string(f.src[tok.Offset:tok.End]),
			)
			prevEnd = tok.End
		}
	}

	if ar.Block != nil {
		f.buf.WriteString(" {\n")
		f.indent++
		for i, child := range ar.Block.Children {
			if i > 0 {
				f.buf.WriteByte('\n')
			}
			switch n := child.(type) {
			case *parser.Ruleset:
				f.formatRuleset(n)
			case *parser.AtRule:
				f.formatAtRule(n)
			case *parser.Comment:
				f.formatComment(n)
			}
		}
		f.indent--
		f.writeIndent()
		f.buf.WriteString("}\n")
	} else {
		f.buf.WriteString(";\n")
	}
}

func (f *formatter) formatComment(c *parser.Comment) {
	f.writeIndent()
	f.buf.WriteString(
		string(f.src[c.StartPos:c.EndPos]),
	)
	f.buf.WriteByte('\n')
}
