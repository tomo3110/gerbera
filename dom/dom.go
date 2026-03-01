package dom

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/expr"
	"github.com/tomo3110/gerbera/property"
)

func Body(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("body", a...)
}

func H1(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("h1", a...)
}

func H2(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("h2", a...)
}

func H3(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("h3", a...)
}

func H4(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("h4", a...)
}

func H5(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("h5", a...)
}

func P(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("p", a...)
}

func Article(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("article", a...)
}

func Section(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("section", a...)
}

func Div(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("div", a...)
}

func A(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("a", a...)
}

func Form(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("form", a...)
}

func Button(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("button", a...)
}

func Submit(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	a = append(a, property.Attr("type", "submit"))
	return Button(a...)
}

func Label(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("label", a...)
}

func Input(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("input", a...)
}

func Textarea(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("textarea", a...)
}

func Ul(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("ul", a...)
}

func Ol(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("ol", a...)
}

func Li(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("li", a...)
}

func Link(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("link", a...)
}

func Title(title string) gerbera.ComponentFunc {
	return gerbera.Tag("title", property.Value(title))
}

func Meta(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("meta", a...)
}

func Head(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("head", a...)
}

func Img(src string, a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	a = append(a, property.Attr("src", src))
	return gerbera.Tag("img", a...)
}

func Script(src string, a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	a = append(a, property.Attr("src", src))
	return gerbera.Tag("script", a...)
}

func Table(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("table", a...)
}

func Tr(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("tr", a...)
}

func Td(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("td", a...)
}

func Th(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("th", a...)
}

func Thead(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("thead", a...)
}

func Tbody(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("tbody", a...)
}

func Tfoot(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("tfoot", a...)
}

func Header(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("header", a...)
}

func Footer(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("footer", a...)
}

func Nav(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("nav", a...)
}

func Address(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("address", a...)
}

func Aside(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("aside", a...)
}

func Area(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("area", a...)
}

func Br(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("br", a...)
}

func Hr(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("hr", a...)
}

func Code(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("code", a...)
}

func Main(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("main", a...)
}

func Select(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("select", a...)
}

func Option(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("option", a...)
}

func B(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("b", a...)
}

func HTML(lang string, a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("html", append(a, property.Attr("lang", lang))...)
}

func InputCheckbox(initCheck bool, value any, a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attr := append(a,
		property.Attr("type", "checkbox"),
		property.Name("done"),
		expr.If(initCheck, property.Attr("checked", "checked")),
		property.Value(value),
	)
	return Input(attr...)
}

// Span creates a <span> element.
func Span(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("span", a...)
}

// Details creates a <details> element.
func Details(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("details", a...)
}

// Summary creates a <summary> element.
func Summary(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("summary", a...)
}

// Style creates a <style> element.
func Style(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("style", a...)
}

// Dialog creates a <dialog> element.
func Dialog(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("dialog", a...)
}

// Progress creates a <progress> element.
func Progress(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("progress", a...)
}

// Meter creates a <meter> element.
func Meter(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("meter", a...)
}

// Time creates a <time> element.
func Time(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("time", a...)
}

// Mark creates a <mark> element.
func Mark(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("mark", a...)
}

// Abbr creates an <abbr> element.
func Abbr(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("abbr", a...)
}

// Figure creates a <figure> element.
func Figure(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("figure", a...)
}

// Figcaption creates a <figcaption> element.
func Figcaption(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("figcaption", a...)
}

// Video creates a <video> element.
func Video(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("video", a...)
}

// Audio creates an <audio> element.
func Audio(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("audio", a...)
}

// Source creates a <source> element.
func Source(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("source", a...)
}

// Canvas creates a <canvas> element.
func Canvas(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("canvas", a...)
}

// Picture creates a <picture> element.
func Picture(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("picture", a...)
}

// Tmpl creates a <template> element.
// Named Tmpl to avoid collision with Go's template package.
func Tmpl(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("template", a...)
}

// Slot creates a <slot> element.
func Slot(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("slot", a...)
}

// H6 creates a <h6> element.
func H6(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("h6", a...)
}

// Pre creates a <pre> element.
func Pre(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("pre", a...)
}

// Blockquote creates a <blockquote> element.
func Blockquote(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("blockquote", a...)
}

// Em creates an <em> element.
func Em(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("em", a...)
}

// Strong creates a <strong> element.
func Strong(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("strong", a...)
}

// Small creates a <small> element.
func Small(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("small", a...)
}

// Sub creates a <sub> element.
func Sub(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("sub", a...)
}

// Sup creates a <sup> element.
func Sup(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("sup", a...)
}

// Dl creates a <dl> (definition list) element.
func Dl(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("dl", a...)
}

// Dt creates a <dt> (definition term) element.
func Dt(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("dt", a...)
}

// Dd creates a <dd> (definition description) element.
func Dd(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("dd", a...)
}

// Fieldset creates a <fieldset> element.
func Fieldset(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("fieldset", a...)
}

// Legend creates a <legend> element.
func Legend(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("legend", a...)
}

// Caption creates a <caption> element.
func Caption(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("caption", a...)
}

// Colgroup creates a <colgroup> element.
func Colgroup(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("colgroup", a...)
}

// Col creates a <col> element.
func Col(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("col", a...)
}

// Iframe creates an <iframe> element.
func Iframe(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("iframe", a...)
}

// Embed creates an <embed> element.
func EmbedEl(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("embed", a...)
}

// Object creates an <object> element.
func Object(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("object", a...)
}

// Param creates a <param> element.
func Param(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("param", a...)
}

// Output creates an <output> element.
func Output(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("output", a...)
}

// Datalist creates a <datalist> element.
func Datalist(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("datalist", a...)
}

// Optgroup creates an <optgroup> element.
func Optgroup(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("optgroup", a...)
}

// Ruby creates a <ruby> element.
func Ruby(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("ruby", a...)
}

// Rt creates an <rt> element.
func Rt(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("rt", a...)
}

// Rp creates an <rp> element.
func Rp(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("rp", a...)
}

// Wbr creates a <wbr> element.
func Wbr(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("wbr", a...)
}

// Map creates a <map> element.
func MapEl(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("map", a...)
}

// Noscript creates a <noscript> element.
func Noscript(a ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return gerbera.Tag("noscript", a...)
}
