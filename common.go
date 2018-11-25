package gerbera

import (
	"io"
	"strings"
)

var (
	SPACE   = []byte(" ")
	NEWLINE = []byte("\n")
)

var emptyElements = []string{
	"doctype",
	"area",
	"base",
	"basefont",
	"br",
	"col",
	"frame",
	"hr",
	"img",
	"input",
	"isindex",
	"link",
	"meta",
	"param",
	"embed",
	"keygen",
	"command",
}

type ComponentFunc func(*Element) error

type ClassMap map[string]bool
type AttrMap map[string]string
type StyleMap map[string]interface{}

func Tag(tagName string, fus []ComponentFunc) ComponentFunc {
	el := &Element{TagName: tagName}
	return func(parent *Element) error {
		el.Children = make([]*Element, 0)
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
	return func(parent *Element) error {
		str := strings.Join(lines, "\n")
		parent.Value = parent.Value + str
		return nil
	}
}

func ExecuteTemplate(w io.Writer, lang string, c ...ComponentFunc) error {
	h := &Template{Lang: lang}
	return h.Execute(w, c...)
}

func CnvToSlice(a ...ComponentFunc) []ComponentFunc {
	return a
}

func isEmptyElement(n string) bool {
	for _, s := range emptyElements {
		if s == n {
			return true
		}
	}
	return false
}

func bytesRepeat(out io.Writer, b []byte, count int) {
	for i := 0; i < count; i++ {
		if _, err := out.Write(b); err != nil {
			continue
		}
	}
}
