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

type ComponentFunc func(*Element) error

type ClassMap map[string]bool
type AttrMap map[string]string
type StyleMap map[string]any

func Tag(tagName string, fus ...ComponentFunc) ComponentFunc {
	el := &Element{TagName: tagName}
	return func(parent *Element) error {
		el.Children = nil
		el.ClassNames = nil
		el.Attr = nil
		el.Value = ""
		for _, ef := range fus {
			if err := ef(el); err != nil {
				return err
			}
		}
		if err := el.AppendTo(parent); err != nil {
			return err
		}
		return nil
	}
}

func Skip() ComponentFunc {
	return func(parent *Element) error {
		return nil
	}
}

func Literal(lines ...string) ComponentFunc {
	joined := strings.Join(lines, "\n")
	return func(parent *Element) error {
		if parent.Value == "" {
			parent.Value = joined
		} else {
			parent.Value += joined
		}
		return nil
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
