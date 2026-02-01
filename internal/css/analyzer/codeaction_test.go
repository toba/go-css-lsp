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
