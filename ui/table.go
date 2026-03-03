package ui

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

// StyledTable renders a <table> with the g-table class applied.
func StyledTable(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return dom.Table(append([]gerbera.ComponentFunc{property.Class("g-table")}, children...)...)
}

// THead renders a <thead> with a single header row from string labels.
func THead(headers ...string) gerbera.ComponentFunc {
	var ths []gerbera.ComponentFunc
	for _, h := range headers {
		ths = append(ths, dom.Th(property.Value(h)))
	}
	return dom.Thead(dom.Tr(ths...))
}
