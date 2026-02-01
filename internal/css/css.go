// Package css provides the public API for CSS parsing,
// analysis, hover, completion, and diagnostics.
package css

import (
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
		line, char := OffsetToLineChar(src, e.StartPos)
		endLine, endChar := OffsetToLineChar(src, e.EndPos)
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
	offset := LineCharToOffset(src, line, char)
	return analyzer.Hover(ss, src, offset)
}

// DocumentColors returns all colors found in the CSS document.
func DocumentColors(
	ss *parser.Stylesheet,
	src []byte,
) []analyzer.DocumentColor {
	return analyzer.FindDocumentColors(ss, src)
}

// ColorPresentations returns alternative representations of a
// color.
func ColorPresentations(
	c analyzer.Color,
) []string {
	return analyzer.ColorPresentation(c)
}

// SelectionRange returns the selection range at the given
// position.
func SelectionRange(
	ss *parser.Stylesheet,
	src []byte,
	line, char int,
) *analyzer.SelectionRange {
	offset := LineCharToOffset(src, line, char)
	return analyzer.FindSelectionRange(ss, offset)
}

// PrepareRename checks if rename is valid at the position.
func PrepareRename(
	ss *parser.Stylesheet,
	src []byte,
	line, char int,
) (analyzer.Location, bool) {
	offset := LineCharToOffset(src, line, char)
	return analyzer.PrepareRename(ss, src, offset)
}

// Rename renames a CSS custom property at the given position.
func Rename(
	ss *parser.Stylesheet,
	src []byte,
	line, char int,
	newName string,
) []analyzer.RenameEdit {
	offset := LineCharToOffset(src, line, char)
	return analyzer.Rename(ss, src, offset, newName)
}

// FormatDocument formats the CSS document.
func FormatDocument(
	ss *parser.Stylesheet,
	src []byte,
	tabSize int,
	insertSpaces bool,
) string {
	return analyzer.Format(ss, src, analyzer.FormatOptions{
		TabSize:      tabSize,
		InsertSpaces: insertSpaces,
	})
}

// FoldingRanges returns foldable ranges in the CSS document.
func FoldingRanges(
	ss *parser.Stylesheet,
	src []byte,
) []analyzer.FoldingRange {
	return analyzer.FindFoldingRanges(ss, src)
}

// DocumentLinks returns links found in the CSS document.
func DocumentLinks(
	ss *parser.Stylesheet,
	src []byte,
) []analyzer.DocumentLink {
	return analyzer.FindDocumentLinks(ss, src)
}

// DocumentHighlights returns highlights for the symbol at the
// given position.
func DocumentHighlights(
	ss *parser.Stylesheet,
	src []byte,
	line, char int,
) []analyzer.DocumentHighlight {
	offset := LineCharToOffset(src, line, char)
	return analyzer.FindDocumentHighlights(ss, src, offset)
}

// CodeActions returns code actions for the given diagnostics.
func CodeActions(
	diags []analyzer.Diagnostic,
	src []byte,
) []analyzer.CodeAction {
	return analyzer.FindCodeActions(diags, src)
}

// References finds all references to the symbol at the given
// position.
func References(
	ss *parser.Stylesheet,
	src []byte,
	line, char int,
) []analyzer.Location {
	offset := LineCharToOffset(src, line, char)
	return analyzer.FindReferences(ss, src, offset)
}

// Definition finds the definition of the symbol at the given
// position.
func Definition(
	ss *parser.Stylesheet,
	src []byte,
	line, char int,
) (loc analyzer.Location, found bool) {
	offset := LineCharToOffset(src, line, char)
	return analyzer.FindDefinition(ss, src, offset)
}

// DocumentSymbols returns a hierarchical list of symbols in
// the CSS document.
func DocumentSymbols(
	ss *parser.Stylesheet,
	src []byte,
) []analyzer.DocumentSymbol {
	return analyzer.FindDocumentSymbols(ss, src)
}

// Completions returns completion items for the given position.
func Completions(
	ss *parser.Stylesheet,
	src []byte,
	line, char int,
) []analyzer.CompletionItem {
	offset := LineCharToOffset(src, line, char)
	return analyzer.Complete(ss, src, offset)
}

// OffsetToLineChar converts a byte offset to line/character.
func OffsetToLineChar(src []byte, offset int) (int, int) {
	return analyzer.OffsetToLineChar(src, offset)
}

// LineCharToOffset converts line/character to byte offset.
func LineCharToOffset(src []byte, line, char int) int {
	return analyzer.LineCharToOffset(src, line, char)
}
