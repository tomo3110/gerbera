package gerbera

import (
	"sort"
	"strings"
)

// Attribute represents a key-value attribute pair on a Node.
type Attribute struct {
	Key   string
	Value string
}

// Node is the interface for building and reading HTML element trees.
// Write methods are used by ComponentFuncs to construct trees.
// Read methods are used by diff to compare trees in a platform-independent way.
type Node interface {
	// --- Write (used by ComponentFunc) ---

	// AppendElement adds a child element with the given tag name and returns it.
	// The returned Node can be further modified by ComponentFuncs.
	AppendElement(tag string) Node

	// SetAttribute sets an attribute key-value pair.
	SetAttribute(key, value string)

	// AddClass adds a CSS class name.
	AddClass(name string)

	// SetText sets the text content.
	SetText(text string)

	// SetKey sets the reconciliation key for diffing.
	SetKey(key string)

	// --- Read (used by diff) ---

	// Tag returns the tag name of this node.
	Tag() string

	// NodeKey returns the reconciliation key for this node.
	NodeKey() string

	// Attributes returns all attributes of this node, including the class attribute.
	Attributes() []Attribute

	// Text returns the text content of this node.
	Text() string

	// Children returns the child nodes of this node.
	Children() []Node
}

// Element is the default implementation of Node.
// It represents an HTML element in the tree.
type Element struct {
	TagName    string
	Key        string
	ClassNames ClassMap
	Attr       AttrMap
	ChildElems []*Element
	Value      string
}

func (el *Element) AppendElement(tag string) Node {
	child := &Element{TagName: tag}
	el.ChildElems = append(el.ChildElems, child)
	return child
}

func (el *Element) SetAttribute(key, value string) {
	if el.Attr == nil {
		el.Attr = make(AttrMap)
	}
	el.Attr[key] = value
}

func (el *Element) AddClass(name string) {
	if el.ClassNames == nil {
		el.ClassNames = make(ClassMap)
	}
	el.ClassNames[name] = false
}

func (el *Element) SetText(text string) {
	if el.Value == "" {
		el.Value = text
	} else {
		el.Value += text
	}
}

func (el *Element) SetKey(key string) {
	el.Key = key
}

// --- Read methods ---

func (el *Element) Tag() string {
	return el.TagName
}

func (el *Element) NodeKey() string {
	return el.Key
}

func (el *Element) Attributes() []Attribute {
	var attrs []Attribute

	// Include class attribute from ClassNames
	if len(el.ClassNames) > 0 {
		names := make([]string, 0, len(el.ClassNames))
		for name := range el.ClassNames {
			names = append(names, name)
		}
		sort.Strings(names)
		attrs = append(attrs, Attribute{Key: "class", Value: strings.Join(names, " ")})
	}

	// Include regular attributes in sorted order for deterministic output
	if len(el.Attr) > 0 {
		keys := make([]string, 0, len(el.Attr))
		for k := range el.Attr {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			attrs = append(attrs, Attribute{Key: k, Value: el.Attr[k]})
		}
	}

	return attrs
}

func (el *Element) Text() string {
	return el.Value
}

func (el *Element) Children() []Node {
	nodes := make([]Node, len(el.ChildElems))
	for i, c := range el.ChildElems {
		nodes[i] = c
	}
	return nodes
}

