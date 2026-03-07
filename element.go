package gerbera

// Node is the interface for building HTML element trees.
// ComponentFuncs operate on Nodes to add children, attributes, and text.
type Node interface {
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
}

// Element is the default implementation of Node.
// It represents an HTML element in the tree.
type Element struct {
	TagName    string
	Key        string
	ClassNames ClassMap
	Attr       AttrMap
	Children   []*Element
	Value      string
}

func (el *Element) AppendElement(tag string) Node {
	child := &Element{TagName: tag}
	el.Children = append(el.Children, child)
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

