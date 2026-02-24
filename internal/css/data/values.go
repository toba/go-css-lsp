package data

// GlobalValues are value keywords valid for any CSS property.
var GlobalValues = []string{
	"inherit",
	"initial",
	"unset",
	"revert",
	"revert-layer",
}

// NamedColors contains all CSS named colors.
var NamedColors = []string{
	"aliceblue", "antiquewhite", "aqua", "aquamarine",
	"azure", "beige", "bisque", "black",
	"blanchedalmond", "blue", "blueviolet", "brown",
	"burlywood", "cadetblue", "chartreuse", "chocolate",
	"coral", "cornflowerblue", "cornsilk", "crimson",
	"cyan", "darkblue", "darkcyan", "darkgoldenrod",
	"darkgray", "darkgreen", "darkgrey", "darkkhaki",
	"darkmagenta", "darkolivegreen", "darkorange",
	"darkorchid", "darkred", "darksalmon",
	"darkseagreen", "darkslateblue", "darkslategray",
	"darkslategrey", "darkturquoise", "darkviolet",
	"deeppink", "deepskyblue", "dimgray", "dimgrey",
	"dodgerblue", "firebrick", "floralwhite",
	"forestgreen", "fuchsia", "gainsboro", "ghostwhite",
	"gold", "goldenrod", "gray", "green", "greenyellow",
	"grey", "honeydew", "hotpink", "indianred", "indigo",
	"ivory", "khaki", "lavender", "lavenderblush",
	"lawngreen", "lemonchiffon", "lightblue",
	"lightcoral", "lightcyan", "lightgoldenrodyellow",
	"lightgray", "lightgreen", "lightgrey", "lightpink",
	"lightsalmon", "lightseagreen", "lightskyblue",
	"lightslategray", "lightslategrey",
	"lightsteelblue", "lightyellow", "lime",
	"limegreen", "linen", "magenta", "maroon",
	"mediumaquamarine", "mediumblue", "mediumorchid",
	"mediumpurple", "mediumseagreen", "mediumslateblue",
	"mediumspringgreen", "mediumturquoise",
	"mediumvioletred", "midnightblue", "mintcream",
	"mistyrose", "moccasin", "navajowhite", "navy",
	"oldlace", "olive", "olivedrab", "orange",
	"orangered", "orchid", "palegoldenrod", "palegreen",
	"paleturquoise", "palevioletred", "papayawhip",
	"peachpuff", "peru", "pink", "plum", "powderblue",
	"purple", "rebeccapurple", "red", "rosybrown",
	"royalblue", "saddlebrown", "salmon", "sandybrown",
	"seagreen", "seashell", "sienna", "silver",
	"skyblue", "slateblue", "slategray", "slategrey",
	"snow", "springgreen", "steelblue", "tan", "teal",
	"thistle", "tomato", "transparent", "turquoise",
	"violet", "wheat", "white", "whitesmoke", "yellow",
	"yellowgreen",
	"currentcolor",
}

// ShorthandLonghands maps shorthand properties to their
// constituent longhand properties, enabling value validation
// by merging longhand values into the shorthand's valid set.
var ShorthandLonghands = map[string][]string{
	"background": {
		"background-attachment", "background-clip",
		"background-color", "background-image",
		"background-origin", "background-position",
		"background-position-x", "background-position-y",
		"background-repeat", "background-size",
	},
}

// Units contains common CSS units.
var Units = []string{
	"px", "em", "rem", "vh", "vw", "vmin", "vmax",
	"%", "ch", "ex", "cm", "mm", "in", "pt", "pc",
	"fr",
	"deg", "rad", "grad", "turn",
	"s", "ms",
	"dpi", "dpcm", "dppx",
	"svh", "svw", "lvh", "lvw", "dvh", "dvw",
	"cqw", "cqh", "cqi", "cqb", "cqmin", "cqmax",
}

// ColorFunctions are CSS functions that produce color values.
var ColorFunctions = []string{
	"rgb", "rgba", "hsl", "hsla",
	"hwb", "lab", "lch", "oklch", "oklab",
	"color", "color-mix", "light-dark",
}

// CommonFunctions are non-color CSS functions.
var CommonFunctions = []string{
	"calc", "min", "max", "clamp",
	"var", "env", "if",
	"url",
	"linear-gradient", "radial-gradient",
	"conic-gradient",
	"repeating-linear-gradient",
	"repeating-radial-gradient",
	"repeating-conic-gradient",
	"image-set",
	"counter", "counters",
	"attr",
	"fit-content", "minmax", "repeat",
	"cubic-bezier", "steps",
	"rotate", "scale", "translate",
	"translateX", "translateY", "translateZ",
	"rotateX", "rotateY", "rotateZ",
	"scaleX", "scaleY", "scaleZ",
	"skew", "skewX", "skewY",
	"matrix", "matrix3d",
	"perspective",
	"blur", "brightness", "contrast",
	"drop-shadow", "grayscale", "hue-rotate",
	"invert", "opacity", "saturate", "sepia",
}
