package live

import (
	"fmt"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	"github.com/tomo3110/gerbera/property"
	"github.com/tomo3110/gerbera/ui"
)

// SearchSelectOpts configures a SearchSelect widget.
type SearchSelectOpts struct {
	Name        string           // form field name
	Query       string           // current search text
	Options     []ui.FormOption  // filtered options to display
	Selected    string           // currently selected value
	Open        bool             // whether the dropdown is open
	Placeholder string           // input placeholder text
	InputEvent  string           // event fired on input change (payload: field value via gerbera-input)
	SelectEvent string           // event fired when an option is selected (payload: {"value": "..."})
	FocusEvent  string           // event fired on input focus (to open dropdown)
	KeydownEvent string          // event fired on keydown (payload: {"key": "..."})
	HighlightIndex int           // index of keyboard-highlighted option (-1 = none)
}

// SearchSelect renders a filterable combobox.
// The caller's View should handle InputEvent to filter options and
// SelectEvent to set the selected value.
func SearchSelect(opts SearchSelectOpts) gerbera.ComponentFunc {
	ph := opts.Placeholder
	if ph == "" {
		ph = "Search..."
	}

	inputAttrs := []gerbera.ComponentFunc{
		property.Class("g-searchselect-input"),
		property.Name(opts.Name),
		property.Placeholder(ph),
		property.Attr("value", opts.Query),
		property.Attr("autocomplete", "off"),
		property.Role("combobox"),
		property.AriaExpanded(opts.Open),
	}
	if opts.InputEvent != "" {
		inputAttrs = append(inputAttrs, gl.Input(opts.InputEvent))
	}
	if opts.FocusEvent != "" {
		inputAttrs = append(inputAttrs, gl.Focus(opts.FocusEvent))
	}
	if opts.KeydownEvent != "" {
		inputAttrs = append(inputAttrs, gl.Keydown(opts.KeydownEvent))
	}

	// aria-activedescendant for keyboard navigation
	if opts.Open && opts.HighlightIndex >= 0 && opts.HighlightIndex < len(opts.Options) {
		inputAttrs = append(inputAttrs,
			property.Attr("aria-activedescendant", fmt.Sprintf("g-ss-%s-%d", opts.Name, opts.HighlightIndex)),
		)
	}

	// Dropdown list
	var listContent gerbera.ComponentFunc
	if len(opts.Options) > 0 {
		var items []gerbera.ComponentFunc
		for i, o := range opts.Options {
			items = append(items, dom.Button(
				property.Class("g-searchselect-option"),
				property.ClassIf(o.Value == opts.Selected, "g-searchselect-option-active"),
				property.ClassIf(i == opts.HighlightIndex, "g-searchselect-option-highlight"),
				property.ID(fmt.Sprintf("g-ss-%s-%d", opts.Name, i)),
				property.Role("option"),
				gl.Click(opts.SelectEvent),
				gl.ClickValue(o.Value),
				property.Value(o.Label),
			))
		}
		listContent = dom.Div(
			append([]gerbera.ComponentFunc{
				property.Class("g-searchselect-list"),
				property.Role("listbox"),
			}, items...)...,
		)
	} else {
		listContent = dom.Div(
			property.Class("g-searchselect-list"),
			dom.Div(property.Class("g-searchselect-empty"), property.Value("No matches")),
		)
	}

	return dom.Div(
		property.Class("g-searchselect"),
		dom.Input(inputAttrs...),
		expr.If(opts.Open, listContent),
	)
}
