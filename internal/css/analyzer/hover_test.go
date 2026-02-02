package analyzer

import (
	"strings"
	"testing"

	"github.com/toba/go-css-lsp/internal/css/parser"
)

func TestHoverProperty(t *testing.T) {
	src := []byte(`body { color: red; }`)
	ss, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	// "color" is at bytes 7-12
	hr := Hover(ss, src, 8)
	if !hr.Found {
		t.Fatal("expected hover to find content")
	}
	if !strings.Contains(hr.Content, "**color**") {
		t.Errorf(
			"expected property name, got %q", hr.Content,
		)
	}
}

func TestHoverUnknownProperty(t *testing.T) {
	src := []byte(`body { foobar: red; }`)
	ss, _ := parser.Parse(src)

	hr := Hover(ss, src, 8)
	if hr.Found {
		t.Error(
			"expected hover not to find unknown property",
		)
	}
}

func TestHoverFunctionSignatures(t *testing.T) {
	src := []byte(`body { color: rgb(255, 0, 0); }`)
	ss, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	// "rgb" function token starts at byte 14
	hr := Hover(ss, src, 14)
	if !hr.Found {
		t.Fatal("expected hover to find rgb function")
	}
	if !strings.Contains(hr.Content, "rgb(") {
		t.Errorf(
			"expected signature lines, got %q", hr.Content,
		)
	}
	if !strings.Contains(hr.Content, "```") {
		t.Errorf(
			"expected code block, got %q", hr.Content,
		)
	}
	if !strings.Contains(hr.Content, "MDN Reference") {
		t.Errorf(
			"expected MDN link, got %q", hr.Content,
		)
	}
	if !strings.Contains(hr.Content, "red, green, blue") {
		t.Errorf(
			"expected description, got %q", hr.Content,
		)
	}
}

func TestHoverFunctionCalc(t *testing.T) {
	src := []byte(`body { width: calc(100% - 20px); }`)
	ss, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	// "calc" function token starts at byte 14
	hr := Hover(ss, src, 14)
	if !hr.Found {
		t.Fatal("expected hover to find calc function")
	}
	if !strings.Contains(hr.Content, "calc(<expression>)") {
		t.Errorf(
			"expected calc signature, got %q", hr.Content,
		)
	}
	if !strings.Contains(hr.Content, "MDN Reference") {
		t.Errorf(
			"expected MDN link, got %q", hr.Content,
		)
	}
}

func TestHoverFunctionUnknown(t *testing.T) {
	src := []byte(
		`body { color: notafunction(1, 2); }`,
	)
	ss, _ := parser.Parse(src)

	hr := Hover(ss, src, 14)
	if hr.Found {
		t.Error(
			"expected hover not to find unknown function",
		)
	}
}

func TestHoverNoPanic(t *testing.T) {
	src := []byte(`body { color: red; }`)
	ss, _ := parser.Parse(src)

	// Just ensure no panic at various offsets
	for i := range src {
		Hover(ss, src, i)
	}
}

func TestHoverCustomPropertyDeclaration(t *testing.T) {
	src := []byte(`:root { --color-link-icon: #0366d6; }`)
	ss, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	// "--color-link-icon" starts at byte 8
	hr := Hover(ss, src, 10)
	if !hr.Found {
		t.Fatal(
			"expected hover for custom property declaration",
		)
	}
	if !strings.Contains(hr.Content, "--color-link-icon") {
		t.Errorf(
			"expected property name, got %q", hr.Content,
		)
	}
	// Range should cover the property name token
	propName := "--color-link-icon"
	start := strings.Index(string(src), propName)
	end := start + len(propName)
	if hr.RangeStart != start || hr.RangeEnd != end {
		t.Errorf(
			"range = [%d,%d], want [%d,%d]",
			hr.RangeStart, hr.RangeEnd, start, end,
		)
	}
}

func TestHoverVarReference(t *testing.T) {
	src := []byte(
		`:root { --brand: blue; }
.a { color: var(--brand); }`,
	)
	ss, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	// Hover on "--brand" inside var()
	varExpr := "var(--brand)"
	varIdx := strings.Index(string(src), varExpr)
	identIdx := strings.Index(
		string(src[varIdx:]), "--brand",
	) + varIdx

	hr := Hover(ss, src, identIdx+1)
	if !hr.Found {
		t.Fatal("expected hover for var() reference")
	}
	if !strings.Contains(hr.Content, "--brand") {
		t.Errorf(
			"expected property name, got %q", hr.Content,
		)
	}
	// Range should cover entire var(--brand)
	if hr.RangeStart != varIdx ||
		hr.RangeEnd != varIdx+len(varExpr) {
		t.Errorf(
			"range = [%d,%d], want [%d,%d]",
			hr.RangeStart, hr.RangeEnd,
			varIdx, varIdx+len(varExpr),
		)
	}
}

func TestHoverVarFunctionToken(t *testing.T) {
	src := []byte(
		`:root { --brand: blue; }
.a { color: var(--brand); }`,
	)
	ss, errs := parser.Parse(src)
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	// Hover on "var" token itself
	varExpr := "var(--brand)"
	varIdx := strings.Index(string(src), varExpr)

	hr := Hover(ss, src, varIdx+1)
	if !hr.Found {
		t.Fatal(
			"expected hover for var() function token",
		)
	}
	if !strings.Contains(hr.Content, "--brand") {
		t.Errorf(
			"expected property name, got %q", hr.Content,
		)
	}
	// Range should cover entire var(--brand)
	if hr.RangeStart != varIdx ||
		hr.RangeEnd != varIdx+len(varExpr) {
		t.Errorf(
			"range = [%d,%d], want [%d,%d]",
			hr.RangeStart, hr.RangeEnd,
			varIdx, varIdx+len(varExpr),
		)
	}
}
