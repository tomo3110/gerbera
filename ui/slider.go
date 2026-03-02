package ui

import (
	"strconv"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

// SliderOpts configures a Slider.
type SliderOpts struct {
	Min        int
	Max        int
	Step       int
	Label      string
	InputEvent string // gerbera-input on range input
}

// Slider renders a range slider with label and current value display.
func Slider(name string, value int, opts SliderOpts, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	if opts.Max == 0 && opts.Min == 0 {
		opts.Max = 100
	}
	step := opts.Step
	if step <= 0 {
		step = 1
	}

	label := opts.Label
	if label == "" {
		label = name
	}

	inputAttrs := []gerbera.ComponentFunc{
		property.Class("g-slider-input"),
		property.Type("range"),
		property.Name(name),
		property.Attr("value", strconv.Itoa(value)),
		property.Attr("min", strconv.Itoa(opts.Min)),
		property.Attr("max", strconv.Itoa(opts.Max)),
		property.Attr("step", strconv.Itoa(step)),
		property.Role("slider"),
		property.AriaValueNow(value),
		property.AriaValueMin(opts.Min),
		property.AriaValueMax(opts.Max),
		property.AriaLabel(label),
	}
	if opts.InputEvent != "" {
		inputAttrs = append(inputAttrs, property.Attr("gerbera-input", opts.InputEvent))
	}

	wrapAttrs := []gerbera.ComponentFunc{
		property.Class("g-slider"),
	}
	wrapAttrs = append(wrapAttrs, extra...)
	wrapAttrs = append(wrapAttrs,
		dom.Div(
			property.Class("g-slider-header"),
			dom.Span(property.Class("g-slider-label"), property.Value(label)),
			dom.Span(property.Class("g-slider-value"), property.Value(strconv.Itoa(value))),
		),
		dom.Input(inputAttrs...),
	)

	return dom.Div(wrapAttrs...)
}
