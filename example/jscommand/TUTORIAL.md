# JS Commands Tutorial

## Overview

Build a demo that executes client-side JavaScript commands from the server. This tutorial covers the following features:

- `gl.CommandQueue` — Embedding a command queue in your View to emit JS commands
- `JSCommander` interface — Automatically satisfied by embedding `CommandQueue`
- JS commands — `Focus`, `Blur`, `ScrollTo`, `AddClass`, `RemoveClass`, `ToggleClass`, `Show`, `Hide`, `Toggle`
- `ge.Each` — Iterating over a list to render repeated elements

This example demonstrates **server-to-client command execution**: when the user clicks a button, the server handles the event and pushes JS commands that the client executes alongside DOM patches.

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

	g  "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	ge "github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
	gs "github.com/tomo3110/gerbera/styles"
)
```

## Step 2: Defining a ConvertToMap Type for Each

```go
type logEntry struct {
	Index   int
	Message string
}

func (e logEntry) ToMap() g.Map {
	return g.Map{"index": e.Index, "message": e.Message}
}
```

`ge.Each` requires items that implement `g.ConvertToMap`. The `logEntry` type wraps a log message with an index, and `ToMap()` exposes the fields for use inside the `Each` callback.

## Step 3: Defining the View Struct

```go
type JSCommandView struct {
	gl.CommandQueue
	Log []logEntry
}
```

Key design pattern: **embed `gl.CommandQueue`** directly in the View struct. This gives you access to all JS command methods (`Focus`, `Blur`, `ScrollTo`, etc.) and automatically satisfies the `JSCommander` interface via the `DrainCommands()` method. The framework detects this after each render cycle and sends queued commands to the client.

`Log` keeps a history of executed commands for display.

## Step 4: Implementing HandleEvent

```go
func (v *JSCommandView) HandleEvent(event string, _ gl.Payload) error {
	switch event {
	case "do_focus":
		v.Focus("#target-input")
		v.appendLog("Focus → #target-input")
	case "do_blur":
		v.Blur("#target-input")
		v.appendLog("Blur → #target-input")
	case "scroll_top":
		v.ScrollTo("#scroll-box", "0", "0")
		v.appendLog("ScrollTo top → #scroll-box")
	case "scroll_bottom":
		v.ScrollTo("#scroll-box", "9999", "0")
		v.appendLog("ScrollTo bottom → #scroll-box")
	case "add_class":
		v.AddClass("#style-box", "highlighted")
		v.appendLog("AddClass 'highlighted' → #style-box")
	case "remove_class":
		v.RemoveClass("#style-box", "highlighted")
		v.appendLog("RemoveClass 'highlighted' → #style-box")
	case "toggle_class":
		v.ToggleClass("#style-box", "highlighted")
		v.appendLog("ToggleClass 'highlighted' → #style-box")
	case "do_show":
		v.Show("#toggle-box")
		v.appendLog("Show → #toggle-box")
	case "do_hide":
		v.Hide("#toggle-box")
		v.appendLog("Hide → #toggle-box")
	case "do_toggle":
		v.Toggle("#toggle-box")
		v.appendLog("Toggle → #toggle-box")
	case "clear_log":
		v.Log = nil
	}
	return nil
}
```

Each event handler calls one `CommandQueue` method and logs the action. The commands are:

| Method | Effect |
|--------|--------|
| `Focus(selector)` | Focuses the matching element |
| `Blur(selector)` | Removes focus from the matching element |
| `ScrollTo(selector, top, left)` | Scrolls the element to the given position |
| `AddClass(selector, class)` | Adds a CSS class |
| `RemoveClass(selector, class)` | Removes a CSS class |
| `ToggleClass(selector, class)` | Toggles a CSS class |
| `Show(selector)` | Shows a hidden element |
| `Hide(selector)` | Hides an element |
| `Toggle(selector)` | Toggles element visibility |

All selectors are standard CSS selectors (e.g., `"#my-id"`, `".my-class"`).

## Step 5: Implementing Render with Each

```go
func (v *JSCommandView) Render() []g.ComponentFunc {
	logItems := make([]g.ConvertToMap, len(v.Log))
	for i, e := range v.Log {
		logItems[i] = e
	}

	return []g.ComponentFunc{
		gd.Head(...),
		gd.Body(
			...
			// Command Log
			ge.If(len(v.Log) > 0,
				gd.Div(
					gp.Class("log"),
					ge.Each(logItems, func(item g.ConvertToMap) g.ComponentFunc {
						m := item.ToMap()
						idx := m.Get("index").(int)
						msg := m.Get("message").(string)
						return gd.Div(
							gp.Class("log-entry"),
							gp.Value(fmt.Sprintf("#%d  %s", idx, msg)),
						)
					}),
				),
				gd.P(gp.Value("No commands executed yet.")),
			),
			...
		),
	}
}
```

Key points:
- Convert the `[]logEntry` slice to `[]g.ConvertToMap` before passing to `ge.Each`
- Inside the callback, call `item.ToMap()` and extract typed values with `.Get("key").(type)`
- `ge.If` shows the log when there are entries, or an empty-state message otherwise

## Step 6: UI Sections

The UI is organized into four demo sections, each targeting a specific element:

```go
// Focus / Blur section
gd.Input(gp.Type("text"), gp.ID("target-input"), gp.Placeholder("Click Focus to focus this input"))
gd.Button(gl.Click("do_focus"), gp.Value("Focus"))
gd.Button(gl.Click("do_blur"), gp.Value("Blur"))

// ScrollTo section
gd.Div(gp.ID("scroll-box"), scrollContent())
gd.Button(gl.Click("scroll_top"), gp.Value("Scroll to Top"))
gd.Button(gl.Click("scroll_bottom"), gp.Value("Scroll to Bottom"))

// Class manipulation section
gd.Div(gp.ID("style-box"), gp.Value("Target Box"))
gd.Button(gl.Click("add_class"), gp.Value("AddClass"))
gd.Button(gl.Click("remove_class"), gp.Value("RemoveClass"))
gd.Button(gl.Click("toggle_class"), gp.Value("ToggleClass"))

// Visibility section
gd.Div(gp.ID("toggle-box"), gp.Value("Toggle Me"))
gd.Button(gl.Click("do_show"), gp.Value("Show"))
gd.Button(gl.Click("do_hide"), gp.Value("Hide"))
gd.Button(gl.Click("do_toggle"), gp.Value("Toggle"))
```

Each section has a target element with a unique `id` and buttons that fire corresponding events.

## How It Works

```
Browser                                   Go Server
  |                                          |
  |  1. User clicks "Focus" button           |
  |  --> {"e":"do_focus","p":{}}             |
  |                                          |
  |      HandleEvent:                        |
  |        v.Focus("#target-input")          |
  |        v.appendLog(...)                  |
  |      -> Render -> Diff -> Patches        |
  |      -> DrainCommands -> JS commands     |
  |                                          |
  |  <-- { patches: [...],                   |
  |        commands: [                        |
  |          {"cmd":"focus",                  |
  |           "target":"#target-input"}       |
  |        ]}                                |
  |                                          |
  |  2. JS applies DOM patches (log update)  |
  |  3. JS executes: focus("#target-input")  |
```

The framework sends both DOM patches and JS commands in the same WebSocket message. The client applies patches first, then executes commands.

## Running

```bash
go run example/jscommand/jscommand.go
```

Open http://localhost:8875 in your browser. Click the buttons in each section and observe the effects. The command log at the bottom tracks every command you execute.

### Debug Mode

```bash
go run example/jscommand/jscommand.go -debug
```

## Exercises

1. Add a `SetAttribute` button that changes the placeholder text of the input
2. Add a `Dispatch` button that dispatches a custom DOM event on an element
3. Combine multiple commands in a single event (e.g., Focus + AddClass)
4. Add a `ScrollIntoPct` demo that scrolls to a percentage position

## API Reference

| Function | Description |
|----------|-------------|
| `gl.CommandQueue` | Embed in View struct to enable JS commands |
| `q.Focus(sel)` | Focus an element |
| `q.Blur(sel)` | Remove focus |
| `q.ScrollTo(sel, top, left)` | Scroll to position |
| `q.ScrollIntoPct(sel, pct)` | Scroll to percentage |
| `q.AddClass(sel, cls)` | Add CSS class |
| `q.RemoveClass(sel, cls)` | Remove CSS class |
| `q.ToggleClass(sel, cls)` | Toggle CSS class |
| `q.Show(sel)` | Show element |
| `q.Hide(sel)` | Hide element |
| `q.Toggle(sel)` | Toggle visibility |
| `q.SetAttribute(sel, key, val)` | Set HTML attribute |
| `q.RemoveAttribute(sel, key)` | Remove HTML attribute |
| `q.SetProperty(sel, key, val)` | Set DOM property |
| `q.Dispatch(sel, event)` | Dispatch DOM event |
| `ge.Each(items, callback)` | Iterate and render a list |
