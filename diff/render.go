package diff

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/tomo3110/gerbera"
)

var emptyElements = map[string]struct{}{
	"doctype": {}, "area": {}, "base": {}, "basefont": {},
	"br": {}, "col": {}, "frame": {}, "hr": {}, "img": {},
	"input": {}, "isindex": {}, "link": {}, "meta": {},
	"param": {}, "embed": {}, "keygen": {}, "command": {},
}

// RenderFragment renders an Element as an HTML fragment (no DOCTYPE).
func RenderFragment(el *gerbera.Element) (string, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	if err := renderNode(w, el, 0); err != nil {
		return "", err
	}
	if err := w.Flush(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderNode(out *bufio.Writer, el *gerbera.Element, indent int) error {
	if _, err := fmt.Fprintf(out, "<%s", el.TagName); err != nil {
		return err
	}
	if err := renderClasses(out, el.ClassNames); err != nil {
		return err
	}
	if err := renderAttr(out, el.Attr); err != nil {
		return err
	}

	if isEmptyElement(el.TagName) {
		if el.Value != "" {
			if _, err := fmt.Fprintf(out, " value=\"%s\"", el.Value); err != nil {
				return err
			}
		}
		_, err := fmt.Fprint(out, ">")
		return err
	}

	if el.Value != "" && len(el.Children) == 0 {
		if _, err := fmt.Fprint(out, ">"); err != nil {
			return err
		}
		if _, err := out.WriteString(el.Value); err != nil {
			return err
		}
		_, err := fmt.Fprintf(out, "</%s>", el.TagName)
		return err
	}

	if _, err := fmt.Fprint(out, ">\n"); err != nil {
		return err
	}
	if el.Value != "" {
		if _, err := out.WriteString(el.Value); err != nil {
			return err
		}
		if _, err := out.Write([]byte("\n")); err != nil {
			return err
		}
	}

	for _, c := range el.Children {
		writeIndent(out, indent+2)
		if err := renderNode(out, c, indent+2); err != nil {
			return err
		}
		if _, err := out.Write([]byte("\n")); err != nil {
			return err
		}
	}
	writeIndent(out, indent)
	_, err := fmt.Fprintf(out, "</%s>", el.TagName)
	return err
}

func renderAttr(out io.Writer, attr gerbera.AttrMap) error {
	for key, val := range attr {
		if _, err := fmt.Fprintf(out, " %s=\"%s\"", key, val); err != nil {
			return err
		}
	}
	return nil
}

func renderClasses(out io.Writer, classes gerbera.ClassMap) error {
	if len(classes) == 0 {
		return nil
	}
	list := make([]string, 0, len(classes))
	for name := range classes {
		list = append(list, name)
	}
	_, err := fmt.Fprintf(out, " class=\"%s\"", strings.Join(list, " "))
	return err
}

func writeIndent(out io.Writer, n int) {
	for i := 0; i < n; i++ {
		out.Write([]byte(" "))
	}
}

func isEmptyElement(n string) bool {
	_, ok := emptyElements[n]
	return ok
}
