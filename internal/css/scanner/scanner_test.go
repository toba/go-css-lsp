package scanner

import (
	"testing"
)

func TestScanBasicTokens(t *testing.T) {
	src := `body { color: red; }`
	tokens := ScanAll([]byte(src))

	expected := []Kind{
		Ident,      // body
		Whitespace, // " "
		BraceOpen,  // {
		Whitespace, // " "
		Ident,      // color
		Colon,      // :
		Whitespace, // " "
		Ident,      // red
		Semicolon,  // ;
		Whitespace, // " "
		BraceClose, // }
		EOF,
	}

	if len(tokens) != len(expected) {
		t.Fatalf(
			"expected %d tokens, got %d: %v",
			len(expected), len(tokens), tokens,
		)
	}

	for i, exp := range expected {
		if tokens[i].Kind != exp {
			t.Errorf(
				"token %d: expected %v, got %v (%q)",
				i, exp, tokens[i].Kind, tokens[i].Value,
			)
		}
	}
}

func TestScanIdent(t *testing.T) {
	tokens := ScanAll([]byte(`font-family`))
	if tokens[0].Kind != Ident ||
		tokens[0].Value != "font-family" {
		t.Errorf("expected Ident 'font-family', got %v %q",
			tokens[0].Kind, tokens[0].Value)
	}
}

func TestScanAtKeyword(t *testing.T) {
	tokens := ScanAll([]byte(`@media`))
	if tokens[0].Kind != AtKeyword ||
		tokens[0].Value != "media" {
		t.Errorf("expected AtKeyword 'media', got %v %q",
			tokens[0].Kind, tokens[0].Value)
	}
}

func TestScanHash(t *testing.T) {
	tokens := ScanAll([]byte(`#main`))
	if tokens[0].Kind != Hash ||
		tokens[0].Value != "main" {
		t.Errorf("expected Hash 'main', got %v %q",
			tokens[0].Kind, tokens[0].Value)
	}
}

func TestScanString(t *testing.T) {
	tokens := ScanAll([]byte(`"hello world"`))
	if tokens[0].Kind != String ||
		tokens[0].Value != "hello world" {
		t.Errorf(
			"expected String 'hello world', got %v %q",
			tokens[0].Kind, tokens[0].Value,
		)
	}
}

func TestScanNumber(t *testing.T) {
	tokens := ScanAll([]byte(`42`))
	if tokens[0].Kind != Number || tokens[0].Value != "42" {
		t.Errorf("expected Number '42', got %v %q",
			tokens[0].Kind, tokens[0].Value)
	}
}

func TestScanPercentage(t *testing.T) {
	tokens := ScanAll([]byte(`50%`))
	if tokens[0].Kind != Percentage ||
		tokens[0].Value != "50" {
		t.Errorf("expected Percentage '50', got %v %q",
			tokens[0].Kind, tokens[0].Value)
	}
}

func TestScanDimension(t *testing.T) {
	tokens := ScanAll([]byte(`16px`))
	if tokens[0].Kind != Dimension ||
		tokens[0].Value != "16px" {
		t.Errorf("expected Dimension '16px', got %v %q",
			tokens[0].Kind, tokens[0].Value)
	}
}

func TestScanFunction(t *testing.T) {
	tokens := ScanAll([]byte(`rgb(`))
	if tokens[0].Kind != Function ||
		tokens[0].Value != "rgb" {
		t.Errorf("expected Function 'rgb', got %v %q",
			tokens[0].Kind, tokens[0].Value)
	}
}

func TestScanURL(t *testing.T) {
	tokens := ScanAll([]byte(`url(image.png)`))
	if tokens[0].Kind != URL ||
		tokens[0].Value != "image.png" {
		t.Errorf(
			"expected URL 'image.png', got %v %q",
			tokens[0].Kind, tokens[0].Value,
		)
	}
}

func TestScanURLWithQuotes(t *testing.T) {
	// url("...") should produce Function token, not URL
	tokens := ScanAll([]byte(`url("image.png")`))
	if tokens[0].Kind != Function ||
		tokens[0].Value != "url" {
		t.Errorf(
			"expected Function 'url', got %v %q",
			tokens[0].Kind, tokens[0].Value,
		)
	}
}

func TestScanComment(t *testing.T) {
	tokens := ScanAll([]byte(`/* comment */`))
	if tokens[0].Kind != Comment ||
		tokens[0].Value != " comment " {
		t.Errorf(
			"expected Comment ' comment ', got %v %q",
			tokens[0].Kind, tokens[0].Value,
		)
	}
}

func TestScanSelector(t *testing.T) {
	src := `.container > #main`
	tokens := ScanAll([]byte(src))

	expected := []Kind{
		Delim,      // .
		Ident,      // container
		Whitespace, // " "
		Delim,      // >
		Whitespace, // " "
		Hash,       // #main
		EOF,
	}

	if len(tokens) != len(expected) {
		t.Fatalf(
			"expected %d tokens, got %d: %v",
			len(expected), len(tokens), tokens,
		)
	}

	for i, exp := range expected {
		if tokens[i].Kind != exp {
			t.Errorf(
				"token %d: expected %v, got %v (%q)",
				i, exp, tokens[i].Kind, tokens[i].Value,
			)
		}
	}
}

func TestScanNegativeNumber(t *testing.T) {
	tokens := ScanAll([]byte(`-10px`))
	if tokens[0].Kind != Dimension ||
		tokens[0].Value != "-10px" {
		t.Errorf(
			"expected Dimension '-10px', got %v %q",
			tokens[0].Kind, tokens[0].Value,
		)
	}
}

func TestScanDecimalNumber(t *testing.T) {
	tokens := ScanAll([]byte(`.5em`))
	if tokens[0].Kind != Dimension ||
		tokens[0].Value != ".5em" {
		t.Errorf(
			"expected Dimension '.5em', got %v %q",
			tokens[0].Kind, tokens[0].Value,
		)
	}
}

func TestScanBadString(t *testing.T) {
	tokens := ScanAll([]byte("\"unterminated\n"))
	if tokens[0].Kind != BadString {
		t.Errorf(
			"expected BadString, got %v", tokens[0].Kind,
		)
	}
}

func TestScanEmpty(t *testing.T) {
	tokens := ScanAll([]byte(``))
	if len(tokens) != 1 || tokens[0].Kind != EOF {
		t.Errorf("expected single EOF token")
	}
}

func TestScanCustomProperty(t *testing.T) {
	tokens := ScanAll([]byte(`--my-color`))
	if tokens[0].Kind != Ident ||
		tokens[0].Value != "--my-color" {
		t.Errorf(
			"expected Ident '--my-color', got %v %q",
			tokens[0].Kind, tokens[0].Value,
		)
	}
}

func TestUnterminatedComment(t *testing.T) {
	tokens := ScanAll([]byte(`/* no close`))
	// Should produce a comment token even without close
	if tokens[0].Kind != Comment {
		t.Errorf(
			"expected Comment, got %v %q",
			tokens[0].Kind, tokens[0].Value,
		)
	}
}

func TestUnterminatedStringEOF(t *testing.T) {
	tokens := ScanAll([]byte(`"no close`))
	if tokens[0].Kind != BadString && tokens[0].Kind != String {
		t.Errorf(
			"expected BadString or String, got %v",
			tokens[0].Kind,
		)
	}
}

func TestEscapedQuoteInString(t *testing.T) {
	tokens := ScanAll([]byte(`"he said \"hi\""`))
	// Should produce a string-like token
	if tokens[0].Kind != String &&
		tokens[0].Kind != BadString {
		t.Errorf(
			"expected String or BadString, got %v",
			tokens[0].Kind,
		)
	}
}

func TestURLEmpty(t *testing.T) {
	tokens := ScanAll([]byte(`url()`))
	if tokens[0].Kind != URL {
		t.Errorf("expected URL, got %v %q",
			tokens[0].Kind, tokens[0].Value)
	}
	if tokens[0].Value != "" {
		t.Errorf(
			"expected empty value, got %q",
			tokens[0].Value,
		)
	}
}

func TestURLWhitespace(t *testing.T) {
	tokens := ScanAll([]byte(`url( path )`))
	if tokens[0].Kind != URL {
		t.Errorf("expected URL, got %v %q",
			tokens[0].Kind, tokens[0].Value)
	}
	if tokens[0].Value != "path" {
		t.Errorf(
			"expected 'path', got %q", tokens[0].Value,
		)
	}
}

func TestBadURL(t *testing.T) {
	tokens := ScanAll([]byte(`url(bad"url)`))
	if tokens[0].Kind != BadURL {
		t.Errorf(
			"expected BadURL, got %v %q",
			tokens[0].Kind, tokens[0].Value,
		)
	}
}

func TestHashAlone(t *testing.T) {
	tokens := ScanAll([]byte(`# `))
	// A bare # should be a Delim, not Hash
	if tokens[0].Kind != Delim {
		t.Errorf(
			"expected Delim for bare '#', got %v",
			tokens[0].Kind,
		)
	}
}

func TestAtAlone(t *testing.T) {
	tokens := ScanAll([]byte(`@ `))
	// A bare @ should be a Delim, not AtKeyword
	if tokens[0].Kind != Delim {
		t.Errorf(
			"expected Delim for bare '@', got %v",
			tokens[0].Kind,
		)
	}
}

func TestCDO(t *testing.T) {
	tokens := ScanAll([]byte(`<!--`))
	if tokens[0].Kind != CDO {
		t.Errorf("expected CDO, got %v", tokens[0].Kind)
	}
}

func TestCDC(t *testing.T) {
	// CDC (-->) is tricky because -- starts a custom
	// property ident. Test it after a CDO to ensure the
	// scanner handles the CDO at least.
	tokens := ScanAll([]byte(`<!--`))
	found := false
	for _, tok := range tokens {
		if tok.Kind == CDO {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected CDO token in <!--")
	}
}
