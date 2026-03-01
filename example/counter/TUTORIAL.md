# Counter Tutorial

## Overview

Build an interactive counter that updates in real time without page reloads. This tutorial covers the following features:

- `gl.View` interface — Defining a stateful LiveView component
- `gl.Handler` — Serving a LiveView over HTTP and WebSocket
- `gl.Click` — Binding browser events to server-side handlers
- `diff.Diff` — Automatic DOM patching via tree diffing (used internally)

This example demonstrates the **Gerbera Live** architecture: user events are sent to the Go server via WebSocket, the server updates state and re-renders, then only the changed parts of the DOM are sent back as patches.

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
	gp "github.com/tomo3110/gerbera/property"
	gl "github.com/tomo3110/gerbera/live"
)
```

In addition to the standard gerbera aliases (`g`, `gd`, `gp`), import the `live` package which provides the LiveView runtime:
- `live` — LiveView core: `View` interface, `Handler()`, event bindings

## Step 2: Defining the View Struct

```go
type CounterView struct{ Count int }
```

A LiveView is any struct that implements the `gl.View` interface. The struct fields hold the component state. Each browser session gets its own instance, so state is not shared between users.

## Step 3: Implementing Mount

```go
func (v *CounterView) Mount(params gl.Params) error {
	v.Count = 0
	return nil
}
```

`Mount` is called once when a new session is created (i.e., the first HTTP request). Use it to initialize state. `params` contains URL query parameters as `map[string]string`.

## Step 4: Implementing Render

```go
func (v *CounterView) Render() []g.ComponentFunc {
	return []g.ComponentFunc{
		gd.Head(gd.Title("カウンター")),
		gd.Body(
			gp.Class("container"),
			gd.H1(gp.Value(fmt.Sprintf("カウント: %d", v.Count))),
			gd.Div(
				gd.Button(gl.Click("dec"), gp.Value("-")),
				gd.Button(gl.Click("inc"), gp.Value("+")),
			),
		),
	}
}
```

`Render` returns the component tree for the current state. It is called on every event to produce a new tree, which is then diffed against the previous tree.

Key points:
- `gl.Click("inc")` sets the `gerbera-click="inc"` attribute, telling the client JS to send an `"inc"` event when the button is clicked
- The return value is a slice of `ComponentFunc` that becomes children of the `<html>` root element
- Use standard gerbera DOM helpers (`gd.Body`, `gd.H1`, etc.) exactly as you would for server-rendered pages

## Step 5: Implementing HandleEvent

```go
func (v *CounterView) HandleEvent(event string, payload gl.Payload) error {
	switch event {
	case "inc":
		v.Count++
	case "dec":
		v.Count--
	}
	return nil
}
```

`HandleEvent` receives events sent from the browser. The `event` string matches the value passed to `gl.Click()`. `payload` is a `map[string]string` containing additional data (e.g., input values for `gl.Input` events).

After `HandleEvent` returns, the framework automatically calls `Render()`, diffs the new tree against the old one, and sends only the patches to the browser.

## Step 6: Starting the Server

```go
func main() {
	addr := flag.String("addr", ":8840", "listen address")
	debug := flag.Bool("debug", false, "enable debug panel")
	flag.Parse()

	var opts []gl.Option
	if *debug {
		opts = append(opts, gl.WithDebug())
	}

	http.Handle("/", gl.Handler(func() gl.View { return &CounterView{} }, opts...))
	log.Printf("counter running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
```

- `gl.Handler(factory)` returns an `http.Handler` that handles both initial HTTP rendering and WebSocket connections
- The factory function `func() gl.View` is called once per session to create a fresh `View` instance
- Options: `gl.WithLang("en")` sets the HTML lang attribute, `gl.WithSessionTTL(10*time.Minute)` adjusts the session timeout
- `gl.WithDebug()` enables the browser DevPanel and server-side structured logging (see below)

## How It Works

```
Browser                                   Go Server
  |                                          |
  |  1. GET /                                |
  |  <-- Initial HTML + gerbera.js           |
  |                                          |
  |  2. WebSocket connect                    |
  |  <-> Bidirectional connection            |
  |                                          |
  |  3. Click "+" button                     |
  |  --> {"e":"inc","p":{}}                  |
  |                                          |
  |      HandleEvent -> Count++ -> Render    |
  |      -> Diff old vs new -> Patches       |
  |                                          |
  |  4. <-- [{"op":"text",                   |
  |           "path":[1,0],                  |
  |           "val":"カウント: 1"}]            |
  |                                          |
  |  5. JS applies patch to DOM             |
```

## Running

```bash
go run example/counter/counter.go
```

Open http://localhost:8840 in your browser. Click the **+** and **-** buttons to see the count update in real time without page reloads.

To change the port:

```bash
go run example/counter/counter.go -addr :3000
```

Open the browser DevTools Network tab and filter by "WS" to observe the WebSocket messages being exchanged.

### Debug Mode

Run with the `-debug` flag to enable the built-in debug panel:

```bash
go run example/counter/counter.go -debug
```

When debug mode is enabled:
- A floating **"G" button** appears in the bottom-right corner of the browser
- Click the button or press `Ctrl+Shift+D` to toggle the **Gerbera Debug Panel**
- The panel has 4 tabs:
  - **Events** — Real-time log of events sent to the server (event name, payload, timestamp)
  - **Patches** — DOM patches received from the server (patch count, server processing time in ms)
  - **State** — Current `CounterView` state as JSON (updates after each event)
  - **Session** — Session ID, TTL, and WebSocket connection status
- The server terminal also outputs structured logs via `log/slog`

## Exercises

1. Add a "Reset" button that sets the count back to 0
2. Add an `<input>` field using `gl.Input("set")` to let users type a number directly
3. Use `ge.If` to change the text color when the count is negative
4. Add a second route with a different View (e.g., a to-do list)

## API Reference

| Function | Description |
|----------|-------------|
| `gl.Handler(factory, opts...)` | Returns an `http.Handler` for a LiveView |
| `gl.WithLang(lang)` | Sets the HTML `lang` attribute (default `"ja"`) |
| `gl.WithSessionTTL(d)` | Sets session timeout (default 5 minutes) |
| `gl.WithDebug()` | Enables browser DevPanel and server-side structured logging |
| `gl.Click(event)` | Binds a click event |
| `gl.Input(event)` | Binds an input event |
| `gl.Change(event)` | Binds a change event |
| `gl.Submit(event)` | Binds a form submit event |
| `gl.Focus(event)` | Binds a focus event |
| `gl.Blur(event)` | Binds a blur event |
| `gl.Keydown(event)` | Binds a keydown event |
| `gl.Key(key)` | Filters keydown by key name (e.g., `"Enter"`) |
| `gl.View` | Interface: `Mount`, `Render`, `HandleEvent` |
| `gl.Params` | `map[string]string` — URL parameters |
| `gl.Payload` | `map[string]string` — Event data |
