package dom

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/expr"
	"github.com/tomo3110/gerbera/property"
)

func Body(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("body", a)
}

func H1(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("h1", a)
}

func H2(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("h2", a)
}

func H3(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("h3", a)
}

func H4(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("h4", a)
}

func H5(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("h5", a)
}

func P(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("p", a)
}

func Article(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("article", a)
}

func Section(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("section", a)
}

func Div(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("div", a)
}

func A(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("a", a)
}

func Form(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("form", a)
}

func Button(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("button", a)
}

func Submit(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	a = append(a, property.Attr("type", "submit"))
	return Button(a...)
}

func Label(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("label", a)
}

func Input(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("input", a)
}

func Textarea(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("textarea", a)
}

func Ul(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("ul", a)
}

func Ol(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("ol", a)
}

func Li(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("li", a)
}

func Link(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("link", a)
}

func Title(title string) gerbera.ComponentFunc {
	return gerbera.Tag("title", gerbera.CnvToSlice(property.Value(title)))
}

func Meta(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("meta", a)
}

func Head(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("head", a)
}

func Img(src string, a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	a = append(a, property.Attr("src", src))
	return gerbera.Tag("img", a)
}

func Script(src string, a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	a = append(a, property.Attr("src", src))
	return gerbera.Tag("script", a)
}

func Table(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("table", a)
}

func Tr(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("tr", a)
}

func Td(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("td", a)
}

func Th(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("th", a)
}

func Thead(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("thead", a)
}

func Tbody(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("tbody", a)
}

func Tfoot(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("tfoot", a)
}

func Header(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("header", a)
}

func Footer(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("footer", a)
}

func Nav(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("nav", a)
}

func Address(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("address", a)
}

func Aside(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("aside", a)
}

func Area(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("area", a)
}

func Br(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("br", a)
}

func Hr(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("hr", a)
}

func Code(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("code", a)
}

func Main(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("main", a)
}

func Select(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("select", a)
}

func Option(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("option", a)
}

func B(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("b", a)
}

func HTML(lang string, a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("html", append(a, property.Attr("lang", lang)))
}

func InputCheckbox(initCheck bool, value interface{}, a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attr := append(a,
		property.Attr("type", "checkbox"),
		property.Name("done"),
		expr.If(initCheck, property.Attr("checked", "checked")),
		property.Value(value),
	)
	return Input(attr...)
}
