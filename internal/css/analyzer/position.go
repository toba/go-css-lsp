package analyzer

import "github.com/toba/lsp/position"

// OffsetToLineChar converts a byte offset to line/character.
func OffsetToLineChar(src []byte, offset int) (int, int) {
	return position.OffsetToLineCol(src, offset)
}

// LineCharToOffset converts line/character to byte offset.
func LineCharToOffset(src []byte, line, char int) int {
	return position.LineColToOffset(src, line, char)
}
