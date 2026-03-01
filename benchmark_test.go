package gerbera

import (
	"io"
	"testing"
)

// buildSimpleTree creates a minimal element tree for baseline measurement.
func buildSimpleTree() *Element {
	return &Element{
		TagName:    "html",
		Attr:       AttrMap{"lang": "en"},
		ClassNames: ClassMap{},
		Children: []*Element{
			{
				TagName: "head", ClassNames: ClassMap{}, Attr: AttrMap{},
				Children: []*Element{
					{TagName: "title", Value: "Hello", ClassNames: ClassMap{}, Attr: AttrMap{}, Children: []*Element{}},
				},
			},
			{
				TagName: "body", ClassNames: ClassMap{"container": true}, Attr: AttrMap{},
				Children: []*Element{
					{TagName: "h1", Value: "Title", ClassNames: ClassMap{}, Attr: AttrMap{}, Children: []*Element{}},
					{TagName: "p", Value: "A paragraph.", ClassNames: ClassMap{}, Attr: AttrMap{}, Children: []*Element{}},
				},
			},
		},
	}
}

// buildDeepTree creates a nested element tree with depth levels and width children per level.
func buildDeepTree(depth, width int) *Element {
	root := &Element{
		TagName:    "div",
		ClassNames: ClassMap{"root": true},
		Attr:       AttrMap{"id": "root"},
		Children:   make([]*Element, 0, width),
	}
	for i := 0; i < width; i++ {
		child := &Element{
			TagName:    "div",
			ClassNames: ClassMap{"level-1": true, "item": true},
			Attr:       AttrMap{"data-index": "0"},
			Children:   make([]*Element, 0),
		}
		node := child
		for d := 2; d <= depth; d++ {
			inner := &Element{
				TagName:    "div",
				ClassNames: ClassMap{"nested": true},
				Attr:       AttrMap{"data-depth": "d"},
				Children:   make([]*Element, 0),
			}
			node.Children = append(node.Children, inner)
			node = inner
		}
		node.Children = append(node.Children, &Element{
			TagName: "span", Value: "leaf", ClassNames: ClassMap{}, Attr: AttrMap{}, Children: []*Element{},
		})
		root.Children = append(root.Children, child)
	}
	return root
}

// buildWideList creates a flat list of n <li> elements inside a <ul>.
func buildWideList(n int) *Element {
	items := make([]*Element, n)
	for i := range items {
		items[i] = &Element{
			TagName: "li", Value: "Item", ClassNames: ClassMap{}, Attr: AttrMap{}, Children: []*Element{},
		}
	}
	return &Element{
		TagName: "ul", ClassNames: ClassMap{"list": true}, Attr: AttrMap{}, Children: items,
	}
}

// --- Render benchmarks ---

func BenchmarkRender_Simple(b *testing.B) {
	tree := buildSimpleTree()
	b.ResetTimer()
	for b.Loop() {
		Render(io.Discard, tree)
	}
}

func BenchmarkRender_Deep5x4(b *testing.B) {
	tree := buildDeepTree(5, 4)
	b.ResetTimer()
	for b.Loop() {
		Render(io.Discard, tree)
	}
}

func BenchmarkRender_Deep10x4(b *testing.B) {
	tree := buildDeepTree(10, 4)
	b.ResetTimer()
	for b.Loop() {
		Render(io.Discard, tree)
	}
}

func BenchmarkRender_Wide100(b *testing.B) {
	tree := buildWideList(100)
	b.ResetTimer()
	for b.Loop() {
		Render(io.Discard, tree)
	}
}

func BenchmarkRender_Wide1000(b *testing.B) {
	tree := buildWideList(1000)
	b.ResetTimer()
	for b.Loop() {
		Render(io.Discard, tree)
	}
}

// --- Parse benchmarks (ComponentFunc composition) ---

func BenchmarkParse_Flat10(b *testing.B) {
	fns := make([]ComponentFunc, 10)
	for i := range fns {
		fns[i] = Tag("p", Literal("text"))
	}
	b.ResetTimer()
	for b.Loop() {
		root := &Element{TagName: "div", Children: make([]*Element, 0)}
		Parse(root, fns...)
	}
}

func BenchmarkParse_Flat100(b *testing.B) {
	fns := make([]ComponentFunc, 100)
	for i := range fns {
		fns[i] = Tag("p", Literal("text"))
	}
	b.ResetTimer()
	for b.Loop() {
		root := &Element{TagName: "div", Children: make([]*Element, 0)}
		Parse(root, fns...)
	}
}

func BenchmarkParse_Nested5(b *testing.B) {
	var fn ComponentFunc
	fn = Tag("span", Literal("leaf"))
	for i := 0; i < 5; i++ {
		fn = Tag("div", fn)
	}
	b.ResetTimer()
	for b.Loop() {
		root := &Element{TagName: "div", Children: make([]*Element, 0)}
		Parse(root, fn)
	}
}

// --- ExecuteTemplate (full pipeline) benchmarks ---

func BenchmarkExecuteTemplate_Simple(b *testing.B) {
	body := Tag("body",
		Tag("h1", Literal("Hello")),
		Tag("p", Literal("World")),
	)
	b.ResetTimer()
	for b.Loop() {
		ExecuteTemplate(io.Discard, "en", Tag("head", Tag("title", Literal("Test"))), body)
	}
}

func BenchmarkExecuteTemplate_Medium(b *testing.B) {
	items := make([]ComponentFunc, 50)
	for i := range items {
		items[i] = Tag("li", Literal("item"))
	}
	body := Tag("body", Tag("ul", items...))
	b.ResetTimer()
	for b.Loop() {
		ExecuteTemplate(io.Discard, "en", Tag("head"), body)
	}
}

func BenchmarkExecuteTemplate_Large(b *testing.B) {
	items := make([]ComponentFunc, 200)
	for i := range items {
		items[i] = Tag("tr",
			Tag("td", Literal("col1")),
			Tag("td", Literal("col2")),
			Tag("td", Literal("col3")),
		)
	}
	body := Tag("body", Tag("table", items...))
	b.ResetTimer()
	for b.Loop() {
		ExecuteTemplate(io.Discard, "en", Tag("head"), body)
	}
}
