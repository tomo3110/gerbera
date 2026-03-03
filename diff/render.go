package diff

import (
	"bufio"
	"bytes"
	"strings"
	"sync"

	"github.com/tomo3110/gerbera"
)

var emptyElements = map[string]struct{}{
	"doctype": {}, "area": {}, "base": {}, "basefont": {},
	"br": {}, "col": {}, "frame": {}, "hr": {}, "img": {},
	"input": {}, "isindex": {}, "link": {}, "meta": {},
	"param": {}, "embed": {}, "keygen": {}, "command": {},
}

var indentSpaces = strings.Repeat(" ", 128)

var bufPool = sync.Pool{
	New: func() any { return &bytes.Buffer{} },
}

var bufioPool = sync.Pool{
	New: func() any { return bufio.NewWriterSize(nil, 4096) },
}

// RenderFragment renders an Element as an HTML fragment (no DOCTYPE).
func RenderFragment(el *gerbera.Element) (string, error) {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	w := bufioPool.Get().(*bufio.Writer)
	w.Reset(buf)

	if err := renderNode(w, el, 0); err != nil {
		bufPool.Put(buf)
		bufioPool.Put(w)
		return "", err
	}
	if err := w.Flush(); err != nil {
		bufPool.Put(buf)
		bufioPool.Put(w)
		return "", err
	}
	s := buf.String()
	bufPool.Put(buf)
	bufioPool.Put(w)
	return s, nil
}

func renderNode(out *bufio.Writer, el *gerbera.Element, indent int) error {
	out.WriteByte('<')
	out.WriteString(el.TagName)

	renderClasses(out, el.ClassNames)
	renderAttr(out, el.Attr)

	if isEmptyElement(el.TagName) {
		if el.Value != "" {
			out.WriteString(" value=\"")
			out.WriteString(el.Value)
			out.WriteByte('"')
		}
		return out.WriteByte('>')
	}

	if el.Value != "" && len(el.Children) == 0 {
		out.WriteByte('>')
		out.WriteString(el.Value)
		out.WriteString("</")
		out.WriteString(el.TagName)
		return out.WriteByte('>')
	}

	if el.Value == "" && len(el.Children) == 0 {
		out.WriteString("></")
		out.WriteString(el.TagName)
		return out.WriteByte('>')
	}

	out.WriteString(">\n")
	if el.Value != "" {
		out.WriteString(el.Value)
		out.WriteByte('\n')
	}

	for _, c := range el.Children {
		writeIndent(out, indent+2)
		if err := renderNode(out, c, indent+2); err != nil {
			return err
		}
		out.WriteByte('\n')
	}
	writeIndent(out, indent)
	out.WriteString("</")
	out.WriteString(el.TagName)
	return out.WriteByte('>')
}

func renderAttr(out *bufio.Writer, attr gerbera.AttrMap) {
	for key, val := range attr {
		out.WriteByte(' ')
		out.WriteString(key)
		out.WriteString("=\"")
		out.WriteString(val)
		out.WriteByte('"')
	}
}

func renderClasses(out *bufio.Writer, classes gerbera.ClassMap) {
	if len(classes) == 0 {
		return
	}
	out.WriteString(" class=\"")
	first := true
	for name := range classes {
		if !first {
			out.WriteByte(' ')
		}
		out.WriteString(name)
		first = false
	}
	out.WriteByte('"')
}

func writeIndent(out *bufio.Writer, n int) {
	if n > 0 {
		if n > len(indentSpaces) {
			n = len(indentSpaces)
		}
		out.WriteString(indentSpaces[:n])
	}
}

func isEmptyElement(n string) bool {
	_, ok := emptyElements[n]
	return ok
}
