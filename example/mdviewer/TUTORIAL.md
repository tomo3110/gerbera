# Markdown Viewer Tutorial

## Overview

Build a real-time Markdown viewer/editor that demonstrates practical Gerbera Live features:

- `gl.View` interface — Stateful LiveView with file I/O
- `gl.TickerView` interface — Server-side periodic file polling (no client JS needed)
- `gl.CommandQueue` — Server-to-client scroll commands via `ScrollIntoPct`
- `gl.Input` — Real-time text input binding
- `g.Literal()` — Injecting raw HTML (rendered Markdown)
- `OpSetHTML` — WebSocket diff patches that use `innerHTML` for HTML content
- `gu.Theme()` — UI component library theme system with CSS custom properties
- `gu.Button()` / `gu.Badge()` / `gu.Row()` — Pre-built UI components replacing inline styles
- `gs.CSS()` — Embedding component-scoped CSS via `<style>` elements
- Independent module — Using `go.mod` with `replace` directive and external dependencies
- `gl.Scroll` / `gl.Throttle` — Server-routed scroll sync via `CommandQueue`

## Prerequisites

- Go 1.22 or later installed
- This repository cloned locally

## Step 1: Project Structure

This example lives in its own Go module with an external dependency (goldmark):

```
example/mdviewer/
├── go.mod          # Independent module with replace directive
├── main.go         # CLI flags and server startup
├── markdown.go     # goldmark wrapper
└── viewer.go       # MarkdownView (gl.View + gl.TickerView implementation)
```

The `go.mod` uses a `replace` directive to reference the local gerbera:

```
module github.com/tomo3110/gerbera/example/mdviewer

require (
    github.com/tomo3110/gerbera v0.0.0
    github.com/yuin/goldmark v1.7.8
)

replace github.com/tomo3110/gerbera => ../..
```

## Step 2: Markdown Rendering

```go
var md = goldmark.New(
    goldmark.WithExtensions(extension.GFM),
    goldmark.WithRendererOptions(html.WithUnsafe()),
)

func renderMarkdown(source string) string {
    var buf bytes.Buffer
    if err := md.Convert([]byte(source), &buf); err != nil {
        return "<p>Markdown render error</p>"
    }
    return buf.String()
}
```

The goldmark parser is configured with GFM (GitHub Flavored Markdown) extensions for tables, strikethrough, and task lists. `html.WithUnsafe()` allows raw HTML passthrough in Markdown source.

## Step 3: View Struct

```go
type MarkdownView struct {
    gl.CommandQueue              // Embed for server-to-client JS commands
    Source     string            // Markdown source text
    HTML       string            // Rendered HTML (cached)
    FilePath   string            // .md file path (empty = no file)
    Preview    bool              // Preview-only mode
    ModTime    time.Time         // File's last modification time
    FileName   string            // Display filename
    SaveStatus string            // "", "saved", "error"
    ScrollPct  float64           // Scroll position (0.0–1.0)
}
```

Key design decisions:
- **`gl.CommandQueue` embed** — Implements the `JSCommander` interface, allowing the view to issue scroll commands to the client without custom JavaScript
- The `HTML` field caches the rendered Markdown to avoid re-rendering on every `Render()` call. It's only updated when `Source` changes.

## Step 4: Mount

```go
func (v *MarkdownView) Mount(_ gl.Params) error {
    if v.FilePath != "" {
        data, err := os.ReadFile(v.FilePath)
        // ... load file, set ModTime and FileName
    }
    if v.Source == "" && !v.Preview {
        v.Source = "# Hello Markdown\n\nStart typing here..."
    }
    v.HTML = renderMarkdown(v.Source)
    return nil
}
```

Unlike the counter example where `Mount` just initializes simple values, here it performs file I/O. Note that `FilePath` and `Preview` are set by the factory function in `main.go` before `Mount` is called.

## Step 5: Event Handling

```go
func (v *MarkdownView) HandleEvent(event string, payload gl.Payload) error {
    switch event {
    case "edit":
        v.Source = payload["value"]
        v.HTML = renderMarkdown(v.Source)
    case "save":
        os.WriteFile(v.FilePath, []byte(v.Source), 0644)
        v.SaveStatus = "saved"
    case "editor-scroll":
        // Compute scroll percentage and send scroll command
    }
    return nil
}
```

Two main events:
- **edit**: Triggered by `gl.Input("edit")` on the textarea. Updates source and re-renders HTML.
- **save**: Writes current source to file. Only available when `FilePath` is set.

Note that `check-file` event handling has been replaced by the `TickerView` interface (see Step 7).

## Step 6: Raw HTML with `g.Literal()` and `OpSetHTML`

```go
gd.Div(
    gp.ID("md-preview"),
    gp.Class("md-preview"),
    g.Literal(v.HTML),  // Raw HTML from goldmark
)
```

`g.Literal()` injects raw HTML into the element's `Value` field. On the initial HTTP response, this renders correctly. For WebSocket updates, the `diff` package detects that the value contains `<` and uses `OpSetHTML` instead of `OpSetText`, which causes the client JS to use `innerHTML` instead of `textContent`.

This is the key difference from simple text updates: without `OpSetHTML`, Markdown rendering would break on WebSocket updates because HTML tags would be escaped.

## Step 7: File Watching via `TickerView`

Instead of using client-side JavaScript polling, the `TickerView` interface provides server-side periodic callbacks:

```go
// TickInterval returns the interval for file change polling.
func (v *MarkdownView) TickInterval() time.Duration {
    if v.FilePath == "" {
        return 0  // disable ticking when no file
    }
    return 2 * time.Second
}

// HandleTick checks for external file changes on each tick.
func (v *MarkdownView) HandleTick() error {
    if v.FilePath == "" {
        return nil
    }
    info, err := os.Stat(v.FilePath)
    if err != nil {
        return nil
    }
    if info.ModTime().After(v.ModTime) {
        data, _ := os.ReadFile(v.FilePath)
        v.Source = string(data)
        v.HTML = renderMarkdown(v.Source)
        v.ModTime = info.ModTime()
    }
    return nil
}
```

Key advantages over the previous client-side polling approach:
- **No JavaScript needed** — No hidden buttons, no `setInterval`, no client-side hacks
- **Zero-overhead when idle** — If `HandleTick` doesn't modify state, no patches are sent
- **Disable dynamically** — Return `0` from `TickInterval()` to disable ticking (e.g., when no file is loaded)

## Step 8: Scroll Sync via `CommandQueue`

Scroll sync uses the `CommandQueue` to send `ScrollIntoPct` commands from the server to the client, eliminating the need for `MutationObserver` and `data-scroll-pct` attribute hacks.

### Binding the scroll event

```go
gd.Textarea(
    gp.ID("md-editor"),
    gp.Class("md-textarea"),
    gl.Input("edit"),
    gl.Scroll("editor-scroll"),  // send scroll events to server
    gl.Throttle(50),             // limit to one event per 50ms
    gp.Value(v.Source),
),
```

`gl.Scroll("editor-scroll")` adds a `gerbera-scroll` attribute. The client JS (`gerbera.js`) listens for `scroll` events and sends `scrollTop`, `scrollHeight`, and `clientHeight` in the payload. `gl.Throttle(50)` limits the event rate.

### Server-side handler with CommandQueue

```go
case "editor-scroll":
    scrollTop, _ := strconv.ParseFloat(payload["scrollTop"], 64)
    scrollHeight, _ := strconv.ParseFloat(payload["scrollHeight"], 64)
    clientHeight, _ := strconv.ParseFloat(payload["clientHeight"], 64)
    max := scrollHeight - clientHeight
    if max > 0 {
        v.ScrollPct = scrollTop / max
        v.ScrollIntoPct(".md-preview-pane", fmt.Sprintf("%.6f", v.ScrollPct))
    }
```

The handler normalizes the scroll position to a 0.0–1.0 range, then issues a `ScrollIntoPct` command. The `CommandQueue` (embedded in the view) queues this command, and the framework sends it to the client after the next render.

### Data flow

```
Browser                                    Go Server
  |                                           |
  | User scrolls textarea                     |
  | gerbera.js: scroll event (throttled 50ms) |
  | --> {"e":"editor-scroll","p":{            |
  |       "scrollTop":"150",                  |
  |       "scrollHeight":"1200",              |
  |       "clientHeight":"400"}}              |
  |                                           |
  |     HandleEvent: ScrollPct = 150/800      |
  |     CommandQueue: ScrollIntoPct(0.1875)   |
  |     Render + Diff + Drain commands        |
  |                                           |
  | <-- {"patches":[...],                     |
  |      "js_commands":[{"cmd":"scroll_into_pct",
  |        "target":".md-preview-pane",       |
  |        "args":{"pct":"0.187500"}}]}       |
  |                                           |
  | gerbera.js executes command:              |
  | el.scrollTop = 0.1875 * maxScroll         |
```

Key advantage: No `MutationObserver`, no `data-scroll-pct` attribute, no custom JavaScript — the framework handles everything.

## Step 9: Theme and CSS

The viewer uses `gu.Theme()` from the UI component library for base styling (fonts, colors, spacing) and `gs.CSS(mdCSS)` for Markdown-specific styles:

```go
gd.Head(
    gd.Title(title),
    gd.Meta(gp.Attr("charset", "utf-8")),
    gd.Meta(gp.Attr("name", "viewport"), ...),
    gu.Theme(),       // Base theme with CSS custom properties
    gs.CSS(mdCSS),    // Markdown-specific styles only
),
```

`gu.Theme()` injects a `<style>` element that defines CSS custom properties (`--g-bg`, `--g-text`, `--g-accent`, `--g-border`, etc.). The Markdown CSS references these variables instead of hardcoding colors:

```css
.md-textarea {
  background: var(--g-bg-inset);    /* instead of #fafbfc */
  color: var(--g-text);             /* instead of #24292e */
  font-family: var(--g-font-mono);  /* instead of hardcoded font stack */
}
.md-preview a { color: var(--g-accent); }           /* instead of #0366d6 */
.md-preview blockquote { color: var(--g-text-secondary); } /* instead of #6a737d */
```

This means the viewer automatically adapts to any theme. Switching to `gu.ThemeWith(gu.DarkTheme())` gives a dark-mode viewer with no CSS changes.

### Header with UI Components

The header uses `gu.Row()`, `gu.Button()`, and `gu.Badge()` instead of inline styles:

```go
func (v *MarkdownView) renderHeader() g.ComponentFunc {
    return gd.Header(
        gp.Class("g-mdviewer-header"),
        gu.Row(gu.AlignCenter, gu.GapSm,
            gd.Span(gp.Class("g-mdviewer-title"), ...),
            ge.If(v.FilePath != "" && !v.Preview,
                gu.Button("Save", gu.ButtonPrimary, gu.ButtonSmall, gl.Click("save")),
                g.Skip(),
            ),
            ge.If(v.SaveStatus == "saved",
                gu.Badge("Saved", "dark"),
                g.Skip(),
            ),
        ),
    )
}
```

Benefits:
- `gu.Button("Save", gu.ButtonPrimary, gu.ButtonSmall)` replaces 7 lines of inline style
- `gu.Badge("Saved", "dark")` replaces a manually styled `<span>`
- `gu.Row(gu.AlignCenter, gu.GapSm)` replaces inline flex CSS

## Step 10: CLI and Server

```go
func main() {
    addr := flag.String("addr", ":8860", "listen address")
    preview := flag.Bool("preview", false, "preview-only mode")
    debug := flag.Bool("debug", false, "enable debug panel")
    flag.Parse()

    factory := func() gl.View {
        return &MarkdownView{
            FilePath: filePath,
            Preview:  *preview,
        }
    }

    http.Handle("/", gl.Handler(factory, opts...))
    log.Fatal(http.ListenAndServe(*addr, nil))
}
```

The factory captures the file path and preview flag from CLI arguments and passes them to each new session's `MarkdownView`.

## How It Works

```
Browser                                  Go Server
  |                                         |
  | 1. GET /                                |
  | <-- HTML (editor + rendered MD)         |
  |                                         |
  | 2. WebSocket connect                    |
  |                                         |
  | 3. Type in textarea                     |
  | --> {"e":"edit","p":{"value":"# Hi"}}   |
  |                                         |
  |     HandleEvent -> renderMarkdown       |
  |     -> Render -> Diff -> Patches        |
  |                                         |
  | 4. <-- [{"op":"html",                   |
  |          "path":[1,1,1,0],              |
  |          "val":"<h1>Hi</h1>"}]          |
  |                                         |
  | 5. JS: node.innerHTML = val             |
  |                                         |
  | 6. (Every 2s) Server tick               |
  |     HandleTick -> check file ModTime    |
  |     -> re-render if changed -> patches  |
```

## Running

```bash
# Editor mode (no file)
cd example/mdviewer && go run .

# Editor mode with file
cd example/mdviewer && go run . README.md

# Preview-only with file watching
cd example/mdviewer && go run . -preview README.md

# With debug panel
cd example/mdviewer && go run . -debug
```

Open http://localhost:8860 and start typing Markdown to see it rendered in real time.

## Exercises

1. Add a "word count" display that updates as you type
2. Add keyboard shortcut (Ctrl+S) to save using `gl.Keydown` and `gl.Key`
3. Add syntax highlighting for code blocks using a client-side library
4. Implement a "split view" toggle button to switch between editor-only, preview-only, and split modes
5. Switch to dark theme with `gu.ThemeWith(gu.DarkTheme())` and observe how all CSS-variable-based styles adapt automatically
