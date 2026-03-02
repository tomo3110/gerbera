package ui

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

// StatCard renders a KPI/metric display card with a label and value.
// Additional opts are applied to the outer container.
func StatCard(label, value string, opts ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	inner := []gerbera.ComponentFunc{
		property.Class("g-stat"),
		dom.P(property.Class("g-stat-label"), property.Value(label)),
		dom.P(property.Class("g-stat-value"), property.Value(value)),
	}
	inner = append(inner, opts...)
	return dom.Div(inner...)
}
