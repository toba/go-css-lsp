package data

// PseudoClasses contains definitions for CSS pseudo-classes.
var PseudoClasses = []PseudoClass{
	{
		Name:        "active",
		Description: "Matches when an element is being activated by the user (e.g., clicked).",
	},
	{
		Name:        "any-link",
		Description: "Matches every <a> or <area> element that has an href attribute.",
	},
	{
		Name:        "autofill",
		Description: "Matches when an input has been autofilled by the browser.",
	},
	{
		Name:        "checked",
		Description: "Matches any radio, checkbox, or option element that is checked or toggled on.",
	},
	{Name: "default", Description: "Matches the default form element in a group."},
	{Name: "defined", Description: "Matches any element that has been defined."},
	{Name: "dir", Description: "Matches elements based on their directionality."},
	{Name: "disabled", Description: "Matches any disabled element."},
	{Name: "empty", Description: "Matches any element that has no children."},
	{Name: "enabled", Description: "Matches any enabled element."},
	{Name: "first-child", Description: "Matches the first child element."},
	{
		Name:        "first-of-type",
		Description: "Matches the first element of its type among siblings.",
	},
	{Name: "focus", Description: "Matches when an element has focus."},
	{
		Name:        "focus-visible",
		Description: "Matches when an element has focus and the focus should be visually indicated.",
	},
	{
		Name:        "focus-within",
		Description: "Matches when an element or any of its descendants has focus.",
	},
	{
		Name:        "fullscreen",
		Description: "Matches an element that is currently in fullscreen mode.",
	},
	{
		Name:        "has",
		Description: "Matches if any of the relative selectors match when anchored against this element.",
	},
	{Name: "hover", Description: "Matches when the user hovers over an element."},
	{
		Name:        "in-range",
		Description: "Matches when an element's value is within its specified range.",
	},
	{
		Name:        "indeterminate",
		Description: "Matches form elements whose state is indeterminate.",
	},
	{
		Name:        "invalid",
		Description: "Matches any element whose contents fail to validate.",
	},
	{
		Name:        "is",
		Description: "Matches any element that can be selected by one of the selectors in its argument list.",
	},
	{Name: "lang", Description: "Matches elements based on their language."},
	{Name: "last-child", Description: "Matches the last child element."},
	{
		Name:        "last-of-type",
		Description: "Matches the last element of its type among siblings.",
	},
	{Name: "link", Description: "Matches unvisited links."},
	{
		Name:        "modal",
		Description: "Matches an element that is in a state in which it excludes all interaction with elements outside it.",
	},
	{
		Name:        "not",
		Description: "Matches elements that do not match the given selector list.",
	},
	{
		Name:        "nth-child",
		Description: "Matches elements based on their position among siblings.",
	},
	{Name: "nth-last-child", Description: "Matches elements counting from the end."},
	{
		Name:        "nth-last-of-type",
		Description: "Matches elements of a type counting from the end.",
	},
	{Name: "nth-of-type", Description: "Matches elements of a type based on position."},
	{Name: "only-child", Description: "Matches an element that is the only child."},
	{
		Name:        "only-of-type",
		Description: "Matches an element that is the only one of its type.",
	},
	{Name: "optional", Description: "Matches input elements that are not required."},
	{
		Name:        "out-of-range",
		Description: "Matches when an element's value is outside its specified range.",
	},
	{
		Name:        "placeholder-shown",
		Description: "Matches an input element that is displaying placeholder text.",
	},
	{
		Name:        "read-only",
		Description: "Matches elements that are not editable by the user.",
	},
	{Name: "read-write", Description: "Matches elements that are editable by the user."},
	{Name: "required", Description: "Matches input elements that are required."},
	{Name: "root", Description: "Matches the root element of the document."},
	{Name: "target", Description: "Matches the element targeted by the URL's fragment."},
	{
		Name:        "valid",
		Description: "Matches any element whose contents validate successfully.",
	},
	{Name: "visited", Description: "Matches visited links."},
	{Name: "where", Description: "Like :is() but with zero specificity."},
}

// PseudoElements contains definitions for CSS pseudo-elements.
var PseudoElements = []PseudoElement{
	{
		Name:        "after",
		Description: "Creates a pseudo-element that is the last child of the selected element.",
	},
	{
		Name:        "backdrop",
		Description: "A box rendered immediately beneath any element rendered in the top layer.",
	},
	{
		Name:        "before",
		Description: "Creates a pseudo-element that is the first child of the selected element.",
	},
	{Name: "cue", Description: "Matches WebVTT cues."},
	{
		Name:        "file-selector-button",
		Description: "Represents the button of an <input> of type=\"file\".",
	},
	{
		Name:        "first-letter",
		Description: "Selects the first letter of the first line of a block-level element.",
	},
	{Name: "first-line", Description: "Selects the first line of a block-level element."},
	{Name: "marker", Description: "Selects the marker box of a list item."},
	{
		Name:        "placeholder",
		Description: "Represents the placeholder text in an input or textarea element.",
	},
	{
		Name:        "selection",
		Description: "Applies styles to the portion of a document highlighted by the user.",
	},
}

var pseudoClassMap = buildPseudoClassMap()

func buildPseudoClassMap() map[string]PseudoClass {
	m := make(map[string]PseudoClass, len(PseudoClasses))
	for _, p := range PseudoClasses {
		m[p.Name] = p
	}
	return m
}

var pseudoElementMap = buildPseudoElementMap()

func buildPseudoElementMap() map[string]PseudoElement {
	m := make(map[string]PseudoElement, len(PseudoElements))
	for _, p := range PseudoElements {
		m[p.Name] = p
	}
	return m
}
