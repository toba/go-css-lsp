package analyzer

import "testing"

func TestAnalyzeUnknownProperty(t *testing.T) {
	src := []byte(`body { colo: red; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags, UnknownPropertyMessage("colo"),
	); !ok {
		t.Error("expected diagnostic for unknown property 'colo'")
	}
}

func TestAnalyzeKnownProperty(t *testing.T) {
	src := []byte(`body { color: red; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags, UnknownPropertyMessage("color"),
	); ok {
		t.Error("should not flag 'color' as unknown")
	}
}

func TestAnalyzeDuplicateProperty(t *testing.T) {
	src := []byte(`body { color: red; color: blue; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags, DuplicatePropertyMessage("color"),
	); !ok {
		t.Error("expected diagnostic for duplicate 'color'")
	}
}

func TestAnalyzeEmptyRuleset(t *testing.T) {
	src := []byte(`body { }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags, EmptyRulesetMsg,
	); !ok {
		t.Error("expected diagnostic for empty ruleset")
	}
}

func TestAnalyzeUnknownAtRule(t *testing.T) {
	src := []byte(`@foobar { }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags, UnknownAtRuleMessage("foobar"),
	); !ok {
		t.Error("expected diagnostic for unknown at-rule")
	}
}

func TestAnalyzeZeroWithUnit(t *testing.T) {
	src := []byte(`body { margin: 0px; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

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
	diags := Analyze(ss, src, LintOptions{})

	for _, d := range diags {
		if d.Message == "unnecessary unit: '0s' can be written as '0'" {
			t.Error("0s should be allowed for time units")
		}
	}
}

func TestAnalyzeImportant(t *testing.T) {
	src := []byte(`body { color: red !important; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags, AvoidImportantMsg,
	); !ok {
		t.Error("expected diagnostic for !important")
	}
}

func TestAnalyzeVendorPrefix(t *testing.T) {
	src := []byte(`body { -webkit-transform: rotate(0); }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags, VendorPrefixMessage("-webkit-transform"),
	); !ok {
		t.Error("expected diagnostic for vendor prefix")
	}
}

func TestAnalyzeCustomProperty(t *testing.T) {
	src := []byte(`:root { --my-color: blue; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags, UnknownPropertyMessage("--my-color"),
	); ok {
		t.Error("custom properties should not be flagged")
	}
}

func TestAnalyzeNesting_NoFalsePositives(t *testing.T) {
	src := []byte(`.parent {
	color: red;
	&:hover { color: blue; }
	.child { font-size: 14px; }
}`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	// Should not flag nested selectors as unknown properties
	for _, d := range diags {
		if d.Severity == SeverityWarning {
			t.Errorf("unexpected warning: %s", d.Message)
		}
	}
}

func TestAnalyzeNesting_NestedNotEmpty(t *testing.T) {
	src := []byte(`.parent { &:hover { color: blue; } }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	// Parent has a nested rule, so it's not empty
	if _, ok := findDiagnostic(
		diags, EmptyRulesetMsg,
	); ok {
		t.Error("parent with nested rules should not be empty")
	}
}

func TestAnalyzeNesting_NestedAtRule(t *testing.T) {
	src := []byte(`.parent { @media (hover) { color: blue; } }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	// Should not produce unknown at-rule for @media
	if _, ok := findDiagnostic(
		diags, UnknownAtRuleMessage("media"),
	); ok {
		t.Error("@media should not be flagged as unknown")
	}
}

func TestAnalyzeUnknownValue_Warn(t *testing.T) {
	src := []byte(`body { justify-content: banana; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags,
		UnknownValueMessage("banana", "justify-content"),
	); !ok {
		t.Error("expected diagnostic for unknown value 'banana'")
	}
}

func TestAnalyzeUnknownValue_ValidValue(t *testing.T) {
	src := []byte(`body { justify-content: center; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags,
		UnknownValueMessage("center", "justify-content"),
	); ok {
		t.Error("'center' is a valid value for justify-content")
	}
}

func TestAnalyzeUnknownValue_GlobalValue(t *testing.T) {
	src := []byte(`body { justify-content: inherit; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags,
		UnknownValueMessage("inherit", "justify-content"),
	); ok {
		t.Error("'inherit' is a global value and should be valid")
	}
}

func TestAnalyzeUnknownValue_NamedColorLenient(t *testing.T) {
	// By default (StrictColorNames=false), named colors are
	// accepted for any property
	src := []byte(`body { justify-content: red; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags,
		UnknownValueMessage("red", "justify-content"),
	); ok {
		t.Error("named colors should be accepted in lenient mode")
	}
}

func TestAnalyzeUnknownValue_NamedColorStrict(t *testing.T) {
	src := []byte(`body { justify-content: red; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{
		StrictColorNames: true,
	})

	if _, ok := findDiagnostic(
		diags,
		UnknownValueMessage("red", "justify-content"),
	); !ok {
		t.Error(
			"'red' should be flagged for justify-content " +
				"in strict color mode",
		)
	}
}

func TestAnalyzeUnknownValue_ColorPropertyAcceptsColor(
	t *testing.T,
) {
	src := []byte(`body { color: red; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{
		StrictColorNames: true,
	})

	if _, ok := findDiagnostic(
		diags,
		UnknownValueMessage("red", "color"),
	); ok {
		t.Error("'red' should be valid for color property")
	}
}

func TestAnalyzeUnknownValue_Ignore(t *testing.T) {
	src := []byte(`body { justify-content: banana; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{
		UnknownValues: UnknownValueIgnore,
	})

	if _, ok := findDiagnostic(
		diags,
		UnknownValueMessage("banana", "justify-content"),
	); ok {
		t.Error("unknown value should be suppressed in ignore mode")
	}
}

func TestAnalyzeUnknownValue_Error(t *testing.T) {
	src := []byte(`body { justify-content: banana; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{
		UnknownValues: UnknownValueError,
	})

	d, ok := findDiagnostic(
		diags,
		UnknownValueMessage("banana", "justify-content"),
	)
	if !ok {
		t.Fatal("expected unknown value diagnostic")
	}
	if d.Severity != SeverityError {
		t.Errorf("expected error severity, got %d", d.Severity)
	}
}

func TestAnalyzeUnknownValue_VarSkipped(t *testing.T) {
	src := []byte(
		`body { justify-content: var(--x); }`,
	)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	for _, d := range diags {
		if d.Message == UnknownValueMessage(
			"var", "justify-content",
		) {
			t.Error("var() should not trigger unknown value")
		}
	}
}

func TestAnalyzeUnknownValue_CustomProperty(t *testing.T) {
	src := []byte(`:root { --foo: banana; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	for _, d := range diags {
		if d.Message == UnknownValueMessage(
			"banana", "--foo",
		) {
			t.Error(
				"custom properties should not have " +
					"value validation",
			)
		}
	}
}

func TestAnalyzeUnknownValue_UnknownProperty(t *testing.T) {
	// Unknown properties should not also get value diagnostics
	src := []byte(`body { foobar: banana; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags,
		UnknownValueMessage("banana", "foobar"),
	); ok {
		t.Error(
			"unknown properties should not get " +
				"value validation",
		)
	}
}

func TestAnalyzeUnknownValue_NoValuesProperty(t *testing.T) {
	// Properties with no defined values (shorthand/complex)
	// should skip value validation
	src := []byte(`body { border: banana; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags,
		UnknownValueMessage("banana", "border"),
	); ok {
		t.Error(
			"shorthand properties with no defined values " +
				"should skip value validation",
		)
	}
}

func TestAnalyzeUnknownValue_TransitionProperty(t *testing.T) {
	src := []byte(
		`body { transition: border-color 0.15s, background 0.15s, color 0.15s; }`,
	)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	for _, d := range diags {
		if d.Message == UnknownValueMessage("border-color", "transition") ||
			d.Message == UnknownValueMessage("background", "transition") ||
			d.Message == UnknownValueMessage("color", "transition") {
			t.Errorf("CSS property names should be accepted in transition: %s", d.Message)
		}
	}
}

func TestAnalyzeUnknownValue_TransitionPropertyProp(
	t *testing.T,
) {
	src := []byte(
		`body { transition-property: opacity, transform; }`,
	)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	for _, d := range diags {
		if d.Message == UnknownValueMessage("opacity", "transition-property") ||
			d.Message == UnknownValueMessage("transform", "transition-property") {
			t.Errorf(
				"CSS property names should be accepted in transition-property: %s",
				d.Message,
			)
		}
	}
}

func TestAnalyzeUnknownValue_WillChange(t *testing.T) {
	src := []byte(`body { will-change: opacity, transform; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	for _, d := range diags {
		if d.Message == UnknownValueMessage("opacity", "will-change") ||
			d.Message == UnknownValueMessage("transform", "will-change") {
			t.Errorf(
				"CSS property names should be accepted in will-change: %s",
				d.Message,
			)
		}
	}
}

func TestAnalyzeUnknownValue_TransitionFakeProperty(
	t *testing.T,
) {
	src := []byte(
		`body { transition: fake-not-a-property 0.15s; }`,
	)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags,
		UnknownValueMessage(
			"fake-not-a-property", "transition",
		),
	); !ok {
		t.Error(
			"fake property names should still be flagged " +
				"in transition",
		)
	}
}

func TestAnalyzeUnknownValue_AnimationName(t *testing.T) {
	src := []byte(`body { animation-name: myAnimation; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	for _, d := range diags {
		if d.Message == UnknownValueMessage("myAnimation", "animation-name") {
			t.Error("animation-name should accept arbitrary identifiers")
		}
	}
}

func TestAnalyzeUnknownValue_Animation(t *testing.T) {
	src := []byte(
		`body { animation: mySlide 0.3s ease-in; }`,
	)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	for _, d := range diags {
		if d.Message == UnknownValueMessage("mySlide", "animation") {
			t.Error("animation should accept arbitrary keyframe names")
		}
	}
}

func TestAnalyzeUnknownValue_OutlineNone(t *testing.T) {
	src := []byte(`body { outline: none; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags,
		UnknownValueMessage("none", "outline"),
	); ok {
		t.Error("'none' is a valid value for outline")
	}
}

func TestAnalyzeUnknownValue_PointerEventsAuto(t *testing.T) {
	src := []byte(`body { pointer-events: auto; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags,
		UnknownValueMessage("auto", "pointer-events"),
	); ok {
		t.Error("'auto' is a valid value for pointer-events")
	}
}

func TestAnalyzeUnknownValue_WhiteSpaceInvalid(t *testing.T) {
	src := []byte(`body { white-space: banana; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags,
		UnknownValueMessage("banana", "white-space"),
	); !ok {
		t.Error("expected diagnostic for unknown value 'banana'")
	}
}

func TestAnalyzeUnknownValue_WhiteSpaceNowrap(t *testing.T) {
	src := []byte(`body { white-space: nowrap; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags,
		UnknownValueMessage("nowrap", "white-space"),
	); ok {
		t.Error("'nowrap' is a valid value for white-space")
	}
}

func TestAnalyzeUnknownValue_GridAreaIdent(t *testing.T) {
	src := []byte(`body { grid-area: header; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags,
		UnknownValueMessage("header", "grid-area"),
	); ok {
		t.Error("grid-area should accept arbitrary grid line names")
	}
}

func TestAnalyzeUnknownValue_FontFamily(t *testing.T) {
	src := []byte(
		`body { font-family: system-ui, -apple-system, Roboto, sans-serif; }`,
	)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	for _, d := range diags {
		if d.Message == UnknownValueMessage("system-ui", "font-family") ||
			d.Message == UnknownValueMessage("-apple-system", "font-family") ||
			d.Message == UnknownValueMessage("Roboto", "font-family") {
			t.Errorf("font-family should accept arbitrary font names: %s", d.Message)
		}
	}
}

func TestAnalyzeUnknownValue_FontFamilyGeneric(t *testing.T) {
	src := []byte(`body { font-family: monospace; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags,
		UnknownValueMessage("monospace", "font-family"),
	); ok {
		t.Error("'monospace' should be valid for font-family")
	}
}

func TestAnalyzeUnknownValue_BackgroundShorthand(t *testing.T) {
	src := []byte(
		`.x { background: url(/img.svg) no-repeat 0.5rem center / 1rem; }`,
	)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	for _, val := range []string{"no-repeat", "center"} {
		if _, ok := findDiagnostic(
			diags,
			UnknownValueMessage(val, "background"),
		); ok {
			t.Errorf(
				"'%s' should be valid in background shorthand",
				val,
			)
		}
	}
}

func TestAnalyzeUnknownValue_BackgroundShorthandInvalid(
	t *testing.T,
) {
	src := []byte(`.x { background: banana; }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{})

	if _, ok := findDiagnostic(
		diags,
		UnknownValueMessage("banana", "background"),
	); !ok {
		t.Error(
			"'banana' should be flagged as unknown in background",
		)
	}
}

func TestAnalyzeDeprecatedProperty_Warn(t *testing.T) {
	src := []byte(`div { clip: rect(0, 0, 0, 0); }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{
		Deprecated: DeprecatedWarn,
	})

	d, ok := findDiagnostic(
		diags, DeprecatedPropertyMessage("clip"),
	)
	if !ok {
		t.Fatal("expected deprecated property diagnostic")
	}
	if d.Severity != SeverityWarning {
		t.Errorf("expected warning severity, got %d", d.Severity)
	}
}

func TestAnalyzeDeprecatedProperty_Ignore(t *testing.T) {
	src := []byte(`div { clip: rect(0, 0, 0, 0); }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{
		Deprecated: DeprecatedIgnore,
	})

	if _, ok := findDiagnostic(
		diags, DeprecatedPropertyMessage("clip"),
	); ok {
		t.Error("deprecated diagnostic should be suppressed")
	}
}

func TestAnalyzeDeprecatedProperty_Error(t *testing.T) {
	src := []byte(`div { clip: rect(0, 0, 0, 0); }`)
	ss := parseCSS(t, src)
	diags := Analyze(ss, src, LintOptions{
		Deprecated: DeprecatedError,
	})

	d, ok := findDiagnostic(
		diags, DeprecatedPropertyMessage("clip"),
	)
	if !ok {
		t.Fatal("expected deprecated property diagnostic")
	}
	if d.Severity != SeverityError {
		t.Errorf("expected error severity, got %d", d.Severity)
	}
}
