package ui

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

// ButtonGroupItem represents a single button in a ButtonGroup.
type ButtonGroupItem struct {
	Label  string
	Value  string
	Active bool
}

// ButtonGroupOpts configures a ButtonGroup component.
type ButtonGroupOpts struct {
	Small      bool
	ClickEvent string // gerbera-click on buttons; payload: {"value": item.Value}
}

// ButtonGroupSmall is a modifier to render a smaller button group.
var ButtonGroupSmall = property.ClassIf(true, "g-btngroup-sm")

// ButtonGroup renders a segmented control / button group.
func ButtonGroup(items []ButtonGroupItem, opts ButtonGroupOpts, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	wrapAttrs := []gerbera.ComponentFunc{
		property.Class("g-btngroup"),
		property.ClassIf(opts.Small, "g-btngroup-sm"),
		property.Role("group"),
	}
	wrapAttrs = append(wrapAttrs, extra...)

	for _, item := range items {
		btnAttrs := []gerbera.ComponentFunc{
			property.Class("g-btngroup-btn"),
			property.ClassIf(item.Active, "g-btngroup-active"),
			property.Attr("type", "button"),
			property.Attr("aria-pressed", boolStr(item.Active)),
			property.Value(item.Label),
		}
		if opts.ClickEvent != "" {
			btnAttrs = append(btnAttrs,
				property.Attr("gerbera-click", opts.ClickEvent),
				property.Attr("gerbera-value", item.Value),
			)
		} else {
			btnAttrs = append(btnAttrs, property.Attr("data-value", item.Value))
		}
		wrapAttrs = append(wrapAttrs, dom.Button(btnAttrs...))
	}

	return dom.Div(wrapAttrs...)
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
