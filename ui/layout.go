package ui

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

// AdminShell renders the top-level admin layout with a sidebar and content area.
// On mobile (<=768px), the sidebar is hidden and a MobileHeader should be used
// together with a Drawer to show navigation.
func AdminShell(sidebar, content gerbera.ComponentFunc) gerbera.ComponentFunc {
	return dom.Div(
		property.Class("g-admin-shell"),
		sidebar,
		dom.Div(
			property.Class("g-admin-content"),
			content,
		),
	)
}

// MobileHeader renders a mobile-only header bar with a hamburger button and title.
// hamburgerAttrs can include live.Click to open a drawer.
func MobileHeader(title string, hamburgerAttrs ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	btnAttrs := []gerbera.ComponentFunc{
		property.Class("g-hamburger"),
		property.AriaLabel("Menu"),
	}
	btnAttrs = append(btnAttrs, hamburgerAttrs...)
	btnAttrs = append(btnAttrs, Icon("menu", "md"))

	return dom.Div(
		property.Class("g-mobile-header"),
		dom.Button(btnAttrs...),
		dom.Span(property.Class("g-mobile-title"), property.Value(title)),
	)
}

// PageHeader renders a page header bar with title and optional action elements.
func PageHeader(title string, children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	inner := []gerbera.ComponentFunc{
		property.Class("g-page-header"),
		dom.H1(property.Class("g-page-header-title"), property.Value(title)),
	}
	inner = append(inner, children...)
	return dom.Div(inner...)
}
