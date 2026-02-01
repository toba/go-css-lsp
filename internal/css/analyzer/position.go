package analyzer

import "strings"

// OffsetToLineChar converts a byte offset to line/character.
func OffsetToLineChar(src []byte, offset int) (int, int) {
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

// LineCharToOffset converts line/character to byte offset.
func LineCharToOffset(src []byte, line, char int) int {
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
