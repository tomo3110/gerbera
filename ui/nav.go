package ui

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

// Sidebar renders a vertical navigation sidebar.
func Sidebar(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return dom.Nav(append(gerbera.Components{property.Class("g-sidebar")}, children...)...)
}

// SidebarHeader renders the title section of the sidebar.
func SidebarHeader(title string) gerbera.ComponentFunc {
	return dom.Div(
		property.Class("g-sidebar-header"),
		property.Value(title),
	)
}

// SidebarLink renders a navigation link in the sidebar.
// If active is true, the active style is applied.
func SidebarLink(href, label string, active bool, attrs ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	a := gerbera.Components{
		property.Class("g-sidebar-link"),
		property.ClassIf(active, "g-sidebar-link-active"),
		property.Href(href),
		property.Value(label),
	}
	if active {
		a = append(a, property.AriaCurrent("page"))
	}
	a = append(a, attrs...)
	return dom.A(a...)
}

// SidebarDivider renders a horizontal divider line in the sidebar.
func SidebarDivider() gerbera.ComponentFunc {
	return dom.Div(property.Class("g-sidebar-divider"))
}

// BreadcrumbItem defines one segment of a breadcrumb trail.
type BreadcrumbItem struct {
	Label string
	Href  string // empty string means current page (no link)
}

// Breadcrumb renders a breadcrumb navigation trail.
func Breadcrumb(items ...BreadcrumbItem) gerbera.ComponentFunc {
	var parts gerbera.Components
	for i, item := range items {
		if i > 0 {
			parts = append(parts, dom.Li(
				property.Class("g-breadcrumb-sep"),
				property.AriaHidden(true),
				property.Value("/"),
			))
		}
		if item.Href == "" {
			parts = append(parts, dom.Li(
				property.Class("g-breadcrumb-current"),
				property.AriaCurrent("page"),
				property.Value(item.Label),
			))
		} else {
			parts = append(parts, dom.Li(
				dom.A(property.Href(item.Href), property.Value(item.Label)),
			))
		}
	}
	return dom.Ul(append(gerbera.Components{
		property.Class("g-breadcrumb"),
		property.AriaLabel("Breadcrumb"),
	}, parts...)...)
}
