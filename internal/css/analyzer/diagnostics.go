package analyzer

import (
	"slices"
	"strings"

	"github.com/toba/go-css-lsp/internal/css/data"
	"github.com/toba/go-css-lsp/internal/css/parser"
	"github.com/toba/go-css-lsp/internal/css/scanner"
)

// vendorPrefixes lists common CSS vendor prefixes.
var vendorPrefixes = []string{
	"-webkit-", "-moz-", "-ms-", "-o-",
}

func hasVendorPrefix(name string) bool {
	for _, p := range vendorPrefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

// Analyze returns diagnostics for the parsed stylesheet.
func Analyze(
	ss *parser.Stylesheet,
	src []byte,
	opts LintOptions,
) []Diagnostic {
	if ss == nil {
		return nil
	}
	a := &diagAnalyzer{src: src, opts: opts}
	a.analyzeStylesheet(ss)
	return a.diags
}

type diagAnalyzer struct {
	src   []byte
	opts  LintOptions
	diags []Diagnostic
}

func (a *diagAnalyzer) analyzeStylesheet(ss *parser.Stylesheet) {
	for _, child := range ss.Children {
		switch n := child.(type) {
		case *parser.Ruleset:
			a.analyzeRuleset(n)
		case *parser.AtRule:
			a.analyzeAtRule(n)
		}
	}
}

func (a *diagAnalyzer) analyzeRuleset(rs *parser.Ruleset) {
	// Check for empty rulesets
	if len(rs.Children) == 0 {
		line, char := OffsetToLineChar(
			a.src, rs.Offset(),
		)
		endLine, endChar := OffsetToLineChar(
			a.src, rs.End(),
		)
		a.diags = append(a.diags, Diagnostic{
			Message:   EmptyRulesetMsg,
			StartLine: line,
			StartChar: char,
			EndLine:   endLine,
			EndChar:   endChar,
			Severity:  SeverityHint,
		})
	}

	seen := make(map[string]bool)

	for _, child := range rs.Children {
		switch n := child.(type) {
		case *parser.Declaration:
			a.analyzeDeclaration(n, seen)
		case *parser.Ruleset:
			a.analyzeRuleset(n)
		case *parser.AtRule:
			a.analyzeAtRule(n)
		}
	}
}

func (a *diagAnalyzer) analyzeDeclaration(
	decl *parser.Declaration,
	seen map[string]bool,
) {
	propName := decl.Property.Value

	// Skip custom properties
	if IsCustomProperty(propName) {
		return
	}

	// Check for unknown properties
	if !data.IsKnownProperty(propName) {
		line, char := OffsetToLineChar(
			a.src, decl.Property.Offset,
		)
		endLine, endChar := OffsetToLineChar(
			a.src, decl.Property.End,
		)
		a.diags = append(a.diags, Diagnostic{
			Message:   UnknownPropertyMessage(propName),
			StartLine: line,
			StartChar: char,
			EndLine:   endLine,
			EndChar:   endChar,
			Severity:  SeverityWarning,
		})
	} else {
		prop := data.LookupProperty(propName)
		if prop != nil && a.opts.Experimental != ExperimentalIgnore &&
			prop.IsExperimental() {
			sev := SeverityWarning
			if a.opts.Experimental == ExperimentalError {
				sev = SeverityError
			}
			line, char := OffsetToLineChar(
				a.src, decl.Property.Offset,
			)
			endLine, endChar := OffsetToLineChar(
				a.src, decl.Property.End,
			)
			a.diags = append(a.diags, Diagnostic{
				Message:   ExperimentalPropertyMessage(propName),
				StartLine: line,
				StartChar: char,
				EndLine:   endLine,
				EndChar:   endChar,
				Severity:  sev,
			})
		} else if prop != nil && a.opts.Deprecated != DeprecatedIgnore &&
			prop.IsDeprecated() {
			sev := SeverityWarning
			if a.opts.Deprecated == DeprecatedError {
				sev = SeverityError
			}
			line, char := OffsetToLineChar(
				a.src, decl.Property.Offset,
			)
			endLine, endChar := OffsetToLineChar(
				a.src, decl.Property.End,
			)
			a.diags = append(a.diags, Diagnostic{
				Message:   DeprecatedPropertyMessage(propName),
				StartLine: line,
				StartChar: char,
				EndLine:   endLine,
				EndChar:   endChar,
				Severity:  sev,
			})
		}
	}

	// Check for duplicate properties
	if seen[propName] {
		line, char := OffsetToLineChar(
			a.src, decl.Property.Offset,
		)
		endLine, endChar := OffsetToLineChar(
			a.src, decl.Property.End,
		)
		a.diags = append(a.diags, Diagnostic{
			Message:   DuplicatePropertyMessage(propName),
			StartLine: line,
			StartChar: char,
			EndLine:   endLine,
			EndChar:   endChar,
			Severity:  SeverityWarning,
		})
	}
	seen[propName] = true

	// Check for zero with units (e.g. 0px -> 0)
	a.checkZeroWithUnit(decl)

	// Check for unknown values
	a.checkUnknownValues(decl)

	// Check for !important usage
	if decl.Important {
		line, char := OffsetToLineChar(
			a.src, decl.StartPos,
		)
		endLine, endChar := OffsetToLineChar(
			a.src, decl.EndPos,
		)
		a.diags = append(a.diags, Diagnostic{
			Message:   AvoidImportantMsg,
			StartLine: line,
			StartChar: char,
			EndLine:   endLine,
			EndChar:   endChar,
			Severity:  SeverityHint,
		})
	}

	// Check for vendor prefixes
	if hasVendorPrefix(propName) {
		line, char := OffsetToLineChar(
			a.src, decl.Property.Offset,
		)
		endLine, endChar := OffsetToLineChar(
			a.src, decl.Property.End,
		)
		a.diags = append(a.diags, Diagnostic{
			Message:   VendorPrefixMessage(propName),
			StartLine: line,
			StartChar: char,
			EndLine:   endLine,
			EndChar:   endChar,
			Severity:  SeverityHint,
		})
	}
}

func (a *diagAnalyzer) checkZeroWithUnit(
	decl *parser.Declaration,
) {
	if decl.Value == nil {
		return
	}
	for _, tok := range decl.Value.Tokens {
		if tok.Kind != scanner.Dimension {
			continue
		}
		val := tok.Value
		if len(val) < 2 || val[0] != '0' {
			continue
		}
		// Check it's actually "0" + unit, not "0.5px" etc.
		unit := val[1:]
		if unit != "" && (unit[0] >= '0' && unit[0] <= '9' || unit[0] == '.') {
			continue
		}
		// Skip time units where 0 is meaningful
		if unit == "s" || unit == "ms" {
			continue
		}
		line, char := OffsetToLineChar(
			a.src, tok.Offset,
		)
		endLine, endChar := OffsetToLineChar(
			a.src, tok.End,
		)
		a.diags = append(a.diags, Diagnostic{
			Message: "unnecessary unit: '" +
				val + "' can be written as '0'",
			StartLine: line,
			StartChar: char,
			EndLine:   endLine,
			EndChar:   endChar,
			Severity:  SeverityHint,
		})
	}
}

func (a *diagAnalyzer) checkUnknownValues(
	decl *parser.Declaration,
) {
	if a.opts.UnknownValues == UnknownValueIgnore {
		return
	}
	if decl.Value == nil || len(decl.Value.Tokens) == 0 {
		return
	}

	propName := decl.Property.Value
	prop := data.LookupProperty(propName)
	if prop == nil || len(prop.Values) == 0 {
		return
	}

	// Skip if value contains var() or any function token
	for _, tok := range decl.Value.Tokens {
		if tok.Kind == scanner.Function {
			return
		}
	}

	// Build lookup set for property values
	validValues := make(map[string]bool, len(prop.Values)+len(data.GlobalValues))
	for _, v := range prop.Values {
		validValues[v] = true
	}
	for _, v := range data.GlobalValues {
		validValues[v] = true
	}

	sev := SeverityWarning
	if a.opts.UnknownValues == UnknownValueError {
		sev = SeverityError
	}

	for _, tok := range decl.Value.Tokens {
		if tok.Kind != scanner.Ident {
			continue
		}
		val := strings.ToLower(tok.Value)
		if validValues[val] {
			continue
		}
		// In lenient color mode, accept named colors
		// for any property
		if !a.opts.StrictColorNames && isNamedColor(val) {
			continue
		}
		line, char := OffsetToLineChar(
			a.src, tok.Offset,
		)
		endLine, endChar := OffsetToLineChar(
			a.src, tok.End,
		)
		a.diags = append(a.diags, Diagnostic{
			Message: UnknownValueMessage(
				tok.Value, propName,
			),
			StartLine: line,
			StartChar: char,
			EndLine:   endLine,
			EndChar:   endChar,
			Severity:  sev,
		})
	}
}

// isNamedColor returns true if the value is a CSS named color.
func isNamedColor(val string) bool {
	return slices.Contains(data.NamedColors, val)
}

func (a *diagAnalyzer) analyzeAtRule(rule *parser.AtRule) {
	if !data.IsKnownAtRule(rule.Name) {
		line, char := OffsetToLineChar(
			a.src, rule.Offset(),
		)
		endLine, endChar := OffsetToLineChar(
			a.src, rule.Offset()+len(rule.Name)+1,
		)
		a.diags = append(a.diags, Diagnostic{
			Message:   UnknownAtRuleMessage(rule.Name),
			StartLine: line,
			StartChar: char,
			EndLine:   endLine,
			EndChar:   endChar,
			Severity:  SeverityWarning,
		})
	}

	if rule.Block != nil {
		a.analyzeStylesheet(rule.Block)
	}
}

// nodeAtOffset finds what kind of context the offset is in.
func nodeAtOffset(
	ss *parser.Stylesheet,
	offset int,
) (node parser.Node, inDeclarationBlock bool) {
	var found parser.Node
	inBlock := false

	parser.Walk(ss, func(n parser.Node) bool {
		if n.Offset() > offset || n.End() < offset {
			return false
		}
		found = n

		if _, ok := n.(*parser.Ruleset); ok {
			inBlock = true
		}

		return true
	})

	return found, inBlock
}

// tokenAtOffset finds the token at the given byte offset in
// the value tokens or declaration.
func tokenAtOffset(
	ss *parser.Stylesheet,
	offset int,
) *scanner.Token {
	var result *scanner.Token

	parser.Walk(ss, func(n parser.Node) bool {
		if n.Offset() > offset || n.End() < offset {
			return false
		}

		switch node := n.(type) {
		case *parser.Declaration:
			if node.Property.Offset <= offset &&
				offset < node.Property.End {
				result = &node.Property
				return false
			}
		case *parser.Value:
			for i := range node.Tokens {
				t := &node.Tokens[i]
				if t.Offset <= offset && offset < t.End {
					result = t
					return false
				}
			}
		}

		return true
	})

	return result
}
