// Package data provides CSS property, at-rule, pseudo-class,
// and pseudo-element definitions as compiled Go literals.
package data

//go:generate go run ../../../cmd/generate-data

// StatusInfo holds the status field shared by all CSS data
// types, providing common status-checking methods.
type StatusInfo struct{ Status string }

// IsDeprecated returns true for obsolete entries.
func (s StatusInfo) IsDeprecated() bool { return s.Status == "obsolete" }

// IsExperimental returns true for experimental entries.
func (s StatusInfo) IsExperimental() bool { return s.Status == "experimental" }

// IsNonstandard returns true for nonstandard entries.
func (s StatusInfo) IsNonstandard() bool { return s.Status == "nonstandard" }

// Property describes a CSS property.
type Property struct {
	Name        string
	Description string
	MDN         string
	Values      []string // common value keywords
	StatusInfo
}

// AtRuleDef describes a CSS at-rule.
type AtRuleDef struct {
	Name        string
	Description string
	StatusInfo
}

// PseudoClass describes a CSS pseudo-class.
type PseudoClass struct {
	Name        string
	Description string
	StatusInfo
}

// PseudoElement describes a CSS pseudo-element.
type PseudoElement struct {
	Name        string
	Description string
	StatusInfo
}

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
