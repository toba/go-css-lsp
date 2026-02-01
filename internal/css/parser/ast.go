// Package parser implements a recursive descent CSS3 parser.
package parser

import (
	"github.com/toba/go-css-lsp/internal/css/scanner"
)

// NodeKind identifies the type of AST node.
type NodeKind int

const (
	NodeStylesheet NodeKind = iota
	NodeRuleset
	NodeSelectorList
	NodeSelector
	NodeSimpleSelector
	NodeDeclaration
	NodeValue
	NodeFunction
	NodeAtRule
	NodeMediaQuery
	NodeComment
)

// Node is the interface for all AST nodes.
type Node interface {
	Kind() NodeKind
	Offset() int
	End() int
}

// Stylesheet is the root AST node.
type Stylesheet struct {
	Children []Node
	EndPos   int
}

func (n *Stylesheet) Kind() NodeKind { return NodeStylesheet }
func (n *Stylesheet) Offset() int    { return 0 }
func (n *Stylesheet) End() int       { return n.EndPos }

// Ruleset represents a CSS rule: selectors + declarations.
type Ruleset struct {
	Selectors    *SelectorList
	Declarations []*Declaration
	StartPos     int
	EndPos       int
}

func (n *Ruleset) Kind() NodeKind { return NodeRuleset }
func (n *Ruleset) Offset() int    { return n.StartPos }
func (n *Ruleset) End() int       { return n.EndPos }

// SelectorList represents a comma-separated list of selectors.
type SelectorList struct {
	Selectors []*Selector
	StartPos  int
	EndPos    int
}

func (n *SelectorList) Kind() NodeKind {
	return NodeSelectorList
}
func (n *SelectorList) Offset() int { return n.StartPos }
func (n *SelectorList) End() int    { return n.EndPos }

// Selector represents a single selector (sequence of simple
// selectors with combinators).
type Selector struct {
	Parts    []SelectorPart
	StartPos int
	EndPos   int
}

func (n *Selector) Kind() NodeKind { return NodeSelector }
func (n *Selector) Offset() int    { return n.StartPos }
func (n *Selector) End() int       { return n.EndPos }

// SelectorPart is one piece of a selector.
type SelectorPart struct {
	Token scanner.Token
	// Combinator is set for combinators: ' ', '>', '+', '~'
	Combinator string
}

// Declaration represents a property: value pair.
type Declaration struct {
	Property  scanner.Token
	Value     *Value
	Important bool
	StartPos  int
	EndPos    int
	Semicolon bool // whether followed by ;
}

func (n *Declaration) Kind() NodeKind {
	return NodeDeclaration
}
func (n *Declaration) Offset() int { return n.StartPos }
func (n *Declaration) End() int    { return n.EndPos }

// Value represents declaration value tokens.
type Value struct {
	Tokens   []scanner.Token
	StartPos int
	EndPos   int
}

func (n *Value) Kind() NodeKind { return NodeValue }
func (n *Value) Offset() int    { return n.StartPos }
func (n *Value) End() int       { return n.EndPos }

// AtRule represents an at-rule (@media, @import, etc.).
type AtRule struct {
	Name     string // without @
	Prelude  []scanner.Token
	Block    *Stylesheet // nil for statement at-rules
	StartPos int
	EndPos   int
}

func (n *AtRule) Kind() NodeKind { return NodeAtRule }
func (n *AtRule) Offset() int    { return n.StartPos }
func (n *AtRule) End() int       { return n.EndPos }

// Comment represents a CSS comment.
type Comment struct {
	Text     string
	StartPos int
	EndPos   int
}

func (n *Comment) Kind() NodeKind { return NodeComment }
func (n *Comment) Offset() int    { return n.StartPos }
func (n *Comment) End() int       { return n.EndPos }

// Error represents a parse error with position information.
type Error struct {
	Message  string
	StartPos int
	EndPos   int
}

func (e *Error) Error() string { return e.Message }
