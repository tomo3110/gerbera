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
go run example/hello/hello.go         # Run hello example on :8800
go run example/wiki/wiki.go           # Run wiki example on :8880
```

No external dependencies — standard library only.

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

### Key patterns

- All DOM helpers return `ComponentFunc` — they compose via variadic `children ...ComponentFunc` parameters
- `Tag(tagName, children...)` is the universal element factory; `dom/` functions are thin wrappers
- Tests use table-driven patterns with a shared helper in `dom/dom_test.go`
- `property.Attr()` applies `html.EscapeString` to attribute values for XSS prevention
