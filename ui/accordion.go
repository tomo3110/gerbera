package ui

import (
	"fmt"
	"strconv"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	"github.com/tomo3110/gerbera/property"
)

// AccordionItem represents a single collapsible section.
type AccordionItem struct {
	Title   string
	Content gerbera.ComponentFunc
	Open    bool
}

// AccordionOpts configures an Accordion component.
type AccordionOpts struct {
	Exclusive   bool
	ToggleEvent string // gerbera-click on header; payload: {"value": "N"} (0-based index)
}

// Accordion renders a set of collapsible sections.
// When ToggleEvent is empty, uses native <details>/<summary> (browser-controlled).
// When ToggleEvent is set, uses <button>/<div> with ARIA attributes (server-controlled).
func Accordion(items []AccordionItem, opts AccordionOpts, extra ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	wrapAttrs := []gerbera.ComponentFunc{
		property.Class("g-accordion"),
	}
	wrapAttrs = append(wrapAttrs, extra...)

	if opts.ToggleEvent == "" {
		// Native details/summary mode
		for _, item := range items {
			detailAttrs := []gerbera.ComponentFunc{
				property.Class("g-accordion-item"),
			}
			if item.Open {
				detailAttrs = append(detailAttrs, property.Attr("open", "open"))
			}
			detailAttrs = append(detailAttrs,
				dom.Summary(
					property.Class("g-accordion-header"),
					property.Value(item.Title),
				),
				dom.Div(
					property.Class("g-accordion-body"),
					item.Content,
				),
			)
			wrapAttrs = append(wrapAttrs, dom.Details(detailAttrs...))
		}
	} else {
		// Server-controlled mode with ARIA
		for i, item := range items {
			idx := strconv.Itoa(i)
			headerID := fmt.Sprintf("g-accordion-header-%s", idx)
			bodyID := fmt.Sprintf("g-accordion-body-%s", idx)

			headerAttrs := []gerbera.ComponentFunc{
				property.Class("g-accordion-header"),
				property.ID(headerID),
				property.Attr("type", "button"),
				property.AriaExpanded(item.Open),
				property.AriaControls(bodyID),
				property.Value(item.Title),
				property.Attr("gerbera-click", opts.ToggleEvent),
				property.Attr("gerbera-value", idx),
			}

			bodyAttrs := []gerbera.ComponentFunc{
				property.Class("g-accordion-body"),
				property.ID(bodyID),
				property.Role("region"),
				property.AriaLabelledBy(headerID),
			}

			itemAttrs := []gerbera.ComponentFunc{
				property.Class("g-accordion-item"),
				property.ClassIf(item.Open, "g-accordion-open"),
				dom.Button(headerAttrs...),
				expr.If(item.Open, dom.Div(append(bodyAttrs, item.Content)...)),
			}

			wrapAttrs = append(wrapAttrs, dom.Div(itemAttrs...))
		}
	}

	return dom.Div(wrapAttrs...)
}
