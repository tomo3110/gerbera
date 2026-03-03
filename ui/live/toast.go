package live

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	gl "github.com/tomo3110/gerbera/live"
	"github.com/tomo3110/gerbera/property"
)

// Toast renders a notification message in the top-right corner.
// Clicking the dismiss button fires dismissEvent.
// variant: "info", "success", "warning", "danger".
func Toast(message, variant, dismissEvent string) gerbera.ComponentFunc {
	if variant == "" {
		variant = "info"
	}
	return dom.Div(
		property.Class("g-toast", "g-toast-"+variant),
		property.Role("alert"),
		property.AriaLive("polite"),
		dom.Span(property.Class("g-toast-msg"), property.Value(message)),
		dom.Button(
			property.Class("g-toast-close"),
			gl.Click(dismissEvent),
			property.AriaLabel("Dismiss"),
			property.Value("\u00d7"),
		),
	)
}
