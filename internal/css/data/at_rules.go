package data

// AtRules contains definitions for CSS at-rules.
var AtRules = []AtRuleDef{
	{Name: "charset", Description: "Specifies the character encoding of the stylesheet."},
	{
		Name:        "container",
		Description: "A conditional group rule that applies styles based on the size of a containment context.",
	},
	{
		Name:        "counter-style",
		Description: "Defines how to convert a counter value into a string representation.",
	},
	{Name: "font-face", Description: "Specifies a custom font to be used for text."},
	{
		Name:        "font-feature-values",
		Description: "Defines named values for font-variant-alternates.",
	},
	{Name: "import", Description: "Imports an external stylesheet."},
	{Name: "keyframes", Description: "Defines the keyframes for an animation sequence."},
	{Name: "layer", Description: "Declares a cascade layer."},
	{
		Name:        "media",
		Description: "A conditional group rule that applies styles based on media queries.",
	},
	{
		Name:        "namespace",
		Description: "Defines XML namespaces to be used in the stylesheet.",
	},
	{Name: "page", Description: "Specifies styles for printed pages."},
	{
		Name:        "property",
		Description: "Defines a custom CSS property with type checking, inheritance, and initial value.",
	},
	{Name: "scope", Description: "Defines a scoping context for selectors."},
	{
		Name:        "starting-style",
		Description: "Defines starting styles for elements that are being animated.",
	},
	{
		Name:        "supports",
		Description: "A conditional group rule that applies styles based on feature support.",
	},
}

var atRuleMap = buildAtRuleMap()

func buildAtRuleMap() map[string]AtRuleDef {
	m := make(map[string]AtRuleDef, len(AtRules))
	for _, a := range AtRules {
		m[a.Name] = a
	}
	return m
}
