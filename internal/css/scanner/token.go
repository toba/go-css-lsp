package scanner

//go:generate stringer -type=Kind
type Kind int

const (
	EOF Kind = iota
	Whitespace
	Comment
	Ident
	AtKeyword // @foo
	Hash      // #foo
	String    // "..." or '...'
	BadString
	Number
	Percentage // 42%
	Dimension  // 42px
	URL        // url(...)
	BadURL
	Function // foo(
	UnicodeRange
	CDO // <!--
	CDC // -->
	Colon
	Semicolon
	Comma
	BraceOpen    // {
	BraceClose   // }
	ParenOpen    // (
	ParenClose   // )
	BracketOpen  // [
	BracketClose // ]
	Delim        // any other single character
)

// Token represents a CSS token with its kind, byte range, and
// value.
type Token struct {
	Kind   Kind
	Offset int    // byte offset in source
	End    int    // byte offset of end (exclusive)
	Value  string // semantic value (ident name, string content, etc.)
}

// Len returns the byte length of the token.
func (t Token) Len() int {
	return t.End - t.Offset
}
