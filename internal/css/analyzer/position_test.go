package analyzer

import "testing"

func TestOffsetToLineChar(t *testing.T) {
	tests := []struct {
		src      string
		offset   int
		wantLine int
		wantChar int
	}{
		{"hello", 0, 0, 0},
		{"hello", 3, 0, 3},
		{"hello\nworld", 6, 1, 0},
		{"hello\nworld", 8, 1, 2},
		{"a\nb\nc", 4, 2, 0},
		{"", 0, 0, 0},
	}
	for _, tt := range tests {
		line, char := OffsetToLineChar(
			[]byte(tt.src), tt.offset,
		)
		if line != tt.wantLine || char != tt.wantChar {
			t.Errorf(
				"OffsetToLineChar(%q, %d) = (%d, %d), "+
					"want (%d, %d)",
				tt.src, tt.offset,
				line, char,
				tt.wantLine, tt.wantChar,
			)
		}
	}
}

func TestLineCharToOffset(t *testing.T) {
	tests := []struct {
		src     string
		line    int
		char    int
		wantOff int
	}{
		{"hello", 0, 0, 0},
		{"hello", 0, 3, 3},
		{"hello\nworld", 1, 0, 6},
		{"hello\nworld", 1, 2, 8},
		{"a\nb\nc", 2, 0, 4},
	}
	for _, tt := range tests {
		off := LineCharToOffset(
			[]byte(tt.src), tt.line, tt.char,
		)
		if off != tt.wantOff {
			t.Errorf(
				"LineCharToOffset(%q, %d, %d) = %d, "+
					"want %d",
				tt.src, tt.line, tt.char,
				off, tt.wantOff,
			)
		}
	}
}

func TestOffsetToLineChar_RoundTrip(t *testing.T) {
	src := []byte("hello\nworld\nfoo bar\n")
	offsets := []int{0, 3, 5, 6, 10, 12, 18}

	for _, off := range offsets {
		line, char := OffsetToLineChar(src, off)
		got := LineCharToOffset(src, line, char)
		if got != off {
			t.Errorf(
				"round-trip failed: offset %d -> "+
					"(%d,%d) -> %d",
				off, line, char, got,
			)
		}
	}
}

func TestOffsetToLineChar_BeyondEnd(t *testing.T) {
	src := []byte("ab")
	line, char := OffsetToLineChar(src, 10)
	// Should not panic; clamps at end
	if line < 0 || char < 0 {
		t.Error("negative line/char for beyond-end offset")
	}
}
