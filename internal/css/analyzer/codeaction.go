package analyzer

import (
	"strings"

	"github.com/toba/go-css-lsp/internal/css/data"
)

// CodeActionKind constants.
const (
	CodeActionQuickFix     = "quickfix"
	CodeActionRefactor     = "refactor"
	CodeActionSourceFixAll = "source.fixAll"
)

// UnnecessaryUnitPrefix is the message prefix for the
// unnecessary unit diagnostic.
const UnnecessaryUnitPrefix = "unnecessary unit: '"

// CodeAction represents a code action (quick fix).
type CodeAction struct {
	Title       string
	Kind        string
	StartLine   int
	StartChar   int
	EndLine     int
	EndChar     int
	ReplaceWith string
}

// FindCodeActions returns code actions for the given
// diagnostics range.
func FindCodeActions(
	diags []Diagnostic,
	src []byte,
) []CodeAction {
	var actions []CodeAction

	for _, d := range diags {
		if strings.HasPrefix(d.Message, UnknownPropertyPrefix) {
			propName := extractQuotedName(d.Message)
			if propName == "" {
				continue
			}
			suggestions := findSimilarProperties(propName)
			for _, suggestion := range suggestions {
				actions = append(actions, CodeAction{
					Title:       "Replace with '" + suggestion + "'",
					Kind:        CodeActionQuickFix,
					StartLine:   d.StartLine,
					StartChar:   d.StartChar,
					EndLine:     d.EndLine,
					EndChar:     d.EndChar,
					ReplaceWith: suggestion,
				})
			}
		}
	}

	return actions
}

func extractQuotedName(msg string) string {
	start := strings.Index(msg, "'")
	if start < 0 {
		return ""
	}
	end := strings.LastIndex(msg, "'")
	if end <= start {
		return ""
	}
	return msg[start+1 : end]
}

// findSimilarProperties returns up to 3 property names
// similar to the given unknown property.
func findSimilarProperties(name string) []string {
	type scored struct {
		name  string
		score int
	}

	var candidates []scored

	for _, prop := range data.AllProperties() {
		d := editDistance(name, prop.Name)
		// Only suggest if distance is small relative to name
		// length
		maxDist := max(len(name)/3, 2)
		if d <= maxDist {
			candidates = append(candidates, scored{
				name: prop.Name, score: d,
			})
		}
	}

	// Sort by score (simple selection sort, list is small)
	for i := range candidates {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].score < candidates[i].score {
				candidates[i], candidates[j] =
					candidates[j], candidates[i]
			}
		}
	}

	limit := min(len(candidates), 3)

	result := make([]string, limit)
	for i := range limit {
		result[i] = candidates[i].name
	}
	return result
}

// FindFixAllActions returns quick-fix code actions for all
// auto-fixable diagnostics.
func FindFixAllActions(diags []Diagnostic) []CodeAction {
	var actions []CodeAction
	for _, d := range diags {
		if a, ok := fixForDiagnostic(d); ok {
			actions = append(actions, a)
		}
	}
	return actions
}

// fixForDiagnostic returns a code action if the diagnostic is
// auto-fixable.
func fixForDiagnostic(d Diagnostic) (CodeAction, bool) {
	if strings.HasPrefix(d.Message, UnnecessaryUnitPrefix) {
		return CodeAction{
			Title:       "Remove unnecessary unit",
			Kind:        CodeActionQuickFix,
			StartLine:   d.StartLine,
			StartChar:   d.StartChar,
			EndLine:     d.EndLine,
			EndChar:     d.EndChar,
			ReplaceWith: "0",
		}, true
	}
	return CodeAction{}, false
}

// editDistance computes the Levenshtein distance between two
// strings.
func editDistance(a, b string) int {
	la := len(a)
	lb := len(b)

	// Use single row for space efficiency
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := range lb + 1 {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost

			m := min(sub, min(ins, del))
			curr[j] = m
		}
		prev, curr = curr, prev
	}

	return prev[lb]
}
