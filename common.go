package gerbera

import (
	"bufio"
	"io"
	"strings"
)

var indentSpaces = strings.Repeat(" ", 128)

var emptyElements = map[string]struct{}{
	"doctype":  {},
	"area":     {},
	"base":     {},
	"basefont": {},
	"br":       {},
	"col":      {},
	"frame":    {},
	"hr":       {},
	"img":      {},
	"input":    {},
	"isindex":  {},
	"link":     {},
	"meta":     {},
	"param":    {},
	"embed":    {},
	"keygen":   {},
	"command":  {},
	"source":   {},
	"track":    {},
	"wbr":      {},
}

// ComponentFunc is a function that builds HTML elements.
// It operates on a Node to add children, attributes, and text content.
// ComponentFuncs are pure construction functions — they do not perform I/O.
type ComponentFunc func(Node)

type ClassMap map[string]bool
type AttrMap map[string]string
type StyleMap map[string]any

func Tag(tagName string, fus ...ComponentFunc) ComponentFunc {
	return func(parent Node) {
		child := parent.AppendElement(tagName)
		for _, f := range fus {
			f(child)
		}
	}
}

func Skip() ComponentFunc {
	return func(parent Node) {}
}

func Literal(lines ...string) ComponentFunc {
	joined := strings.Join(lines, "\n")
	return func(parent Node) {
		parent.SetText(joined)
	}
}

func ExecuteTemplate(w io.Writer, lang string, c ...ComponentFunc) error {
	h := &Template{Lang: lang}
	return h.Execute(w, c...)
}

func isEmptyElement(n string) bool {
	_, ok := emptyElements[n]
	return ok
}

func writeIndent(out *bufio.Writer, n int) {
	if n > 0 {
		if n > len(indentSpaces) {
			n = len(indentSpaces)
		}
		out.WriteString(indentSpaces[:n])
	}
}
