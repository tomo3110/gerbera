package gerbera

type Element struct {
	TagName    string
	ClassNames ClassMap
	Attr       AttrMap
	Children   []*Element
	Value      string
}

func (el *Element) AppendTo(parent *Element) error {
	parent.Children = append(parent.Children, el)
	return nil
}
