package analyzer

import (
	"math"
	"strconv"
	"strings"

	"github.com/toba/go-css-lsp/internal/css/data"
	"github.com/toba/go-css-lsp/internal/css/parser"
	"github.com/toba/go-css-lsp/internal/css/scanner"
)

// Color represents an RGBA color with components in [0,1].
type Color struct {
	Red   float64
	Green float64
	Blue  float64
	Alpha float64
}

// DocumentColor represents a color found in a document with
// its byte range.
type DocumentColor struct {
	Color    Color
	StartPos int
	EndPos   int
}

// VariableResolver resolves CSS custom property names to their
// raw value text.
type VariableResolver interface {
	ResolveVariable(name string) (rawValue string, ok bool)
}

const maxVarDepth = 5

// depthLimitedResolver wraps a VariableResolver with a depth
// counter to prevent infinite recursion through chained var()
// references.
type depthLimitedResolver struct {
	inner VariableResolver
	depth int
}

func (d *depthLimitedResolver) ResolveVariable(
	name string,
) (string, bool) {
	if d.depth <= 0 {
		return "", false
	}
	d.depth--
	return d.inner.ResolveVariable(name)
}

// FindDocumentColors returns all colors found in the CSS
// document.
func FindDocumentColors(
	ss *parser.Stylesheet,
	src []byte,
) []DocumentColor {
	return FindDocumentColorsResolved(ss, src, nil)
}

// FindDocumentColorsResolved returns all colors found in the
// CSS document, resolving var() references through the given
// resolver.
func FindDocumentColorsResolved(
	ss *parser.Stylesheet,
	src []byte,
	resolver VariableResolver,
) []DocumentColor {
	var colors []DocumentColor

	parser.Walk(ss, func(n parser.Node) bool {
		decl, ok := n.(*parser.Declaration)
		if !ok {
			return true
		}
		if decl.Value == nil {
			return true
		}
		colors = append(
			colors,
			findColorsInTokens(
				decl.Value.Tokens, src, resolver,
			)...,
		)
		return true
	})

	return colors
}

func findColorsInTokens(
	tokens []scanner.Token,
	src []byte,
	resolver VariableResolver,
) []DocumentColor {
	var colors []DocumentColor

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		switch tok.Kind {
		case scanner.Hash:
			if c, ok := parseHexColor(tok.Value); ok {
				colors = append(colors, DocumentColor{
					Color:    c,
					StartPos: tok.Offset,
					EndPos:   tok.End,
				})
			}

		case scanner.Ident:
			name := strings.ToLower(tok.Value)
			if name == "currentcolor" {
				continue
			}
			if c, ok := namedColorMap[name]; ok {
				colors = append(colors, DocumentColor{
					Color:    c,
					StartPos: tok.Offset,
					EndPos:   tok.End,
				})
			}

		case scanner.Function:
			name := strings.ToLower(tok.Value)
			switch name {
			case "rgb", "rgba",
				"hsl", "hsla",
				"hwb",
				"lab", "lch",
				"oklab", "oklch":
				if dc, ok := parseColorFunction(
					name, tokens[i:], src, resolver,
				); ok {
					colors = append(colors, dc)
					// Skip past the closing paren to
					// avoid matching inner tokens.
					i = skipPastCloseParen(tokens, i)
				}
			case "var":
				if resolver == nil {
					continue
				}
				if dc, ok := resolveVarColor(
					tokens[i:], resolver,
				); ok {
					colors = append(colors, dc)
					i = skipPastCloseParen(tokens, i)
				}
			}
		}
	}

	return colors
}

// skipPastCloseParen advances past the matching close paren for
// the function token at tokens[start]. Returns the index of the
// closing paren (the loop will increment past it).
func skipPastCloseParen(
	tokens []scanner.Token,
	start int,
) int {
	depth := 1
	for j := start + 1; j < len(tokens); j++ {
		switch tokens[j].Kind {
		case scanner.Function, scanner.ParenOpen:
			depth++
		case scanner.ParenClose:
			depth--
			if depth == 0 {
				return j
			}
		}
	}
	return start
}

// resolveVarColor attempts to resolve a var() call to a color
// through the resolver. tokens[0] must be the var( function
// token.
func resolveVarColor(
	tokens []scanner.Token,
	resolver VariableResolver,
) (DocumentColor, bool) {
	if len(tokens) < 2 {
		return DocumentColor{}, false
	}

	startPos := tokens[0].Offset
	var endPos int
	var varName string

	// Find the variable name and closing paren
	for j := 1; j < len(tokens); j++ {
		tok := tokens[j]
		if tok.Kind == scanner.Whitespace {
			continue
		}
		if tok.Kind == scanner.ParenClose {
			endPos = tok.End
			break
		}
		if tok.Kind == scanner.Ident && varName == "" {
			varName = tok.Value
			continue
		}
		// Hit comma (fallback) or other token — find the
		// closing paren for the span
		if tok.Kind == scanner.Comma {
			for k := j + 1; k < len(tokens); k++ {
				if tokens[k].Kind == scanner.ParenClose {
					endPos = tokens[k].End
					goto resolve
				}
			}
			return DocumentColor{}, false
		}
	}

resolve:
	if varName == "" || endPos == 0 {
		return DocumentColor{}, false
	}

	rawValue, ok := resolver.ResolveVariable(varName)
	if !ok {
		return DocumentColor{}, false
	}

	// Scan the raw value as CSS tokens and look for a color.
	// Use a depth-limited resolver to allow chained var()
	// resolution while preventing infinite recursion.
	valTokens := scanner.ScanAll([]byte(rawValue))
	var remaining int
	if dl, ok := resolver.(*depthLimitedResolver); ok {
		remaining = dl.depth
	} else {
		remaining = maxVarDepth
	}
	nextResolver := &depthLimitedResolver{
		inner: resolver,
		depth: remaining,
	}
	if dl, ok := resolver.(*depthLimitedResolver); ok {
		nextResolver.inner = dl.inner
	}
	resolved := findColorsInTokens(valTokens, []byte(rawValue), nextResolver)
	if len(resolved) == 0 {
		return DocumentColor{}, false
	}

	return DocumentColor{
		Color:    resolved[0].Color,
		StartPos: startPos,
		EndPos:   endPos,
	}, true
}

// parseHexColor parses a hex color value (without the # prefix
// which the scanner strips).
func parseHexColor(value string) (Color, bool) {
	hex := value
	if hex == "" {
		return Color{}, false
	}

	switch len(hex) {
	case 3: // #rgb
		r := hexVal(hex[0])
		g := hexVal(hex[1])
		b := hexVal(hex[2])
		if r < 0 || g < 0 || b < 0 {
			return Color{}, false
		}
		return Color{
			Red:   float64(r*17) / 255.0,
			Green: float64(g*17) / 255.0,
			Blue:  float64(b*17) / 255.0,
			Alpha: 1.0,
		}, true

	case 4: // #rgba
		r := hexVal(hex[0])
		g := hexVal(hex[1])
		b := hexVal(hex[2])
		a := hexVal(hex[3])
		if r < 0 || g < 0 || b < 0 || a < 0 {
			return Color{}, false
		}
		return Color{
			Red:   float64(r*17) / 255.0,
			Green: float64(g*17) / 255.0,
			Blue:  float64(b*17) / 255.0,
			Alpha: float64(a*17) / 255.0,
		}, true

	case 6: // #rrggbb
		r, ok1 := hexByte(hex[0:2])
		g, ok2 := hexByte(hex[2:4])
		b, ok3 := hexByte(hex[4:6])
		if !ok1 || !ok2 || !ok3 {
			return Color{}, false
		}
		return Color{
			Red:   float64(r) / 255.0,
			Green: float64(g) / 255.0,
			Blue:  float64(b) / 255.0,
			Alpha: 1.0,
		}, true

	case 8: // #rrggbbaa
		r, ok1 := hexByte(hex[0:2])
		g, ok2 := hexByte(hex[2:4])
		b, ok3 := hexByte(hex[4:6])
		a, ok4 := hexByte(hex[6:8])
		if !ok1 || !ok2 || !ok3 || !ok4 {
			return Color{}, false
		}
		return Color{
			Red:   float64(r) / 255.0,
			Green: float64(g) / 255.0,
			Blue:  float64(b) / 255.0,
			Alpha: float64(a) / 255.0,
		}, true
	}

	return Color{}, false
}

func hexVal(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}

func hexByte(s string) (int, bool) {
	h := hexVal(s[0])
	l := hexVal(s[1])
	if h < 0 || l < 0 {
		return 0, false
	}
	return h*16 + l, true
}

// parseColorFunction parses rgb(), rgba(), hsl(), hsla(), hwb()
// from a token slice starting at the function token.
func parseColorFunction(
	name string,
	tokens []scanner.Token,
	src []byte,
	resolver VariableResolver,
) (DocumentColor, bool) {
	// Check for relative color syntax: func(from <origin> ...)
	if isRelativeColor(tokens) {
		return parseRelativeColor(
			name, tokens, src, resolver,
		)
	}

	// Collect numeric arguments until closing paren.
	startPos := tokens[0].Offset
	var endPos int
	var args []float64
	var isPercent []bool
	hasSlash := false

	negateNext := false
	for j := 1; j < len(tokens); j++ {
		tok := tokens[j]
		switch tok.Kind {
		case scanner.ParenClose:
			endPos = tok.End
			goto done
		case scanner.Number:
			v, err := strconv.ParseFloat(tok.Value, 64)
			if err != nil {
				return DocumentColor{}, false
			}
			if negateNext {
				v = -v
				negateNext = false
			}
			args = append(args, v)
			isPercent = append(isPercent, false)
		case scanner.Percentage:
			v, err := strconv.ParseFloat(tok.Value, 64)
			if err != nil {
				return DocumentColor{}, false
			}
			if negateNext {
				v = -v
				negateNext = false
			}
			args = append(args, v)
			isPercent = append(isPercent, true)
		case scanner.Delim:
			switch tok.Value {
			case "/":
				hasSlash = true
			case "-":
				negateNext = true
			}
		case scanner.Whitespace, scanner.Comma:
			continue
		default:
			continue
		}
	}
	return DocumentColor{}, false

done:
	if len(args) < 3 {
		return DocumentColor{}, false
	}

	var c Color
	var ok bool

	switch name {
	case "rgb", "rgba":
		c, ok = buildRGB(args, isPercent, hasSlash)
	case "hsl", "hsla":
		c, ok = buildHSL(args, isPercent, hasSlash)
	case "hwb":
		c, ok = buildHWB(args, isPercent, hasSlash)
	case "lab":
		c, ok = buildLab(args, isPercent)
	case "lch":
		c, ok = buildLCH(args, isPercent)
	case "oklab":
		c, ok = buildOklab(args, isPercent)
	case "oklch":
		c, ok = buildOklch(args, isPercent)
	default:
		return DocumentColor{}, false
	}

	if !ok {
		return DocumentColor{}, false
	}

	return DocumentColor{
		Color:    c,
		StartPos: startPos,
		EndPos:   endPos,
	}, true
}

func buildRGB(
	args []float64,
	isPercent []bool,
	hasSlash bool,
) (Color, bool) {
	r, g, b := args[0], args[1], args[2]

	if isPercent[0] {
		r = r / 100.0 * 255.0
	}
	if isPercent[1] {
		g = g / 100.0 * 255.0
	}
	if isPercent[2] {
		b = b / 100.0 * 255.0
	}

	a := 1.0
	if len(args) >= 4 {
		a = args[3]
		if isPercent[3] {
			a /= 100.0
		}
		// If no slash and alpha > 1, it's likely 0-255
		if !hasSlash && a > 1.0 {
			a /= 255.0
		}
	}

	return Color{
		Red:   clamp01(r / 255.0),
		Green: clamp01(g / 255.0),
		Blue:  clamp01(b / 255.0),
		Alpha: clamp01(a),
	}, true
}

func buildHSL(
	args []float64,
	isPercent []bool,
	_ bool,
) (Color, bool) {
	h := args[0]
	s := args[1]
	l := args[2]

	// Normalize hue to [0,360)
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}

	// S and L should be percentages
	if isPercent[1] {
		s /= 100.0
	} else if s > 1.0 {
		s /= 100.0
	}
	if isPercent[2] {
		l /= 100.0
	} else if l > 1.0 {
		l /= 100.0
	}

	a := 1.0
	if len(args) >= 4 {
		a = args[3]
		if len(isPercent) >= 4 && isPercent[3] {
			a /= 100.0
		}
	}

	r, g, b := hslToRGB(h/360.0, s, l)

	return Color{
		Red:   clamp01(r),
		Green: clamp01(g),
		Blue:  clamp01(b),
		Alpha: clamp01(a),
	}, true
}

func buildHWB(
	args []float64,
	isPercent []bool,
	_ bool,
) (Color, bool) {
	h := args[0]
	w := args[1]
	bk := args[2]

	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}

	if isPercent[1] {
		w /= 100.0
	} else if w > 1.0 {
		w /= 100.0
	}
	if isPercent[2] {
		bk /= 100.0
	} else if bk > 1.0 {
		bk /= 100.0
	}

	a := 1.0
	if len(args) >= 4 {
		a = args[3]
		if len(isPercent) >= 4 && isPercent[3] {
			a /= 100.0
		}
	}

	// HWB to RGB via HSL
	r, g, b := hslToRGB(h/360.0, 1.0, 0.5)

	// Apply whiteness and blackness
	total := w + bk
	if total > 1.0 {
		w /= total
		bk /= total
	}

	r = r*(1.0-w-bk) + w
	g = g*(1.0-w-bk) + w
	b = b*(1.0-w-bk) + w

	return Color{
		Red:   clamp01(r),
		Green: clamp01(g),
		Blue:  clamp01(b),
		Alpha: clamp01(a),
	}, true
}

// buildLab converts CIE Lab to sRGB.
// L: [0,100], a: [-125,125], b: [-125,125]
func buildLab(
	args []float64,
	isPercent []bool,
) (Color, bool) {
	// L percentage maps to [0,100] — same scale, no conversion
	lVal := args[0]
	a := args[1]
	b := args[2]

	alpha := 1.0
	if len(args) >= 4 {
		alpha = args[3]
		if isPercent[3] {
			alpha /= 100.0
		}
	}

	r, g, bl := labToSRGB(lVal, a, b)
	return Color{
		Red:   clamp01(r),
		Green: clamp01(g),
		Blue:  clamp01(bl),
		Alpha: clamp01(alpha),
	}, true
}

// buildLCH converts CIE LCH to sRGB.
// L: [0,100], C: [0,150], H: [0,360)
func buildLCH(
	args []float64,
	isPercent []bool,
) (Color, bool) {
	// L percentage maps to [0,100] — same scale, no conversion
	lVal := args[0]
	chroma := args[1]
	hue := args[2]

	alpha := 1.0
	if len(args) >= 4 {
		alpha = args[3]
		if isPercent[3] {
			alpha /= 100.0
		}
	}

	// LCH to Lab
	hRad := hue * math.Pi / 180.0
	a := chroma * math.Cos(hRad)
	b := chroma * math.Sin(hRad)

	r, g, bl := labToSRGB(lVal, a, b)
	return Color{
		Red:   clamp01(r),
		Green: clamp01(g),
		Blue:  clamp01(bl),
		Alpha: clamp01(alpha),
	}, true
}

// buildOklab converts Oklab to sRGB.
// L: [0,1], a: [-0.4,0.4], b: [-0.4,0.4]
func buildOklab(
	args []float64,
	isPercent []bool,
) (Color, bool) {
	lVal := args[0]
	if isPercent[0] {
		lVal /= 100.0
	}
	a := args[1]
	b := args[2]

	alpha := 1.0
	if len(args) >= 4 {
		alpha = args[3]
		if isPercent[3] {
			alpha /= 100.0
		}
	}

	r, g, bl := oklabToSRGB(lVal, a, b)
	return Color{
		Red:   clamp01(r),
		Green: clamp01(g),
		Blue:  clamp01(bl),
		Alpha: clamp01(alpha),
	}, true
}

// buildOklch converts Oklch to sRGB.
// L: [0,1], C: [0,0.4], H: [0,360)
func buildOklch(
	args []float64,
	isPercent []bool,
) (Color, bool) {
	lVal := args[0]
	if isPercent[0] {
		lVal /= 100.0
	}
	chroma := args[1]
	hue := args[2]

	alpha := 1.0
	if len(args) >= 4 {
		alpha = args[3]
		if isPercent[3] {
			alpha /= 100.0
		}
	}

	// Oklch to Oklab
	hRad := hue * math.Pi / 180.0
	a := chroma * math.Cos(hRad)
	b := chroma * math.Sin(hRad)

	r, g, bl := oklabToSRGB(lVal, a, b)
	return Color{
		Red:   clamp01(r),
		Green: clamp01(g),
		Blue:  clamp01(bl),
		Alpha: clamp01(alpha),
	}, true
}

// labToSRGB converts CIE Lab to linear sRGB via XYZ D65.
func labToSRGB(
	lVal, a, b float64,
) (float64, float64, float64) {
	// Lab to XYZ (D65 white point)
	fy := (lVal + 16.0) / 116.0
	fx := a/500.0 + fy
	fz := fy - b/200.0

	const epsilon = 216.0 / 24389.0
	const kappa = 24389.0 / 27.0

	var x, y, z float64
	if fx*fx*fx > epsilon {
		x = fx * fx * fx
	} else {
		x = (116.0*fx - 16.0) / kappa
	}
	if lVal > kappa*epsilon {
		y = fy * fy * fy
	} else {
		y = lVal / kappa
	}
	if fz*fz*fz > epsilon {
		z = fz * fz * fz
	} else {
		z = (116.0*fz - 16.0) / kappa
	}

	// D65 white point
	x *= 0.95047
	z *= 1.08883

	// XYZ to linear sRGB
	r := x*3.2404542 - y*1.5371385 - z*0.4985314
	g := -x*0.9692660 + y*1.8760108 + z*0.0415560
	bl := x*0.0556434 - y*0.2040259 + z*1.0572252

	return linearToSRGB(r), linearToSRGB(g), linearToSRGB(bl)
}

// oklabToSRGB converts Oklab to sRGB.
func oklabToSRGB(
	lVal, a, b float64,
) (float64, float64, float64) {
	// Oklab to LMS
	l := lVal + 0.3963377774*a + 0.2158037573*b
	m := lVal - 0.1055613458*a - 0.0638541728*b
	s := lVal - 0.0894841775*a - 1.2914855480*b

	// Cube
	l = l * l * l
	m = m * m * m
	s = s * s * s

	// LMS to linear sRGB
	r := +4.0767416621*l - 3.3077115913*m + 0.2309699292*s
	g := -1.2684380046*l + 2.6097574011*m - 0.3413193965*s
	bl := -0.0041960863*l - 0.7034186147*m + 1.7076147010*s

	return linearToSRGB(r), linearToSRGB(g), linearToSRGB(bl)
}

// linearToSRGB applies sRGB gamma.
func linearToSRGB(c float64) float64 {
	if c <= 0.0031308 {
		return 12.92 * c
	}
	return 1.055*math.Pow(c, 1.0/2.4) - 0.055
}

func hslToRGB(h, s, l float64) (float64, float64, float64) {
	if s == 0 {
		return l, l, l
	}

	var q float64
	if l < 0.5 {
		q = l * (1.0 + s)
	} else {
		q = l + s - l*s
	}
	p := 2.0*l - q

	r := hueToRGB(p, q, h+1.0/3.0)
	g := hueToRGB(p, q, h)
	b := hueToRGB(p, q, h-1.0/3.0)

	return r, g, b
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t++
	}
	if t > 1 {
		t--
	}
	switch {
	case t < 1.0/6.0:
		return p + (q-p)*6.0*t
	case t < 1.0/2.0:
		return q
	case t < 2.0/3.0:
		return p + (q-p)*(2.0/3.0-t)*6.0
	}
	return p
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// --- Relative color syntax support ---

// srgbToLinear inverts the sRGB gamma curve.
func srgbToLinear(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

// srgbToXYZ converts linear sRGB to CIE XYZ (D65).
func srgbToXYZ(
	r, g, b float64,
) (float64, float64, float64) {
	rl := srgbToLinear(r)
	gl := srgbToLinear(g)
	bl := srgbToLinear(b)

	x := 0.4124564*rl + 0.3575761*gl + 0.1804375*bl
	y := 0.2126729*rl + 0.7151522*gl + 0.0721750*bl
	z := 0.0193339*rl + 0.1191920*gl + 0.9503041*bl

	return x, y, z
}

// srgbToLab converts sRGB [0,1] to CIE Lab.
func srgbToLab(
	r, g, b float64,
) (float64, float64, float64) {
	x, y, z := srgbToXYZ(r, g, b)

	// D65 white point
	x /= 0.95047
	z /= 1.08883

	const epsilon = 216.0 / 24389.0
	const kappa = 24389.0 / 27.0

	f := func(t float64) float64 {
		if t > epsilon {
			return math.Cbrt(t)
		}
		return (kappa*t + 16.0) / 116.0
	}

	fx := f(x)
	fy := f(y)
	fz := f(z)

	lVal := 116.0*fy - 16.0
	aVal := 500.0 * (fx - fy)
	bVal := 200.0 * (fy - fz)

	return lVal, aVal, bVal
}

// srgbToOklab converts sRGB [0,1] to Oklab.
func srgbToOklab(
	r, g, b float64,
) (float64, float64, float64) {
	rl := srgbToLinear(r)
	gl := srgbToLinear(g)
	bl := srgbToLinear(b)

	l := 0.4122214708*rl + 0.5363325363*gl + 0.0514459929*bl
	m := 0.2119034982*rl + 0.6806995451*gl + 0.1073969566*bl
	s := 0.0883024619*rl + 0.2817188376*gl + 0.6299787005*bl

	l = math.Cbrt(l)
	m = math.Cbrt(m)
	s = math.Cbrt(s)

	lVal := 0.2104542553*l + 0.7936177850*m - 0.0040720468*s
	aVal := 1.9779984951*l - 2.4285922050*m + 0.4505937099*s
	bVal := 0.0259040371*l + 0.7827717662*m - 0.8086757660*s

	return lVal, aVal, bVal
}

// colorToRGB decomposes sRGB Color to [r,g,b,alpha] in
// RGB native ranges (r,g,b: 0-255).
func colorToRGB(c Color) [4]float64 {
	return [4]float64{
		c.Red * 255, c.Green * 255, c.Blue * 255, c.Alpha,
	}
}

// colorToHSL decomposes sRGB Color to [h,s,l,alpha] in
// HSL native ranges (h: 0-360, s,l: 0-100).
func colorToHSL(c Color) [4]float64 {
	h, s, l := rgbToHSL(c.Red, c.Green, c.Blue)
	return [4]float64{h * 360, s * 100, l * 100, c.Alpha}
}

// colorToHWB decomposes sRGB Color to [h,w,b,alpha] in
// HWB native ranges (h: 0-360, w,b: 0-100).
func colorToHWB(c Color) [4]float64 {
	h, _, _ := rgbToHSL(c.Red, c.Green, c.Blue)
	w := math.Min(c.Red, math.Min(c.Green, c.Blue))
	bk := 1 - math.Max(c.Red, math.Max(c.Green, c.Blue))
	return [4]float64{h * 360, w * 100, bk * 100, c.Alpha}
}

// colorToLab decomposes sRGB Color to [L,a,b,alpha] in
// CIE Lab native ranges.
func colorToLab(c Color) [4]float64 {
	l, a, b := srgbToLab(c.Red, c.Green, c.Blue)
	return [4]float64{l, a, b, c.Alpha}
}

// labToLCH converts L,a,b coordinates to L,C,H where C is
// chroma and H is hue in degrees [0,360).
func labToLCH(l, a, b float64) (float64, float64, float64) {
	c := math.Sqrt(a*a + b*b)
	h := math.Atan2(b, a) * 180 / math.Pi
	if h < 0 {
		h += 360
	}
	return l, c, h
}

// colorToLCH decomposes sRGB Color to [L,C,H,alpha] in
// CIE LCH native ranges.
func colorToLCH(c Color) [4]float64 {
	l, a, b := srgbToLab(c.Red, c.Green, c.Blue)
	lch, ch, h := labToLCH(l, a, b)
	return [4]float64{lch, ch, h, c.Alpha}
}

// colorToOklab decomposes sRGB Color to [L,a,b,alpha] in
// Oklab native ranges.
func colorToOklab(c Color) [4]float64 {
	l, a, b := srgbToOklab(c.Red, c.Green, c.Blue)
	return [4]float64{l, a, b, c.Alpha}
}

// colorToOklch decomposes sRGB Color to [L,C,H,alpha] in
// Oklch native ranges.
func colorToOklch(c Color) [4]float64 {
	l, a, b := srgbToOklab(c.Red, c.Green, c.Blue)
	lch, ch, h := labToLCH(l, a, b)
	return [4]float64{lch, ch, h, c.Alpha}
}

// colorSpaceChannels maps color function names to their channel
// name identifiers.
var colorSpaceChannels = map[string][]string{
	"rgb":   {"r", "g", "b", "alpha"},
	"rgba":  {"r", "g", "b", "alpha"},
	"hsl":   {"h", "s", "l", "alpha"},
	"hsla":  {"h", "s", "l", "alpha"},
	"hwb":   {"h", "w", "b", "alpha"},
	"lab":   {"l", "a", "b", "alpha"},
	"lch":   {"l", "c", "h", "alpha"},
	"oklab": {"l", "a", "b", "alpha"},
	"oklch": {"l", "c", "h", "alpha"},
}

// colorSpaceDecompose maps color function names to their
// decomposition functions.
var colorSpaceDecompose = map[string]func(Color) [4]float64{
	"rgb":   colorToRGB,
	"rgba":  colorToRGB,
	"hsl":   colorToHSL,
	"hsla":  colorToHSL,
	"hwb":   colorToHWB,
	"lab":   colorToLab,
	"lch":   colorToLCH,
	"oklab": colorToOklab,
	"oklch": colorToOklch,
}

// isRelativeColor checks if tokens[1] (after function open) is
// the "from" keyword, indicating relative color syntax.
func isRelativeColor(tokens []scanner.Token) bool {
	for i := 1; i < len(tokens); i++ {
		if tokens[i].Kind == scanner.Whitespace {
			continue
		}
		return tokens[i].Kind == scanner.Ident &&
			strings.ToLower(tokens[i].Value) == "from"
	}
	return false
}

// parseOriginColor parses the origin color in a relative color
// expression. It handles: hex, named color, color function, or
// var(). Returns the color, the token index after the origin,
// and success.
func parseOriginColor(
	tokens []scanner.Token,
	start int,
	src []byte,
	resolver VariableResolver,
) (Color, int, bool) {
	for i := start; i < len(tokens); i++ {
		tok := tokens[i]
		if tok.Kind == scanner.Whitespace {
			continue
		}

		switch tok.Kind {
		case scanner.Hash:
			c, ok := parseHexColor(tok.Value)
			if !ok {
				return Color{}, i, false
			}
			return c, i + 1, true

		case scanner.Ident:
			c, ok := namedColorMap[strings.ToLower(tok.Value)]
			if !ok {
				return Color{}, i, false
			}
			return c, i + 1, true

		case scanner.Function:
			name := strings.ToLower(tok.Value)
			switch name {
			case "rgb", "rgba", "hsl", "hsla", "hwb",
				"lab", "lch", "oklab", "oklch":
				j := skipPastCloseParen(tokens, i)
				if j == i {
					return Color{}, i, false
				}
				dc, ok := parseColorFunction(
					name,
					tokens[i:j+1],
					src,
					resolver,
				)
				if !ok {
					return Color{}, j + 1, false
				}
				return dc.Color, j + 1, true

			case "var":
				if resolver == nil {
					return Color{}, i, false
				}
				j := skipPastCloseParen(tokens, i)
				if j == i {
					return Color{}, i, false
				}
				dc, ok := resolveVarColor(
					tokens[i:j+1],
					resolver,
				)
				if !ok {
					return Color{}, j + 1, false
				}
				return dc.Color, j + 1, true
			}
		}

		return Color{}, i, false
	}
	return Color{}, start, false
}

// parseRelativeColor parses a relative color expression like
// rgb(from red r g b / 50%).
func parseRelativeColor(
	name string,
	tokens []scanner.Token,
	src []byte,
	resolver VariableResolver,
) (DocumentColor, bool) {
	channels, ok := colorSpaceChannels[name]
	if !ok {
		return DocumentColor{}, false
	}
	decompose, ok := colorSpaceDecompose[name]
	if !ok {
		return DocumentColor{}, false
	}

	startPos := tokens[0].Offset

	// Skip past the "from" keyword.
	i := 1
	for i < len(tokens) && tokens[i].Kind == scanner.Whitespace {
		i++
	}
	if i >= len(tokens) ||
		tokens[i].Kind != scanner.Ident ||
		strings.ToLower(tokens[i].Value) != "from" {
		return DocumentColor{}, false
	}
	i++

	// Parse origin color.
	origin, nextIdx, ok := parseOriginColor(
		tokens, i, src, resolver,
	)
	if !ok {
		return DocumentColor{}, false
	}
	i = nextIdx

	// Decompose origin into channel variables.
	vals := decompose(origin)
	channelVars := make(map[string]float64, 4)
	for ci, ch := range channels {
		channelVars[ch] = vals[ci]
	}

	// Parse 3 channel expressions + optional alpha.
	args := make([]float64, 0, 4)
	isPercent := make([]bool, 0, 4)
	hasSlash := false

	for len(args) < 3 && i < len(tokens) {
		if tokens[i].Kind == scanner.Whitespace {
			i++
			continue
		}
		if tokens[i].Kind == scanner.ParenClose {
			break
		}

		val, pct, newI, ok := parseChannelExpr(
			tokens, i, channelVars,
		)
		if !ok {
			return DocumentColor{}, false
		}
		args = append(args, val)
		isPercent = append(isPercent, pct)
		i = newI
	}

	if len(args) < 3 {
		return DocumentColor{}, false
	}

	// Optional: / alpha
	for i < len(tokens) && tokens[i].Kind == scanner.Whitespace {
		i++
	}
	if i < len(tokens) &&
		tokens[i].Kind == scanner.Delim &&
		tokens[i].Value == "/" {
		hasSlash = true
		i++
		for i < len(tokens) &&
			tokens[i].Kind == scanner.Whitespace {
			i++
		}
		if i < len(tokens) {
			val, pct, newI, ok := parseChannelExpr(
				tokens, i, channelVars,
			)
			if !ok {
				return DocumentColor{}, false
			}
			args = append(args, val)
			isPercent = append(isPercent, pct)
			i = newI
		}
	}

	// Find closing paren.
	var endPos int
	for i < len(tokens) {
		if tokens[i].Kind == scanner.ParenClose {
			endPos = tokens[i].End
			break
		}
		i++
	}
	if endPos == 0 {
		return DocumentColor{}, false
	}

	// Build color via existing build functions.
	var c Color
	switch name {
	case "rgb", "rgba":
		c, ok = buildRGB(args, isPercent, hasSlash)
	case "hsl", "hsla":
		c, ok = buildHSL(args, isPercent, hasSlash)
	case "hwb":
		c, ok = buildHWB(args, isPercent, hasSlash)
	case "lab":
		c, ok = buildLab(args, isPercent)
	case "lch":
		c, ok = buildLCH(args, isPercent)
	case "oklab":
		c, ok = buildOklab(args, isPercent)
	case "oklch":
		c, ok = buildOklch(args, isPercent)
	default:
		return DocumentColor{}, false
	}

	if !ok {
		return DocumentColor{}, false
	}

	return DocumentColor{
		Color:    c,
		StartPos: startPos,
		EndPos:   endPos,
	}, true
}

// parseChannelExpr parses a single channel value: a number,
// percentage, channel ident, "none", or calc() expression.
// Returns value, whether it's a percentage, the next token
// index, and success.
func parseChannelExpr(
	tokens []scanner.Token,
	start int,
	channelVars map[string]float64,
) (float64, bool, int, bool) {
	i := start
	for i < len(tokens) && tokens[i].Kind == scanner.Whitespace {
		i++
	}
	if i >= len(tokens) {
		return 0, false, i, false
	}

	tok := tokens[i]

	// Number literal
	if tok.Kind == scanner.Number {
		v, err := strconv.ParseFloat(tok.Value, 64)
		if err != nil {
			return 0, false, i, false
		}
		return v, false, i + 1, true
	}

	// Percentage literal
	if tok.Kind == scanner.Percentage {
		v, err := strconv.ParseFloat(tok.Value, 64)
		if err != nil {
			return 0, false, i, false
		}
		return v, true, i + 1, true
	}

	// "none" keyword
	if tok.Kind == scanner.Ident &&
		strings.ToLower(tok.Value) == "none" {
		return 0, false, i + 1, true
	}

	// Channel name ident
	if tok.Kind == scanner.Ident {
		if v, ok := channelVars[strings.ToLower(tok.Value)]; ok {
			return v, false, i + 1, true
		}
		return 0, false, i, false
	}

	// calc() function
	if tok.Kind == scanner.Function &&
		strings.ToLower(tok.Value) == "calc" {
		// Collect tokens until matching close paren.
		depth := 1
		j := i + 1
		var calcTokens []scanner.Token
		for j < len(tokens) {
			if tokens[j].Kind == scanner.Function ||
				tokens[j].Kind == scanner.ParenOpen {
				depth++
			} else if tokens[j].Kind == scanner.ParenClose {
				depth--
				if depth == 0 {
					break
				}
			}
			calcTokens = append(calcTokens, tokens[j])
			j++
		}
		if depth != 0 {
			return 0, false, j, false
		}
		val, ok := evalCalc(calcTokens, channelVars)
		if !ok {
			return 0, false, j, false
		}
		return val, false, j + 1, true
	}

	// Unary minus before number/ident
	if tok.Kind == scanner.Delim && tok.Value == "-" {
		val, pct, newI, ok := parseChannelExpr(
			tokens, i+1, channelVars,
		)
		if !ok {
			return 0, false, newI, false
		}
		return -val, pct, newI, true
	}

	return 0, false, i, false
}

// ColorPresentation returns alternative representations of a
// color value.
func ColorPresentation(c Color) []string {
	r := int(math.Round(c.Red * 255))
	g := int(math.Round(c.Green * 255))
	b := int(math.Round(c.Blue * 255))

	var presentations []string

	// Hex
	if c.Alpha >= 1.0 {
		presentations = append(presentations,
			"#"+hexStr(r)+hexStr(g)+hexStr(b),
		)
	} else {
		a := int(math.Round(c.Alpha * 255))
		presentations = append(presentations,
			"#"+hexStr(r)+hexStr(g)+hexStr(b)+hexStr(a),
		)
	}

	// RGB (modern space-separated syntax)
	if c.Alpha >= 1.0 {
		presentations = append(presentations,
			"rgb("+strconv.Itoa(r)+" "+
				strconv.Itoa(g)+" "+
				strconv.Itoa(b)+")",
		)
	} else {
		presentations = append(presentations,
			"rgb("+strconv.Itoa(r)+" "+
				strconv.Itoa(g)+" "+
				strconv.Itoa(b)+" / "+
				formatAlpha(c.Alpha)+")",
		)
	}

	// HSL (modern space-separated syntax)
	h, s, l := rgbToHSL(c.Red, c.Green, c.Blue)
	hDeg := int(math.Round(h * 360))
	sPct := int(math.Round(s * 100))
	lPct := int(math.Round(l * 100))

	if c.Alpha >= 1.0 {
		presentations = append(presentations,
			"hsl("+strconv.Itoa(hDeg)+" "+
				strconv.Itoa(sPct)+"% "+
				strconv.Itoa(lPct)+"%)",
		)
	} else {
		presentations = append(presentations,
			"hsl("+strconv.Itoa(hDeg)+" "+
				strconv.Itoa(sPct)+"% "+
				strconv.Itoa(lPct)+"% / "+
				formatAlpha(c.Alpha)+")",
		)
	}

	return presentations
}

func hexStr(v int) string {
	s := strconv.FormatInt(int64(v), 16)
	if len(s) == 1 {
		return "0" + s
	}
	return s
}

func formatAlpha(a float64) string {
	pct := int(math.Round(a * 100))
	return strconv.Itoa(pct) + "%"
}

func rgbToHSL(
	r, g, b float64,
) (float64, float64, float64) {
	cMax := math.Max(r, math.Max(g, b))
	cMin := math.Min(r, math.Min(g, b))
	l := (cMax + cMin) / 2.0

	if cMax == cMin {
		return 0, 0, l
	}

	d := cMax - cMin
	var s float64
	if l > 0.5 {
		s = d / (2.0 - cMax - cMin)
	} else {
		s = d / (cMax + cMin)
	}

	var h float64
	switch cMax {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	case b:
		h = (r-g)/d + 4
	}
	h /= 6.0

	return h, s, l
}

// detectColorFormat inspects source text at [start, end) and
// returns "hex", "rgb", "hsl", or "other".
func detectColorFormat(src []byte, start, end int) string {
	if start < 0 || start >= len(src) {
		return "other"
	}
	if src[start] == '#' {
		return "hex"
	}
	text := strings.ToLower(string(src[start:end]))
	if strings.HasPrefix(text, "rgb") {
		return "rgb"
	}
	if strings.HasPrefix(text, "hsl") {
		return "hsl"
	}
	return "other"
}

// FindColorCodeActions returns refactor code actions to convert
// the color at the given byte offset to other formats.
func FindColorCodeActions(
	ss *parser.Stylesheet,
	src []byte,
	offset int,
) []CodeAction {
	colors := FindDocumentColors(ss, src)

	var dc *DocumentColor
	for i := range colors {
		if offset >= colors[i].StartPos &&
			offset <= colors[i].EndPos {
			dc = &colors[i]
			break
		}
	}
	if dc == nil {
		return nil
	}

	format := detectColorFormat(src, dc.StartPos, dc.EndPos)
	presentations := ColorPresentation(dc.Color)

	// presentations: [0]=hex, [1]=rgb, [2]=hsl
	type entry struct {
		label  string
		format string
		idx    int
	}
	entries := []entry{
		{"Convert to hex", "hex", 0},
		{"Convert to rgb", "rgb", 1},
		{"Convert to hsl", "hsl", 2},
	}

	startLine, startChar := OffsetToLineChar(
		src, dc.StartPos,
	)
	endLine, endChar := OffsetToLineChar(src, dc.EndPos)

	var actions []CodeAction
	for _, e := range entries {
		if e.format == format {
			continue
		}
		if e.idx >= len(presentations) {
			continue
		}
		actions = append(actions, CodeAction{
			Title:       e.label,
			Kind:        CodeActionRefactor,
			StartLine:   startLine,
			StartChar:   startChar,
			EndLine:     endLine,
			EndChar:     endChar,
			ReplaceWith: presentations[e.idx],
		})
	}

	return actions
}

// namedColorMap maps CSS named colors to their RGBA values.
var namedColorMap map[string]Color

func init() {
	namedColorMap = make(map[string]Color, len(data.NamedColors))
	for _, name := range data.NamedColors {
		if c, ok := namedColorRGBA[name]; ok {
			namedColorMap[name] = c
		}
	}
}

// namedColorRGBA provides RGB values for named colors.
// This covers a representative set; transparent and
// currentcolor are handled specially.
var namedColorRGBA = map[string]Color{
	"black":                {0, 0, 0, 1},
	"silver":               {0.753, 0.753, 0.753, 1},
	"gray":                 {0.502, 0.502, 0.502, 1},
	"grey":                 {0.502, 0.502, 0.502, 1},
	"white":                {1, 1, 1, 1},
	"maroon":               {0.502, 0, 0, 1},
	"red":                  {1, 0, 0, 1},
	"purple":               {0.502, 0, 0.502, 1},
	"fuchsia":              {1, 0, 1, 1},
	"green":                {0, 0.502, 0, 1},
	"lime":                 {0, 1, 0, 1},
	"olive":                {0.502, 0.502, 0, 1},
	"yellow":               {1, 1, 0, 1},
	"navy":                 {0, 0, 0.502, 1},
	"blue":                 {0, 0, 1, 1},
	"teal":                 {0, 0.502, 0.502, 1},
	"aqua":                 {0, 1, 1, 1},
	"cyan":                 {0, 1, 1, 1},
	"orange":               {1, 0.647, 0, 1},
	"aliceblue":            {0.941, 0.973, 1, 1},
	"antiquewhite":         {0.980, 0.922, 0.843, 1},
	"aquamarine":           {0.498, 1, 0.831, 1},
	"azure":                {0.941, 1, 1, 1},
	"beige":                {0.961, 0.961, 0.863, 1},
	"bisque":               {1, 0.894, 0.769, 1},
	"blanchedalmond":       {1, 0.922, 0.804, 1},
	"blueviolet":           {0.541, 0.169, 0.886, 1},
	"brown":                {0.647, 0.165, 0.165, 1},
	"burlywood":            {0.871, 0.722, 0.529, 1},
	"cadetblue":            {0.373, 0.620, 0.627, 1},
	"chartreuse":           {0.498, 1, 0, 1},
	"chocolate":            {0.824, 0.412, 0.118, 1},
	"coral":                {1, 0.498, 0.314, 1},
	"cornflowerblue":       {0.392, 0.584, 0.929, 1},
	"cornsilk":             {1, 0.973, 0.863, 1},
	"crimson":              {0.863, 0.078, 0.235, 1},
	"darkblue":             {0, 0, 0.545, 1},
	"darkcyan":             {0, 0.545, 0.545, 1},
	"darkgoldenrod":        {0.722, 0.525, 0.043, 1},
	"darkgray":             {0.663, 0.663, 0.663, 1},
	"darkgrey":             {0.663, 0.663, 0.663, 1},
	"darkgreen":            {0, 0.392, 0, 1},
	"darkkhaki":            {0.741, 0.718, 0.420, 1},
	"darkmagenta":          {0.545, 0, 0.545, 1},
	"darkolivegreen":       {0.333, 0.420, 0.184, 1},
	"darkorange":           {1, 0.549, 0, 1},
	"darkorchid":           {0.600, 0.196, 0.800, 1},
	"darkred":              {0.545, 0, 0, 1},
	"darksalmon":           {0.914, 0.588, 0.478, 1},
	"darkseagreen":         {0.561, 0.737, 0.561, 1},
	"darkslateblue":        {0.282, 0.239, 0.545, 1},
	"darkslategray":        {0.184, 0.310, 0.310, 1},
	"darkslategrey":        {0.184, 0.310, 0.310, 1},
	"darkturquoise":        {0, 0.808, 0.820, 1},
	"darkviolet":           {0.580, 0, 0.827, 1},
	"deeppink":             {1, 0.078, 0.576, 1},
	"deepskyblue":          {0, 0.749, 1, 1},
	"dimgray":              {0.412, 0.412, 0.412, 1},
	"dimgrey":              {0.412, 0.412, 0.412, 1},
	"dodgerblue":           {0.118, 0.565, 1, 1},
	"firebrick":            {0.698, 0.133, 0.133, 1},
	"floralwhite":          {1, 0.980, 0.941, 1},
	"forestgreen":          {0.133, 0.545, 0.133, 1},
	"gainsboro":            {0.863, 0.863, 0.863, 1},
	"ghostwhite":           {0.973, 0.973, 1, 1},
	"gold":                 {1, 0.843, 0, 1},
	"goldenrod":            {0.855, 0.647, 0.125, 1},
	"greenyellow":          {0.678, 1, 0.184, 1},
	"honeydew":             {0.941, 1, 0.941, 1},
	"hotpink":              {1, 0.412, 0.706, 1},
	"indianred":            {0.804, 0.361, 0.361, 1},
	"indigo":               {0.294, 0, 0.510, 1},
	"ivory":                {1, 1, 0.941, 1},
	"khaki":                {0.941, 0.902, 0.549, 1},
	"lavender":             {0.902, 0.902, 0.961, 1},
	"lavenderblush":        {1, 0.941, 0.961, 1},
	"lawngreen":            {0.486, 0.988, 0, 1},
	"lemonchiffon":         {1, 0.980, 0.804, 1},
	"lightblue":            {0.678, 0.847, 0.902, 1},
	"lightcoral":           {0.941, 0.502, 0.502, 1},
	"lightcyan":            {0.878, 1, 1, 1},
	"lightgoldenrodyellow": {0.980, 0.980, 0.824, 1},
	"lightgray":            {0.827, 0.827, 0.827, 1},
	"lightgrey":            {0.827, 0.827, 0.827, 1},
	"lightgreen":           {0.565, 0.933, 0.565, 1},
	"lightpink":            {1, 0.714, 0.757, 1},
	"lightsalmon":          {1, 0.627, 0.478, 1},
	"lightseagreen":        {0.125, 0.698, 0.667, 1},
	"lightskyblue":         {0.529, 0.808, 0.980, 1},
	"lightslategray":       {0.467, 0.533, 0.600, 1},
	"lightslategrey":       {0.467, 0.533, 0.600, 1},
	"lightsteelblue":       {0.690, 0.769, 0.871, 1},
	"lightyellow":          {1, 1, 0.878, 1},
	"limegreen":            {0.196, 0.804, 0.196, 1},
	"linen":                {0.980, 0.941, 0.902, 1},
	"magenta":              {1, 0, 1, 1},
	"mediumaquamarine":     {0.400, 0.804, 0.667, 1},
	"mediumblue":           {0, 0, 0.804, 1},
	"mediumorchid":         {0.729, 0.333, 0.827, 1},
	"mediumpurple":         {0.576, 0.439, 0.859, 1},
	"mediumseagreen":       {0.235, 0.702, 0.443, 1},
	"mediumslateblue":      {0.482, 0.408, 0.933, 1},
	"mediumspringgreen":    {0, 0.980, 0.604, 1},
	"mediumturquoise":      {0.282, 0.820, 0.800, 1},
	"mediumvioletred":      {0.780, 0.082, 0.522, 1},
	"midnightblue":         {0.098, 0.098, 0.439, 1},
	"mintcream":            {0.961, 1, 0.980, 1},
	"mistyrose":            {1, 0.894, 0.882, 1},
	"moccasin":             {1, 0.894, 0.710, 1},
	"navajowhite":          {1, 0.871, 0.678, 1},
	"oldlace":              {0.992, 0.961, 0.902, 1},
	"olivedrab":            {0.420, 0.557, 0.137, 1},
	"orangered":            {1, 0.271, 0, 1},
	"orchid":               {0.855, 0.439, 0.839, 1},
	"palegoldenrod":        {0.933, 0.910, 0.667, 1},
	"palegreen":            {0.596, 0.984, 0.596, 1},
	"paleturquoise":        {0.686, 0.933, 0.933, 1},
	"palevioletred":        {0.859, 0.439, 0.576, 1},
	"papayawhip":           {1, 0.937, 0.835, 1},
	"peachpuff":            {1, 0.855, 0.725, 1},
	"peru":                 {0.804, 0.522, 0.247, 1},
	"pink":                 {1, 0.753, 0.796, 1},
	"plum":                 {0.867, 0.627, 0.867, 1},
	"powderblue":           {0.690, 0.878, 0.902, 1},
	"rebeccapurple":        {0.400, 0.200, 0.600, 1},
	"rosybrown":            {0.737, 0.561, 0.561, 1},
	"royalblue":            {0.255, 0.412, 0.882, 1},
	"saddlebrown":          {0.545, 0.271, 0.075, 1},
	"salmon":               {0.980, 0.502, 0.447, 1},
	"sandybrown":           {0.957, 0.643, 0.376, 1},
	"seagreen":             {0.180, 0.545, 0.341, 1},
	"seashell":             {1, 0.961, 0.933, 1},
	"sienna":               {0.627, 0.322, 0.176, 1},
	"skyblue":              {0.529, 0.808, 0.922, 1},
	"slateblue":            {0.416, 0.353, 0.804, 1},
	"slategray":            {0.439, 0.502, 0.565, 1},
	"slategrey":            {0.439, 0.502, 0.565, 1},
	"snow":                 {1, 0.980, 0.980, 1},
	"springgreen":          {0, 1, 0.498, 1},
	"steelblue":            {0.275, 0.510, 0.706, 1},
	"tan":                  {0.824, 0.706, 0.549, 1},
	"thistle":              {0.847, 0.749, 0.847, 1},
	"tomato":               {1, 0.388, 0.278, 1},
	"transparent":          {0, 0, 0, 0},
	"turquoise":            {0.251, 0.878, 0.816, 1},
	"violet":               {0.933, 0.510, 0.933, 1},
	"wheat":                {0.961, 0.871, 0.702, 1},
	"whitesmoke":           {0.961, 0.961, 0.961, 1},
	"yellowgreen":          {0.604, 0.804, 0.196, 1},
	"currentcolor":         {0, 0, 0, 1},
}
