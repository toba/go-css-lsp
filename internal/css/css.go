// Package css provides the public API for CSS parsing,
// analysis, hover, completion, and diagnostics.
package css

import (
	"strings"

	"github.com/toba/go-css-lsp/internal/css/analyzer"
	"github.com/toba/go-css-lsp/internal/css/parser"
)

// ParseResult holds the parsed AST and any errors.
type ParseResult struct {
	Stylesheet *parser.Stylesheet
	Errors     []*parser.Error
}

// Parse parses CSS source into an AST.
func Parse(src []byte) *ParseResult {
	ss, errs := parser.Parse(src)
	return &ParseResult{Stylesheet: ss, Errors: errs}
}

// Diagnostics returns diagnostic messages for the given CSS.
func Diagnostics(
	src []byte,
) ([]analyzer.Diagnostic, *parser.Stylesheet) {
	result := Parse(src)
	diags := analyzer.Analyze(result.Stylesheet, src)

	// Add parse errors as diagnostics
	for _, e := range result.Errors {
		line, char := offsetToLineChar(src, e.StartPos)
		endLine, endChar := offsetToLineChar(src, e.EndPos)
		diags = append(diags, analyzer.Diagnostic{
			Message:   e.Message,
			StartLine: line,
			StartChar: char,
			EndLine:   endLine,
			EndChar:   endChar,
			Severity:  analyzer.SeverityError,
		})
	}

	return diags, result.Stylesheet
}

// Hover returns hover information for the given position.
func Hover(
	ss *parser.Stylesheet,
	src []byte,
	line, char int,
) (content string, found bool) {
	offset := lineCharToOffset(src, line, char)
	return analyzer.Hover(ss, src, offset)
}

// Completions returns completion items for the given position.
func Completions(
	ss *parser.Stylesheet,
	src []byte,
	line, char int,
) []analyzer.CompletionItem {
	offset := lineCharToOffset(src, line, char)
	return analyzer.Complete(ss, src, offset)
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

// lineCharToOffset converts line/character to byte offset.
func lineCharToOffset(src []byte, line, char int) int {
	text := string(src)
	lines := strings.SplitAfter(text, "\n")
	offset := 0
	for i, l := range lines {
		if i == line {
			offset += char
			break
		}
		offset += len(l)
	}
	if offset > len(src) {
		offset = len(src)
	}
	return offset
}
