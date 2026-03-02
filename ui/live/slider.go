package live

import (
	"strconv"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	gl "github.com/tomo3110/gerbera/live"
	"github.com/tomo3110/gerbera/property"
)

// SliderOpts configures a live Slider.
type SliderOpts struct {
	Name       string
	Value      int
	Min        int
	Max        int
	Step       int
	Label      string
	InputEvent string // event fired on input change
}

// Slider renders a live range slider with label and current value display.
func Slider(opts SliderOpts) gerbera.ComponentFunc {
	if opts.Max == 0 && opts.Min == 0 {
		opts.Max = 100
	}
	step := opts.Step
	if step <= 0 {
		step = 1
	}

	label := opts.Label
	if label == "" {
		label = opts.Name
	}

	inputAttrs := []gerbera.ComponentFunc{
		property.Class("g-slider-input"),
		property.Type("range"),
		property.Name(opts.Name),
		property.Attr("value", strconv.Itoa(opts.Value)),
		property.Attr("min", strconv.Itoa(opts.Min)),
		property.Attr("max", strconv.Itoa(opts.Max)),
		property.Attr("step", strconv.Itoa(step)),
		property.Role("slider"),
		property.AriaValueNow(opts.Value),
		property.AriaValueMin(opts.Min),
		property.AriaValueMax(opts.Max),
		property.AriaLabel(label),
	}
	if opts.InputEvent != "" {
		inputAttrs = append(inputAttrs, gl.Input(opts.InputEvent))
	}

	return dom.Div(
		property.Class("g-slider"),
		dom.Div(
			property.Class("g-slider-header"),
			dom.Span(property.Class("g-slider-label"), property.Value(label)),
			dom.Span(property.Class("g-slider-value"), property.Value(strconv.Itoa(opts.Value))),
		),
		dom.Input(inputAttrs...),
	)
}
