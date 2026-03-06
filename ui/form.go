package ui

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	"github.com/tomo3110/gerbera/property"
)

// FormGroup renders a form field wrapper with bottom margin.
func FormGroup(children ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	return dom.Div(append(gerbera.Components{property.Class("g-form-group")}, children...)...)
}

// FormLabel renders a styled <label> element.
func FormLabel(text, forID string) gerbera.ComponentFunc {
	return dom.Label(
		property.Class("g-form-label"),
		property.For(forID),
		property.Value(text),
	)
}

// FormInput renders a styled <input> element.
// opts can include property.Type, property.Placeholder, live.Input, etc.
func FormInput(name string, opts ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := gerbera.Components{
		property.Class("g-form-input"),
		property.Name(name),
	}
	attrs = append(attrs, opts...)
	return dom.Input(attrs...)
}

// FormOption defines a value-label pair for FormSelect.
type FormOption struct {
	Value string
	Label string
}

// FormSelect renders a styled <select> element.
func FormSelect(name string, options []FormOption, opts ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	var optEls gerbera.Components
	for _, o := range options {
		optEls = append(optEls, dom.Option(
			property.Attr("value", o.Value),
			property.Value(o.Label),
		))
	}
	attrs := gerbera.Components{
		property.Class("g-form-select"),
		property.Name(name),
	}
	attrs = append(attrs, opts...)
	attrs = append(attrs, optEls...)
	return dom.Select(attrs...)
}

// FormTextarea renders a styled <textarea> element.
func FormTextarea(name string, opts ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	attrs := gerbera.Components{
		property.Class("g-form-textarea"),
		property.Name(name),
	}
	attrs = append(attrs, opts...)
	return dom.Textarea(attrs...)
}

// FormError renders a validation error message below a form field.
// If message is empty, nothing is rendered.
func FormError(message string) gerbera.ComponentFunc {
	return expr.If(message != "",
		dom.Div(
			property.Class("g-form-error"),
			property.Role("alert"),
			property.Value(message),
		),
	)
}

// FormInputError is a convenience option to apply error styling to a FormInput.
var FormInputError gerbera.ComponentFunc = property.ClassIf(true, "g-form-input-error")

// FormTextareaError is a convenience option to apply error styling to a FormTextarea.
var FormTextareaError gerbera.ComponentFunc = property.ClassIf(true, "g-form-textarea-error")

// FormSelectError is a convenience option to apply error styling to a FormSelect.
var FormSelectError gerbera.ComponentFunc = property.ClassIf(true, "g-form-select-error")

// Checkbox renders a styled checkbox with label.
func Checkbox(name, label string, checked bool, opts ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	inputAttrs := gerbera.Components{
		property.Attr("type", "checkbox"),
		property.Name(name),
	}
	if checked {
		inputAttrs = append(inputAttrs, property.Attr("checked", "checked"))
	}
	inputAttrs = append(inputAttrs, opts...)

	return dom.Label(
		property.Class("g-form-check"),
		dom.Input(inputAttrs...),
		dom.Span(property.Class("g-form-check-label"), property.Value(label)),
	)
}

// Radio renders a styled radio button with label.
func Radio(name, value, label string, checked bool, opts ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	inputAttrs := gerbera.Components{
		property.Attr("type", "radio"),
		property.Name(name),
		property.Attr("value", value),
	}
	if checked {
		inputAttrs = append(inputAttrs, property.Attr("checked", "checked"))
	}
	inputAttrs = append(inputAttrs, opts...)

	return dom.Label(
		property.Class("g-form-check"),
		dom.Input(inputAttrs...),
		dom.Span(property.Class("g-form-check-label"), property.Value(label)),
	)
}
