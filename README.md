# Gerbera

A Go HTML template engine that uses functional composition instead of traditional template files.

HTML is built programmatically by composing `ComponentFunc` functions (`func(*Element) error`), providing type safety and the full power of Go at template construction time.

[日本語版 README はこちら](README.ja.md)

## Features

- **Functional composition** — Build HTML by composing small, reusable functions instead of writing template strings
- **Type-safe** — Compile-time guarantees with no template parsing errors at runtime
- **LiveView** — Phoenix LiveView-style real-time DOM updates via WebSocket (`live/` package)
- **DOM diffing** — Automatic minimal DOM patching through element tree diffing (`diff/` package)
- **XSS prevention** — Attribute values are automatically escaped via `html.EscapeString`
- **Zero dependencies for core** — Only `github.com/gorilla/websocket` is required, and only by the `live/` package
- **Optimized rendering** — Zero-allocation render layer with `sync.Pool` reuse, pre-computed indentation, and direct `bufio.Writer` writes
- **Accessibility** — Comprehensive ARIA attribute helpers in `property/` package
- **Testing utilities** — `live.TestView` for unit-testing LiveView implementations without a browser

## Quick Start

### Installation

```bash
go get github.com/tomo3110/gerbera
```

### Hello World

```go
package main

import (
	"log"
	"net/http"

	g  "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	gp "github.com/tomo3110/gerbera/property"
)

func main() {
	mux := g.NewServeMux(
		gd.Head(gd.Title("Hello Gerbera")),
		gd.Body(
			gd.H1(gp.Value("Hello, World!")),
			gd.P(gp.Value("Built with Gerbera")),
		),
	)
	log.Fatal(http.ListenAndServe(":8800", mux))
}
```

### LiveView Counter

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	g  "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	gp "github.com/tomo3110/gerbera/property"
	gl "github.com/tomo3110/gerbera/live"
)

type CounterView struct{ Count int }

func (v *CounterView) Mount(_ gl.Params) error { v.Count = 0; return nil }

func (v *CounterView) Render() []g.ComponentFunc {
	return []g.ComponentFunc{
		gd.Head(gd.Title("Counter")),
		gd.Body(
			gd.H1(gp.Value(fmt.Sprintf("Count: %d", v.Count))),
			gd.Button(gl.Click("dec"), gp.Value("-")),
			gd.Button(gl.Click("inc"), gp.Value("+")),
		),
	}
}

func (v *CounterView) HandleEvent(event string, _ gl.Payload) error {
	switch event {
	case "inc":
		v.Count++
	case "dec":
		v.Count--
	}
	return nil
}

func main() {
	http.Handle("/", gl.Handler(func() gl.View { return &CounterView{} }))
	log.Fatal(http.ListenAndServe(":8840", nil))
}
```

Open http://localhost:8840 — the counter updates in real time without page reloads.

## Architecture

### Rendering Pipeline

```
ComponentFunc composition → Template.Mount() → Parse() → Render() → HTML output
```

### LiveView Pipeline

```
Browser click → WebSocket → HandleEvent() → Render() → Diff() → DOM patches → Browser
```

### Package Structure

```
gerbera/
├── common.go          # ComponentFunc type, Tag(), Literal(), ExecuteTemplate()
├── element.go         # Element struct (TagName, ClassNames, Attr, Children, Value)
├── template.go        # Template: mount/render orchestration, io.Reader
├── render.go          # Recursive HTML rendering
├── server.go          # NewServeMux() HTTP helper
├── map.go             # ConvertToMap interface, Map type
│
├── dom/               # HTML element wrappers (H1, Div, Body, Span, Details, etc.)
├── property/          # Attribute setters (Class, Attr, ID, ARIA helpers, etc.)
├── expr/              # Control flow: If(), Unless(), Each()
├── styles/            # Style(), CSS(), ScopedCSS() for CSS handling
├── components/        # Pre-built components (Bootstrap CDN, Tabs, etc.)
├── diff/              # Element tree diffing and fragment rendering
├── live/              # LiveView: View interface, WebSocket handler, client JS
│
└── example/           # Example applications
    ├── hello/         # Minimal hello world
    ├── portfolio/     # Portfolio page
    ├── dashboard/     # Dashboard with tables
    ├── survey/        # Survey form
    ├── report/        # Report output (stdout)
    ├── wiki/          # Multi-route Wiki (server-side rendering)
    ├── counter/       # LiveView counter
    ├── clock/         # LiveView clock (TickerView)
    ├── jscommand/     # LiveView JS command examples
    ├── upload/        # LiveView file upload
    ├── tabview/       # LiveView tab component demo
    ├── wiki_live/     # LiveView Wiki (SPA-style)
    └── mdviewer/      # LiveView Markdown viewer/editor
```

### Core Concepts

**ComponentFunc** — The fundamental building block. Every DOM helper, attribute setter, and control flow function returns `ComponentFunc`:

```go
type ComponentFunc func(*Element) error
```

**Tag()** — Universal element factory. All `dom/` functions are thin wrappers around `Tag()`:

```go
gd.Div(children...)  // equivalent to g.Tag("div", children...)
```

**Functional composition** — Components compose via variadic `children ...ComponentFunc` parameters:

```go
gd.Body(
    gp.Class("container"),           // sets class attribute
    gd.H1(gp.Value("Title")),       // child element with text
    ge.If(showDetails,               // conditional rendering
        gd.P(gp.Value("Details")),
    ),
)
```

### LiveView (`live/` package)

Gerbera Live provides Phoenix LiveView-style real-time updates. Implement the `gl.View` interface:

| Method | Description |
|--------|-------------|
| `Mount(params)` | Called once per session to initialize state |
| `Render()` | Returns component tree for current state |
| `HandleEvent(event, payload)` | Processes browser events via WebSocket |

Event bindings:

| Function | Description |
|----------|-------------|
| `gl.Click(event)` | Binds a click event |
| `gl.ClickValue(value)` | Sets a value to send with click events |
| `gl.Input(event)` | Binds an input event (sends `value` in payload) |
| `gl.Change(event)` | Binds a change event |
| `gl.Submit(event)` | Binds a form submit event |
| `gl.Focus(event)` / `gl.Blur(event)` | Focus/blur events |
| `gl.Keydown(event)` | Binds a keydown event |
| `gl.Key(key)` | Filters keydown by key name |
| `gl.Scroll(event)` | Binds a scroll event (sends scroll position in payload) |
| `gl.Throttle(ms)` | Sets throttle interval in ms for scroll events (default 100ms) |
| `gl.Debounce(ms)` | Debounces any event — waits for idle before sending |
| `gl.DblClick(event)` | Binds a double-click event |
| `gl.MouseEnter(event)` / `gl.MouseLeave(event)` | Mouse enter/leave events |
| `gl.TouchStart(event)` / `gl.TouchEnd(event)` / `gl.TouchMove(event)` | Touch events |
| `gl.Hook(name)` | Attaches a client-side JS hook |
| `gl.LiveLink(href)` | Client-side navigation without full page reload |
| `gl.Loading(cssClass)` | Sets a CSS class during event processing |

### TickerView

Implement the `TickerView` interface for periodic server-side updates (e.g., polling a file):

```go
type MyView struct { /* ... */ }

func (v *MyView) TickInterval() time.Duration { return 2 * time.Second }
func (v *MyView) HandleTick() error {
    // Check external state, update fields — view is re-rendered automatically
    return nil
}
```

### CommandQueue (Server-to-Client Commands)

Embed `gl.CommandQueue` in your View to issue JS commands without writing JavaScript:

```go
type MyView struct {
    gl.CommandQueue
    // ...
}

func (v *MyView) HandleEvent(event string, payload gl.Payload) error {
    v.ScrollTo(".target", "100", "0")   // Scroll to position
    v.Focus("#input")                    // Focus an element
    v.AddClass(".item", "active")        // Add CSS class
    v.Hide(".modal")                     // Hide an element
    return nil
}
```

Available commands: `ScrollTo`, `ScrollIntoPct`, `Focus`, `Blur`, `SetAttribute`, `RemoveAttribute`, `AddClass`, `RemoveClass`, `ToggleClass`, `SetProperty`, `Dispatch`, `Show`, `Hide`, `Toggle`.

### Upload Support

Implement the `UploadHandler` interface for file upload handling:

```go
type MyView struct {
    gl.CommandQueue
    // ...
}

func (v *MyView) HandleUpload(event string, files []gl.UploadedFile) error {
    for _, f := range files {
        // f.Name, f.Size, f.ContentType, f.Data
    }
    return nil
}
```

Event bindings: `gl.Upload(event)`, `gl.UploadAccept(accept)`, `gl.UploadMultiple()`, `gl.UploadMaxSize(bytes)`.

### Form Utilities

The `live/` package includes form helpers for building accessible forms:

```go
gl.Field("email", gl.FieldOpts{
    Label: "Email", Type: "email", Placeholder: "you@example.com",
})
gl.TextareaField("bio", gl.FieldOpts{Label: "Bio"})
gl.SelectField("role", []gl.SelectOption{
    {Value: "admin", Label: "Admin"},
    {Value: "user", Label: "User"},
}, gl.FieldOpts{Label: "Role"})
```

### Testing LiveViews

Use `live.TestView` to unit-test LiveView implementations without a real WebSocket connection:

```go
tv := gl.NewTestView(&MyView{})
html, _ := tv.RenderHTML()
patches, _ := tv.SimulateEvent("inc", gl.Payload{})
fmt.Println(tv.PatchCount(), tv.HasPatch(diff.OpSetText))
```

### Property Helpers

Common HTML attributes have dedicated helpers in the `property/` package:

| Function | Equivalent |
|----------|------------|
| `gp.Href(url)` | `gp.Attr("href", url)` |
| `gp.For(id)` | `gp.Attr("for", id)` |
| `gp.Type(t)` | `gp.Attr("type", t)` |
| `gp.Placeholder(text)` | `gp.Attr("placeholder", text)` |
| `gp.Required(true)` | `gp.Attr("required", "required")` |
| `gp.Disabled(true)` | `gp.Attr("disabled", "disabled")` |
| `gp.Src(url)` | `gp.Attr("src", url)` |
| `gp.ClassIf(cond, name)` | Conditional CSS class |
| `gp.ClassMap(map)` | Multiple conditional CSS classes |
| `gp.DataAttr(key, val)` | `gp.Attr("data-"+key, val)` |
| `gp.Tabindex(n)` | `gp.Attr("tabindex", n)` |
| `gp.Readonly(true)` | `gp.Attr("readonly", "readonly")` |
| `gp.Title(text)` | `gp.Attr("title", text)` |

ARIA accessibility helpers:

| Function | Equivalent |
|----------|------------|
| `gp.Role(role)` | `gp.Attr("role", role)` |
| `gp.AriaLabel(label)` | `gp.Attr("aria-label", label)` |
| `gp.AriaHidden(bool)` | `gp.Attr("aria-hidden", "true"/"false")` |
| `gp.AriaExpanded(bool)` | `gp.Attr("aria-expanded", "true"/"false")` |
| `gp.AriaSelected(bool)` | `gp.Attr("aria-selected", "true"/"false")` |
| `gp.AriaControls(id)` | `gp.Attr("aria-controls", id)` |
| `gp.AriaLive(mode)` | `gp.Attr("aria-live", mode)` |

And more: `AriaDescribedBy`, `AriaLabelledBy`, `AriaDisabled`, `AriaAtomic`, `AriaCurrent`, `AriaHasPopup`, `AriaInvalid`, `AriaRequired`, `AriaValueNow`, `AriaValueMin`, `AriaValueMax`.

### Styles Helpers

The `styles/` package provides additional CSS utilities:

| Function | Description |
|----------|-------------|
| `gs.Style(styleMap)` | Inline CSS from a map |
| `gs.CSS(cssText)` | `<style>` element with raw CSS |
| `gs.ScopedCSS(scope, rules)` | CSS rules scoped to a selector |
| `gs.StyleIf(cond, styleMap)` | Conditional inline styles |

### Additional DOM Elements

The `dom/` package includes helpers for many HTML elements beyond the basics:

`Span`, `Details`, `Summary`, `Dialog`, `Progress`, `Meter`, `Time`, `Mark`, `Figure`, `Figcaption`, `Video`, `Audio`, `Source`, `Canvas`, `Pre`, `Blockquote`, `Em`, `Strong`, `Small`, `Dl`, `Dt`, `Dd`, `Fieldset`, `Legend`, `Iframe`, and more.

Handler options:

| Function | Description |
|----------|-------------|
| `gl.WithLang(lang)` | Sets HTML `lang` attribute (default `"ja"`) |
| `gl.WithSessionTTL(d)` | Sets session timeout (default 5 minutes) |
| `gl.WithDebug()` | Enables browser DevPanel and server-side structured logging |
| `gl.WithMiddleware(mw)` | Adds HTTP middleware to the handler |

### Debug Mode

Enable debug mode by passing `gl.WithDebug()` to the handler:

```go
http.Handle("/", gl.Handler(factory, gl.WithDebug()))
```

When enabled:
- **Browser DevPanel** — A floating overlay (toggle with `Ctrl+Shift+D` or the "G" button) with 4 tabs:
  - **Events** — Real-time log of sent events with name, payload, and timestamp
  - **Patches** — Received DOM patches with count and server processing time
  - **State** — Current View struct state as JSON (updated after each event)
  - **Session** — Session ID, TTL, and connection status
- **Server-side logging** — Structured logs via `log/slog` for event processing, patch generation, and session lifecycle

When `WithDebug()` is not set, there is zero overhead — no debug JS is injected and no timing is measured.

## Examples

Each example includes bilingual tutorials (`TUTORIAL.md` / `TUTORIAL.ja.md`).

| Example | Description | Port | Command |
|---------|-------------|------|---------|
| [hello](example/hello/) | Minimal hello world | :8800 | `go run example/hello/hello.go` |
| [portfolio](example/portfolio/) | Portfolio page | :8810 | `go run example/portfolio/portfolio.go` |
| [dashboard](example/dashboard/) | Dashboard with tables | :8820 | `go run example/dashboard/dashboard.go` |
| [survey](example/survey/) | Survey form | :8830 | `go run example/survey/survey.go` |
| [report](example/report/) | Report output | stdout | `go run example/report/report.go` |
| [wiki](example/wiki/) | Multi-route Wiki (SSR) | :8880 | `go run example/wiki/wiki.go` |
| [counter](example/counter/) | LiveView counter | :8840 | `go run example/counter/counter.go` |
| [clock](example/clock/) | LiveView clock (TickerView) | :8870 | `go run example/clock/clock.go` |
| [jscommand](example/jscommand/) | LiveView JS commands | :8875 | `go run example/jscommand/jscommand.go` |
| [upload](example/upload/) | LiveView file upload | :8885 | `go run example/upload/upload.go` |
| [tabview](example/tabview/) | LiveView tab component | :8890 | `go run example/tabview/tabview.go` |
| [wiki_live](example/wiki_live/) | LiveView Wiki (SPA) | :8850 | `go run example/wiki_live/wiki_live.go` |
| [mdviewer](example/mdviewer/) | LiveView Markdown Viewer | :8860 | `cd example/mdviewer && go run .` |

## Development

```bash
go build ./...    # Build all packages
go test ./...     # Run all tests
go test -v ./...  # Verbose test output
```

Requires Go 1.22 or later.

### Benchmarks

```bash
go test -bench=. -benchmem ./...   # Run all benchmarks with memory stats
```

The rendering pipeline is optimized with `sync.Pool` buffer reuse, zero-allocation core render path, and pre-computed indentation strings.

## License

[MIT](LICENSE)
