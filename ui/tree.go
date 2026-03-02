package ui

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	"github.com/tomo3110/gerbera/property"
)

// TreeNode defines one node in a tree view.
type TreeNode struct {
	Label    string
	Icon     string // icon name (optional)
	Children []TreeNode
	Open     bool   // whether children are expanded
	Active   bool   // highlight as selected
	Attrs    []gerbera.ComponentFunc // additional attrs (e.g., live.Click)
}

// Tree renders a hierarchical tree view from a slice of nodes.
func Tree(nodes []TreeNode) gerbera.ComponentFunc {
	return renderTreeLevel(nodes)
}

func renderTreeLevel(nodes []TreeNode) gerbera.ComponentFunc {
	var items []gerbera.ComponentFunc
	for _, n := range nodes {
		items = append(items, renderTreeNode(n))
	}
	return dom.Ul(append([]gerbera.ComponentFunc{
		property.Class("g-tree"),
		property.Role("tree"),
	}, items...)...)
}

func renderTreeNode(n TreeNode) gerbera.ComponentFunc {
	hasChildren := len(n.Children) > 0

	// Node content row
	var rowParts []gerbera.ComponentFunc
	rowParts = append(rowParts,
		property.Class("g-tree-node"),
		property.ClassIf(n.Active, "g-tree-node-active"),
	)
	rowParts = append(rowParts, n.Attrs...)

	if hasChildren {
		toggleCls := "g-tree-toggle"
		if n.Open {
			toggleCls += " g-tree-toggle-open"
		}
		rowParts = append(rowParts,
			dom.Span(property.Class(toggleCls), property.AriaHidden(true), property.Value("\u25b6")),
		)
	} else {
		rowParts = append(rowParts, dom.Span(property.Class("g-tree-spacer")))
	}

	if n.Icon != "" {
		rowParts = append(rowParts, Icon(n.Icon, "sm"))
	}

	rowParts = append(rowParts,
		dom.Span(property.Class("g-tree-label"), property.Value(n.Label)),
	)

	row := dom.Div(rowParts...)

	// Item with optional children
	inner := []gerbera.ComponentFunc{
		property.Class("g-tree-item"),
		property.Role("treeitem"),
		row,
	}

	if hasChildren {
		inner = append(inner,
			property.AriaExpanded(n.Open),
			expr.If(n.Open, renderSubTree(n.Children)),
		)
	}

	return dom.Li(inner...)
}

func renderSubTree(nodes []TreeNode) gerbera.ComponentFunc {
	var items []gerbera.ComponentFunc
	for _, n := range nodes {
		items = append(items, renderTreeNode(n))
	}
	return dom.Ul(append([]gerbera.ComponentFunc{
		property.Class("g-tree"),
		property.Role("group"),
	}, items...)...)
}
