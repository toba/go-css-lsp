// Package scanner implements a CSS tokenizer per W3C CSS Syntax
// Module Level 3.
package scanner

// Scanner tokenizes CSS source text.
type Scanner struct {
	src []byte
	pos int
}

// New creates a Scanner for the given CSS source.
func New(src []byte) *Scanner {
	return &Scanner{src: src}
}

// ScanAll returns all tokens from the source.
func ScanAll(src []byte) []Token {
	s := New(src)
	var tokens []Token
	for {
		t := s.Next()
		tokens = append(tokens, t)
		if t.Kind == EOF {
			break
		}
	}
	return tokens
}

// Next returns the next token.
func (s *Scanner) Next() Token {
	if s.pos >= len(s.src) {
		return Token{Kind: EOF, Offset: s.pos, End: s.pos}
	}

	ch := s.src[s.pos]

	// Whitespace
	if isWhitespace(ch) {
		return s.scanWhitespace()
	}

	// Comments
	if ch == '/' && s.pos+1 < len(s.src) &&
		s.src[s.pos+1] == '*' {
		return s.scanComment()
	}

	// Strings
	if ch == '"' || ch == '\'' {
		return s.scanString(ch)
	}

	// Numbers
	if isDigit(ch) ||
		(ch == '.' && s.pos+1 < len(s.src) &&
			isDigit(s.src[s.pos+1])) {
		return s.scanNumeric()
	}

	// +/- can start a number
	if (ch == '+' || ch == '-') && s.pos+1 < len(s.src) {
		next := s.src[s.pos+1]
		if isDigit(next) ||
			(next == '.' && s.pos+2 < len(s.src) &&
				isDigit(s.src[s.pos+2])) {
			return s.scanNumeric()
		}
	}

	// Hash
	if ch == '#' {
		start := s.pos
		s.pos++
		if s.pos < len(s.src) && isNameChar(s.src[s.pos]) {
			for s.pos < len(s.src) &&
				isNameChar(s.src[s.pos]) {
				s.pos++
			}
			return Token{
				Kind:   Hash,
				Offset: start,
				End:    s.pos,
				Value:  string(s.src[start+1 : s.pos]),
			}
		}
		return Token{
			Kind:   Delim,
			Offset: start,
			End:    s.pos,
			Value:  "#",
		}
	}

	// At-keyword
	if ch == '@' {
		start := s.pos
		s.pos++
		if s.pos < len(s.src) && isNameStart(s.src[s.pos]) {
			for s.pos < len(s.src) &&
				isNameChar(s.src[s.pos]) {
				s.pos++
			}
			return Token{
				Kind:   AtKeyword,
				Offset: start,
				End:    s.pos,
				Value:  string(s.src[start+1 : s.pos]),
			}
		}
		return Token{
			Kind:   Delim,
			Offset: start,
			End:    s.pos,
			Value:  "@",
		}
	}

	// Ident or Function or URL
	if isNameStart(ch) || ch == '-' {
		return s.scanIdentLike()
	}

	// CDO <!--
	if ch == '<' && s.pos+3 < len(s.src) &&
		s.src[s.pos+1] == '!' &&
		s.src[s.pos+2] == '-' &&
		s.src[s.pos+3] == '-' {
		t := Token{
			Kind:   CDO,
			Offset: s.pos,
			End:    s.pos + 4,
		}
		s.pos += 4
		return t
	}

	// CDC -->
	if ch == '-' && s.pos+2 < len(s.src) &&
		s.src[s.pos+1] == '-' &&
		s.src[s.pos+2] == '>' {
		t := Token{
			Kind:   CDC,
			Offset: s.pos,
			End:    s.pos + 3,
		}
		s.pos += 3
		return t
	}

	// Single-character tokens
	start := s.pos
	s.pos++

	switch ch {
	case ':':
		return Token{
			Kind: Colon, Offset: start, End: s.pos,
		}
	case ';':
		return Token{
			Kind: Semicolon, Offset: start, End: s.pos,
		}
	case ',':
		return Token{
			Kind: Comma, Offset: start, End: s.pos,
		}
	case '{':
		return Token{
			Kind: BraceOpen, Offset: start, End: s.pos,
		}
	case '}':
		return Token{
			Kind: BraceClose, Offset: start, End: s.pos,
		}
	case '(':
		return Token{
			Kind: ParenOpen, Offset: start, End: s.pos,
		}
	case ')':
		return Token{
			Kind: ParenClose, Offset: start, End: s.pos,
		}
	case '[':
		return Token{
			Kind: BracketOpen, Offset: start, End: s.pos,
		}
	case ']':
		return Token{
			Kind: BracketClose, Offset: start, End: s.pos,
		}
	default:
		return Token{
			Kind:   Delim,
			Offset: start,
			End:    s.pos,
			Value:  string(s.src[start:s.pos]),
		}
	}
}

func (s *Scanner) scanWhitespace() Token {
	start := s.pos
	for s.pos < len(s.src) && isWhitespace(s.src[s.pos]) {
		s.pos++
	}
	return Token{Kind: Whitespace, Offset: start, End: s.pos}
}

func (s *Scanner) scanComment() Token {
	start := s.pos
	s.pos += 2 // skip /*
	for s.pos+1 < len(s.src) {
		if s.src[s.pos] == '*' && s.src[s.pos+1] == '/' {
			s.pos += 2
			return Token{
				Kind:   Comment,
				Offset: start,
				End:    s.pos,
				Value: string(
					s.src[start+2 : s.pos-2],
				),
			}
		}
		s.pos++
	}
	// Unterminated comment
	s.pos = len(s.src)
	return Token{
		Kind:   Comment,
		Offset: start,
		End:    s.pos,
		Value:  string(s.src[start+2:]),
	}
}

func (s *Scanner) scanString(quote byte) Token {
	start := s.pos
	s.pos++ // skip opening quote
	for s.pos < len(s.src) {
		ch := s.src[s.pos]
		if ch == quote {
			s.pos++
			return Token{
				Kind:   String,
				Offset: start,
				End:    s.pos,
				Value: string(
					s.src[start+1 : s.pos-1],
				),
			}
		}
		if ch == '\\' && s.pos+1 < len(s.src) {
			s.pos += 2 // skip escaped char
			continue
		}
		if ch == '\n' {
			// Unescaped newline in string = bad string
			return Token{
				Kind:   BadString,
				Offset: start,
				End:    s.pos,
				Value: string(
					s.src[start+1 : s.pos],
				),
			}
		}
		s.pos++
	}
	// Unterminated string
	return Token{
		Kind:   BadString,
		Offset: start,
		End:    s.pos,
		Value:  string(s.src[start+1:]),
	}
}

func (s *Scanner) scanNumeric() Token {
	start := s.pos

	// Optional sign
	if s.pos < len(s.src) &&
		(s.src[s.pos] == '+' || s.src[s.pos] == '-') {
		s.pos++
	}

	// Integer part
	for s.pos < len(s.src) && isDigit(s.src[s.pos]) {
		s.pos++
	}

	// Decimal part
	if s.pos+1 < len(s.src) && s.src[s.pos] == '.' &&
		isDigit(s.src[s.pos+1]) {
		s.pos++ // skip '.'
		for s.pos < len(s.src) && isDigit(s.src[s.pos]) {
			s.pos++
		}
	}

	numEnd := s.pos
	numValue := string(s.src[start:numEnd])

	// Percentage
	if s.pos < len(s.src) && s.src[s.pos] == '%' {
		s.pos++
		return Token{
			Kind:   Percentage,
			Offset: start,
			End:    s.pos,
			Value:  numValue,
		}
	}

	// Dimension (number followed by ident)
	if s.pos < len(s.src) && isNameStart(s.src[s.pos]) {
		unitStart := s.pos
		for s.pos < len(s.src) && isNameChar(s.src[s.pos]) {
			s.pos++
		}
		return Token{
			Kind:   Dimension,
			Offset: start,
			End:    s.pos,
			Value: numValue + string(
				s.src[unitStart:s.pos],
			),
		}
	}

	return Token{
		Kind:   Number,
		Offset: start,
		End:    numEnd,
		Value:  numValue,
	}
}

func (s *Scanner) scanIdentLike() Token {
	start := s.pos
	s.consumeName()
	name := string(s.src[start:s.pos])

	// Check for function token: ident followed by (
	if s.pos < len(s.src) && s.src[s.pos] == '(' {
		// Special case: url(
		if len(name) == 3 &&
			(name[0] == 'u' || name[0] == 'U') &&
			(name[1] == 'r' || name[1] == 'R') &&
			(name[2] == 'l' || name[2] == 'L') {
			s.pos++ // skip (
			return s.scanURLOrFunction(start, name)
		}
		s.pos++ // skip (
		return Token{
			Kind:   Function,
			Offset: start,
			End:    s.pos,
			Value:  name,
		}
	}

	return Token{
		Kind:   Ident,
		Offset: start,
		End:    s.pos,
		Value:  name,
	}
}

func (s *Scanner) scanURLOrFunction(
	start int,
	name string,
) Token {
	// Skip whitespace after url(
	wsStart := s.pos
	for s.pos < len(s.src) && isWhitespace(s.src[s.pos]) {
		s.pos++
	}

	// If next char is quote, it's a regular function call
	if s.pos < len(s.src) &&
		(s.src[s.pos] == '"' || s.src[s.pos] == '\'') {
		s.pos = wsStart
		return Token{
			Kind:   Function,
			Offset: start,
			End:    s.pos,
			Value:  name,
		}
	}

	// Otherwise consume URL token value
	urlStart := s.pos
	for s.pos < len(s.src) {
		ch := s.src[s.pos]
		if ch == ')' {
			urlValue := string(s.src[urlStart:s.pos])
			s.pos++ // skip )
			return Token{
				Kind:   URL,
				Offset: start,
				End:    s.pos,
				Value:  urlValue,
			}
		}
		if isWhitespace(ch) {
			urlValue := string(s.src[urlStart:s.pos])
			// Skip trailing whitespace
			for s.pos < len(s.src) &&
				isWhitespace(s.src[s.pos]) {
				s.pos++
			}
			if s.pos < len(s.src) && s.src[s.pos] == ')' {
				s.pos++
				return Token{
					Kind:   URL,
					Offset: start,
					End:    s.pos,
					Value:  urlValue,
				}
			}
			// Bad URL
			for s.pos < len(s.src) &&
				s.src[s.pos] != ')' {
				s.pos++
			}
			if s.pos < len(s.src) {
				s.pos++
			}
			return Token{
				Kind:   BadURL,
				Offset: start,
				End:    s.pos,
			}
		}
		if ch == '\\' && s.pos+1 < len(s.src) {
			s.pos += 2
			continue
		}
		if ch == '(' || ch == '"' || ch == '\'' {
			// Bad URL
			for s.pos < len(s.src) &&
				s.src[s.pos] != ')' {
				s.pos++
			}
			if s.pos < len(s.src) {
				s.pos++
			}
			return Token{
				Kind:   BadURL,
				Offset: start,
				End:    s.pos,
			}
		}
		s.pos++
	}

	return Token{
		Kind:   BadURL,
		Offset: start,
		End:    s.pos,
	}
}

func (s *Scanner) consumeName() {
	for s.pos < len(s.src) && isNameChar(s.src[s.pos]) {
		if s.src[s.pos] == '\\' && s.pos+1 < len(s.src) {
			s.pos += 2
			continue
		}
		s.pos++
	}
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' ||
		ch == '\n' || ch == '\r' || ch == '\f'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isNameStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		ch == '_' || ch >= 0x80
}

func isNameChar(ch byte) bool {
	return isNameStart(ch) || isDigit(ch) || ch == '-'
}
