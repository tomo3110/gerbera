package live

import (
	"fmt"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	gl "github.com/tomo3110/gerbera/live"
	"github.com/tomo3110/gerbera/property"
)

// Tab defines a single tab with its label and panel content.
type Tab struct {
	Label   string
	Content gerbera.ComponentFunc
}

// Tabs renders an accessible tab UI with LiveView event bindings.
// id is used as the ARIA ID prefix.
// activeIndex is the 0-based index of the currently active tab.
// changeEvent is the event name fired when a tab is clicked;
// the payload contains {"value": "<index>"}.
func Tabs(id string, activeIndex int, tabs []Tab, changeEvent string) gerbera.ComponentFunc {
	// Build tab buttons
	var buttons gerbera.Components
	for i, tab := range tabs {
		active := i == activeIndex
		tabID := fmt.Sprintf("%s-tab-%d", id, i)
		panelID := fmt.Sprintf("%s-panel-%d", id, i)

		attrs := gerbera.Components{
			property.Class("g-tab"),
			property.ClassIf(active, "g-tab-active"),
			property.Role("tab"),
			property.ID(tabID),
			property.AriaSelected(active),
			property.AriaControls(panelID),
			gl.Click(changeEvent),
			gl.ClickValue(fmt.Sprintf("%d", i)),
		}
		if active {
			attrs = append(attrs, property.Tabindex(0))
		} else {
			attrs = append(attrs, property.Tabindex(-1))
		}
		attrs = append(attrs, property.Value(tab.Label))
		buttons = append(buttons, dom.Button(attrs...))
	}

	tablist := dom.Div(
		append(gerbera.Components{
			property.Class("g-tablist"),
			property.Role("tablist"),
			property.AriaLabel("Tabs"),
		}, buttons...)...,
	)

	// Build tab panels
	var panels gerbera.Components
	for i, tab := range tabs {
		active := i == activeIndex
		tabID := fmt.Sprintf("%s-tab-%d", id, i)
		panelID := fmt.Sprintf("%s-panel-%d", id, i)

		attrs := gerbera.Components{
			property.Class("g-tabpanel"),
			property.Role("tabpanel"),
			property.ID(panelID),
			property.AriaLabelledBy(tabID),
		}
		if !active {
			attrs = append(attrs, property.Attr("hidden", "hidden"))
		}
		if tab.Content != nil {
			attrs = append(attrs, tab.Content)
		}
		panels = append(panels, dom.Div(attrs...))
	}

	// Assemble wrapper
	wrapper := gerbera.Components{
		property.Class("g-tabs"),
		property.ID(id),
	}
	wrapper = append(wrapper, tablist)
	wrapper = append(wrapper, panels...)

	return dom.Div(wrapper...)
}
