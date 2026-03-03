package ui

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

// ButtonPrimary is an option that applies the primary (dark) button style.
var ButtonPrimary gerbera.ComponentFunc = property.ClassIf(true, "g-btn-primary")

// ButtonOutline is an option that applies the outline button style.
var ButtonOutline gerbera.ComponentFunc = property.ClassIf(true, "g-btn-outline")

// ButtonDanger is an option that applies the danger button style.
var ButtonDanger gerbera.ComponentFunc = property.ClassIf(true, "g-btn-danger")

// ButtonSmall is an option that applies the small size button style.
var ButtonSmall gerbera.ComponentFunc = property.ClassIf(true, "g-btn-sm")

// Button renders a styled button element.
// opts can include ButtonPrimary, ButtonOutline, ButtonDanger, ButtonSmall,
// or any ComponentFunc (e.g., live.Click).
func Button(label string, opts ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := []gerbera.ComponentFunc{
		property.Class("g-btn"),
	}
	attrs = append(attrs, opts...)
	attrs = append(attrs, property.Value(label))
	return dom.Button(attrs...)
}
