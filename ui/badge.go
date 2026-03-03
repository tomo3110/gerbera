package ui

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

// Badge renders a small status indicator label.
// variant: "default" (default), "dark", "outline", "light".
func Badge(text string, variant ...string) gerbera.ComponentFunc {
	v := "default"
	if len(variant) > 0 && variant[0] != "" {
		v = variant[0]
	}
	return dom.Span(
		property.Class("g-badge", "g-badge-"+v),
		property.Value(text),
	)
}
