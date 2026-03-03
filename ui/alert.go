package ui

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

// Alert renders a notification message box.
// variant: "info" (default), "success", "warning", "danger".
func Alert(text string, variant ...string) gerbera.ComponentFunc {
	v := "info"
	if len(variant) > 0 && variant[0] != "" {
		v = variant[0]
	}
	return dom.Div(
		property.Class("g-alert", "g-alert-"+v),
		property.Role("alert"),
		property.Value(text),
	)
}
