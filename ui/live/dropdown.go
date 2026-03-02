package live

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	"github.com/tomo3110/gerbera/property"
)

// Dropdown renders a toggle-able dropdown menu.
// trigger is the element that opens the dropdown (usually a button).
// menu is the dropdown content shown when open is true.
// toggleEvent fires when the trigger is clicked.
func Dropdown(open bool, toggleEvent string, trigger, menu gerbera.ComponentFunc) gerbera.ComponentFunc {
	return dom.Div(
		property.Class("g-dropdown"),
		dom.Div(
			gl.Click(toggleEvent),
			property.AriaHasPopup("true"),
			property.AriaExpanded(open),
			trigger,
		),
		expr.If(open, dom.Div(
			property.Class("g-dropdown-menu"),
			property.Role("menu"),
			menu,
		)),
	)
}
