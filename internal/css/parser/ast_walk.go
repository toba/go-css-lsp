package parser

// Visitor is called for each AST node during a walk.
// Return false to skip children.
type Visitor func(node Node) bool

// Walk traverses the AST depth-first, calling visit for each
// node.
func Walk(node Node, visit Visitor) {
	if node == nil || !visit(node) {
		return
	}

	switch n := node.(type) {
	case *Stylesheet:
		for _, child := range n.Children {
			Walk(child, visit)
		}
	case *Ruleset:
		if n.Selectors != nil {
			Walk(n.Selectors, visit)
		}
		for _, child := range n.Children {
			Walk(child, visit)
		}
	case *SelectorList:
		for _, sel := range n.Selectors {
			Walk(sel, visit)
		}
	case *AtRule:
		if n.Block != nil {
			Walk(n.Block, visit)
		}
	case *Declaration:
		if n.Value != nil {
			Walk(n.Value, visit)
		}
	}
}
