package gerbera

import "testing"

func TestElement_AppendTo(t *testing.T) {
	el := &Element{TagName: "div"}
	child := &Element{TagName: "h1"}
	if err := child.AppendTo(el); err != nil {
		t.Error(err)
	}
	if len(el.Children) != 1 {
		t.Errorf("")
	}
	child1 := &Element{TagName: "p"}
	if err := child1.AppendTo(el); err != nil {
		t.Error(err)
	}
}
