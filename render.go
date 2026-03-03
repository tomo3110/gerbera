package gerbera

import (
	"bufio"
	"io"
	"sync"
)

var bufioPool = sync.Pool{
	New: func() any { return bufio.NewWriterSize(nil, 4096) },
}

func Render(w io.Writer, el *Element) error {
	buf := bufioPool.Get().(*bufio.Writer)
	buf.Reset(w)
	defer bufioPool.Put(buf)

	if _, err := buf.WriteString("<!DOCTYPE html>\n"); err != nil {
		return err
	}
	if err := render(buf, el, 0); err != nil {
		return err
	}
	return buf.Flush()
}

func render(out *bufio.Writer, el *Element, indent int) error {
	writeIndent(out, indent)
	out.WriteByte('<')
	out.WriteString(el.TagName)

	if err := renderClasses(out, el.ClassNames); err != nil {
		return err
	}
	if err := renderAttr(out, el.Attr); err != nil {
		return err
	}

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
		if err := renderValue(out, el); err != nil {
			return err
		}
	}

	if len(el.Children) > 0 {
		for _, c := range el.Children {
			if err := render(out, c, indent+2); err != nil {
				return err
			}
			out.WriteByte('\n')
		}
	}
	writeIndent(out, indent)
	out.WriteString("</")
	out.WriteString(el.TagName)
	return out.WriteByte('>')
}

func renderAttr(out *bufio.Writer, attr AttrMap) error {
	for key, val := range attr {
		out.WriteByte(' ')
		out.WriteString(key)
		out.WriteString("=\"")
		out.WriteString(val)
		out.WriteByte('"')
	}
	return nil
}

func renderClasses(out *bufio.Writer, classes ClassMap) error {
	if len(classes) == 0 {
		return nil
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
	return nil
}

func renderValue(out *bufio.Writer, el *Element) error {
	if !isEmptyElement(el.TagName) {
		if el.Value != "" {
			out.WriteString(el.Value)
			if len(el.Children) > 0 {
				out.WriteByte('\n')
			}
		}
	}
	return nil
}
