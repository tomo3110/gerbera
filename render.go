package gerbera

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func Render(w io.Writer, el *Element) error {
	buf := bufio.NewWriter(w)
	if _, err := buf.Write([]byte("<!DOCTYPE html>\n")); err != nil {
		return err
	}
	if err := render(buf, el, 0); err != nil {
		return err
	}
	return buf.Flush()
}

func render(out *bufio.Writer, el *Element, indent int) error {
	bytesRepeat(out, space, indent)
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
		// img, input etc...
		if el.Value != "" {
			if _, err := fmt.Fprintf(out, " value=\"%s\"", el.Value); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprint(out, ">"); err != nil {
			return err
		}
		return nil
	}

	if el.Value != "" && len(el.Children) == 0 {
		if _, err := fmt.Fprint(out, ">"); err != nil {
			return err
		}
		if err := renderValue(out, el); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(out, "</%s>", el.TagName); err != nil {
			return err
		}
		return nil
	}

	if _, err := fmt.Fprint(out, ">\n"); err != nil {
		return err
	}
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
			if _, err := out.Write(newline); err != nil {
				return err
			}
		}
	}
	bytesRepeat(out, space, indent)
	if _, err := fmt.Fprintf(out, "</%s>", el.TagName); err != nil {
		return err
	}
	return nil
}

func renderAttr(out io.Writer, attr AttrMap) error {
	for key, val := range attr {
		if _, err := fmt.Fprintf(out, " %s=\"%s\"", key, val); err != nil {
			return err
		}
	}
	return nil
}

func renderClasses(out io.Writer, classes ClassMap) error {
	if len(classes) > 0 {
		list := make([]string, 0, len(classes))
		if _, err := out.Write([]byte(" class=\"")); err != nil {
			return err
		}
		for name := range classes {
			list = append(list, name)
		}
		if _, err := fmt.Fprint(out, strings.Join(list, string(space))); err != nil {
			return err
		}
		if _, err := out.Write([]byte("\"")); err != nil {
			return err
		}
	}
	return nil
}

func renderValue(out io.Writer, el *Element) error {
	if !isEmptyElement(el.TagName) {
		if el.Value != "" {
			if len(el.Children) > 0 {
				if _, err := out.Write([]byte(el.Value)); err != nil {
					return err
				}
				if _, err := out.Write(newline); err != nil {
					return err
				}
			} else {
				if _, err := out.Write([]byte(el.Value)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
