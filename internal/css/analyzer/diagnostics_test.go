package analyzer

import "testing"

func TestAnalyzeUnknownProperty(t *testing.T) {
	src := []byte(`body { colo: red; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src)

	if _, ok := findDiagnostic(
		diags, UnknownPropertyMessage("colo"),
	); !ok {
		t.Error("expected diagnostic for unknown property 'colo'")
	}
}

func TestAnalyzeKnownProperty(t *testing.T) {
	src := []byte(`body { color: red; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src)

	if _, ok := findDiagnostic(
		diags, UnknownPropertyMessage("color"),
	); ok {
		t.Error("should not flag 'color' as unknown")
	}
}

func TestAnalyzeDuplicateProperty(t *testing.T) {
	src := []byte(`body { color: red; color: blue; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src)

	if _, ok := findDiagnostic(
		diags, DuplicatePropertyMessage("color"),
	); !ok {
		t.Error("expected diagnostic for duplicate 'color'")
	}
}

func TestAnalyzeEmptyRuleset(t *testing.T) {
	src := []byte(`body { }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src)

	if _, ok := findDiagnostic(
		diags, EmptyRulesetMsg,
	); !ok {
		t.Error("expected diagnostic for empty ruleset")
	}
}

func TestAnalyzeUnknownAtRule(t *testing.T) {
	src := []byte(`@foobar { }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src)

	if _, ok := findDiagnostic(
		diags, UnknownAtRuleMessage("foobar"),
	); !ok {
		t.Error("expected diagnostic for unknown at-rule")
	}
}

func TestAnalyzeZeroWithUnit(t *testing.T) {
	src := []byte(`body { margin: 0px; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src)

	found := false
	for _, d := range diags {
		if d.Message == "unnecessary unit: '0px' can be written as '0'" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected diagnostic for 0px")
	}
}

func TestAnalyzeZeroWithUnit_TimeAllowed(t *testing.T) {
	src := []byte(`body { transition-duration: 0s; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src)

	for _, d := range diags {
		if d.Message == "unnecessary unit: '0s' can be written as '0'" {
			t.Error("0s should be allowed for time units")
		}
	}
}

func TestAnalyzeImportant(t *testing.T) {
	src := []byte(`body { color: red !important; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src)

	if _, ok := findDiagnostic(
		diags, AvoidImportantMsg,
	); !ok {
		t.Error("expected diagnostic for !important")
	}
}

func TestAnalyzeVendorPrefix(t *testing.T) {
	src := []byte(`body { -webkit-transform: rotate(0); }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src)

	if _, ok := findDiagnostic(
		diags, VendorPrefixMessage("-webkit-transform"),
	); !ok {
		t.Error("expected diagnostic for vendor prefix")
	}
}

func TestAnalyzeCustomProperty(t *testing.T) {
	src := []byte(`:root { --my-color: blue; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src)

	if _, ok := findDiagnostic(
		diags, UnknownPropertyMessage("--my-color"),
	); ok {
		t.Error("custom properties should not be flagged")
	}
}
