package gerbera

func Parse(root *Element, fn ...ComponentFunc) (*Element, error) {
	for _, f := range fn {
		if err := f(root); err != nil {
			return nil, err
		}
	}
	return root, nil
}
