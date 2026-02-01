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

// FindDocumentColors returns all colors found in the CSS
// document.
func FindDocumentColors(
	ss *parser.Stylesheet,
	src []byte,
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
			findColorsInTokens(decl.Value.Tokens, src)...,
		)
		return true
	})

	return colors
}

func findColorsInTokens(
	tokens []scanner.Token,
	src []byte,
) []DocumentColor {
	var colors []DocumentColor

	for i, tok := range tokens {
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
			if c, ok := namedColorMap[strings.ToLower(tok.Value)]; ok {
				colors = append(colors, DocumentColor{
					Color:    c,
					StartPos: tok.Offset,
					EndPos:   tok.End,
				})
			}

		case scanner.Function:
			name := strings.ToLower(tok.Value)
			switch name {
			case "rgb", "rgba":
				if dc, ok := parseColorFunction(
					name, tokens[i:], src,
				); ok {
					colors = append(colors, dc)
				}
			case "hsl", "hsla":
				if dc, ok := parseColorFunction(
					name, tokens[i:], src,
				); ok {
					colors = append(colors, dc)
				}
			case "hwb":
				if dc, ok := parseColorFunction(
					name, tokens[i:], src,
				); ok {
					colors = append(colors, dc)
				}
			}
		}
	}

	return colors
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
	_ []byte,
) (DocumentColor, bool) {
	// Collect numeric arguments until closing paren.
	startPos := tokens[0].Offset
	var endPos int
	var args []float64
	var isPercent []bool
	hasSlash := false

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
			args = append(args, v)
			isPercent = append(isPercent, false)
		case scanner.Percentage:
			v, err := strconv.ParseFloat(tok.Value, 64)
			if err != nil {
				return DocumentColor{}, false
			}
			args = append(args, v)
			isPercent = append(isPercent, true)
		case scanner.Delim:
			if tok.Value == "/" {
				hasSlash = true
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

	// RGB
	if c.Alpha >= 1.0 {
		presentations = append(presentations,
			"rgb("+strconv.Itoa(r)+", "+
				strconv.Itoa(g)+", "+
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

	// HSL
	h, s, l := rgbToHSL(c.Red, c.Green, c.Blue)
	hDeg := int(math.Round(h * 360))
	sPct := int(math.Round(s * 100))
	lPct := int(math.Round(l * 100))

	if c.Alpha >= 1.0 {
		presentations = append(presentations,
			"hsl("+strconv.Itoa(hDeg)+", "+
				strconv.Itoa(sPct)+"%, "+
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
