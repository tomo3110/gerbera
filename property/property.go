package property

import (
	"fmt"
	"html"

	"github.com/tomo3110/gerbera"
)

func Class(c ...string) gerbera.ComponentFunc {
	return func(el *gerbera.Element) error {
		el.ClassNames = make(gerbera.ClassMap)
		for _, s := range c {
			el.ClassNames[s] = false
		}
		return nil
	}
}

func Attr(key string, val interface{}) gerbera.ComponentFunc {
	return func(el *gerbera.Element) error {
		if el.Attr == nil {
			el.Attr = make(map[string]string)
		}
		el.Attr[key] = html.EscapeString(fmt.Sprint(val))
		return nil
	}
}

func ID(text string) gerbera.ComponentFunc {
	return Attr("id", text)
}

func Name(text string) gerbera.ComponentFunc {
	return Attr("name", text)
}

func Value(value interface{}) gerbera.ComponentFunc {
	return func(el *gerbera.Element) error {
		el.Value = html.EscapeString(fmt.Sprint(value))
		return nil
	}
}
