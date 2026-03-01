package diff

import (
	"testing"

	"github.com/tomo3110/gerbera"
)

func benchEl(tag, value string, children ...*gerbera.Element) *gerbera.Element {
	return &gerbera.Element{
		TagName:    tag,
		Value:      value,
		Children:   children,
		ClassNames: make(gerbera.ClassMap),
		Attr:       make(gerbera.AttrMap),
	}
}

func benchElFull(tag, value string, classes []string, attrs gerbera.AttrMap, children ...*gerbera.Element) *gerbera.Element {
	cm := make(gerbera.ClassMap)
	for _, c := range classes {
		cm[c] = false
	}
	return &gerbera.Element{
		TagName:    tag,
		Value:      value,
		ClassNames: cm,
		Attr:       attrs,
		Children:   children,
	}
}

// buildList creates a <ul> with n <li> children.
func buildList(n int, prefix string) *gerbera.Element {
	items := make([]*gerbera.Element, n)
	for i := range items {
		items[i] = benchEl("li", prefix)
	}
	return benchEl("ul", "", items...)
}

// --- Diff benchmarks ---

func BenchmarkDiff_Identical(b *testing.B) {
	oldEl := benchEl("div", "text",
		benchEl("p", "child1"),
		benchEl("p", "child2"),
	)
	newEl := benchEl("div", "text",
		benchEl("p", "child1"),
		benchEl("p", "child2"),
	)
	b.ResetTimer()
	for b.Loop() {
		Diff(oldEl, newEl)
	}
}

func BenchmarkDiff_TextChange(b *testing.B) {
	oldEl := benchEl("div", "old")
	newEl := benchEl("div", "new")
	b.ResetTimer()
	for b.Loop() {
		Diff(oldEl, newEl)
	}
}

func BenchmarkDiff_AttrChange(b *testing.B) {
	oldEl := benchElFull("div", "", nil, gerbera.AttrMap{"id": "a", "data-x": "1"})
	newEl := benchElFull("div", "", nil, gerbera.AttrMap{"id": "b", "data-x": "1"})
	b.ResetTimer()
	for b.Loop() {
		Diff(oldEl, newEl)
	}
}

func BenchmarkDiff_ClassChange(b *testing.B) {
	oldEl := benchElFull("div", "", []string{"active", "primary"}, gerbera.AttrMap{})
	newEl := benchElFull("div", "", []string{"inactive", "primary"}, gerbera.AttrMap{})
	b.ResetTimer()
	for b.Loop() {
		Diff(oldEl, newEl)
	}
}

func BenchmarkDiff_ChildInsert(b *testing.B) {
	oldEl := buildList(10, "item")
	newEl := buildList(11, "item")
	b.ResetTimer()
	for b.Loop() {
		Diff(oldEl, newEl)
	}
}

func BenchmarkDiff_ChildRemove(b *testing.B) {
	oldEl := buildList(11, "item")
	newEl := buildList(10, "item")
	b.ResetTimer()
	for b.Loop() {
		Diff(oldEl, newEl)
	}
}

func BenchmarkDiff_DeepTree(b *testing.B) {
	makeDeep := func(leafVal string) *gerbera.Element {
		leaf := benchEl("span", leafVal)
		node := leaf
		for i := 0; i < 10; i++ {
			node = benchEl("div", "", node)
		}
		return node
	}
	oldEl := makeDeep("old-leaf")
	newEl := makeDeep("new-leaf")
	b.ResetTimer()
	for b.Loop() {
		Diff(oldEl, newEl)
	}
}

func BenchmarkDiff_WideList50(b *testing.B) {
	oldEl := buildList(50, "item")
	newEl := buildList(50, "item")
	// Change the last item
	newEl.Children[49].Value = "changed"
	b.ResetTimer()
	for b.Loop() {
		Diff(oldEl, newEl)
	}
}

func BenchmarkDiff_Replace(b *testing.B) {
	oldEl := benchEl("div", "", benchEl("p", "text"))
	newEl := benchEl("div", "", benchEl("span", "text"))
	b.ResetTimer()
	for b.Loop() {
		Diff(oldEl, newEl)
	}
}

// --- RenderFragment benchmarks ---

func BenchmarkRenderFragment_Simple(b *testing.B) {
	el := benchEl("div", "hello")
	b.ResetTimer()
	for b.Loop() {
		RenderFragment(el)
	}
}

func BenchmarkRenderFragment_WithChildren(b *testing.B) {
	el := benchEl("div", "",
		benchEl("h1", "Title"),
		benchEl("p", "Paragraph"),
		benchEl("ul", "",
			benchEl("li", "item 1"),
			benchEl("li", "item 2"),
			benchEl("li", "item 3"),
		),
	)
	b.ResetTimer()
	for b.Loop() {
		RenderFragment(el)
	}
}

func BenchmarkRenderFragment_WithAttrsAndClasses(b *testing.B) {
	el := benchElFull("div", "text",
		[]string{"container", "active", "primary"},
		gerbera.AttrMap{"id": "main", "data-role": "content", "aria-label": "Main"},
	)
	b.ResetTimer()
	for b.Loop() {
		RenderFragment(el)
	}
}
