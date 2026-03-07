package ui

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

// SpinnerInline is a convenience option to display the spinner inline.
var SpinnerInline = property.ClassIf(true, "g-spinner-inline")

// Spinner renders a CSS-only loading spinner.
// size can be "sm" (16px), "md" (24px, default), or "lg" (40px).
func Spinner(size string, opts ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	if size == "" {
		size = "md"
	}
	attrs := gerbera.Components{
		property.Class("g-spinner", "g-spinner-"+size),
		property.Role("status"),
		property.AriaLabel("Loading"),
	}
	attrs = append(attrs, opts...)
	attrs = append(attrs, dom.Span(property.Class("g-spinner-arc")))
	return dom.Div(attrs...)
}
