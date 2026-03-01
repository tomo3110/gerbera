package live

import (
	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
)

// FieldError represents a validation error for a form field.
type FieldError struct {
	Field   string
	Message string
}

// Errors is a collection of field validation errors.
type Errors []FieldError

// HasErrors returns true if there are any errors.
func (e Errors) HasErrors() bool {
	return len(e) > 0
}

// ErrorFor returns the error message for a given field name, or empty string.
func (e Errors) ErrorFor(field string) string {
	for _, err := range e {
		if err.Field == field {
			return err.Message
		}
	}
	return ""
}

// HasError returns true if the given field has an error.
func (e Errors) HasError(field string) bool {
	return e.ErrorFor(field) != ""
}

// AddError adds a validation error for a field.
func (e *Errors) AddError(field, message string) {
	*e = append(*e, FieldError{Field: field, Message: message})
}

// Clear removes all errors.
func (e *Errors) Clear() {
	*e = nil
}

// FieldOpts configures a form field.
type FieldOpts struct {
	// Type is the HTML input type (default "text").
	Type string
	// Label is the text label for the field.
	Label string
	// Placeholder is the input placeholder text.
	Placeholder string
	// Value is the current field value.
	Value string
	// Required marks the field as required.
	Required bool
	// Event is the LiveView event name for input changes (e.g., "validate").
	Event string
	// Error is the error message to display (if any).
	Error string
	// Extra attributes/children to apply to the input element.
	Extra []gerbera.ComponentFunc
	// LabelClass is the CSS class for the label element.
	LabelClass string
	// InputClass is the CSS class for the input element.
	InputClass string
	// ErrorClass is the CSS class for the error message element.
	ErrorClass string
	// WrapperClass is the CSS class for the wrapper div.
	WrapperClass string
}

// Field creates a form field group with label, input, and optional error message.
func Field(name string, opts FieldOpts) gerbera.ComponentFunc {
	inputType := opts.Type
	if inputType == "" {
		inputType = "text"
	}
	errorClass := opts.ErrorClass
	if errorClass == "" {
		errorClass = "gerbera-field-error"
	}
	wrapperClass := opts.WrapperClass
	if wrapperClass == "" {
		wrapperClass = "gerbera-field"
	}

	var children []gerbera.ComponentFunc

	// Label
	if opts.Label != "" {
		labelAttrs := []gerbera.ComponentFunc{
			property.Attr("for", name),
			property.Value(opts.Label),
		}
		if opts.LabelClass != "" {
			labelAttrs = append(labelAttrs, property.Class(opts.LabelClass))
		}
		children = append(children, dom.Label(labelAttrs...))
	}

	// Input
	inputAttrs := []gerbera.ComponentFunc{
		property.Attr("type", inputType),
		property.Name(name),
		property.ID(name),
	}
	if opts.Value != "" {
		inputAttrs = append(inputAttrs, property.Attr("value", opts.Value))
	}
	if opts.Placeholder != "" {
		inputAttrs = append(inputAttrs, property.Attr("placeholder", opts.Placeholder))
	}
	if opts.Required {
		inputAttrs = append(inputAttrs, property.Attr("required", "required"))
	}
	if opts.Event != "" {
		inputAttrs = append(inputAttrs, Input(opts.Event))
	}
	if opts.InputClass != "" {
		inputAttrs = append(inputAttrs, property.Class(opts.InputClass))
	}
	if opts.Error != "" {
		inputAttrs = append(inputAttrs, property.AriaInvalid(true))
		inputAttrs = append(inputAttrs, property.AriaDescribedBy(name+"-error"))
	}
	inputAttrs = append(inputAttrs, opts.Extra...)
	children = append(children, dom.Input(inputAttrs...))

	// Error message
	if opts.Error != "" {
		children = append(children, dom.Span(
			property.Class(errorClass),
			property.ID(name+"-error"),
			property.Role("alert"),
			property.AriaLive("polite"),
			property.Value(opts.Error),
		))
	}

	return dom.Div(
		append([]gerbera.ComponentFunc{property.Class(wrapperClass)}, children...)...,
	)
}

// TextareaField creates a form field group with label, textarea, and optional error message.
func TextareaField(name string, opts FieldOpts) gerbera.ComponentFunc {
	errorClass := opts.ErrorClass
	if errorClass == "" {
		errorClass = "gerbera-field-error"
	}
	wrapperClass := opts.WrapperClass
	if wrapperClass == "" {
		wrapperClass = "gerbera-field"
	}

	var children []gerbera.ComponentFunc

	// Label
	if opts.Label != "" {
		labelAttrs := []gerbera.ComponentFunc{
			property.Attr("for", name),
			property.Value(opts.Label),
		}
		if opts.LabelClass != "" {
			labelAttrs = append(labelAttrs, property.Class(opts.LabelClass))
		}
		children = append(children, dom.Label(labelAttrs...))
	}

	// Textarea
	textareaAttrs := []gerbera.ComponentFunc{
		property.Name(name),
		property.ID(name),
	}
	if opts.Value != "" {
		textareaAttrs = append(textareaAttrs, property.Value(opts.Value))
	}
	if opts.Placeholder != "" {
		textareaAttrs = append(textareaAttrs, property.Attr("placeholder", opts.Placeholder))
	}
	if opts.Required {
		textareaAttrs = append(textareaAttrs, property.Attr("required", "required"))
	}
	if opts.Event != "" {
		textareaAttrs = append(textareaAttrs, Input(opts.Event))
	}
	if opts.InputClass != "" {
		textareaAttrs = append(textareaAttrs, property.Class(opts.InputClass))
	}
	if opts.Error != "" {
		textareaAttrs = append(textareaAttrs, property.AriaInvalid(true))
		textareaAttrs = append(textareaAttrs, property.AriaDescribedBy(name+"-error"))
	}
	textareaAttrs = append(textareaAttrs, opts.Extra...)
	children = append(children, dom.Textarea(textareaAttrs...))

	// Error message
	if opts.Error != "" {
		children = append(children, dom.Span(
			property.Class(errorClass),
			property.ID(name+"-error"),
			property.Role("alert"),
			property.AriaLive("polite"),
			property.Value(opts.Error),
		))
	}

	return dom.Div(
		append([]gerbera.ComponentFunc{property.Class(wrapperClass)}, children...)...,
	)
}

// SelectField creates a form field group with label, select, and optional error message.
func SelectField(name string, options []SelectOption, opts FieldOpts) gerbera.ComponentFunc {
	errorClass := opts.ErrorClass
	if errorClass == "" {
		errorClass = "gerbera-field-error"
	}
	wrapperClass := opts.WrapperClass
	if wrapperClass == "" {
		wrapperClass = "gerbera-field"
	}

	var children []gerbera.ComponentFunc

	// Label
	if opts.Label != "" {
		labelAttrs := []gerbera.ComponentFunc{
			property.Attr("for", name),
			property.Value(opts.Label),
		}
		if opts.LabelClass != "" {
			labelAttrs = append(labelAttrs, property.Class(opts.LabelClass))
		}
		children = append(children, dom.Label(labelAttrs...))
	}

	// Select + options
	selectAttrs := []gerbera.ComponentFunc{
		property.Name(name),
		property.ID(name),
	}
	if opts.Required {
		selectAttrs = append(selectAttrs, property.Attr("required", "required"))
	}
	if opts.Event != "" {
		selectAttrs = append(selectAttrs, Change(opts.Event))
	}
	if opts.InputClass != "" {
		selectAttrs = append(selectAttrs, property.Class(opts.InputClass))
	}
	if opts.Error != "" {
		selectAttrs = append(selectAttrs, property.AriaInvalid(true))
		selectAttrs = append(selectAttrs, property.AriaDescribedBy(name+"-error"))
	}
	for _, opt := range options {
		optAttrs := []gerbera.ComponentFunc{
			property.Attr("value", opt.Value),
			property.Value(opt.Label),
		}
		if opt.Value == opts.Value {
			optAttrs = append(optAttrs, property.Attr("selected", "selected"))
		}
		if opt.Disabled {
			optAttrs = append(optAttrs, property.Attr("disabled", "disabled"))
		}
		selectAttrs = append(selectAttrs, dom.Option(optAttrs...))
	}
	selectAttrs = append(selectAttrs, opts.Extra...)
	children = append(children, dom.Select(selectAttrs...))

	// Error message
	if opts.Error != "" {
		children = append(children, dom.Span(
			property.Class(errorClass),
			property.ID(name+"-error"),
			property.Role("alert"),
			property.AriaLive("polite"),
			property.Value(opts.Error),
		))
	}

	return dom.Div(
		append([]gerbera.ComponentFunc{property.Class(wrapperClass)}, children...)...,
	)
}

// SelectOption represents an option in a select field.
type SelectOption struct {
	Value    string
	Label    string
	Disabled bool
}
