package styles

import (
	"fmt"
	"strings"

	"github.com/tomo3110/gerbera"
)

func Style(styleMap gerbera.StyleMap) gerbera.ComponentFunc {
	return func(n gerbera.Node) {
		var b strings.Builder
		for key, val := range styleMap {
			fmt.Fprintf(&b, "%s: %s; ", key, fmt.Sprint(val))
		}
		n.SetAttribute("style", strings.TrimRight(b.String(), " "))
	}
}

// StyleIf conditionally applies inline styles.
func StyleIf(condition bool, styleMap gerbera.StyleMap) gerbera.ComponentFunc {
	if condition {
		return Style(styleMap)
	}
	return func(n gerbera.Node) {}
}

// CSS generates a <style> element with raw CSS content.
// Useful for embedding component-scoped CSS.
func CSS(cssText string) gerbera.ComponentFunc {
	return gerbera.Tag("style", func(n gerbera.Node) {
		n.SetText(cssText)
	})
}

// ScopedCSS generates CSS rules scoped to a given CSS selector prefix.
// Each rule in the map has the selector as key and declaration block as value.
func ScopedCSS(scope string, rules map[string]string) gerbera.ComponentFunc {
	var b strings.Builder
	for selector, declarations := range rules {
		if selector == "" {
			fmt.Fprintf(&b, "%s { %s } ", scope, declarations)
		} else {
			fmt.Fprintf(&b, "%s %s { %s } ", scope, selector, declarations)
		}
	}
	return CSS(b.String())
}

// MediaQuery generates a @media rule wrapping the given CSS.
func MediaQuery(query string, cssText string) string {
	return fmt.Sprintf("@media %s { %s }", query, cssText)
}

// Keyframes generates a @keyframes rule.
func Keyframes(name string, frames map[string]string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "@keyframes %s { ", name)
	for selector, declarations := range frames {
		fmt.Fprintf(&b, "%s { %s } ", selector, declarations)
	}
	b.WriteString("}")
	return b.String()
}
