package analyzer

import (
	"math"
	"testing"

	"github.com/toba/go-css-lsp/internal/css/parser"
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
