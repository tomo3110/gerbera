package gerbera

// Parse executes ComponentFuncs on a root Element to build the element tree.
func Parse(root *Element, fn ...ComponentFunc) *Element {
	for _, f := range fn {
		f(root)
	}
	return root
}
