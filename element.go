package gerbera

type Element struct {
	Index      int
	TagName    string
	ClassNames ClassMap
	Attr       AttrMap
	Children   []*Element
	Value      string
}

func (el *Element) AppendTo(parent *Element) error {
	length := len(parent.Children)
	if length > 0 {
		parent.Index = length
	}
	parent.Children = append(parent.Children, el)
	return nil
}
