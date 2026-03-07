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

	hasMetaInjection := el.meta != nil && (el.TagName == "head" || el.TagName == "body")
	hasChildren := len(el.ChildElems) > 0

	if el.Value != "" && !hasChildren && !hasMetaInjection {
		out.WriteByte('>')
		out.WriteString(el.Value)
		out.WriteString("</")
		out.WriteString(el.TagName)
		return out.WriteByte('>')
	}

	if el.Value == "" && !hasChildren && !hasMetaInjection {
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

	if hasChildren {
		for _, c := range el.ChildElems {
			if err := render(out, c, indent+2); err != nil {
				return err
			}
			out.WriteByte('\n')
		}
	}

	// Inject stylesheets before </head>
	if el.TagName == "head" {
		if err := renderMetaStyleSheets(out, el, indent+2); err != nil {
			return err
		}
	}

	// Inject scripts before </body>
	if el.TagName == "body" {
		if err := renderMetaScripts(out, el, indent+2); err != nil {
			return err
		}
	}

	writeIndent(out, indent)
	out.WriteString("</")
	out.WriteString(el.TagName)
	return out.WriteByte('>')
}

const (
	metaKeyScripts     = "gerbera.web.scripts"
	metaKeyStyleSheets = "gerbera.web.stylesheets"
)

func renderMetaStyleSheets(out *bufio.Writer, el *Element, indent int) error {
	if el.meta == nil {
		return nil
	}
	val := el.Meta(metaKeyStyleSheets)
	if val == nil {
		return nil
	}
	set, ok := val.(map[string]bool)
	if !ok {
		return nil
	}
	for path := range set {
		writeIndent(out, indent)
		out.WriteString(`<link rel="stylesheet" href="`)
		out.WriteString(path)
		out.WriteString("\">\n")
	}
	return nil
}

func renderMetaScripts(out *bufio.Writer, el *Element, indent int) error {
	if el.meta == nil {
		return nil
	}
	val := el.Meta(metaKeyScripts)
	if val == nil {
		return nil
	}
	set, ok := val.(map[string]bool)
	if !ok {
		return nil
	}
	for path := range set {
		writeIndent(out, indent)
		out.WriteString(`<script src="`)
		out.WriteString(path)
		out.WriteString("\" defer></script>\n")
	}
	return nil
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
			if len(el.ChildElems) > 0 {
				out.WriteByte('\n')
			}
		}
	}
	return nil
}
