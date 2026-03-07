package live

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	gl "github.com/tomo3110/gerbera/live"
	"github.com/tomo3110/gerbera/property"
)

// Modal renders a dialog overlay panel.
// The wrapper element is always present in the DOM; when open is false
// it carries the hidden attribute so the positional diff stays stable.
// Clicking the backdrop or the close button fires closeEvent.
func Modal(open bool, closeEvent string, children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	wrapper := gerbera.Components{
		dom.Div(
			property.Class("g-modal-overlay"),
			property.Role("dialog"),
			property.AriaLabel("Modal"),
			gl.Click(closeEvent),
			dom.Div(
				append(gerbera.Components{
					property.Class("g-modal-panel"),
					// Stop click propagation by re-binding click to a no-op
					// The panel itself doesn't fire the close event
					property.Attr("onclick", "event.stopPropagation()"),
				}, children...)...,
			),
		),
	}
	if !open {
		wrapper = append(gerbera.Components{property.Attr("hidden", "")}, wrapper...)
	}
	return dom.Div(wrapper...)
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
	return dom.Div(append(gerbera.Components{property.Class("g-modal-body")}, children...)...)
}

// ModalFooter renders a modal footer for action buttons.
func ModalFooter(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return dom.Div(append(gerbera.Components{property.Class("g-modal-footer")}, children...)...)
}
