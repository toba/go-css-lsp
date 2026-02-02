package analyzer

import "testing"

func TestFindCodeActions_UnknownProperty(t *testing.T) {
	diags := []Diagnostic{
		{
			Message:   "unknown property 'colr'",
			StartLine: 1,
			StartChar: 2,
			EndLine:   1,
			EndChar:   6,
			Severity:  SeverityWarning,
		},
	}

	actions := FindCodeActions(diags, nil)

	if len(actions) == 0 {
		t.Fatal("expected at least 1 code action")
	}

	// "color" should be the top suggestion
	found := false
	for _, a := range actions {
		if a.ReplaceWith == "color" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'color' as a suggestion for 'colr'")
	}
}

func TestFindCodeActions_NoActions(t *testing.T) {
	diags := []Diagnostic{
		{
			Message:  "empty ruleset",
			Severity: SeverityHint,
		},
	}

	actions := FindCodeActions(diags, nil)

	if len(actions) != 0 {
		t.Fatalf("expected 0 actions, got %d", len(actions))
	}
}

func TestEditDistance(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"colr", "color", 1},
		{"bakground", "background", 1},
		{"cat", "hat", 1},
		{"kitten", "sitting", 3},
	}

	for _, tt := range tests {
		got := editDistance(tt.a, tt.b)
		if got != tt.want {
			t.Errorf(
				"editDistance(%q, %q) = %d, want %d",
				tt.a, tt.b, got, tt.want,
			)
		}
	}
}

func TestFindFixAllActions_UnnecessaryUnit(t *testing.T) {
	diags := []Diagnostic{
		{
			Message:   "unnecessary unit: '0px' can be written as '0'",
			StartLine: 1,
			StartChar: 10,
			EndLine:   1,
			EndChar:   13,
			Severity:  SeverityHint,
		},
		{
			Message:   "unnecessary unit: '0deg' can be written as '0'",
			StartLine: 2,
			StartChar: 15,
			EndLine:   2,
			EndChar:   19,
			Severity:  SeverityHint,
		},
	}

	actions := FindFixAllActions(diags)

	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}
	for _, a := range actions {
		if a.ReplaceWith != "0" {
			t.Errorf(
				"expected ReplaceWith '0', got %q",
				a.ReplaceWith,
			)
		}
		if a.Kind != CodeActionQuickFix {
			t.Errorf(
				"expected kind %q, got %q",
				CodeActionQuickFix, a.Kind,
			)
		}
	}

	// Verify positions are preserved
	if actions[0].StartLine != 1 || actions[0].StartChar != 10 {
		t.Errorf("first action has wrong start position")
	}
	if actions[1].StartLine != 2 || actions[1].StartChar != 15 {
		t.Errorf("second action has wrong start position")
	}
}

func TestFindFixAllActions_NonFixable(t *testing.T) {
	diags := []Diagnostic{
		{
			Message:  "empty ruleset",
			Severity: SeverityHint,
		},
		{
			Message:  "unknown property 'colr'",
			Severity: SeverityWarning,
		},
	}

	actions := FindFixAllActions(diags)

	if len(actions) != 0 {
		t.Fatalf("expected 0 actions, got %d", len(actions))
	}
}

func TestFindFixAllActions_Mixed(t *testing.T) {
	diags := []Diagnostic{
		{
			Message:   "unnecessary unit: '0px' can be written as '0'",
			StartLine: 1,
			StartChar: 10,
			EndLine:   1,
			EndChar:   13,
			Severity:  SeverityHint,
		},
		{
			Message:  "empty ruleset",
			Severity: SeverityHint,
		},
		{
			Message:  "unknown property 'colr'",
			Severity: SeverityWarning,
		},
	}

	actions := FindFixAllActions(diags)

	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].ReplaceWith != "0" {
		t.Errorf(
			"expected ReplaceWith '0', got %q",
			actions[0].ReplaceWith,
		)
	}
}

func TestFindSimilarProperties(t *testing.T) {
	suggestions := findSimilarProperties("colr")

	if len(suggestions) == 0 {
		t.Fatal("expected suggestions for 'colr'")
	}

	if suggestions[0] != "color" {
		t.Errorf(
			"expected 'color' as top suggestion, got %s",
			suggestions[0],
		)
	}
}
