package diff

import (
	"testing"

	"github.com/tomo3110/gerbera"
)

func el(tag string, value string, children ...*gerbera.Element) *gerbera.Element {
	return &gerbera.Element{
		TagName:    tag,
		Value:      value,
		Children:   children,
		ClassNames: make(gerbera.ClassMap),
		Attr:       make(gerbera.AttrMap),
	}
}

func elWithAttr(tag string, attr gerbera.AttrMap) *gerbera.Element {
	return &gerbera.Element{
		TagName:    tag,
		Attr:       attr,
		ClassNames: make(gerbera.ClassMap),
		Children:   []*gerbera.Element{},
	}
}

func elWithClass(tag string, classes ...string) *gerbera.Element {
	cm := make(gerbera.ClassMap)
	for _, c := range classes {
		cm[c] = false
	}
	return &gerbera.Element{
		TagName:    tag,
		ClassNames: cm,
		Attr:       make(gerbera.AttrMap),
		Children:   []*gerbera.Element{},
	}
}

func TestDiffIdentical(t *testing.T) {
	oldEl := el("div", "hello")
	newEl := el("div", "hello")
	patches := Diff(oldEl, newEl)
	if len(patches) != 0 {
		t.Errorf("expected 0 patches for identical elements, got %d", len(patches))
	}
}

func TestDiffSetText(t *testing.T) {
	oldEl := el("div", "hello")
	newEl := el("div", "world")
	patches := Diff(oldEl, newEl)
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}
	if patches[0].Op != OpSetText {
		t.Errorf("expected OpSetText, got %s", patches[0].Op)
	}
	if patches[0].Value != "world" {
		t.Errorf("expected value 'world', got '%s'", patches[0].Value)
	}
}

func TestDiffSetHTML(t *testing.T) {
	oldEl := el("div", "hello")
	newEl := el("div", "<strong>hello</strong>")
	patches := Diff(oldEl, newEl)
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}
	if patches[0].Op != OpSetHTML {
		t.Errorf("expected OpSetHTML, got %s", patches[0].Op)
	}
	if patches[0].Value != "<strong>hello</strong>" {
		t.Errorf("expected value '<strong>hello</strong>', got '%s'", patches[0].Value)
	}
}

func TestDiffSetTextNoHTML(t *testing.T) {
	oldEl := el("div", "hello")
	newEl := el("div", "world")
	patches := Diff(oldEl, newEl)
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}
	if patches[0].Op != OpSetText {
		t.Errorf("expected OpSetText for plain text, got %s", patches[0].Op)
	}
}

func TestDiffReplace(t *testing.T) {
	oldEl := el("div", "")
	newEl := el("span", "text")
	patches := Diff(oldEl, newEl)
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}
	if patches[0].Op != OpReplace {
		t.Errorf("expected OpReplace, got %s", patches[0].Op)
	}
	if patches[0].HTML == "" {
		t.Error("expected non-empty HTML fragment")
	}
}

func TestDiffSetAttr(t *testing.T) {
	oldEl := elWithAttr("div", gerbera.AttrMap{"id": "a"})
	newEl := elWithAttr("div", gerbera.AttrMap{"id": "b"})
	patches := Diff(oldEl, newEl)
	found := false
	for _, p := range patches {
		if p.Op == OpSetAttr && p.Key == "id" && p.Value == "b" {
			found = true
		}
	}
	if !found {
		t.Error("expected OpSetAttr patch for id=b")
	}
}

func TestDiffAddAttr(t *testing.T) {
	oldEl := elWithAttr("div", gerbera.AttrMap{})
	newEl := elWithAttr("div", gerbera.AttrMap{"data-x": "1"})
	patches := Diff(oldEl, newEl)
	found := false
	for _, p := range patches {
		if p.Op == OpSetAttr && p.Key == "data-x" && p.Value == "1" {
			found = true
		}
	}
	if !found {
		t.Error("expected OpSetAttr patch for data-x=1")
	}
}

func TestDiffRemoveAttr(t *testing.T) {
	oldEl := elWithAttr("div", gerbera.AttrMap{"id": "a"})
	newEl := elWithAttr("div", gerbera.AttrMap{})
	patches := Diff(oldEl, newEl)
	found := false
	for _, p := range patches {
		if p.Op == OpRemoveAttr && p.Key == "id" {
			found = true
		}
	}
	if !found {
		t.Error("expected OpRemoveAttr patch for id")
	}
}

func TestDiffSetClass(t *testing.T) {
	oldEl := elWithClass("div", "a")
	newEl := elWithClass("div", "b")
	patches := Diff(oldEl, newEl)
	found := false
	for _, p := range patches {
		if p.Op == OpSetClass && p.Value == "b" {
			found = true
		}
	}
	if !found {
		t.Error("expected OpSetClass patch with value 'b'")
	}
}

func TestDiffInsertChild(t *testing.T) {
	oldEl := el("div", "", el("p", "first"))
	newEl := el("div", "", el("p", "first"), el("p", "second"))
	patches := Diff(oldEl, newEl)
	found := false
	for _, p := range patches {
		if p.Op == OpInsert && p.Index == 1 {
			found = true
		}
	}
	if !found {
		t.Error("expected OpInsert patch at index 1")
	}
}

func TestDiffRemoveChild(t *testing.T) {
	oldEl := el("div", "", el("p", "first"), el("p", "second"))
	newEl := el("div", "", el("p", "first"))
	patches := Diff(oldEl, newEl)
	found := false
	for _, p := range patches {
		if p.Op == OpRemove && p.Index == 1 {
			found = true
		}
	}
	if !found {
		t.Error("expected OpRemove patch at index 1")
	}
}

func TestDiffNestedChange(t *testing.T) {
	oldEl := el("div", "",
		el("p", "unchanged"),
		el("span", "old"),
	)
	newEl := el("div", "",
		el("p", "unchanged"),
		el("span", "new"),
	)
	patches := Diff(oldEl, newEl)
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}
	p := patches[0]
	if p.Op != OpSetText {
		t.Errorf("expected OpSetText, got %s", p.Op)
	}
	if len(p.Path) != 1 || p.Path[0] != 1 {
		t.Errorf("expected path [1], got %v", p.Path)
	}
	if p.Value != "new" {
		t.Errorf("expected value 'new', got '%s'", p.Value)
	}
}

func TestDiffMultipleRemoves(t *testing.T) {
	oldEl := el("div", "", el("p", "a"), el("p", "b"), el("p", "c"))
	newEl := el("div", "", el("p", "a"))
	patches := Diff(oldEl, newEl)
	removeCount := 0
	for _, p := range patches {
		if p.Op == OpRemove {
			removeCount++
		}
	}
	if removeCount != 2 {
		t.Errorf("expected 2 remove patches, got %d", removeCount)
	}
}

func TestRenderFragment(t *testing.T) {
	e := el("div", "", el("p", "hello"))
	html, err := RenderFragment(e)
	if err != nil {
		t.Fatal(err)
	}
	if html == "" {
		t.Error("expected non-empty HTML fragment")
	}
	// Should not contain DOCTYPE
	if len(html) > 9 && html[:9] == "<!DOCTYPE" {
		t.Error("fragment should not contain DOCTYPE")
	}
}

func TestRenderFragmentVoidElement(t *testing.T) {
	e := &gerbera.Element{
		TagName:    "input",
		Attr:       gerbera.AttrMap{"type": "text"},
		ClassNames: make(gerbera.ClassMap),
	}
	html, err := RenderFragment(e)
	if err != nil {
		t.Fatal(err)
	}
	if html == "" {
		t.Error("expected non-empty HTML")
	}
}
