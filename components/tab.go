package components

import (
	"fmt"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
	"github.com/tomo3110/gerbera/styles"
)

// Tab defines a single tab with its label, content, and optional button attributes.
// ButtonAttrs can be used to inject event bindings (e.g., live.Click) without
// depending on the live package.
type Tab struct {
	Label       string                  // Tab button text
	Content     gerbera.ComponentFunc   // Panel content
	ButtonAttrs gerbera.Components // Additional attributes for the tab button
}

// Tabs renders an accessible tab UI.
// id is the ARIA ID prefix, activeIndex is the 0-based active tab index.
// wrapperAttrs are additional ComponentFuncs applied to the outer wrapper div.
func Tabs(id string, activeIndex int, tabs []Tab, wrapperAttrs ...gerbera.ComponentFunc) gerbera.ComponentFunc {
	// Build tab buttons
	var buttons gerbera.Components
	for i, tab := range tabs {
		active := i == activeIndex
		tabID := fmt.Sprintf("%s-tab-%d", id, i)
		panelID := fmt.Sprintf("%s-panel-%d", id, i)

		attrs := gerbera.Components{
			property.Role("tab"),
			property.ID(tabID),
			property.AriaSelected(active),
			property.AriaControls(panelID),
			property.ClassIf(true, "gerbera-tab"),
			property.ClassIf(active, "gerbera-tab-active"),
		}
		if active {
			attrs = append(attrs, property.Tabindex(0))
		} else {
			attrs = append(attrs, property.Tabindex(-1))
		}
		attrs = append(attrs, tab.ButtonAttrs...)
		attrs = append(attrs, property.Value(tab.Label))

		buttons = append(buttons, dom.Button(attrs...))
	}

	tablist := dom.Div(
		append(gerbera.Components{
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
			property.Role("tabpanel"),
			property.ID(panelID),
			property.AriaLabelledBy(tabID),
			property.ClassIf(true, "gerbera-tabpanel"),
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
		property.ClassIf(true, "gerbera-tabs"),
		property.ID(id),
	}
	wrapper = append(wrapper, wrapperAttrs...)
	wrapper = append(wrapper, tablist)
	wrapper = append(wrapper, panels...)

	return dom.Div(wrapper...)
}

// TabsDefaultCSS returns a <style> element with minimal default CSS for the tab component.
func TabsDefaultCSS() gerbera.ComponentFunc {
	return styles.CSS(
		`.gerbera-tabs { width: 100%; }` +
			` [role="tablist"] { display: flex; gap: 0; border-bottom: 2px solid #ddd; }` +
			` .gerbera-tab { padding: 8px 16px; border: none; background: transparent; cursor: pointer; border-bottom: 2px solid transparent; margin-bottom: -2px; font-size: inherit; }` +
			` .gerbera-tab:hover { background: #f0f0f0; }` +
			` .gerbera-tab-active { border-bottom-color: #0066cc; color: #0066cc; font-weight: bold; }` +
			` .gerbera-tabpanel { padding: 16px 0; }`)
}
