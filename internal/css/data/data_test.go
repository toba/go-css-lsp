package data

import "testing"

func TestLookupProperty(t *testing.T) {
	p := LookupProperty("color")
	if p == nil {
		t.Fatal("expected to find 'color' property")
	}
	if p.Name != "color" {
		t.Errorf("expected Name 'color', got %q", p.Name)
	}
	if p.Description == "" {
		t.Error("expected non-empty description")
	}
}

func TestLookupPropertyUnknown(t *testing.T) {
	p := LookupProperty("not-a-property")
	if p != nil {
		t.Error("expected nil for unknown property")
	}
}

func TestLookupPropertyVendorPrefixed(t *testing.T) {
	// Vendor-prefixed properties are filtered out during
	// generation, so lookup should return nil.
	p := LookupProperty("-webkit-transform")
	if p != nil {
		t.Error("expected nil for vendor-prefixed property")
	}
}

func TestLookupAtRule(t *testing.T) {
	a := LookupAtRule("media")
	if a == nil {
		t.Fatal("expected to find 'media' at-rule")
	}
	if a.Name != "media" {
		t.Errorf("expected Name 'media', got %q", a.Name)
	}
}

func TestLookupAtRuleUnknown(t *testing.T) {
	a := LookupAtRule("not-a-rule")
	if a != nil {
		t.Error("expected nil for unknown at-rule")
	}
}

func TestLookupPseudoClass(t *testing.T) {
	p := LookupPseudoClass("hover")
	if p == nil {
		t.Fatal("expected to find 'hover' pseudo-class")
	}
	if p.Name != "hover" {
		t.Errorf("expected Name 'hover', got %q", p.Name)
	}
}

func TestLookupPseudoClassUnknown(t *testing.T) {
	p := LookupPseudoClass("not-a-pseudo")
	if p != nil {
		t.Error("expected nil for unknown pseudo-class")
	}
}

func TestLookupPseudoElement(t *testing.T) {
	p := LookupPseudoElement("before")
	if p == nil {
		t.Fatal("expected to find 'before' pseudo-element")
	}
	if p.Name != "before" {
		t.Errorf(
			"expected Name 'before', got %q", p.Name,
		)
	}
}

func TestLookupPseudoElementUnknown(t *testing.T) {
	p := LookupPseudoElement("not-a-pseudo")
	if p != nil {
		t.Error("expected nil for unknown pseudo-element")
	}
}

func TestIsKnownProperty(t *testing.T) {
	if !IsKnownProperty("display") {
		t.Error("expected 'display' to be known")
	}
	if IsKnownProperty("fake-prop") {
		t.Error("expected 'fake-prop' to not be known")
	}
}

func TestIsKnownAtRule(t *testing.T) {
	if !IsKnownAtRule("media") {
		t.Error("expected 'media' to be known")
	}
	if IsKnownAtRule("fake-rule") {
		t.Error("expected 'fake-rule' to not be known")
	}
}

func TestIsKnownPseudoClass(t *testing.T) {
	if !IsKnownPseudoClass("hover") {
		t.Error("expected 'hover' to be known")
	}
	if IsKnownPseudoClass("fake-pseudo") {
		t.Error("expected 'fake-pseudo' to not be known")
	}
}

func TestIsKnownPseudoElement(t *testing.T) {
	if !IsKnownPseudoElement("before") {
		t.Error("expected 'before' to be known")
	}
	if IsKnownPseudoElement("fake-pseudo") {
		t.Error("expected 'fake-pseudo' to not be known")
	}
}

func TestAllProperties(t *testing.T) {
	all := AllProperties()
	if len(all) == 0 {
		t.Fatal("expected non-empty property list")
	}
	// Verify consistency with Lookup
	for _, p := range all {
		if LookupProperty(p.Name) == nil {
			t.Errorf(
				"AllProperties contains %q but Lookup returns nil",
				p.Name,
			)
		}
	}
}

func TestAllAtRules(t *testing.T) {
	all := AllAtRules()
	if len(all) == 0 {
		t.Fatal("expected non-empty at-rule list")
	}
	for _, a := range all {
		if LookupAtRule(a.Name) == nil {
			t.Errorf(
				"AllAtRules contains %q but Lookup returns nil",
				a.Name,
			)
		}
	}
}

func TestAllPseudoClasses(t *testing.T) {
	all := AllPseudoClasses()
	if len(all) == 0 {
		t.Fatal("expected non-empty pseudo-class list")
	}
	for _, p := range all {
		if LookupPseudoClass(p.Name) == nil {
			t.Errorf(
				"AllPseudoClasses contains %q but Lookup returns nil",
				p.Name,
			)
		}
	}
}

func TestAllPseudoElements(t *testing.T) {
	all := AllPseudoElements()
	if len(all) == 0 {
		t.Fatal("expected non-empty pseudo-element list")
	}
	for _, p := range all {
		if LookupPseudoElement(p.Name) == nil {
			t.Errorf(
				"AllPseudoElements contains %q but Lookup returns nil",
				p.Name,
			)
		}
	}
}

func TestStatusInfo_IsDeprecated(t *testing.T) {
	s := StatusInfo{Status: "obsolete"}
	if !s.IsDeprecated() {
		t.Error("expected IsDeprecated() for 'obsolete'")
	}
	s = StatusInfo{Status: "experimental"}
	if s.IsDeprecated() {
		t.Error("unexpected IsDeprecated() for 'experimental'")
	}
	s = StatusInfo{}
	if s.IsDeprecated() {
		t.Error("unexpected IsDeprecated() for empty status")
	}
}

func TestStatusInfo_IsExperimental(t *testing.T) {
	s := StatusInfo{Status: "experimental"}
	if !s.IsExperimental() {
		t.Error("expected IsExperimental() for 'experimental'")
	}
	s = StatusInfo{Status: "obsolete"}
	if s.IsExperimental() {
		t.Error("unexpected IsExperimental() for 'obsolete'")
	}
}

func TestStatusInfo_IsNonstandard(t *testing.T) {
	s := StatusInfo{Status: "nonstandard"}
	if !s.IsNonstandard() {
		t.Error("expected IsNonstandard() for 'nonstandard'")
	}
	s = StatusInfo{Status: "experimental"}
	if s.IsNonstandard() {
		t.Error("unexpected IsNonstandard() for 'experimental'")
	}
}
