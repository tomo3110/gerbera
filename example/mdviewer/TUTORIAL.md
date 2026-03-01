# Markdown Viewer Tutorial

## Overview

Build a real-time Markdown viewer/editor that demonstrates practical Gerbera Live features:

- `gl.View` interface — Stateful LiveView with file I/O
- `gl.Input` — Real-time text input binding
- `g.Literal()` — Injecting raw HTML (rendered Markdown)
- `OpSetHTML` — WebSocket diff patches that use `innerHTML` for HTML content
- Independent module — Using `go.mod` with `replace` directive and external dependencies
- Client-side polling — File change detection without server push

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
└── viewer.go       # MarkdownView (gl.View implementation)
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
    Source     string    // Markdown source text
    HTML       string    // Rendered HTML (cached)
    FilePath   string    // .md file path (empty = no file)
    Preview    bool      // Preview-only mode
    ModTime    time.Time // File's last modification time
    FileName   string    // Display filename
    SaveStatus string    // "", "saved", "error"
}
```

The `HTML` field caches the rendered Markdown to avoid re-rendering on every `Render()` call. It's only updated when `Source` changes.

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
    case "check-file":
        // Compare ModTime, reload if changed
    }
    return nil
}
```

Three events:
- **edit**: Triggered by `gl.Input("edit")` on the textarea. Updates source and re-renders HTML.
- **save**: Writes current source to file. Only available when `FilePath` is set.
- **check-file**: Polls the file's modification time. If the file changed externally, reloads it.

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

## Step 7: File Watching via Polling

Since Gerbera Live doesn't have server push, file watching is implemented with client-side polling:

```go
// Hidden button in the DOM
gd.Button(
    gp.ID("file-check-trigger"),
    gl.Click("check-file"),
    gs.Style(g.StyleMap{"display": "none"}),
)
```

```javascript
// Auto-click every 2 seconds
setInterval(function() {
    document.getElementById('file-check-trigger').click();
}, 2000);
```

When no file changes are detected, `HandleEvent` returns without modifying state, producing zero patches, and the handler skips sending a WebSocket message (see `handler.go` L191).

## Step 8: Scroll Sync

```javascript
ta.addEventListener('scroll', function() {
    var pct = ta.scrollTop / (ta.scrollHeight - ta.clientHeight);
    pv.scrollTop = pct * (pv.scrollHeight - pv.clientHeight);
});
```

The editor's scroll position is mapped to a percentage and applied to the preview pane. This is injected as an inline `<script>` via `g.Literal()`.

## Step 9: CLI and Server

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
