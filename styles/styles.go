package styles

import (
	"bytes"
	"fmt"

	"github.com/tomo3110/gerbera"
)

func Style(styleMap gerbera.StyleMap) gerbera.ComponentFunc {
	return func(el *gerbera.Element) error {
		if el.Attr == nil {
			el.Attr = make(map[string]string)
		}
		buf := &bytes.Buffer{}
		for key, val := range styleMap {
			if _, err := fmt.Fprintf(buf, "%s: %s; ", key, fmt.Sprint(val)); err != nil {
				return err
			}
		}
		if _, err := buf.Write(bytes.TrimRight(buf.Bytes(), " ")); err != nil {
			return err
		}
		el.Attr["style"] = buf.String()
		return nil
	}
}
