package ui

import (
	"fmt"
	"strconv"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

// NumberInputOpts configures a NumberInput.
type NumberInputOpts struct {
	Min      *int
	Max      *int
	Step     int
	Disabled bool
}

// NumberInput renders a number input with decrement and increment buttons.
func NumberInput(name string, value int, opts NumberInputOpts, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	step := opts.Step
	if step <= 0 {
		step = 1
	}

	minAttr := []gerbera.ComponentFunc{}
	maxAttr := []gerbera.ComponentFunc{}
	ariaAttrs := []gerbera.ComponentFunc{
		property.AriaValueNow(value),
	}

	if opts.Min != nil {
		minAttr = append(minAttr, property.Attr("min", strconv.Itoa(*opts.Min)))
		ariaAttrs = append(ariaAttrs, property.AriaValueMin(*opts.Min))
	}
	if opts.Max != nil {
		maxAttr = append(maxAttr, property.Attr("max", strconv.Itoa(*opts.Max)))
		ariaAttrs = append(ariaAttrs, property.AriaValueMax(*opts.Max))
	}

	decDisabled := opts.Disabled || (opts.Min != nil && value <= *opts.Min)
	incDisabled := opts.Disabled || (opts.Max != nil && value >= *opts.Max)

	inputAttrs := []gerbera.ComponentFunc{
		property.Class("g-numberinput-field"),
		property.Type("number"),
		property.Name(name),
		property.Attr("value", strconv.Itoa(value)),
		property.Attr("step", strconv.Itoa(step)),
		property.Disabled(opts.Disabled),
	}
	inputAttrs = append(inputAttrs, minAttr...)
	inputAttrs = append(inputAttrs, maxAttr...)

	wrapAttrs := []gerbera.ComponentFunc{
		property.Class("g-numberinput"),
		property.Role("spinbutton"),
		property.AriaLabel(name),
	}
	wrapAttrs = append(wrapAttrs, ariaAttrs...)
	wrapAttrs = append(wrapAttrs, extra...)
	wrapAttrs = append(wrapAttrs,
		dom.Button(
			property.Class("g-numberinput-btn", "g-numberinput-dec"),
			property.Attr("type", "button"),
			property.Attr("tabindex", "-1"),
			property.AriaLabel(fmt.Sprintf("Decrease %s", name)),
			property.Disabled(decDisabled),
			property.Value("\u2212"),
		),
		dom.Input(inputAttrs...),
		dom.Button(
			property.Class("g-numberinput-btn", "g-numberinput-inc"),
			property.Attr("type", "button"),
			property.Attr("tabindex", "-1"),
			property.AriaLabel(fmt.Sprintf("Increase %s", name)),
			property.Disabled(incDisabled),
			property.Value("+"),
		),
	)

	return dom.Div(wrapAttrs...)
}
