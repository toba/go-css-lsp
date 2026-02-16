package analyzer

import (
	"strings"

	"github.com/toba/go-css-lsp/internal/css/data"
	"github.com/toba/go-css-lsp/internal/css/parser"
)

// Complete returns completion items for the given byte offset.
func Complete(
	ss *parser.Stylesheet,
	src []byte,
	offset int,
	opts LintOptions,
) []CompletionItem {
	if ss == nil {
		return nil
	}

	tag := opts.Experimental != ExperimentalIgnore
	tagDep := opts.Deprecated != DeprecatedIgnore
	ctx := determineContext(ss, src, offset)

	switch ctx.kind {
	case contextNone:
		return nil
	case contextProperty:
		return completeProperties(ctx.prefix, tag, tagDep)
	case contextValue:
		return completeValues(ctx.propertyName, ctx.prefix)
	case contextAtRule:
		return completeAtRules(ctx.prefix, tag, tagDep)
	case contextPseudoClass:
		return completePseudoClasses(ctx.prefix, tag, tagDep)
	case contextPseudoElement:
		return completePseudoElements(ctx.prefix, tag, tagDep)
	case contextSelector:
		return completeSelectorStart(ctx.prefix)
	case contextMediaFeature:
		return completeMediaFeatures(ctx.prefix)
	case contextMediaValue:
		return completeMediaValues(ctx.mediaFeatureName, ctx.prefix)
	default:
		return completeTopLevel(ctx.prefix, tag, tagDep)
	}
}

type contextKind int

const (
	contextUnknown contextKind = iota
	contextNone
	contextProperty
	contextValue
	contextAtRule
	contextPseudoClass
	contextPseudoElement
	contextSelector
	contextMediaFeature
	contextMediaValue
)

type completionContext struct {
	kind             contextKind
	prefix           string
	propertyName     string
	mediaFeatureName string // for contextMediaValue
}

func determineContext(
	ss *parser.Stylesheet,
	src []byte,
	offset int,
) completionContext {
	if offset > len(src) {
		offset = len(src)
	}

	// No completions inside comments
	if isInsideComment(src, offset) {
		return completionContext{kind: contextNone}
	}

	// Look backwards from offset for context clues
	text := string(src[:offset])

	// Check for media feature/value context: @media (...|...)
	if ctx, ok := detectMediaContext(text); ok {
		return ctx
	}

	// Check for @ at-rule context
	atIdx := strings.LastIndex(text, "@")
	if atIdx >= 0 {
		after := text[atIdx+1:]
		if !strings.ContainsAny(after, " \t\n\r{;},") {
			return completionContext{
				kind:   contextAtRule,
				prefix: after,
			}
		}
	}

	// Check for :: pseudo-element
	if strings.HasSuffix(
		text, "::",
	) || (len(text) > 2 && findLastDoubleColon(text)) {
		prefix := extractPseudoPrefix(text, true)
		return completionContext{
			kind:   contextPseudoElement,
			prefix: prefix,
		}
	}

	// Check for : pseudo-class (but not :: or property:)
	colonIdx := strings.LastIndex(text, ":")
	if colonIdx >= 0 {
		after := text[colonIdx+1:]
		if !strings.ContainsAny(
			after, " \t\n\r{;},",
		) && colonIdx > 0 &&
			text[colonIdx-1] != ':' {
			// Check if we're in a selector context
			// (not after a property name)
			if !isAfterPropertyColon(text, colonIdx) {
				return completionContext{
					kind:   contextPseudoClass,
					prefix: after,
				}
			}
		}
	}

	// Check if we're inside a declaration block
	node, inBlock := nodeAtOffset(ss, offset)
	if inBlock {
		// Check if we're in a value context
		if decl, ok := node.(*parser.Declaration); ok {
			if offset > decl.Property.End {
				prefix := extractValuePrefix(
					src, offset,
				)
				return completionContext{
					kind:         contextValue,
					prefix:       prefix,
					propertyName: decl.Property.Value,
				}
			}
		}

		// Inside block but not in a value = property context
		prefix := extractWordPrefix(src, offset)
		return completionContext{
			kind:   contextProperty,
			prefix: prefix,
		}
	}

	// Top level = selector context
	prefix := extractWordPrefix(src, offset)
	return completionContext{
		kind:   contextSelector,
		prefix: prefix,
	}
}

// tagCompletionItem annotates a completion item with
// experimental/deprecated markers based on its status.
func tagCompletionItem(
	item *CompletionItem,
	status data.StatusInfo,
	tagExperimental, tagDeprecated bool,
) {
	if tagExperimental && status.IsExperimental() {
		item.Detail = "(experimental) " + item.Detail
	}
	if tagDeprecated && status.IsDeprecated() {
		item.Detail = "(deprecated) " + item.Detail
		item.Deprecated = true
	}
}

func completeProperties(
	prefix string, tagExperimental, tagDeprecated bool,
) []CompletionItem {
	var items []CompletionItem
	prefix = strings.ToLower(prefix)

	for _, prop := range data.AllProperties() {
		if prefix != "" &&
			!strings.HasPrefix(prop.Name, prefix) {
			continue
		}
		item := CompletionItem{
			Label:         prop.Name,
			Kind:          KindProperty,
			Detail:        prop.Description,
			Documentation: prop.MDN,
			InsertText:    prop.Name + ": ",
		}
		tagCompletionItem(
			&item, prop.StatusInfo,
			tagExperimental, tagDeprecated,
		)
		items = append(items, item)
	}

	return items
}

func completeValues(
	propertyName, prefix string,
) []CompletionItem {
	var items []CompletionItem
	prefix = strings.ToLower(prefix)

	// Add global values
	for _, v := range data.GlobalValues {
		if prefix != "" &&
			!strings.HasPrefix(v, prefix) {
			continue
		}
		items = append(items, CompletionItem{
			Label: v,
			Kind:  KindKeyword,
		})
	}

	// Add property-specific values
	prop := data.LookupProperty(propertyName)
	if prop != nil {
		for _, v := range prop.Values {
			if prefix != "" &&
				!strings.HasPrefix(v, prefix) {
				continue
			}
			items = append(items, CompletionItem{
				Label: v,
				Kind:  KindValue,
			})
		}
	}

	// Add common functions
	for _, f := range data.CommonFunctions {
		if prefix != "" &&
			!strings.HasPrefix(f, prefix) {
			continue
		}
		items = append(items, CompletionItem{
			Label:      f,
			Kind:       KindFunction,
			InsertText: f + "(",
		})
	}

	// Add color functions
	for _, f := range data.ColorFunctions {
		if prefix != "" &&
			!strings.HasPrefix(f, prefix) {
			continue
		}
		items = append(items, CompletionItem{
			Label:      f,
			Kind:       KindColor,
			InsertText: f + "(",
		})
	}

	return items
}

func completeAtRules(
	prefix string, tagExperimental, tagDeprecated bool,
) []CompletionItem {
	var items []CompletionItem
	prefix = strings.ToLower(prefix)

	for _, rule := range data.AllAtRules() {
		if prefix != "" &&
			!strings.HasPrefix(rule.Name, prefix) {
			continue
		}
		item := CompletionItem{
			Label:  "@" + rule.Name,
			Kind:   KindKeyword,
			Detail: rule.Description,
		}
		tagCompletionItem(
			&item, rule.StatusInfo,
			tagExperimental, tagDeprecated,
		)
		items = append(items, item)
	}

	return items
}

func completePseudoClasses(
	prefix string, tagExperimental, tagDeprecated bool,
) []CompletionItem {
	var items []CompletionItem
	prefix = strings.ToLower(prefix)

	for _, pc := range data.AllPseudoClasses() {
		if prefix != "" &&
			!strings.HasPrefix(pc.Name, prefix) {
			continue
		}
		item := CompletionItem{
			Label:  ":" + pc.Name,
			Kind:   KindKeyword,
			Detail: pc.Description,
		}
		tagCompletionItem(
			&item, pc.StatusInfo,
			tagExperimental, tagDeprecated,
		)
		items = append(items, item)
	}

	return items
}

func completePseudoElements(
	prefix string, tagExperimental, tagDeprecated bool,
) []CompletionItem {
	var items []CompletionItem
	prefix = strings.ToLower(prefix)

	for _, pe := range data.AllPseudoElements() {
		if prefix != "" &&
			!strings.HasPrefix(pe.Name, prefix) {
			continue
		}
		item := CompletionItem{
			Label:  "::" + pe.Name,
			Kind:   KindKeyword,
			Detail: pe.Description,
		}
		tagCompletionItem(
			&item, pe.StatusInfo,
			tagExperimental, tagDeprecated,
		)
		items = append(items, item)
	}

	return items
}

// htmlElements lists common HTML element names for selector
// completion.
var htmlElements = []string{
	"a", "article", "aside", "body", "button",
	"div", "footer", "form", "h1", "h2", "h3",
	"h4", "h5", "h6", "header", "html", "img",
	"input", "label", "li", "main", "nav", "ol",
	"p", "section", "select", "span", "table",
	"textarea", "ul",
}

func completeSelectorStart(
	prefix string,
) []CompletionItem {
	var items []CompletionItem

	for _, el := range htmlElements {
		if prefix != "" &&
			!strings.HasPrefix(el, prefix) {
			continue
		}
		items = append(items, CompletionItem{
			Label: el,
			Kind:  KindKeyword,
		})
	}

	return items
}

func completeTopLevel(
	prefix string, tagExperimental, tagDeprecated bool,
) []CompletionItem {
	items := completeSelectorStart(prefix)
	items = append(items, completeAtRules(prefix, tagExperimental, tagDeprecated)...)
	return items
}

// detectMediaContext checks if the cursor is inside a
// @media (...) expression and returns the appropriate
// context for feature name or value completion.
func detectMediaContext(
	text string,
) (completionContext, bool) {
	// Find the last unmatched ( before cursor
	parenDepth := 0
	openIdx := -1
	for i := len(text) - 1; i >= 0; i-- {
		switch text[i] {
		case ')':
			parenDepth++
		case '(':
			if parenDepth > 0 {
				parenDepth--
			} else {
				openIdx = i
			}
		}
		if openIdx >= 0 {
			break
		}
	}
	if openIdx < 0 {
		return completionContext{}, false
	}

	// Check if @media precedes the paren (possibly with
	// whitespace, "not", "and", "or", other parens).
	before := strings.TrimSpace(text[:openIdx])
	if !isInMediaContext(before) {
		return completionContext{}, false
	}

	// Inside @media (...). Check if there's a colon
	// (feature: value context).
	inside := text[openIdx+1:]
	colonIdx := strings.LastIndex(inside, ":")
	if colonIdx >= 0 {
		featureName := strings.TrimSpace(inside[:colonIdx])
		valuePrefix := strings.TrimSpace(inside[colonIdx+1:])
		return completionContext{
			kind:             contextMediaValue,
			prefix:           valuePrefix,
			mediaFeatureName: featureName,
		}, true
	}

	// Feature name context
	prefix := strings.TrimSpace(inside)
	return completionContext{
		kind:   contextMediaFeature,
		prefix: prefix,
	}, true
}

// isInMediaContext checks if the text before an opening
// paren is part of a @media query.
func isInMediaContext(before string) bool {
	// Walk backwards through tokens like "and", "or",
	// "not", "(", ")" to find @media.
	s := before
	for {
		s = strings.TrimSpace(s)
		if s == "" {
			return false
		}
		// Check for @media directly
		if strings.HasSuffix(s, "@media") {
			return true
		}
		// Strip trailing keyword: and, or, not, only
		for _, kw := range []string{
			"and", "or", "not", "only",
		} {
			if strings.HasSuffix(s, kw) {
				trimmed := strings.TrimSpace(
					s[:len(s)-len(kw)],
				)
				if trimmed != s[:len(s)-len(kw)] ||
					len(trimmed) < len(s) {
					s = trimmed
					continue
				}
			}
		}
		// Strip trailing ) and balanced parens
		if strings.HasSuffix(s, ")") {
			depth := 0
			for i := len(s) - 1; i >= 0; i-- {
				if s[i] == ')' {
					depth++
				} else if s[i] == '(' {
					depth--
					if depth == 0 {
						s = strings.TrimSpace(s[:i])
						break
					}
				}
			}
			if depth != 0 {
				return false
			}
			continue
		}
		// Strip trailing comma
		if strings.HasSuffix(s, ",") {
			s = strings.TrimSpace(
				s[:len(s)-1],
			)
			continue
		}
		return false
	}
}

func completeMediaFeatures(
	prefix string,
) []CompletionItem {
	var items []CompletionItem
	prefix = strings.ToLower(prefix)

	for _, f := range data.AllMediaFeatures() {
		if prefix != "" &&
			!strings.HasPrefix(f.Name, prefix) {
			continue
		}
		items = append(items, CompletionItem{
			Label:  f.Name,
			Kind:   KindProperty,
			Detail: f.Description,
		})
	}

	return items
}

func completeMediaValues(
	featureName, prefix string,
) []CompletionItem {
	var items []CompletionItem
	prefix = strings.ToLower(prefix)

	f := data.LookupMediaFeature(featureName)
	if f == nil {
		return nil
	}

	for _, v := range f.Values {
		if prefix != "" &&
			!strings.HasPrefix(v, prefix) {
			continue
		}
		items = append(items, CompletionItem{
			Label: v,
			Kind:  KindValue,
		})
	}

	return items
}

func extractWordPrefix(src []byte, offset int) string {
	i := offset - 1
	for i >= 0 && isNameChar(src[i]) {
		i--
	}
	if i+1 >= offset {
		return ""
	}
	return string(src[i+1 : offset])
}

func extractValuePrefix(src []byte, offset int) string {
	i := offset - 1
	for i >= 0 && !isBreakChar(src[i]) {
		i--
	}
	prefix := strings.TrimSpace(string(src[i+1 : offset]))
	return prefix
}

func extractPseudoPrefix(text string, _ bool) string {
	// Find the last :: and return everything after
	idx := strings.LastIndex(text, "::")
	if idx >= 0 {
		return text[idx+2:]
	}
	return ""
}

func findLastDoubleColon(text string) bool {
	idx := strings.LastIndex(text, "::")
	if idx < 0 {
		return false
	}
	after := text[idx+2:]
	return !strings.ContainsAny(after, " \t\n\r{;},")
}

func isAfterPropertyColon(text string, colonIdx int) bool {
	// Walk backwards from colon to see if there's an ident
	// before it (property name pattern)
	i := colonIdx - 1
	for i >= 0 && (text[i] == ' ' || text[i] == '\t') {
		i--
	}
	if i < 0 {
		return false
	}

	// Check if preceded by an ident and we're inside { }
	braceDepth := 0
	for j := colonIdx; j >= 0; j-- {
		if text[j] == '}' {
			braceDepth++
		}
		if text[j] == '{' {
			braceDepth--
			if braceDepth < 0 {
				return true // inside a block
			}
		}
	}

	return false
}

func isNameChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '-' || ch == '_'
}

func isBreakChar(ch byte) bool {
	return ch == ':' || ch == ';' || ch == '{' ||
		ch == '}' || ch == '(' || ch == ',' ||
		ch == '\n'
}

// isInsideComment checks if offset is within a /* */ comment.
func isInsideComment(src []byte, offset int) bool {
	text := src[:offset]
	// Find the last /* and check there's no */ after it
	for i := len(text) - 1; i >= 1; i-- {
		if text[i] == '/' && text[i-1] == '*' {
			// Found a closing */, so we're not in a comment
			return false
		}
		if text[i] == '*' && i >= 1 && text[i-1] == '/' {
			// Found an opening /*, so we're in a comment
			return true
		}
	}
	return false
}
