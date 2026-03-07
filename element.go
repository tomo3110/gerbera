package gerbera

import (
	"sort"
	"strings"
	"sync"
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

	// --- Metadata (shared across tree) ---

	// SetMeta sets tree-wide metadata under the given key.
	// All nodes in the same tree share the same metadata store,
	// so SetMeta called on any node is visible from any other node.
	SetMeta(key string, val any)

	// Meta returns the metadata value for the given key, or nil if not set.
	Meta(key string) any

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

// renderMeta is a tree-wide metadata store shared by all nodes in the same tree.
type renderMeta struct {
	mu   sync.Mutex
	data map[string]any
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
	meta       *renderMeta
}

func (el *Element) AppendElement(tag string) Node {
	child := &Element{TagName: tag, meta: el.ensureMeta()}
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

func (el *Element) ensureMeta() *renderMeta {
	if el.meta == nil {
		el.meta = &renderMeta{data: map[string]any{}}
	}
	return el.meta
}

func (el *Element) SetMeta(key string, val any) {
	m := el.ensureMeta()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = val
}

func (el *Element) Meta(key string) any {
	m := el.ensureMeta()
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.data[key]
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

