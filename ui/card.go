package ui

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

// Card renders a styled card container.
func Card(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return dom.Div(append(gerbera.Components{property.Class("g-card")}, children...)...)
}

// CardHeader renders a card header with a title and optional action buttons.
func CardHeader(title string, actions ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	header := gerbera.Components{
		property.Class("g-card-header"),
		dom.H3(property.Class("g-card-header-title"), property.Value(title)),
	}
	if len(actions) > 0 {
		header = append(header, dom.Div(append(gerbera.Components{property.Class("g-card-header-actions")}, actions...)...))
	}
	return dom.Div(header...)
}

// CardBody renders a card body section with standard padding.
func CardBody(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return dom.Div(append(gerbera.Components{property.Class("g-card-body")}, children...)...)
}

// CardFooter renders a card footer.
func CardFooter(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return dom.Div(append(gerbera.Components{property.Class("g-card-footer")}, children...)...)
}
