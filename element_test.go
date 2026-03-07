package gerbera

import "testing"

func TestSetMeta_BasicGetSet(t *testing.T) {
	root := &Element{TagName: "html"}
	root.SetMeta("key1", "value1")

	got := root.Meta("key1")
	if got != "value1" {
		t.Errorf("Meta(key1) = %v, want value1", got)
	}
}

func TestMeta_ReturnsNilForUnset(t *testing.T) {
	root := &Element{TagName: "html"}
	if got := root.Meta("nonexistent"); got != nil {
		t.Errorf("Meta(nonexistent) = %v, want nil", got)
	}
}

func TestSetMeta_SharedAcrossTree(t *testing.T) {
	root := &Element{TagName: "html"}
	child := root.AppendElement("body")
	grandchild := child.AppendElement("div")

	// Set from grandchild
	grandchild.SetMeta("shared", "hello")

	// Read from root
	if got := root.Meta("shared"); got != "hello" {
		t.Errorf("root.Meta(shared) = %v, want hello", got)
	}

	// Read from child
	if got := child.Meta("shared"); got != "hello" {
		t.Errorf("child.Meta(shared) = %v, want hello", got)
	}
}

func TestSetMeta_OverwritesPreviousValue(t *testing.T) {
	root := &Element{TagName: "html"}
	root.SetMeta("key", "v1")
	root.SetMeta("key", "v2")

	if got := root.Meta("key"); got != "v2" {
		t.Errorf("Meta(key) = %v, want v2", got)
	}
}

func TestMeta_InitializedByTemplate(t *testing.T) {
	tmpl := &Template{Lang: "en"}
	tmpl.init()

	tmpl.el.SetMeta("test", 42)
	if got := tmpl.el.Meta("test"); got != 42 {
		t.Errorf("Meta(test) = %v, want 42", got)
	}
}
