package data

// Functions contains definitions for all known CSS functions.
var Functions = []Function{
	// Color functions
	{
		Name:        "rgb",
		Description: "Specifies a color using red, green, blue (and alpha) values.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/color_value/rgb",
		Signatures: []string{
			"rgb(<red>, <green>, <blue>)",
			"rgb(<red> <green> <blue>)",
			"rgb(<red> <green> <blue> / <alpha>)",
			"rgb(from <color> r g b / <alpha>)",
		},
	},
	{
		Name:        "rgba",
		Description: "Specifies a color using red, green, blue (and alpha) values.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/color_value/rgb",
		Signatures: []string{
			"rgba(<red>, <green>, <blue>)",
			"rgba(<red>, <green>, <blue>, <alpha>)",
		},
	},
	{
		Name:        "hsl",
		Description: "Specifies a color using hue, saturation, lightness (and alpha) values.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/color_value/hsl",
		Signatures: []string{
			"hsl(<hue>, <saturation>, <lightness>)",
			"hsl(<hue> <saturation> <lightness>)",
			"hsl(<hue> <saturation> <lightness> / <alpha>)",
			"hsl(from <color> h s l / <alpha>)",
		},
	},
	{
		Name:        "hsla",
		Description: "Specifies a color using hue, saturation, lightness (and alpha) values.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/color_value/hsl",
		Signatures: []string{
			"hsla(<hue>, <saturation>, <lightness>)",
			"hsla(<hue>, <saturation>, <lightness>, <alpha>)",
		},
	},
	{
		Name:        "hwb",
		Description: "Specifies a color using hue, whiteness, blackness values.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/color_value/hwb",
		Signatures: []string{
			"hwb(<hue> <whiteness> <blackness>)",
			"hwb(<hue> <whiteness> <blackness> / <alpha>)",
			"hwb(from <color> h w b / <alpha>)",
		},
	},
	{
		Name:        "lab",
		Description: "Specifies a color in the CIE Lab color space.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/color_value/lab",
		Signatures: []string{
			"lab(<lightness> <a> <b>)",
			"lab(<lightness> <a> <b> / <alpha>)",
			"lab(from <color> l a b / <alpha>)",
		},
	},
	{
		Name:        "lch",
		Description: "Specifies a color in the CIE LCH color space.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/color_value/lch",
		Signatures: []string{
			"lch(<lightness> <chroma> <hue>)",
			"lch(<lightness> <chroma> <hue> / <alpha>)",
			"lch(from <color> l c h / <alpha>)",
		},
	},
	{
		Name:        "oklch",
		Description: "Specifies a color in the OKLCH color space.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/color_value/oklch",
		Signatures: []string{
			"oklch(<lightness> <chroma> <hue>)",
			"oklch(<lightness> <chroma> <hue> / <alpha>)",
			"oklch(from <color> l c h / <alpha>)",
		},
	},
	{
		Name:        "oklab",
		Description: "Specifies a color in the OKLAB color space.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/color_value/oklab",
		Signatures: []string{
			"oklab(<lightness> <a> <b>)",
			"oklab(<lightness> <a> <b> / <alpha>)",
			"oklab(from <color> l a b / <alpha>)",
		},
	},
	{
		Name:        "color",
		Description: "Specifies a color in a given color space.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/color_value/color",
		Signatures: []string{
			"color(<colorspace> <c1> <c2> <c3>)",
			"color(<colorspace> <c1> <c2> <c3> / <alpha>)",
			"color(from <color> <colorspace> c1 c2 c3 / <alpha>)",
		},
	},
	{
		Name:        "color-mix",
		Description: "Mixes two colors in a given color space.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/color_value/color-mix",
		Signatures: []string{
			"color-mix(in <colorspace>, <color1>, <color2>)",
			"color-mix(in <colorspace>, <color1> <percentage>, <color2> <percentage>)",
		},
	},
	{
		Name:        "light-dark",
		Description: "Returns one of two colors depending on the user's color scheme preference.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/color_value/light-dark",
		Signatures: []string{
			"light-dark(<light-color>, <dark-color>)",
		},
	},

	// Math functions
	{
		Name:        "calc",
		Description: "Performs calculations to determine CSS property values.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/calc",
		Signatures: []string{
			"calc(<expression>)",
		},
	},
	{
		Name:        "min",
		Description: "Returns the smallest of the given values.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/min",
		Signatures: []string{
			"min(<value>, <value>, ...)",
		},
	},
	{
		Name:        "max",
		Description: "Returns the largest of the given values.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/max",
		Signatures: []string{
			"max(<value>, <value>, ...)",
		},
	},
	{
		Name:        "clamp",
		Description: "Clamps a value between a minimum and maximum.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/clamp",
		Signatures: []string{
			"clamp(<min>, <preferred>, <max>)",
		},
	},

	// Conditional
	{
		Name:        "if",
		Description: "Evaluates a conditional expression, returning one of the given values based on a condition.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/if",
		Signatures: []string{
			"if(<condition>: <value>; else: <value>)",
			"if(<condition>: <value>; <condition>: <value>; else: <value>)",
		},
	},

	// Custom properties
	{
		Name:        "var",
		Description: "Inserts the value of a CSS custom property.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/var",
		Signatures: []string{
			"var(<custom-property>)",
			"var(<custom-property>, <fallback>)",
		},
	},
	{
		Name:        "env",
		Description: "Inserts the value of a user-agent defined environment variable.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/env",
		Signatures: []string{
			"env(<variable>)",
			"env(<variable>, <fallback>)",
		},
	},

	// Resource
	{
		Name:        "url",
		Description: "References a resource by URL.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/url",
		Signatures: []string{
			"url(<string>)",
		},
	},

	// Gradients
	{
		Name:        "linear-gradient",
		Description: "Creates a linear gradient image.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/gradient/linear-gradient",
		Signatures: []string{
			"linear-gradient(<color-stop>, <color-stop>, ...)",
			"linear-gradient(<angle>, <color-stop>, <color-stop>, ...)",
			"linear-gradient(to <direction>, <color-stop>, <color-stop>, ...)",
		},
	},
	{
		Name:        "radial-gradient",
		Description: "Creates a radial gradient image.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/gradient/radial-gradient",
		Signatures: []string{
			"radial-gradient(<color-stop>, <color-stop>, ...)",
			"radial-gradient(<shape> <size> at <position>, <color-stop>, ...)",
		},
	},
	{
		Name:        "conic-gradient",
		Description: "Creates a conic gradient image.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/gradient/conic-gradient",
		Signatures: []string{
			"conic-gradient(<color-stop>, <color-stop>, ...)",
			"conic-gradient(from <angle> at <position>, <color-stop>, ...)",
		},
	},
	{
		Name:        "repeating-linear-gradient",
		Description: "Creates a repeating linear gradient image.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/gradient/repeating-linear-gradient",
		Signatures: []string{
			"repeating-linear-gradient(<color-stop>, <color-stop>, ...)",
			"repeating-linear-gradient(<angle>, <color-stop>, <color-stop>, ...)",
		},
	},
	{
		Name:        "repeating-radial-gradient",
		Description: "Creates a repeating radial gradient image.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/gradient/repeating-radial-gradient",
		Signatures: []string{
			"repeating-radial-gradient(<color-stop>, <color-stop>, ...)",
			"repeating-radial-gradient(<shape> <size> at <position>, <color-stop>, ...)",
		},
	},
	{
		Name:        "repeating-conic-gradient",
		Description: "Creates a repeating conic gradient image.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/gradient/repeating-conic-gradient",
		Signatures: []string{
			"repeating-conic-gradient(<color-stop>, <color-stop>, ...)",
			"repeating-conic-gradient(from <angle> at <position>, <color-stop>, ...)",
		},
	},

	// Image
	{
		Name:        "image-set",
		Description: "Provides a set of images for the browser to choose from based on resolution.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/image/image-set",
		Signatures: []string{
			"image-set(<image> <resolution>, ...)",
		},
	},

	// Counter
	{
		Name:        "counter",
		Description: "Returns the current value of a CSS counter.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/counter",
		Signatures: []string{
			"counter(<name>)",
			"counter(<name>, <style>)",
		},
	},
	{
		Name:        "counters",
		Description: "Returns the values of nested CSS counters.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/counters",
		Signatures: []string{
			"counters(<name>, <string>)",
			"counters(<name>, <string>, <style>)",
		},
	},

	// Attribute
	{
		Name:        "attr",
		Description: "Returns the value of an HTML attribute as a CSS value.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/attr",
		Signatures: []string{
			"attr(<attribute-name>)",
			"attr(<attribute-name> <type-or-unit>)",
			"attr(<attribute-name> <type-or-unit>, <fallback>)",
		},
	},

	// Grid
	{
		Name:        "fit-content",
		Description: "Clamps a size to a maximum, using available space down to a minimum.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/fit-content",
		Signatures: []string{
			"fit-content(<length-percentage>)",
		},
	},
	{
		Name:        "minmax",
		Description: "Defines a size range for grid tracks.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/minmax",
		Signatures: []string{
			"minmax(<min>, <max>)",
		},
	},
	{
		Name:        "repeat",
		Description: "Repeats a track list pattern for grid layouts.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/repeat",
		Signatures: []string{
			"repeat(<count>, <tracks>)",
			"repeat(auto-fill, <tracks>)",
			"repeat(auto-fit, <tracks>)",
		},
	},

	// Easing
	{
		Name:        "cubic-bezier",
		Description: "Defines a cubic Bézier easing curve.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/easing-function/cubic-bezier",
		Signatures: []string{
			"cubic-bezier(<x1>, <y1>, <x2>, <y2>)",
		},
	},
	{
		Name:        "steps",
		Description: "Defines a stepped easing function.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/easing-function/steps",
		Signatures: []string{
			"steps(<count>)",
			"steps(<count>, <direction>)",
		},
	},

	// Transform
	{
		Name:        "rotate",
		Description: "Rotates an element around a fixed point.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/transform-function/rotate",
		Signatures: []string{
			"rotate(<angle>)",
		},
	},
	{
		Name:        "rotateX",
		Description: "Rotates an element around the X axis.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/transform-function/rotateX",
		Signatures: []string{
			"rotateX(<angle>)",
		},
	},
	{
		Name:        "rotateY",
		Description: "Rotates an element around the Y axis.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/transform-function/rotateY",
		Signatures: []string{
			"rotateY(<angle>)",
		},
	},
	{
		Name:        "rotateZ",
		Description: "Rotates an element around the Z axis.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/transform-function/rotateZ",
		Signatures: []string{
			"rotateZ(<angle>)",
		},
	},
	{
		Name:        "scale",
		Description: "Scales an element up or down.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/transform-function/scale",
		Signatures: []string{
			"scale(<sx>)",
			"scale(<sx>, <sy>)",
		},
	},
	{
		Name:        "scaleX",
		Description: "Scales an element along the X axis.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/transform-function/scaleX",
		Signatures: []string{
			"scaleX(<sx>)",
		},
	},
	{
		Name:        "scaleY",
		Description: "Scales an element along the Y axis.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/transform-function/scaleY",
		Signatures: []string{
			"scaleY(<sy>)",
		},
	},
	{
		Name:        "scaleZ",
		Description: "Scales an element along the Z axis.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/transform-function/scaleZ",
		Signatures: []string{
			"scaleZ(<sz>)",
		},
	},
	{
		Name:        "translate",
		Description: "Translates an element's position.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/transform-function/translate",
		Signatures: []string{
			"translate(<tx>)",
			"translate(<tx>, <ty>)",
		},
	},
	{
		Name:        "translateX",
		Description: "Translates an element along the X axis.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/transform-function/translateX",
		Signatures: []string{
			"translateX(<tx>)",
		},
	},
	{
		Name:        "translateY",
		Description: "Translates an element along the Y axis.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/transform-function/translateY",
		Signatures: []string{
			"translateY(<ty>)",
		},
	},
	{
		Name:        "translateZ",
		Description: "Translates an element along the Z axis.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/transform-function/translateZ",
		Signatures: []string{
			"translateZ(<tz>)",
		},
	},
	{
		Name:        "skew",
		Description: "Skews an element on the 2D plane.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/transform-function/skew",
		Signatures: []string{
			"skew(<ax>)",
			"skew(<ax>, <ay>)",
		},
	},
	{
		Name:        "skewX",
		Description: "Skews an element along the X axis.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/transform-function/skewX",
		Signatures: []string{
			"skewX(<angle>)",
		},
	},
	{
		Name:        "skewY",
		Description: "Skews an element along the Y axis.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/transform-function/skewY",
		Signatures: []string{
			"skewY(<angle>)",
		},
	},
	{
		Name:        "matrix",
		Description: "Defines a 2D transformation matrix.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/transform-function/matrix",
		Signatures: []string{
			"matrix(<a>, <b>, <c>, <d>, <tx>, <ty>)",
		},
	},
	{
		Name:        "matrix3d",
		Description: "Defines a 3D transformation matrix.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/transform-function/matrix3d",
		Signatures: []string{
			"matrix3d(<a1>, <b1>, <c1>, <d1>, <a2>, <b2>, <c2>, <d2>, <a3>, <b3>, <c3>, <d3>, <a4>, <b4>, <c4>, <d4>)",
		},
	},
	{
		Name:        "perspective",
		Description: "Sets the distance between the user and the z=0 plane.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/transform-function/perspective",
		Signatures: []string{
			"perspective(<length>)",
		},
	},

	// Filter functions
	{
		Name:        "blur",
		Description: "Applies a Gaussian blur to an element.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/filter-function/blur",
		Signatures: []string{
			"blur(<radius>)",
		},
	},
	{
		Name:        "brightness",
		Description: "Adjusts the brightness of an element.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/filter-function/brightness",
		Signatures: []string{
			"brightness(<amount>)",
		},
	},
	{
		Name:        "contrast",
		Description: "Adjusts the contrast of an element.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/filter-function/contrast",
		Signatures: []string{
			"contrast(<amount>)",
		},
	},
	{
		Name:        "drop-shadow",
		Description: "Applies a drop shadow to an element.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/filter-function/drop-shadow",
		Signatures: []string{
			"drop-shadow(<offset-x> <offset-y> <blur-radius> <color>)",
		},
	},
	{
		Name:        "grayscale",
		Description: "Converts an element to grayscale.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/filter-function/grayscale",
		Signatures: []string{
			"grayscale(<amount>)",
		},
	},
	{
		Name:        "hue-rotate",
		Description: "Rotates the hue of an element.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/filter-function/hue-rotate",
		Signatures: []string{
			"hue-rotate(<angle>)",
		},
	},
	{
		Name:        "invert",
		Description: "Inverts the colors of an element.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/filter-function/invert",
		Signatures: []string{
			"invert(<amount>)",
		},
	},
	{
		Name:        "opacity",
		Description: "Adjusts the opacity of an element.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/filter-function/opacity",
		Signatures: []string{
			"opacity(<amount>)",
		},
	},
	{
		Name:        "saturate",
		Description: "Adjusts the saturation of an element.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/filter-function/saturate",
		Signatures: []string{
			"saturate(<amount>)",
		},
	},
	{
		Name:        "sepia",
		Description: "Applies a sepia tone to an element.",
		MDN:         "https://developer.mozilla.org/docs/Web/CSS/filter-function/sepia",
		Signatures: []string{
			"sepia(<amount>)",
		},
	},
}

var functionMap = buildFunctionMap()

func buildFunctionMap() map[string]Function {
	m := make(map[string]Function, len(Functions))
	for _, f := range Functions {
		m[f.Name] = f
	}
	return m
}
