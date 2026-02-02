package analyzer

import (
	"math"
	"strings"
	"testing"

	"github.com/toba/go-css-lsp/internal/css/parser"
	"github.com/toba/go-css-lsp/internal/css/scanner"
)

func TestFindDocumentColors_HexColors(t *testing.T) {
	src := []byte(`
.foo {
  color: #ff0000;
  background: #0f0;
  border-color: #00ff00ff;
  outline-color: #abcd;
}`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 4 {
		t.Fatalf("expected 4 colors, got %d", len(colors))
	}

	// #ff0000 -> red
	assertColorClose(t, colors[0].Color, 1.0, 0.0, 0.0, 1.0)
	// #0f0 -> green
	assertColorClose(t, colors[1].Color, 0.0, 1.0, 0.0, 1.0)
	// #00ff00ff -> green, full alpha
	assertColorClose(t, colors[2].Color, 0.0, 1.0, 0.0, 1.0)
	// #abcd -> short rgba
	assertColorClose(t, colors[3].Color,
		float64(0xaa)/255.0,
		float64(0xbb)/255.0,
		float64(0xcc)/255.0,
		float64(0xdd)/255.0,
	)
}

func TestFindDocumentColors_NamedColors(t *testing.T) {
	src := []byte(`.foo { color: red; background: blue; }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 2 {
		t.Fatalf("expected 2 colors, got %d", len(colors))
	}

	assertColorClose(t, colors[0].Color, 1.0, 0.0, 0.0, 1.0)
	assertColorClose(t, colors[1].Color, 0.0, 0.0, 1.0, 1.0)
}

func TestFindDocumentColors_Transparent(t *testing.T) {
	src := []byte(`.foo { color: transparent; }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}

	assertColorClose(t, colors[0].Color, 0.0, 0.0, 0.0, 0.0)
}

func TestFindDocumentColors_RGBFunction(t *testing.T) {
	src := []byte(`.foo { color: rgb(255, 128, 0); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}

	assertColorClose(t, colors[0].Color, 1.0, 128.0/255.0, 0.0, 1.0)
}

func TestFindDocumentColors_RGBModernSyntax(t *testing.T) {
	src := []byte(`.foo { color: rgb(255 128 0 / 50%); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}

	assertColorClose(t, colors[0].Color, 1.0, 128.0/255.0, 0.0, 0.5)
}

func TestFindDocumentColors_HSLFunction(t *testing.T) {
	src := []byte(`.foo { color: hsl(0, 100%, 50%); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}

	// hsl(0, 100%, 50%) = pure red
	assertColorClose(t, colors[0].Color, 1.0, 0.0, 0.0, 1.0)
}

func TestFindDocumentColors_HSLBlue(t *testing.T) {
	src := []byte(`.foo { color: hsl(240, 100%, 50%); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}

	// hsl(240, 100%, 50%) = pure blue
	assertColorClose(t, colors[0].Color, 0.0, 0.0, 1.0, 1.0)
}

func TestFindDocumentColors_HWBFunction(t *testing.T) {
	src := []byte(`.foo { color: hwb(0 0% 0%); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}

	// hwb(0 0% 0%) = pure red
	assertColorClose(t, colors[0].Color, 1.0, 0.0, 0.0, 1.0)
}

func TestFindDocumentColors_NoColors(t *testing.T) {
	src := []byte(`.foo { display: block; margin: 10px; }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 0 {
		t.Fatalf("expected 0 colors, got %d", len(colors))
	}
}

func TestFindDocumentColors_MultipleDeclarations(t *testing.T) {
	src := []byte(`
.foo {
  color: red;
  background-color: #00ff00;
  border: 1px solid rgb(0, 0, 255);
}`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 3 {
		t.Fatalf("expected 3 colors, got %d", len(colors))
	}
}

func TestColorPresentation(t *testing.T) {
	c := Color{Red: 1.0, Green: 0.0, Blue: 0.0, Alpha: 1.0}
	presentations := ColorPresentation(c)

	if len(presentations) != 3 {
		t.Fatalf(
			"expected 3 presentations, got %d",
			len(presentations),
		)
	}

	if presentations[0] != "#ff0000" {
		t.Errorf("expected #ff0000, got %s", presentations[0])
	}
	if presentations[1] != "rgb(255 0 0)" {
		t.Errorf(
			"expected rgb(255 0 0), got %s",
			presentations[1],
		)
	}
	if presentations[2] != "hsl(0 100% 50%)" {
		t.Errorf(
			"expected hsl(0 100%% 50%%), got %s",
			presentations[2],
		)
	}
}

func TestColorPresentation_WithAlpha(t *testing.T) {
	c := Color{Red: 1.0, Green: 0.0, Blue: 0.0, Alpha: 0.5}
	presentations := ColorPresentation(c)

	if len(presentations) != 3 {
		t.Fatalf(
			"expected 3 presentations, got %d",
			len(presentations),
		)
	}

	if presentations[0] != "#ff000080" {
		t.Errorf("expected #ff000080, got %s", presentations[0])
	}
	if presentations[1] != "rgb(255 0 0 / 50%)" {
		t.Errorf(
			"expected rgb(255 0 0 / 50%%), got %s",
			presentations[1],
		)
	}
	if presentations[2] != "hsl(0 100% 50% / 50%)" {
		t.Errorf(
			"expected hsl(0 100%% 50%% / 50%%), got %s",
			presentations[2],
		)
	}
}

func TestFindDocumentColors_LabFunction(t *testing.T) {
	src := []byte(`.foo { color: lab(50 40 -20); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	// lab(50 40 -20) should produce some color
	c := colors[0].Color
	if c.Alpha != 1.0 {
		t.Errorf("expected alpha 1.0, got %f", c.Alpha)
	}
}

func TestFindDocumentColors_LchFunction(t *testing.T) {
	src := []byte(`.foo { color: lch(50 30 270); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	c := colors[0].Color
	if c.Alpha != 1.0 {
		t.Errorf("expected alpha 1.0, got %f", c.Alpha)
	}
}

func TestFindDocumentColors_OklabFunction(t *testing.T) {
	src := []byte(`.foo { color: oklab(0.5 0.1 -0.1); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	c := colors[0].Color
	if c.Alpha != 1.0 {
		t.Errorf("expected alpha 1.0, got %f", c.Alpha)
	}
}

func TestFindDocumentColors_OklchFunction(t *testing.T) {
	src := []byte(`.foo { color: oklch(0.5 0.15 270); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	c := colors[0].Color
	if c.Alpha != 1.0 {
		t.Errorf("expected alpha 1.0, got %f", c.Alpha)
	}
}

func TestFindDocumentColors_LabWithAlpha(t *testing.T) {
	src := []byte(`.foo { color: lab(50 40 -20 / 50%); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	assertColorClose(t, colors[0].Color,
		colors[0].Color.Red,
		colors[0].Color.Green,
		colors[0].Color.Blue,
		0.5,
	)
}

func TestFindDocumentColors_OklchRed(t *testing.T) {
	// oklch(0.6279 0.2577 29.23) ≈ red-ish
	src := []byte(`.foo { color: oklch(0.6279 0.2577 29.23); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	c := colors[0].Color
	// Should be reddish: R should be high
	if c.Red < 0.5 {
		t.Errorf("expected reddish color, got R=%.3f", c.Red)
	}
}

func TestFindDocumentColors_OklabPercentL(t *testing.T) {
	// oklab(50% 0.1 -0.1) — L as percentage
	src := []byte(`.foo { color: oklab(50% 0.1 -0.1); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	c := colors[0].Color
	if c.Alpha != 1.0 {
		t.Errorf("expected alpha 1.0, got %f", c.Alpha)
	}
}

func TestFindDocumentColors_OklchPercentL(t *testing.T) {
	// oklch(50% 0.15 270) — L as percentage
	src := []byte(`.foo { color: oklch(50% 0.15 270); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	c := colors[0].Color
	if c.Alpha != 1.0 {
		t.Errorf("expected alpha 1.0, got %f", c.Alpha)
	}
}

func TestFindDocumentColors_LchWithAlpha(t *testing.T) {
	src := []byte(`.foo { color: lch(50 30 270 / 75%); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	assertColorClose(t, colors[0].Color,
		colors[0].Color.Red,
		colors[0].Color.Green,
		colors[0].Color.Blue,
		0.75,
	)
}

func TestFindDocumentColors_OklchWithAlpha(t *testing.T) {
	src := []byte(`.foo { color: oklch(0.5 0.15 270 / 25%); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	assertColorClose(t, colors[0].Color,
		colors[0].Color.Red,
		colors[0].Color.Green,
		colors[0].Color.Blue,
		0.25,
	)
}

func TestFindDocumentColors_LabBlack(t *testing.T) {
	// lab(0 0 0) = black
	src := []byte(`.foo { color: lab(0 0 0); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	assertColorClose(t, colors[0].Color, 0.0, 0.0, 0.0, 1.0)
}

func TestFindDocumentColors_LabWhite(t *testing.T) {
	// lab(100 0 0) = white
	src := []byte(`.foo { color: lab(100 0 0); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	assertColorClose(t, colors[0].Color, 1.0, 1.0, 1.0, 1.0)
}

func TestFindDocumentColors_OklabBlack(t *testing.T) {
	// oklab(0 0 0) = black
	src := []byte(`.foo { color: oklab(0 0 0); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	assertColorClose(t, colors[0].Color, 0.0, 0.0, 0.0, 1.0)
}

func TestFindDocumentColors_OklabWhite(t *testing.T) {
	// oklab(1 0 0) = white
	src := []byte(`.foo { color: oklab(1 0 0); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	assertColorClose(t, colors[0].Color, 1.0, 1.0, 1.0, 1.0)
}

func TestFindDocumentColors_HSLWithAlpha(t *testing.T) {
	src := []byte(`.foo { color: hsl(0 100% 50% / 50%); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	assertColorClose(t, colors[0].Color, 1.0, 0.0, 0.0, 0.5)
}

func TestFindDocumentColors_HWBWithAlpha(t *testing.T) {
	src := []byte(`.foo { color: hwb(0 0% 0% / 50%); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	assertColorClose(t, colors[0].Color, 1.0, 0.0, 0.0, 0.5)
}

func TestFindDocumentColors_RGBPercentage(t *testing.T) {
	src := []byte(`.foo { color: rgb(100%, 0%, 0%); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	assertColorClose(t, colors[0].Color, 1.0, 0.0, 0.0, 1.0)
}

func TestColorPresentation_Green(t *testing.T) {
	c := Color{Red: 0.0, Green: 1.0, Blue: 0.0, Alpha: 1.0}
	presentations := ColorPresentation(c)

	if presentations[0] != "#00ff00" {
		t.Errorf("expected #00ff00, got %s", presentations[0])
	}
	if presentations[1] != "rgb(0 255 0)" {
		t.Errorf("expected rgb(0 255 0), got %s",
			presentations[1])
	}
	if presentations[2] != "hsl(120 100% 50%)" {
		t.Errorf("expected hsl(120 100%% 50%%), got %s",
			presentations[2])
	}
}

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		input      string
		ok         bool
		r, g, b, a float64
	}{
		{"fff", true, 1, 1, 1, 1},
		{"000", true, 0, 0, 0, 1},
		{"ff0000", true, 1, 0, 0, 1},
		{"00ff0080", true, 0, 1, 0, 128.0 / 255.0},
		{"xyz", false, 0, 0, 0, 0},
		{"", false, 0, 0, 0, 0},
		{"12345", false, 0, 0, 0, 0}, // invalid length
	}

	for _, tt := range tests {
		c, ok := parseHexColor(tt.input)
		if ok != tt.ok {
			t.Errorf("parseHexColor(%q): ok=%v, want %v",
				tt.input, ok, tt.ok)
			continue
		}
		if !ok {
			continue
		}
		assertColorClose(t, c, tt.r, tt.g, tt.b, tt.a)
	}
}

// mockResolver implements VariableResolver for testing.
type mockResolver struct {
	vars map[string]string
}

func (m *mockResolver) ResolveVariable(
	name string,
) (string, bool) {
	v, ok := m.vars[name]
	return v, ok
}

func TestFindDocumentColorsWithVar(t *testing.T) {
	src := []byte(`.foo { color: var(--primary); }`)
	ss, _ := parser.Parse(src)
	resolver := &mockResolver{
		vars: map[string]string{
			"--primary": "#ff0000",
		},
	}
	colors := FindDocumentColorsResolved(ss, src, resolver)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	assertColorClose(t, colors[0].Color, 1.0, 0.0, 0.0, 1.0)
}

func TestFindDocumentColorsVarUnknown(t *testing.T) {
	src := []byte(`.foo { color: var(--unknown); }`)
	ss, _ := parser.Parse(src)
	resolver := &mockResolver{
		vars: map[string]string{},
	}
	colors := FindDocumentColorsResolved(ss, src, resolver)

	if len(colors) != 0 {
		t.Fatalf("expected 0 colors, got %d", len(colors))
	}
}

func TestFindDocumentColorsVarNonColor(t *testing.T) {
	src := []byte(`.foo { margin: var(--spacing); }`)
	ss, _ := parser.Parse(src)
	resolver := &mockResolver{
		vars: map[string]string{
			"--spacing": "10px",
		},
	}
	colors := FindDocumentColorsResolved(ss, src, resolver)

	if len(colors) != 0 {
		t.Fatalf("expected 0 colors, got %d", len(colors))
	}
}

func TestFindDocumentColorsVarNested(t *testing.T) {
	// var() inside a variable value should not be recursively
	// resolved (nil resolver on recursion).
	src := []byte(`.foo { color: var(--alias); }`)
	ss, _ := parser.Parse(src)
	resolver := &mockResolver{
		vars: map[string]string{
			"--alias": "var(--real)",
			"--real":  "#00ff00",
		},
	}
	colors := FindDocumentColorsResolved(ss, src, resolver)

	// --alias resolves to "var(--real)" which contains a var()
	// but since recursion uses nil resolver, no color found.
	if len(colors) != 0 {
		t.Fatalf("expected 0 colors, got %d", len(colors))
	}
}

func TestFindDocumentColorsVarRGBValue(t *testing.T) {
	src := []byte(`.foo { background: var(--bg); }`)
	ss, _ := parser.Parse(src)
	resolver := &mockResolver{
		vars: map[string]string{
			"--bg": "rgb(0, 255, 0)",
		},
	}
	colors := FindDocumentColorsResolved(ss, src, resolver)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	assertColorClose(t, colors[0].Color, 0.0, 1.0, 0.0, 1.0)
}

func TestFindDocumentColorsVarNamedColor(t *testing.T) {
	src := []byte(`.foo { color: var(--accent); }`)
	ss, _ := parser.Parse(src)
	resolver := &mockResolver{
		vars: map[string]string{
			"--accent": "blue",
		},
	}
	colors := FindDocumentColorsResolved(ss, src, resolver)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	assertColorClose(t, colors[0].Color, 0.0, 0.0, 1.0, 1.0)
}

func TestFindDocumentColorsVarNilResolver(t *testing.T) {
	// With nil resolver, var() should produce no colors (same
	// as original behavior).
	src := []byte(`.foo { color: var(--primary); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 0 {
		t.Fatalf("expected 0 colors, got %d", len(colors))
	}
}

func TestRelativeColor_RGBPassthrough(t *testing.T) {
	src := []byte(`.foo { color: rgb(from red r g b); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	assertColorClose(t, colors[0].Color, 1.0, 0.0, 0.0, 1.0)
}

func TestRelativeColor_HSLFromHex(t *testing.T) {
	src := []byte(`.foo { color: hsl(from #ff0000 h s l); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	assertColorClose(t, colors[0].Color, 1.0, 0.0, 0.0, 1.0)
}

func TestRelativeColor_LiteralSubstitution(t *testing.T) {
	// rgb(from blue 255 g b) → r=255, g=0, b=255 → magenta
	src := []byte(`.foo { color: rgb(from blue 255 g b); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	assertColorClose(t, colors[0].Color, 1.0, 0.0, 1.0, 1.0)
}

func TestRelativeColor_CalcExpression(t *testing.T) {
	// hsl(from red h calc(s - 50) l) → reduced saturation
	src := []byte(`.foo { color: hsl(from red h calc(s - 50) l); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	// red is hsl(0, 100%, 50%), calc(s-50) = 50%
	// hsl(0, 50%, 50%) → rgb(191, 64, 64) approximately
	c := colors[0].Color
	if c.Red < 0.5 {
		t.Errorf("expected reddish, got R=%.3f", c.Red)
	}
	if c.Green > 0.4 || c.Green < 0.15 {
		t.Errorf(
			"expected moderate green channel, got G=%.3f",
			c.Green,
		)
	}
}

func TestRelativeColor_AlphaOverride(t *testing.T) {
	src := []byte(`.foo { color: rgb(from red r g b / 50%); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	assertColorClose(t, colors[0].Color, 1.0, 0.0, 0.0, 0.5)
}

func TestRelativeColor_NoneKeyword(t *testing.T) {
	src := []byte(`.foo { color: rgb(from red none g b); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	assertColorClose(t, colors[0].Color, 0.0, 0.0, 0.0, 1.0)
}

func TestRelativeColor_VarOrigin(t *testing.T) {
	src := []byte(`.foo { color: rgb(from var(--c) r g b / 40%); }`)
	ss, _ := parser.Parse(src)
	resolver := &mockResolver{
		vars: map[string]string{"--c": "#ff0000"},
	}
	colors := FindDocumentColorsResolved(ss, src, resolver)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	assertColorClose(t, colors[0].Color, 1.0, 0.0, 0.0, 0.4)
}

func TestRelativeColor_NestedFunctionOrigin(t *testing.T) {
	src := []byte(
		`.foo { color: rgb(from hsl(0 100% 50%) r g b); }`,
	)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	assertColorClose(t, colors[0].Color, 1.0, 0.0, 0.0, 1.0)
}

func TestRelativeColor_OklchPassthrough(t *testing.T) {
	src := []byte(`.foo { color: oklch(from green l c h); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	// green = {0, 0.502, 0, 1} — oklch passthrough should
	// round-trip close to original
	c := colors[0].Color
	if c.Green < 0.4 {
		t.Errorf("expected greenish, got G=%.3f", c.Green)
	}
	if c.Red > 0.1 {
		t.Errorf("expected low red, got R=%.3f", c.Red)
	}
}

func TestRelativeColor_PercentageChannel(t *testing.T) {
	// rgb(from red 50% g b) → r channel = 50% of 255 = 127.5
	src := []byte(`.foo { color: rgb(from red 50% g b); }`)
	ss, _ := parser.Parse(src)
	colors := FindDocumentColors(ss, src)

	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
	assertColorClose(t, colors[0].Color, 0.5, 0.0, 0.0, 1.0)
}

func TestCalcEval(t *testing.T) {
	tests := []struct {
		expr string
		vars map[string]float64
		want float64
		ok   bool
	}{
		{"10 + 5", nil, 15, true},
		{"10 - 3", nil, 7, true},
		{"2 * 3", nil, 6, true},
		{"10 / 2", nil, 5, true},
		{"2 + 3 * 4", nil, 14, true},
		{"s - 50", map[string]float64{"s": 100}, 50, true},
		{
			"l + 40",
			map[string]float64{"l": 10},
			50,
			true,
		},
		{"(2 + 3) * 4", nil, 20, true},
	}

	for _, tt := range tests {
		tokens := scanner.ScanAll([]byte(tt.expr))
		got, ok := evalCalc(tokens, tt.vars)
		if ok != tt.ok {
			t.Errorf(
				"evalCalc(%q): ok=%v, want %v",
				tt.expr, ok, tt.ok,
			)
			continue
		}
		if !ok {
			continue
		}
		if math.Abs(got-tt.want) > 0.001 {
			t.Errorf(
				"evalCalc(%q) = %f, want %f",
				tt.expr, got, tt.want,
			)
		}
	}
}

func assertColorClose(
	t *testing.T,
	c Color,
	r, g, b, a float64,
) {
	t.Helper()
	const epsilon = 0.02
	if math.Abs(c.Red-r) > epsilon ||
		math.Abs(c.Green-g) > epsilon ||
		math.Abs(c.Blue-b) > epsilon ||
		math.Abs(c.Alpha-a) > epsilon {
		t.Errorf(
			"color mismatch: got (%.3f, %.3f, %.3f, %.3f), "+
				"want (%.3f, %.3f, %.3f, %.3f)",
			c.Red, c.Green, c.Blue, c.Alpha,
			r, g, b, a,
		)
	}
}

func TestFindColorCodeActions_Hex(t *testing.T) {
	src := []byte(`.a { color: #ff0000; }`)
	ss, _ := parser.Parse(src)
	// Place cursor on the hex color
	offset := 12 // on '#'
	actions := FindColorCodeActions(ss, src, offset)

	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}

	// Should offer rgb and hsl, not hex
	for _, a := range actions {
		if a.Kind != CodeActionRefactor {
			t.Errorf("expected refactor kind, got %s", a.Kind)
		}
		if a.Title == "Convert to hex" {
			t.Error("should not offer conversion to same format")
		}
	}
	if actions[0].Title != "Convert to rgb" {
		t.Errorf("expected 'Convert to rgb', got %q",
			actions[0].Title)
	}
	if actions[1].Title != "Convert to hsl" {
		t.Errorf("expected 'Convert to hsl', got %q",
			actions[1].Title)
	}
}

func TestFindColorCodeActions_RGB(t *testing.T) {
	src := []byte(`.a { color: rgb(255 0 0); }`)
	ss, _ := parser.Parse(src)
	offset := 13 // inside 'rgb('
	actions := FindColorCodeActions(ss, src, offset)

	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}
	if actions[0].Title != "Convert to hex" {
		t.Errorf("expected 'Convert to hex', got %q",
			actions[0].Title)
	}
	if actions[1].Title != "Convert to hsl" {
		t.Errorf("expected 'Convert to hsl', got %q",
			actions[1].Title)
	}
}

func TestFindColorCodeActions_HSL(t *testing.T) {
	src := []byte(`.a { color: hsl(0 100% 50%); }`)
	ss, _ := parser.Parse(src)
	offset := 13 // inside 'hsl('
	actions := FindColorCodeActions(ss, src, offset)

	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}
	if actions[0].Title != "Convert to hex" {
		t.Errorf("expected 'Convert to hex', got %q",
			actions[0].Title)
	}
	if actions[1].Title != "Convert to rgb" {
		t.Errorf("expected 'Convert to rgb', got %q",
			actions[1].Title)
	}
}

func TestFindColorCodeActions_Named(t *testing.T) {
	src := []byte(`.a { color: red; }`)
	ss, _ := parser.Parse(src)
	offset := 13 // on 'red'
	actions := FindColorCodeActions(ss, src, offset)

	// Named color is "other" format, so all three offered
	if len(actions) != 3 {
		t.Fatalf("expected 3 actions, got %d", len(actions))
	}
	if actions[0].Title != "Convert to hex" {
		t.Errorf("expected 'Convert to hex', got %q",
			actions[0].Title)
	}
	if actions[1].Title != "Convert to rgb" {
		t.Errorf("expected 'Convert to rgb', got %q",
			actions[1].Title)
	}
	if actions[2].Title != "Convert to hsl" {
		t.Errorf("expected 'Convert to hsl', got %q",
			actions[2].Title)
	}
}

func TestFindColorCodeActions_NotOnColor(t *testing.T) {
	src := []byte(`.a { color: #ff0000; margin: 10px; }`)
	ss, _ := parser.Parse(src)
	offset := 28 // on '10px'
	actions := FindColorCodeActions(ss, src, offset)

	if len(actions) != 0 {
		t.Fatalf("expected 0 actions, got %d", len(actions))
	}
}

func TestFindColorCodeActions_Alpha(t *testing.T) {
	src := []byte(`.a { color: #ff000080; }`)
	ss, _ := parser.Parse(src)
	offset := 12
	actions := FindColorCodeActions(ss, src, offset)

	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}

	// Check that alpha is present in conversions
	for _, a := range actions {
		if a.ReplaceWith == "" {
			t.Error("replacement should not be empty")
		}
		// rgb and hsl with alpha should contain "/"
		if a.Title == "Convert to rgb" ||
			a.Title == "Convert to hsl" {
			if !strings.Contains(a.ReplaceWith, "/") {
				t.Errorf(
					"expected alpha separator in %q",
					a.ReplaceWith,
				)
			}
		}
		// hex with alpha should be 9 chars (#rrggbbaa)
		if a.Title == "Convert to hex" {
			if len(a.ReplaceWith) != 9 {
				t.Errorf(
					"expected 9-char hex with alpha, got %q",
					a.ReplaceWith,
				)
			}
		}
	}
}
