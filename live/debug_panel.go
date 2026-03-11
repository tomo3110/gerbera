package live

import (
	"strings"
	"sync"

	"github.com/tomo3110/gerbera"
	"github.com/tomo3110/gerbera/diff"
	"github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/property"
	"github.com/tomo3110/gerbera/styles"
)

const debugPanelCSS = `.gd-panel { font-family:monospace; font-size:12px; color:#e0e0e0;
  background:#1e1e2e; border:1px solid #444; border-radius:8px 0 0 0;
  width:420px; max-height:70vh; display:flex; flex-direction:column;
  box-shadow:0 0 20px rgba(0,0,0,0.5); }
.gd-panel.gd-collapsed { display:none; }
.gd-header { display:flex; justify-content:space-between; align-items:center;
  padding:8px 12px; background:#2d2d44; border-radius:8px 0 0 0; cursor:pointer; }
.gd-header-title { font-weight:bold; color:#cba6f7; }
.gd-tabs { display:flex; border-bottom:1px solid #444; }
.gd-tab { padding:6px 12px; cursor:pointer; border:none; background:none;
  color:#888; font-family:monospace; font-size:12px; }
.gd-tab:hover { color:#cdd6f4; }
.gd-tab.active { color:#cba6f7; border-bottom:2px solid #cba6f7; }
.gd-content { overflow-y:auto; padding:8px 12px; flex:1; max-height:50vh; }
.gd-entry { padding:4px 0; border-bottom:1px solid #333; word-break:break-all; }
.gd-time { color:#888; }
.gd-event-name { color:#89b4fa; font-weight:bold; }
.gd-payload { color:#a6adc8; }
.gd-badge { display:inline-block; padding:1px 6px; border-radius:4px;
  font-size:10px; margin-left:4px; }
.gd-badge-ok { background:#2d4a2d; color:#a6e3a1; }
.gd-badge-warn { background:#4a3d2d; color:#fab387; }
.gd-json { white-space:pre-wrap; word-break:break-all; color:#cdd6f4;
  background:#181825; padding:8px; border-radius:4px; margin:4px 0;
  max-height:40vh; overflow-y:auto; }
.gd-toggle { width:36px; height:36px; border-radius:50%; border:none;
  background:#cba6f7; color:#1e1e2e; font-size:16px; font-weight:bold;
  cursor:pointer; box-shadow:0 2px 8px rgba(0,0,0,0.3);
  display:flex; align-items:center; justify-content:center; margin:8px;
  padding:0; line-height:0; }
.gd-toggle:hover { background:#b48ef0; }
.gd-toggle svg { display:block; }
.gd-status { display:inline-block; width:8px; height:8px; border-radius:50%;
  margin-right:6px; }
.gd-status-connected { background:#a6e3a1; }
.gd-status-disconnected { background:#f38ba8; }
.gd-session-item { padding:6px 0; border-bottom:1px solid #333; }
.gd-session-label { color:#888; display:inline-block; width:100px; }
.gd-session-value { color:#cdd6f4; }
.gd-empty { color:#666; font-style:italic; padding:16px 0; text-align:center; }`

const bugIconSVG = `<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M8 2l1.88 1.88"/><path d="M14.12 3.88L16 2"/><path d="M9 7.13v-1a3.003 3.003 0 116 0v1"/><path d="M12 20c-3.3 0-6-2.7-6-6v-3a4 4 0 014-4h4a4 4 0 014 4v3c0 3.3-2.7 6-6 6"/><path d="M12 20v-9"/><path d="M6.53 9C4.6 8.8 3 7.1 3 5"/><path d="M6 13H2"/><path d="M3 21c0-2.1 1.7-3.9 3.8-4"/><path d="M20.97 5c0 2.1-1.6 3.8-3.5 4"/><path d="M22 13h-4"/><path d="M17.2 17c2.1.1 3.8 1.9 3.8 4"/></svg>`

// debugPanelComponents returns the ComponentFunc list that builds the debug panel UI structure.
func debugPanelComponents() gerbera.Components {
	tabs := []struct {
		label  string
		active bool
		empty  string
	}{
		{"Events", true, "No data yet"},
		{"Patches", false, "No data yet"},
		{"State", false, "Waiting for first event..."},
		{"Session", false, "Connecting..."},
	}

	tabButtons := make(gerbera.Components, len(tabs))
	for i, t := range tabs {
		classes := []string{"gd-tab"}
		if t.active {
			classes = append(classes, "active")
		}
		tabButtons[i] = dom.Button(
			property.Class(classes...),
			gerbera.Literal(t.label),
		)
	}

	contentDivs := make(gerbera.Components, len(tabs))
	for i, t := range tabs {
		display := "none"
		if i == 0 {
			display = "block"
		}
		contentDivs[i] = dom.Div(
			property.Class("gd-content"),
			property.Attr("style", "display:"+display),
			dom.Div(
				property.Class("gd-empty"),
				gerbera.Literal(t.empty),
			),
		)
	}

	panelChildren := gerbera.Components{
		// Header
		dom.Div(
			property.Class("gd-header"),
			dom.Span(
				property.Class("gd-header-title"),
				gerbera.Literal("Gerbera Debug"),
			),
			dom.Span(
				property.Class("gd-status", "gd-status-connected"),
			),
		),
		// Tab bar
		dom.Div(
			append(gerbera.Components{property.Class("gd-tabs")}, tabButtons...)...,
		),
	}
	panelChildren = append(panelChildren, contentDivs...)

	return gerbera.Components{
		// <style> with CSS
		styles.CSS(debugPanelCSS),
		// Toggle button
		dom.Button(
			property.Class("gd-toggle"),
			property.Attr("title", "Toggle Gerbera Debug Panel (Ctrl+Shift+D)"),
			gerbera.Literal(bugIconSVG),
		),
		// Panel
		dom.Div(
			append(gerbera.Components{property.Class("gd-panel", "gd-collapsed")}, panelChildren...)...,
		),
	}
}

var (
	debugPanelHTMLOnce sync.Once
	debugPanelHTMLStr  string
)

// renderDebugPanelHTML renders the debug panel UI structure to an HTML string.
// The result is cached via sync.Once.
func renderDebugPanelHTML() string {
	debugPanelHTMLOnce.Do(func() {
		wrapper := &gerbera.Element{
			TagName:    "div",
			ClassNames: make(gerbera.ClassMap),
			Attr:       make(gerbera.AttrMap),
			ChildElems:   make([]*gerbera.Element, 0),
		}

		wrapper = gerbera.Parse(wrapper, debugPanelComponents()...)

		var sb strings.Builder
		for _, child := range wrapper.ChildElems {
			html, err := diff.RenderFragment(child)
			if err != nil {
				continue
			}
			sb.WriteString(html)
		}
		debugPanelHTMLStr = sb.String()
	})
	return debugPanelHTMLStr
}
