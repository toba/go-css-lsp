package analyzer

import (
	"testing"

	"github.com/toba/go-css-lsp/internal/css/parser"
)

func TestAnalyzeUnknownProperty(t *testing.T) {
	src := []byte(`body { colo: red; }`)
	ss, _ := parser.Parse(src)
	diags := Analyze(ss, src)

	found := false
	for _, d := range diags {
		if d.Message == "unknown property 'colo'" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected diagnostic for unknown property 'colo'")
	}
}

func TestAnalyzeKnownProperty(t *testing.T) {
	src := []byte(`body { color: red; }`)
	ss, _ := parser.Parse(src)
	diags := Analyze(ss, src)

	for _, d := range diags {
		if d.Message == "unknown property 'color'" {
			t.Error(
				"should not flag 'color' as unknown",
			)
		}
	}
}

func TestAnalyzeDuplicateProperty(t *testing.T) {
	src := []byte(
		`body { color: red; color: blue; }`,
	)
	ss, _ := parser.Parse(src)
	diags := Analyze(ss, src)

	found := false
	for _, d := range diags {
		if d.Message == "duplicate property 'color'" {
			found = true
			break
		}
	}
	if !found {
		t.Error(
			"expected diagnostic for duplicate 'color'",
		)
	}
}

func TestAnalyzeEmptyRuleset(t *testing.T) {
	src := []byte(`body { }`)
	ss, _ := parser.Parse(src)
	diags := Analyze(ss, src)

	found := false
	for _, d := range diags {
		if d.Message == "empty ruleset" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected diagnostic for empty ruleset")
	}
}

func TestAnalyzeUnknownAtRule(t *testing.T) {
	src := []byte(`@foobar { }`)
	ss, _ := parser.Parse(src)
	diags := Analyze(ss, src)

	found := false
	for _, d := range diags {
		if d.Message == "unknown at-rule '@foobar'" {
			found = true
			break
		}
	}
	if !found {
		t.Error(
			"expected diagnostic for unknown at-rule",
		)
	}
}

func TestAnalyzeCustomProperty(t *testing.T) {
	src := []byte(`:root { --my-color: blue; }`)
	ss, _ := parser.Parse(src)
	diags := Analyze(ss, src)

	for _, d := range diags {
		if d.Message == "unknown property '--my-color'" {
			t.Error(
				"custom properties should not be flagged",
			)
		}
	}
}
