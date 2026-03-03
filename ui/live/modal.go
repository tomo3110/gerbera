package live

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	"github.com/tomo3110/gerbera/property"
)

// Modal renders a dialog overlay panel.
// When open is true, the overlay and panel are shown.
// Clicking the backdrop or the close button fires closeEvent.
func Modal(open bool, closeEvent string, children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return expr.If(open,
		dom.Div(
			property.Class("g-modal-overlay"),
			property.Role("dialog"),
			property.AriaLabel("Modal"),
			gl.Click(closeEvent),
			dom.Div(
				append([]gerbera.ComponentFunc{
					property.Class("g-modal-panel"),
					// Stop click propagation by re-binding click to a no-op
					// The panel itself doesn't fire the close event
					property.Attr("onclick", "event.stopPropagation()"),
				}, children...)...,
			),
		),
	)
}

// ModalHeader renders a modal header with a title and close button.
func ModalHeader(title, closeEvent string) gerbera.ComponentFunc {
	return dom.Div(
		property.Class("g-modal-header"),
		dom.H3(property.Class("g-modal-title"), property.Value(title)),
		dom.Button(
			property.Class("g-modal-close"),
			gl.Click(closeEvent),
			property.AriaLabel("Close"),
			property.Value("\u00d7"),
		),
	)
}

// ModalBody renders the main content area of a modal.
func ModalBody(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return dom.Div(append([]gerbera.ComponentFunc{property.Class("g-modal-body")}, children...)...)
}

// ModalFooter renders a modal footer for action buttons.
func ModalFooter(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return dom.Div(append([]gerbera.ComponentFunc{property.Class("g-modal-footer")}, children...)...)
}
