package analyzer

// Diagnostic message prefixes and builders.
const (
	UnknownPropertyPrefix = "unknown property '"
	DuplicatePropertyMsg  = "duplicate property '"
	EmptyRulesetMsg       = "empty ruleset"
	AvoidImportantMsg     = "avoid using !important"
	VendorPrefixPrefix    = "vendor prefix '"
	UnknownAtRulePrefix   = "unknown at-rule '@"
)

// UnknownPropertyMessage returns a diagnostic message for an
// unknown CSS property.
func UnknownPropertyMessage(name string) string {
	return UnknownPropertyPrefix + name + "'"
}

// DuplicatePropertyMessage returns a diagnostic message for a
// duplicate CSS property.
func DuplicatePropertyMessage(name string) string {
	return DuplicatePropertyMsg + name + "'"
}

// VendorPrefixMessage returns a diagnostic message for a
// vendor-prefixed property.
func VendorPrefixMessage(name string) string {
	return VendorPrefixPrefix + name + "' may not be needed"
}

// UnknownAtRuleMessage returns a diagnostic message for an
// unknown at-rule.
func UnknownAtRuleMessage(name string) string {
	return UnknownAtRulePrefix + name + "'"
}

// ExperimentalPropertyMessage returns a diagnostic message for
// an experimental CSS property.
func ExperimentalPropertyMessage(name string) string {
	return "experimental property '" + name + "'"
}

// DeprecatedPropertyMessage returns a diagnostic message for
// a deprecated (obsolete) CSS property.
func DeprecatedPropertyMessage(name string) string {
	return "deprecated property '" + name + "'"
}

// UnknownValueMessage returns a diagnostic message for an
// unrecognized value keyword.
func UnknownValueMessage(value, property string) string {
	return "unknown value '" + value +
		"' for property '" + property + "'"
}
