package diff

import (
	"testing"

	"github.com/tomo3110/gerbera"
)

func el(tag string, value string, children ...*gerbera.Element) *gerbera.Element {
	return &gerbera.Element{
		TagName:    tag,
		Value:      value,
		ChildElems:   children,
		ClassNames: make(gerbera.ClassMap),
		Attr:       make(gerbera.AttrMap),
	}
}

func elWithAttr(tag string, attr gerbera.AttrMap) *gerbera.Element {
	return &gerbera.Element{
		TagName:    tag,
		Attr:       attr,
		ClassNames: make(gerbera.ClassMap),
		ChildElems:   []*gerbera.Element{},
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
		ChildElems:   []*gerbera.Element{},
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

func elWithKey(tag, key, value string, children ...*gerbera.Element) *gerbera.Element {
	return &gerbera.Element{
		TagName:    tag,
		Key:        key,
		Value:      value,
		ChildElems:   children,
		ClassNames: make(gerbera.ClassMap),
		Attr:       make(gerbera.AttrMap),
	}
}

func TestDiffKeyReplace(t *testing.T) {
	oldEl := elWithKey("div", "list", "content A")
	newEl := elWithKey("div", "chat", "content B")
	patches := Diff(oldEl, newEl)
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}
	if patches[0].Op != OpReplace {
		t.Errorf("expected OpReplace, got %s", patches[0].Op)
	}
}

func TestDiffKeySame(t *testing.T) {
	oldEl := elWithKey("div", "same", "old text")
	newEl := elWithKey("div", "same", "new text")
	patches := Diff(oldEl, newEl)
	// Should do a recursive diff, not a replace
	for _, p := range patches {
		if p.Op == OpReplace {
			t.Error("expected no OpReplace for same key, but got one")
		}
	}
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch (text change), got %d", len(patches))
	}
	if patches[0].Op != OpSetText {
		t.Errorf("expected OpSetText, got %s", patches[0].Op)
	}
}

func TestDiffKeyOneEmpty(t *testing.T) {
	oldEl := elWithKey("div", "list", "content")
	newEl := el("div", "content") // no key
	patches := Diff(oldEl, newEl)
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}
	if patches[0].Op != OpReplace {
		t.Errorf("expected OpReplace when one side has key and other doesn't, got %s", patches[0].Op)
	}
}

// --- TestNode: lightweight Node implementation for platform-independent diff testing ---

type testNode struct {
	tag      string
	key      string
	text     string
	attrs    []gerbera.Attribute
	children []*testNode
}

func (n *testNode) AppendElement(tag string) gerbera.Node {
	child := &testNode{tag: tag}
	n.children = append(n.children, child)
	return child
}
func (n *testNode) SetAttribute(key, value string) {
	n.attrs = append(n.attrs, gerbera.Attribute{Key: key, Value: value})
}
func (n *testNode) AddClass(name string)  {}
func (n *testNode) SetText(text string)   { n.text = text }
func (n *testNode) SetKey(key string)     { n.key = key }
func (n *testNode) Tag() string           { return n.tag }
func (n *testNode) NodeKey() string       { return n.key }
func (n *testNode) Text() string          { return n.text }
func (n *testNode) Attributes() []gerbera.Attribute { return n.attrs }
func (n *testNode) Children() []gerbera.Node {
	nodes := make([]gerbera.Node, len(n.children))
	for i, c := range n.children {
		nodes[i] = c
	}
	return nodes
}

func tn(tag, text string, children ...*testNode) *testNode {
	return &testNode{tag: tag, text: text, children: children}
}

func tnWithAttr(tag string, attrs ...gerbera.Attribute) *testNode {
	return &testNode{tag: tag, attrs: attrs}
}

// TestDiffWithTestNode verifies that diff works with a non-Element Node implementation.
func TestDiffWithTestNode(t *testing.T) {
	old := tn("box", "hello")
	new := tn("box", "world")

	patches := Diff(old, new)
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}
	if patches[0].Op != OpSetText {
		t.Errorf("expected OpSetText, got %s", patches[0].Op)
	}
	if patches[0].Value != "world" {
		t.Errorf("expected 'world', got '%s'", patches[0].Value)
	}
}

// TestDiffTestNodeReplace verifies tag-change detection on a custom Node.
func TestDiffTestNodeReplace(t *testing.T) {
	old := tn("box", "")
	new := tn("panel", "text")

	patches := Diff(old, new)
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}
	if patches[0].Op != OpReplace {
		t.Errorf("expected OpReplace, got %s", patches[0].Op)
	}
}

// TestDiffTestNodeAttributes verifies attribute comparison on a custom Node.
func TestDiffTestNodeAttributes(t *testing.T) {
	old := tnWithAttr("box", gerbera.Attribute{Key: "color", Value: "red"})
	new := tnWithAttr("box", gerbera.Attribute{Key: "color", Value: "blue"})

	patches := Diff(old, new)
	found := false
	for _, p := range patches {
		if p.Op == OpSetAttr && p.Key == "color" && p.Value == "blue" {
			found = true
		}
	}
	if !found {
		t.Error("expected OpSetAttr patch for color=blue")
	}
}

// TestDiffTestNodeChildren verifies child-node comparison on a custom Node.
func TestDiffTestNodeChildren(t *testing.T) {
	old := tn("container", "", tn("item", "a"))
	new := tn("container", "", tn("item", "a"), tn("item", "b"))

	patches := Diff(old, new)
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

// TestDiffMixedNodeTypes verifies that diff works when comparing
// Element against TestNode (both implement Node).
func TestDiffMixedNodeTypes(t *testing.T) {
	old := tn("div", "hello")
	new := el("div", "world")

	patches := Diff(old, new)
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}
	if patches[0].Op != OpSetText {
		t.Errorf("expected OpSetText, got %s", patches[0].Op)
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
