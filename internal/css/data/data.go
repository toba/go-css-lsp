// Package data provides CSS property, at-rule, pseudo-class,
// and pseudo-element definitions as compiled Go literals.
package data

//go:generate go run ../../../cmd/generate-data

// Property describes a CSS property.
type Property struct {
	Name        string
	Description string
	MDN         string
	Values      []string // common value keywords
	Status      string   // "obsolete", "experimental", "nonstandard", or "" (standard)
}

// IsDeprecated returns true for obsolete properties.
func (p *Property) IsDeprecated() bool { return p.Status == "obsolete" }

// IsExperimental returns true for experimental properties.
func (p *Property) IsExperimental() bool { return p.Status == "experimental" }

// IsNonstandard returns true for nonstandard properties.
func (p *Property) IsNonstandard() bool { return p.Status == "nonstandard" }

// AtRuleDef describes a CSS at-rule.
type AtRuleDef struct {
	Name        string
	Description string
	Status      string
}

// IsExperimental returns true for experimental at-rules.
func (a AtRuleDef) IsExperimental() bool { return a.Status == "experimental" }

// IsDeprecated returns true for obsolete at-rules.
func (a AtRuleDef) IsDeprecated() bool { return a.Status == "obsolete" }

// PseudoClass describes a CSS pseudo-class.
type PseudoClass struct {
	Name        string
	Description string
	Status      string
}

// IsExperimental returns true for experimental pseudo-classes.
func (p PseudoClass) IsExperimental() bool { return p.Status == "experimental" }

// IsDeprecated returns true for obsolete pseudo-classes.
func (p PseudoClass) IsDeprecated() bool { return p.Status == "obsolete" }

// PseudoElement describes a CSS pseudo-element.
type PseudoElement struct {
	Name        string
	Description string
	Status      string
}

// IsExperimental returns true for experimental pseudo-elements.
func (p PseudoElement) IsExperimental() bool { return p.Status == "experimental" }

// IsDeprecated returns true for obsolete pseudo-elements.
func (p PseudoElement) IsDeprecated() bool { return p.Status == "obsolete" }

// Function describes a CSS function.
type Function struct {
	Name        string
	Description string
	MDN         string
	Signatures  []string // human-readable overloads
}

// LookupProperty returns the property definition or nil.
func LookupProperty(name string) *Property {
	p, ok := propertyMap[name]
	if !ok {
		return nil
	}
	return &p
}

// LookupAtRule returns the at-rule definition or nil.
func LookupAtRule(name string) *AtRuleDef {
	a, ok := atRuleMap[name]
	if !ok {
		return nil
	}
	return &a
}

// LookupPseudoClass returns the pseudo-class definition or nil.
func LookupPseudoClass(name string) *PseudoClass {
	p, ok := pseudoClassMap[name]
	if !ok {
		return nil
	}
	return &p
}

// LookupPseudoElement returns the pseudo-element definition or
// nil.
func LookupPseudoElement(name string) *PseudoElement {
	p, ok := pseudoElementMap[name]
	if !ok {
		return nil
	}
	return &p
}

// AllProperties returns all known property definitions.
func AllProperties() []Property {
	return Properties
}

// AllAtRules returns all known at-rule definitions.
func AllAtRules() []AtRuleDef {
	return AtRules
}

// AllPseudoClasses returns all known pseudo-class definitions.
func AllPseudoClasses() []PseudoClass {
	return PseudoClasses
}

// AllPseudoElements returns all known pseudo-element
// definitions.
func AllPseudoElements() []PseudoElement {
	return PseudoElements
}

// IsKnownProperty returns whether the name is a known CSS
// property.
func IsKnownProperty(name string) bool {
	_, ok := propertyMap[name]
	return ok
}

// IsKnownAtRule returns whether the name is a known at-rule.
func IsKnownAtRule(name string) bool {
	_, ok := atRuleMap[name]
	return ok
}

// IsKnownPseudoClass returns whether the name is a known
// pseudo-class.
func IsKnownPseudoClass(name string) bool {
	_, ok := pseudoClassMap[name]
	return ok
}

// IsKnownPseudoElement returns whether the name is a known
// pseudo-element.
func IsKnownPseudoElement(name string) bool {
	_, ok := pseudoElementMap[name]
	return ok
}

// LookupFunction returns the function definition or nil.
func LookupFunction(name string) *Function {
	f, ok := functionMap[name]
	if !ok {
		return nil
	}
	return &f
}

// AllFunctions returns all known function definitions.
func AllFunctions() []Function {
	return Functions
}

// IsKnownFunction returns whether the name is a known CSS
// function.
func IsKnownFunction(name string) bool {
	_, ok := functionMap[name]
	return ok
}
