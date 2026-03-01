# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Gerbera is a Go HTML template engine that uses functional composition instead of traditional template files. HTML is built programmatically by composing `ComponentFunc` functions (`func(*Element) error`).

## Build & Test Commands

```bash
go build ./...                        # Build all packages
go test ./...                         # Run all tests
go test -v ./...                      # Verbose test output
go test -run TestFuncName ./...       # Run a single test
go run example/hello/hello.go           # Run hello example on :8800
go run example/portfolio/portfolio.go   # Run portfolio example on :8810
go run example/dashboard/dashboard.go   # Run dashboard example on :8820
go run example/survey/survey.go         # Run survey example on :8830
go run example/report/report.go         # Run report example (stdout, no server)
go run example/wiki/wiki.go             # Run wiki example on :8880
go run example/counter/counter.go       # Run counter LiveView example on :8840
go run example/counter/counter.go -debug  # Run counter with debug panel enabled
go run example/wiki_live/wiki_live.go   # Run wiki_live LiveView example on :8850
cd example/mdviewer && go run .           # Run mdviewer LiveView example on :8860
cd example/mdviewer && go run . -preview README.md  # Run mdviewer in preview-only mode
```

External dependencies: `github.com/gorilla/websocket` (only used by `live/` package), `github.com/yuin/goldmark` (only used by `example/mdviewer`).

## Architecture

**Rendering pipeline:** `ComponentFunc composition → Template.Mount() → Parse() → Render() → HTML output`

### Core types and files

- **`common.go`** — `ComponentFunc` type (`func(*Element) error`), `Tag()` factory, `Literal()`, `ExecuteTemplate()`
- **`element.go`** — `Element` struct (TagName, ClassNames, Attr, Children, Value) with `AppendTo()`
- **`template.go`** — `Template` struct that orchestrates mount/render, implements `io.Reader`
- **`render.go`** — Recursive HTML rendering with indentation, DOCTYPE, void element handling
- **`server.go`** — `NewServeMux()` helper for HTTP serving (default lang "ja")

### Sub-packages

- **`dom/`** — HTML element wrappers (`H1()`, `Div()`, `Body()`, `Form()`, etc.) that call `Tag()` with the appropriate tag name
- **`property/`** — Attribute setters (`Class()`, `Attr()`, `ID()`, `Name()`, `Value()`) with HTML escaping
- **`expr/`** — Control flow: `If()`, `Unless()`, `Each()` for conditional/iterative rendering
- **`styles/`** — `Style()` for inline CSS from a map
- **`components/`** — Pre-built head components (Bootstrap CDN, Materialize CSS CDN)
- **`diff/`** — Element tree diffing: `Diff()` compares two `*Element` trees and returns `[]Patch`; `RenderFragment()` for HTML fragments without DOCTYPE
- **`live/`** — Phoenix LiveView-style real-time updates: `View` interface, `Handler()`, WebSocket event loop, session management, client JS (`gerbera.js` via `go:embed`), debug panel (`WithDebug()`, `gerbera_debug.js`)

### Key patterns

- All DOM helpers return `ComponentFunc` — they compose via variadic `children ...ComponentFunc` parameters
- `Tag(tagName, children...)` is the universal element factory; `dom/` functions are thin wrappers
- Tests use table-driven patterns with a shared helper in `dom/dom_test.go`
- `property.Attr()` applies `html.EscapeString` to attribute values for XSS prevention

## Git Branching Rules

**NEVER commit directly to the `main` branch.** Always create a working branch and commit there.

`main` ブランチへの直接コミットは禁止。必ず作業ブランチを作成し、そちらにコミットすること。

```bash
git checkout -b feature/my-feature   # Create a working branch
# ... make changes ...
git add <files>
git commit -m "..."
```

## Commit Message Convention

Commit messages MUST be written in both English and Japanese (bilingual).
Format the message with English first, followed by Japanese on the next line.

コミットメッセージは必ず英語と日本語を併記すること。英語を先に、日本語を次の行に記載する。

Example:
```
Add user authentication feature
ユーザー認証機能を追加
```
