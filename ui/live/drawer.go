package live

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	"github.com/tomo3110/gerbera/property"
)

// Drawer renders a slide-out panel from the left or right edge.
// side: "left" (default) or "right".
// When open is true, the overlay and panel are shown.
// Clicking the backdrop or the close button fires closeEvent.
func Drawer(open bool, closeEvent, side string, children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	if side == "" {
		side = "left"
	}
	return expr.If(open,
		dom.Div(
			dom.Div(
				property.Class("g-drawer-overlay"),
				gl.Click(closeEvent),
			),
			dom.Div(
				append([]gerbera.ComponentFunc{
					property.Class("g-drawer", "g-drawer-"+side),
					property.Role("dialog"),
					property.AriaLabel("Drawer"),
				}, children...)...,
			),
		),
	)
}

// DrawerHeader renders a drawer header with title and close button.
func DrawerHeader(title, closeEvent string) gerbera.ComponentFunc {
	return dom.Div(
		property.Class("g-drawer-header"),
		dom.H3(property.Class("g-drawer-title"), property.Value(title)),
		dom.Button(
			property.Class("g-drawer-close"),
			gl.Click(closeEvent),
			property.AriaLabel("Close"),
			property.Value("\u00d7"),
		),
	)
}

// DrawerBody renders the main scrollable content area of a drawer.
func DrawerBody(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return dom.Div(append([]gerbera.ComponentFunc{property.Class("g-drawer-body")}, children...)...)
}
