package analyzer

import (
	"bytes"
	"strings"

	"github.com/toba/go-css-lsp/internal/css/parser"
	"github.com/toba/go-css-lsp/internal/css/scanner"
)

// FormatMode controls how rulesets are laid out.
type FormatMode int

const (
	// FormatExpanded puts one declaration per line (default).
	FormatExpanded FormatMode = iota
	// FormatCompact puts rulesets on a single line when they
	// fit within PrintWidth.
	FormatCompact
	// FormatPreserve keeps original single/multi-line layout
	// and normalizes whitespace only.
	FormatPreserve
)

// FormatOptions controls CSS formatting behavior.
type FormatOptions struct {
	TabSize      int
	InsertSpaces bool
	Mode         FormatMode
	PrintWidth   int
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
	if opts.PrintWidth == 0 {
		opts.PrintWidth = 80
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
			f.dispatchRuleset(n)
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

// dispatchRuleset formats a ruleset using the configured mode.
func (f *formatter) dispatchRuleset(rs *parser.Ruleset) {
	switch f.opts.Mode {
	case FormatCompact:
		f.formatRulesetCompact(rs)
	case FormatPreserve:
		f.formatRulesetPreserve(rs)
	default:
		f.formatRuleset(rs)
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

	for i, child := range rs.Children {
		switch n := child.(type) {
		case *parser.Declaration:
			f.formatDeclaration(n)
		case *parser.Ruleset:
			if i > 0 {
				f.buf.WriteByte('\n')
			}
			f.dispatchRuleset(n)
		case *parser.AtRule:
			if i > 0 {
				f.buf.WriteByte('\n')
			}
			f.formatAtRule(n)
		case *parser.Comment:
			f.formatComment(n)
		}
	}

	f.indent--
	f.writeIndent()
	f.buf.WriteString("}\n")
}

// hasNestedRules reports whether a ruleset contains any nested
// rulesets or at-rules (not just declarations).
func hasNestedRules(rs *parser.Ruleset) bool {
	for _, child := range rs.Children {
		switch child.(type) {
		case *parser.Ruleset, *parser.AtRule:
			return true
		}
	}
	return false
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
	if decl.Value != nil {
		// Measure the single-line length.
		var tmp strings.Builder
		f.writeValueTo(&tmp, decl.Value)
		valStr := tmp.String()

		lineLen := f.indentWidth() + len(decl.Property.Value) + 2 + len(valStr) + 1
		if decl.Important {
			lineLen += 11 // " !important"
		}

		commaIndices := f.topLevelCommaIndices(decl.Value)

		if lineLen > f.opts.PrintWidth && len(commaIndices) > 0 {
			f.writeIndent()
			f.buf.WriteString(decl.Property.Value)
			f.buf.WriteString(":\n")
			f.writeValueMultiLine(decl.Value, commaIndices)
			if decl.Important {
				f.buf.WriteString(" !important")
			}
			f.buf.WriteString(";\n")
			return
		}
	}

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

// topLevelCommaIndices returns token indices of commas that
// are not inside parentheses (depth 0).
func (f *formatter) topLevelCommaIndices(
	v *parser.Value,
) []int {
	var indices []int
	depth := 0
	for i, tok := range v.Tokens {
		switch tok.Kind {
		case scanner.Function, scanner.ParenOpen:
			depth++
		case scanner.ParenClose:
			if depth > 0 {
				depth--
			}
		case scanner.Comma:
			if depth == 0 {
				indices = append(indices, i)
			}
		}
	}
	return indices
}

// writeValueMultiLine writes a value across multiple lines,
// breaking at each top-level comma. Each segment is indented
// one level deeper than the current indent.
func (f *formatter) writeValueMultiLine(
	v *parser.Value,
	commaIndices []int,
) {
	commaSet := make(map[int]bool, len(commaIndices))
	for _, idx := range commaIndices {
		commaSet[idx] = true
	}

	f.indent++
	f.writeIndent()

	prevEnd := -1
	afterBreak := true
	for i, tok := range v.Tokens {
		if tok.Kind == scanner.Whitespace {
			continue
		}
		if commaSet[i] {
			// Write the comma, then newline + indent
			f.buf.WriteByte(',')
			f.buf.WriteByte('\n')
			f.writeIndent()
			prevEnd = tok.End
			afterBreak = true
			continue
		}
		if !afterBreak && prevEnd >= 0 && tok.Offset > prevEnd {
			if tok.Kind != scanner.ParenClose {
				f.buf.WriteByte(' ')
			}
		}
		afterBreak = false
		f.buf.WriteString(
			string(f.src[tok.Offset:tok.End]),
		)
		prevEnd = tok.End
	}

	f.indent--
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
				if _, isDecl := child.(*parser.Declaration); !isDecl {
					f.buf.WriteByte('\n')
				}
			}
			switch n := child.(type) {
			case *parser.Declaration:
				f.formatDeclaration(n)
			case *parser.Ruleset:
				f.dispatchRuleset(n)
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

// indentWidth returns the character width of the current
// indent level.
func (f *formatter) indentWidth() int {
	if f.opts.InsertSpaces {
		return f.indent * f.opts.TabSize
	}
	return f.indent
}

// buildSingleLine renders a ruleset as a single-line string:
// "selector { decl; decl; }" or "selector {}" if empty.
// It does not include a trailing newline.
func (f *formatter) buildSingleLine(
	rs *parser.Ruleset,
) string {
	var sb strings.Builder

	if rs.Selectors != nil {
		f.writeSelectorTo(&sb, rs.Selectors)
	}

	decls := rs.Declarations()

	if len(decls) == 0 && !hasNestedRules(rs) {
		sb.WriteString(" {}")
		return sb.String()
	}

	sb.WriteString(" { ")
	for i, decl := range decls {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(decl.Property.Value)
		sb.WriteString(": ")
		if decl.Value != nil {
			f.writeValueTo(&sb, decl.Value)
		}
		if decl.Important {
			sb.WriteString(" !important")
		}
		sb.WriteByte(';')
	}
	sb.WriteString(" }")
	return sb.String()
}

// writeSelectorTo writes the selector list to a builder,
// separating multiple selectors with ", ".
func (f *formatter) writeSelectorTo(
	sb *strings.Builder,
	sl *parser.SelectorList,
) {
	for i, sel := range sl.Selectors {
		if i > 0 {
			sb.WriteString(", ")
		}
		f.writeSingleSelectorTo(sb, sel)
	}
}

// writeSingleSelectorTo writes a single selector to a builder.
func (f *formatter) writeSingleSelectorTo(
	sb *strings.Builder,
	sel *parser.Selector,
) {
	for i, part := range sel.Parts {
		if part.Combinator != "" && part.Combinator != " " {
			sb.WriteByte(' ')
			sb.WriteString(part.Combinator)
			sb.WriteByte(' ')
		} else if i > 0 && part.Combinator == " " {
			sb.WriteByte(' ')
		}

		if part.Token.Kind != scanner.EOF {
			sb.WriteString(
				string(f.src[part.Token.Offset:part.Token.End]),
			)
		}
	}
}

// writeValueTo writes a value to a string builder.
func (f *formatter) writeValueTo(
	sb *strings.Builder,
	v *parser.Value,
) {
	prevEnd := -1
	for _, tok := range v.Tokens {
		if tok.Kind == scanner.Whitespace {
			continue
		}
		if prevEnd >= 0 && tok.Offset > prevEnd {
			if tok.Kind != scanner.ParenClose {
				sb.WriteByte(' ')
			}
		}
		sb.WriteString(
			string(f.src[tok.Offset:tok.End]),
		)
		prevEnd = tok.End
	}
}

// rulesetHasCommentInSource checks whether the source bytes
// between the opening brace and closing brace of a ruleset
// contain a CSS comment.
func (f *formatter) rulesetHasCommentInSource(
	rs *parser.Ruleset,
) bool {
	return bytes.Contains(
		f.src[rs.StartPos:rs.EndPos], []byte("/*"),
	)
}

// formatRulesetCompact tries to render the ruleset on a single
// line. Falls back to expanded if it doesn't fit within
// printWidth or if the ruleset contains comments.
func (f *formatter) formatRulesetCompact(
	rs *parser.Ruleset,
) {
	if hasNestedRules(rs) || f.rulesetHasCommentInSource(rs) {
		f.formatRuleset(rs)
		return
	}

	line := f.buildSingleLine(rs)
	totalWidth := f.indentWidth() + len(line)

	if totalWidth <= f.opts.PrintWidth {
		f.writeIndent()
		f.buf.WriteString(line)
		f.buf.WriteByte('\n')
		return
	}

	f.formatRuleset(rs)
}

// isOriginalSingleLine checks whether the source between
// startPos and endPos contains no newline characters.
func (f *formatter) isOriginalSingleLine(
	startPos, endPos int,
) bool {
	if startPos < 0 {
		startPos = 0
	}
	if endPos > len(f.src) {
		endPos = len(f.src)
	}
	return !bytes.ContainsRune(
		f.src[startPos:endPos], '\n',
	)
}

// formatRulesetPreserve keeps the original single-line vs
// multi-line structure, normalizing whitespace only.
func (f *formatter) formatRulesetPreserve(
	rs *parser.Ruleset,
) {
	if !hasNestedRules(rs) &&
		f.isOriginalSingleLine(rs.StartPos, rs.EndPos) {
		line := f.buildSingleLine(rs)
		f.writeIndent()
		f.buf.WriteString(line)
		f.buf.WriteByte('\n')
		return
	}

	f.formatRuleset(rs)
}
