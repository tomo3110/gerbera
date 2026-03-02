package live

import (
	"fmt"
	"strconv"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	gl "github.com/tomo3110/gerbera/live"
	"github.com/tomo3110/gerbera/property"
)

// NumberInputOpts configures a live NumberInput.
type NumberInputOpts struct {
	Name           string
	Value          int
	Min            *int
	Max            *int
	Step           int
	Disabled       bool
	ChangeEvent    string // event fired when input value changes
	IncrementEvent string // event fired when + button is clicked
	DecrementEvent string // event fired when − button is clicked
}

// NumberInput renders a live number input with decrement and increment buttons.
func NumberInput(opts NumberInputOpts) gerbera.ComponentFunc {
	step := opts.Step
	if step <= 0 {
		step = 1
	}

	decDisabled := opts.Disabled || (opts.Min != nil && opts.Value <= *opts.Min)
	incDisabled := opts.Disabled || (opts.Max != nil && opts.Value >= *opts.Max)

	inputAttrs := []gerbera.ComponentFunc{
		property.Class("g-numberinput-field"),
		property.Type("number"),
		property.Name(opts.Name),
		property.Attr("value", strconv.Itoa(opts.Value)),
		property.Attr("step", strconv.Itoa(step)),
		property.Disabled(opts.Disabled),
	}
	if opts.Min != nil {
		inputAttrs = append(inputAttrs, property.Attr("min", strconv.Itoa(*opts.Min)))
	}
	if opts.Max != nil {
		inputAttrs = append(inputAttrs, property.Attr("max", strconv.Itoa(*opts.Max)))
	}
	if opts.ChangeEvent != "" {
		inputAttrs = append(inputAttrs, gl.Change(opts.ChangeEvent))
	}

	ariaAttrs := []gerbera.ComponentFunc{
		property.AriaValueNow(opts.Value),
	}
	if opts.Min != nil {
		ariaAttrs = append(ariaAttrs, property.AriaValueMin(*opts.Min))
	}
	if opts.Max != nil {
		ariaAttrs = append(ariaAttrs, property.AriaValueMax(*opts.Max))
	}

	decBtn := []gerbera.ComponentFunc{
		property.Class("g-numberinput-btn", "g-numberinput-dec"),
		property.Attr("type", "button"),
		property.Attr("tabindex", "-1"),
		property.AriaLabel(fmt.Sprintf("Decrease %s", opts.Name)),
		property.Disabled(decDisabled),
		property.Value("\u2212"),
	}
	if opts.DecrementEvent != "" {
		decBtn = append(decBtn, gl.Click(opts.DecrementEvent))
	}

	incBtn := []gerbera.ComponentFunc{
		property.Class("g-numberinput-btn", "g-numberinput-inc"),
		property.Attr("type", "button"),
		property.Attr("tabindex", "-1"),
		property.AriaLabel(fmt.Sprintf("Increase %s", opts.Name)),
		property.Disabled(incDisabled),
		property.Value("+"),
	}
	if opts.IncrementEvent != "" {
		incBtn = append(incBtn, gl.Click(opts.IncrementEvent))
	}

	wrapAttrs := []gerbera.ComponentFunc{
		property.Class("g-numberinput"),
		property.Role("spinbutton"),
		property.AriaLabel(opts.Name),
	}
	wrapAttrs = append(wrapAttrs, ariaAttrs...)
	wrapAttrs = append(wrapAttrs,
		dom.Button(decBtn...),
		dom.Input(inputAttrs...),
		dom.Button(incBtn...),
	)

	return dom.Div(wrapAttrs...)
}
