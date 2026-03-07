package ui

import (
	"fmt"
	"strconv"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	"github.com/tomo3110/gerbera/property"
)

// StepStatus represents the state of a step.
type StepStatus string

const (
	StepCompleted StepStatus = "completed"
	StepActive    StepStatus = "active"
	StepUpcoming  StepStatus = "upcoming"
)

// Step represents a single step in a Stepper.
type Step struct {
	Label       string
	Description string
	Status      StepStatus
}

// StepperOpts configures a Stepper component.
type StepperOpts struct {
	Vertical   bool
	ClickEvent string // gerbera-click on completed steps; payload: {"value": "N"} (0-based index)
}

// StepperVertical is a modifier to render the stepper in vertical orientation.
var StepperVertical = property.ClassIf(true, "g-stepper-vertical")

// Stepper renders a step-by-step progress indicator.
func Stepper(steps []Step, opts StepperOpts, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	wrapAttrs := gerbera.Components{
		property.Class("g-stepper"),
		property.ClassIf(opts.Vertical, "g-stepper-vertical"),
		property.Role("list"),
	}
	wrapAttrs = append(wrapAttrs, extra...)

	for i, step := range steps {
		status := step.Status
		if status == "" {
			status = StepUpcoming
		}

		stepAttrs := gerbera.Components{
			property.Class("g-stepper-step", fmt.Sprintf("g-stepper-%s", status)),
			property.Role("listitem"),
		}
		if status == StepActive {
			stepAttrs = append(stepAttrs, property.AriaCurrent("step"))
		}

		// Indicator
		var indicatorContent gerbera.ComponentFunc
		if status == StepCompleted {
			indicatorContent = dom.Span(property.Value("\u2713"))
		} else {
			indicatorContent = dom.Span(property.Value(fmt.Sprintf("%d", i+1)))
		}
		indicator := dom.Div(
			property.Class("g-stepper-indicator"),
			indicatorContent,
		)

		// Content (label + optional description)
		contentChildren := gerbera.Components{
			property.Class("g-stepper-content"),
			dom.Span(property.Class("g-stepper-label"), property.Value(step.Label)),
		}
		contentChildren = append(contentChildren,
			expr.If(step.Description != "",
				dom.Span(property.Class("g-stepper-desc"), property.Value(step.Description)),
			),
		)
		content := dom.Div(contentChildren...)

		stepAttrs = append(stepAttrs, indicator, content)

		// Click event on completed steps
		if opts.ClickEvent != "" && status == StepCompleted {
			stepAttrs = append(stepAttrs,
				property.Attr("gerbera-click", opts.ClickEvent),
				property.Attr("gerbera-value", strconv.Itoa(i)),
				property.Attr("style", "cursor:pointer"),
			)
		}

		// Connector (not on last step)
		if i < len(steps)-1 {
			stepAttrs = append(stepAttrs, dom.Div(property.Class("g-stepper-connector")))
		}

		wrapAttrs = append(wrapAttrs, dom.Div(stepAttrs...))
	}

	return dom.Div(wrapAttrs...)
}
