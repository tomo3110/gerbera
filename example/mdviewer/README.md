# Markdown Viewer

A real-time Markdown viewer/editor built with Gerbera Live.

Features:
- **Editor + Preview**: Side-by-side Markdown editing with live preview
- **File mode**: Load, edit, and save `.md` files
- **Preview-only mode**: Watch a file for external changes (auto-refresh every 2s via `TickerView`)
- **Scroll sync**: Editor scroll position is mirrored in the preview pane via `CommandQueue`
- **Download**: Export your Markdown as a `.md` file (when no file is specified)
- **GitHub-style CSS**: Familiar Markdown rendering with GFM extensions (tables, task lists, strikethrough)

## Usage

```bash
# Editor + preview (default)
cd example/mdviewer && go run .

# Open a file for editing (save button writes back to file)
cd example/mdviewer && go run . README.md

# Preview-only mode (watches file for changes)
cd example/mdviewer && go run . -preview README.md

# Custom port + debug panel
cd example/mdviewer && go run . -addr :3000 -debug README.md
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-addr` | `:8860` | Listen address |
| `-preview` | `false` | Preview-only mode (requires file argument) |
| `-debug` | `false` | Enable Gerbera debug panel |

## Dependencies

- [goldmark](https://github.com/yuin/goldmark) — CommonMark-compliant Markdown parser with GFM extensions

## Architecture

This example uses an independent `go.mod` with `replace` directive to reference the local gerbera module, demonstrating how to use Gerbera in a separate module.

File watching is implemented via the `TickerView` interface: the server calls `HandleTick()` every 2 seconds to compare `ModTime` and detect external changes — no client-side JavaScript polling is needed.

Scroll sync uses `CommandQueue.ScrollIntoPct()` to send percentage-based scroll commands from the server to the client, eliminating the need for `MutationObserver` or `data-scroll-pct` attribute hacks.
