package live

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	gl "github.com/tomo3110/gerbera/live"
	"github.com/tomo3110/gerbera/property"
	"github.com/tomo3110/gerbera/ui"
)

// Confirm renders a modal-based confirmation dialog.
// When open is true, a dialog with title, message, and Confirm/Cancel buttons is shown.
func Confirm(open bool, title, message, confirmEvent, cancelEvent string) gerbera.ComponentFunc {
	return Modal(open, cancelEvent,
		ModalHeader(title, cancelEvent),
		ModalBody(
			dom.P(property.Value(message)),
		),
		ModalFooter(
			ui.Button("Cancel", ui.ButtonOutline, gl.Click(cancelEvent)),
			ui.Button("OK", ui.ButtonPrimary, gl.Click(confirmEvent)),
		),
	)
}
