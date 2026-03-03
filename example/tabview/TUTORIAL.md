# Tab Component Tutorial

## Overview

Build an interactive tab UI that switches panels in real time without page reloads. This tutorial covers the following features:

- `gc.Tab` struct — Defining a tab with label, content, and event bindings
- `gc.Tabs()` — Rendering an accessible tab UI with ARIA attributes
- `gc.TabsDefaultCSS()` — Minimal default styling for the tab component
- `gl.Click` / `gl.ClickValue` — Binding events via `ButtonAttrs` injection

The `components.Tabs` function lives in the `components` package, which has **no dependency on `live/`**. LiveView event bindings are injected by the caller through the `ButtonAttrs` field, making the component usable in both static and LiveView contexts.

## Prerequisites

- Go 1.22 or later installed
- This repository cloned locally

## Step 1: Importing Packages

```go
import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	g  "github.com/tomo3110/gerbera"
	gc "github.com/tomo3110/gerbera/components"
	gd "github.com/tomo3110/gerbera/dom"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
)
```

- `gc` (components) — provides `Tab`, `Tabs()`, and `TabsDefaultCSS()`
- `gl` (live) — provides `Click()`, `ClickValue()`, and the LiveView runtime
- `gd` (dom) — HTML element helpers
- `gp` (property) — attribute and value setters

## Step 2: Defining the View Struct

```go
type TabDemoView struct {
	ActiveTab int
}
```

The `ActiveTab` field tracks which tab is currently selected (0-based index). Each browser session gets its own instance.

## Step 3: Implementing Mount

```go
func (v *TabDemoView) Mount(params gl.Params) error {
	v.ActiveTab = 0
	return nil
}
```

`Mount` initializes the active tab to the first one.

## Step 4: Building Tab Definitions

```go
tabs := []gc.Tab{
	{
		Label:       "Home",
		Content:     homeContent(),
		ButtonAttrs: tabClickAttrs("0"),
	},
	{
		Label:       "Profile",
		Content:     profileContent(),
		ButtonAttrs: tabClickAttrs("1"),
	},
	{
		Label:       "Settings",
		Content:     settingsContent(),
		ButtonAttrs: tabClickAttrs("2"),
	},
}
```

Each `gc.Tab` has:
- **Label** — The text displayed on the tab button
- **Content** — A `ComponentFunc` that renders the panel body
- **ButtonAttrs** — Slice of `ComponentFunc` injected into the `<button>` element

The `tabClickAttrs` helper binds a click event:

```go
func tabClickAttrs(index string) []g.ComponentFunc {
	return []g.ComponentFunc{
		gl.Click("switch-tab"),
		gl.ClickValue(index),
	}
}
```

This sets `gerbera-click="switch-tab"` and `gerbera-value="0"` (or `"1"`, `"2"`) on the button, so the client sends `{e: "switch-tab", p: {value: "0"}}` on click.

## Step 5: Rendering with Tabs()

```go
func (v *TabDemoView) Render() []g.ComponentFunc {
	return []g.ComponentFunc{
		gd.Head(
			gd.Title("Tab Demo"),
			gc.TabsDefaultCSS(),
		),
		gd.Body(
			gc.Tabs("demo", v.ActiveTab, tabs),
		),
	}
}
```

- `gc.TabsDefaultCSS()` emits a `<style>` element with minimal tab styling
- `gc.Tabs("demo", v.ActiveTab, tabs)` renders the full tab structure with ARIA attributes
- The first argument `"demo"` is the ID prefix used for `aria-controls` / `aria-labelledby` linking

## Step 6: Handling Events

```go
func (v *TabDemoView) HandleEvent(event string, payload gl.Payload) error {
	switch event {
	case "switch-tab":
		if idx, err := strconv.Atoi(payload["value"]); err == nil {
			v.ActiveTab = idx
		}
	}
	return nil
}
```

When the user clicks a tab button, the server receives `"switch-tab"` with `payload["value"]` set to the tab index. The view updates `ActiveTab`, and the framework re-renders and patches the DOM.

## Step 7: Starting the Server

```go
func main() {
	addr := flag.String("addr", ":8890", "listen address")
	debug := flag.Bool("debug", false, "enable debug panel")
	flag.Parse()

	var opts []gl.Option
	if *debug {
		opts = append(opts, gl.WithDebug())
	}

	http.Handle("/", gl.Handler(func(_ context.Context) gl.View { return &TabDemoView{} }, opts...))
	log.Printf("tabview running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
```

## How It Works

```
Browser                                     Go Server
  |                                            |
  |  1. GET /                                  |
  |  <-- HTML with Tabs (active=0) + JS        |
  |                                            |
  |  2. WebSocket connect                      |
  |                                            |
  |  3. Click "Profile" tab                    |
  |  --> {"e":"switch-tab","p":{"value":"1"}}  |
  |                                            |
  |      HandleEvent -> ActiveTab=1 -> Render  |
  |      -> Diff -> Patches                    |
  |                                            |
  |  4. <-- [aria-selected, hidden, class      |
  |          patches for tabs + panels]         |
  |                                            |
  |  5. JS applies patches to DOM              |
```

The `hidden` attribute on inactive panels and `aria-selected` on buttons are toggled server-side. The diff engine sends only the changed attributes as patches.

## Running

```bash
go run example/tabview/tabview.go
```

Open http://localhost:8890 in your browser. Click the tab buttons to switch panels.

### Debug Mode

```bash
go run example/tabview/tabview.go -debug
```

## Static Usage (without LiveView)

The `Tabs` component works without LiveView for server-rendered pages:

```go
func renderStaticTabs(w io.Writer) error {
	tabs := []gc.Tab{
		{Label: "Tab 1", Content: gp.Value("Content 1")},
		{Label: "Tab 2", Content: gp.Value("Content 2")},
	}
	return g.ExecuteTemplate(w, "en",
		gd.Head(gc.TabsDefaultCSS()),
		gd.Body(gc.Tabs("static", 0, tabs)),
	)
}
```

In static mode, all panels are rendered in the HTML with `hidden` attributes. You can add client-side JavaScript to toggle visibility if needed.

## Exercises

1. Add a 4th tab with dynamic content (e.g., current timestamp)
2. Persist the active tab in the URL using `params` in `Mount`
3. Add custom CSS by passing `wrapperAttrs` to `Tabs()`
4. Create a nested tab component (tabs inside a tab panel)

## API Reference

| Function / Type | Description |
|---|---|
| `gc.Tab` | Struct: `Label`, `Content`, `ButtonAttrs` |
| `gc.Tabs(id, activeIndex, tabs, wrapperAttrs...)` | Renders an accessible tab UI |
| `gc.TabsDefaultCSS()` | Returns a `<style>` element with default CSS |
| `gl.Click(event)` | Binds a click event |
| `gl.ClickValue(value)` | Sets the value sent with click events |
