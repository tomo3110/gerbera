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
├── dom/               # HTML element wrappers (H1, Div, Body, Form, etc.)
├── property/          # Attribute setters (Class, Attr, ID, Name, Value)
├── expr/              # Control flow: If(), Unless(), Each()
├── styles/            # Style() for inline CSS from a map
├── components/        # Pre-built head components (Bootstrap CDN, etc.)
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
    └── wiki_live/     # LiveView Wiki (SPA-style)
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

Handler options:

| Function | Description |
|----------|-------------|
| `gl.WithLang(lang)` | Sets HTML `lang` attribute (default `"ja"`) |
| `gl.WithSessionTTL(d)` | Sets session timeout (default 5 minutes) |

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
| [wiki_live](example/wiki_live/) | LiveView Wiki (SPA) | :8850 | `go run example/wiki_live/wiki_gl.go` |

## Development

```bash
go build ./...    # Build all packages
go test ./...     # Run all tests
go test -v ./...  # Verbose test output
```

Requires Go 1.22 or later.

## License

MIT
