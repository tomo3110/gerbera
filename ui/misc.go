package ui

import (
	"fmt"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
	"github.com/tomo3110/gerbera/styles"
)

// Divider renders a horizontal line separator.
func Divider() gerbera.ComponentFunc {
	return dom.Hr(property.Class("g-divider"))
}

// EmptyState renders a centered placeholder message for empty content areas.
func EmptyState(message string, children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	inner := []gerbera.ComponentFunc{
		property.Class("g-empty-state"),
		dom.P(property.Class("g-empty-state-msg"), property.Value(message)),
	}
	inner = append(inner, children...)
	return dom.Div(inner...)
}

// Progress renders a horizontal progress bar.
// pct is the percentage (0-100).
func Progress(pct int) gerbera.ComponentFunc {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	return dom.Div(
		property.Class("g-progress-track"),
		property.Role("progressbar"),
		property.AriaValueNow(pct),
		property.AriaValueMin(0),
		property.AriaValueMax(100),
		dom.Div(
			property.Class("g-progress-bar"),
			styles.Style(gerbera.StyleMap{"width": fmt.Sprintf("%d%%", pct)}),
		),
	)
}
