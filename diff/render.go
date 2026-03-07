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

// RenderFragment renders a Node as an HTML fragment (no DOCTYPE).
// It operates on the Node interface, making it platform-independent.
func RenderFragment(node gerbera.Node) (string, error) {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	w := bufioPool.Get().(*bufio.Writer)
	w.Reset(buf)

	if err := renderNode(w, node, 0); err != nil {
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

func renderNode(out *bufio.Writer, node gerbera.Node, indent int) error {
	tag := node.Tag()
	text := node.Text()
	attrs := node.Attributes()
	children := node.Children()

	out.WriteByte('<')
	out.WriteString(tag)

	renderAttributes(out, attrs)

	if isEmptyElement(tag) {
		if text != "" {
			out.WriteString(" value=\"")
			out.WriteString(text)
			out.WriteByte('"')
		}
		return out.WriteByte('>')
	}

	if text != "" && len(children) == 0 {
		out.WriteByte('>')
		out.WriteString(text)
		out.WriteString("</")
		out.WriteString(tag)
		return out.WriteByte('>')
	}

	if text == "" && len(children) == 0 {
		out.WriteString("></")
		out.WriteString(tag)
		return out.WriteByte('>')
	}

	out.WriteString(">\n")
	if text != "" {
		out.WriteString(text)
		out.WriteByte('\n')
	}

	for _, c := range children {
		writeIndent(out, indent+2)
		if err := renderNode(out, c, indent+2); err != nil {
			return err
		}
		out.WriteByte('\n')
	}
	writeIndent(out, indent)
	out.WriteString("</")
	out.WriteString(tag)
	return out.WriteByte('>')
}

func renderAttributes(out *bufio.Writer, attrs []gerbera.Attribute) {
	for _, a := range attrs {
		out.WriteByte(' ')
		out.WriteString(a.Key)
		out.WriteString("=\"")
		out.WriteString(a.Value)
		out.WriteByte('"')
	}
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
