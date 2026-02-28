package styles

import (
	"fmt"
	"strings"

	"github.com/tomo3110/gerbera"
)

func Style(styleMap gerbera.StyleMap) gerbera.ComponentFunc {
	return func(el *gerbera.Element) error {
		if el.Attr == nil {
			el.Attr = make(map[string]string)
		}
		var b strings.Builder
		for key, val := range styleMap {
			if _, err := fmt.Fprintf(&b, "%s: %s; ", key, fmt.Sprint(val)); err != nil {
				return err
			}
		}
		el.Attr["style"] = strings.TrimRight(b.String(), " ")
		return nil
	}
}
