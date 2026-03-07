# Hello Tutorial

## Overview

Build a minimal web page using gerbera. This tutorial covers the following features:

- `Handler` — Creating a simple HTTP server
- `BootstrapCSS` — Adding Bootstrap 5 CSS/JS to `<head>`
- `Body`, `H1`, `P` — Building basic DOM elements
- `Class`, `Value` — Setting attributes and text content

## Prerequisites

- Go 1.22 or later installed
- This repository cloned locally

## Step 1: Importing Packages

```go
import (
	"flag"
	"log"
	"net/http"

	g "github.com/tomo3110/gerbera"
	gc "github.com/tomo3110/gerbera/components"
	gd "github.com/tomo3110/gerbera/dom"
	gp "github.com/tomo3110/gerbera/property"
)
```

By convention, gerbera uses the following import aliases:
- `g` — Core package
- `gc` — Pre-built components (Bootstrap CDN, etc.)
- `gd` — DOM elements (`H1`, `P`, `Body`, etc.)
- `gp` — Properties (`Class`, `Value`, etc.)

## Step 2: Defining the body Function

```go
func body() g.ComponentFunc {
	return gd.Body(
		gp.Class("container"),
		gd.H1(gp.Value("Gerbera Template Engine !")),
		gd.P(gp.Value("view html template Test")),
	)
}
```

`ComponentFunc` is the type `func(Node)` and serves as the fundamental unit of gerbera. `Node` is the interface that abstracts over `Element`. All DOM helpers return a `ComponentFunc`.

- `gd.Body(...)` — Creates a `<body>` element and appends children passed as arguments
- `gp.Class("container")` — Sets the `class="container"` attribute
- `gp.Value("...")` — Sets the text content

## Step 3: Starting the Server

```go
func main() {
	addr := flag.String("addr", ":8800", "running address")
	flag.Parse()
	mux := http.NewServeMux()
	mux.Handle("GET /", g.Handler(
		gd.Head(
			gd.Title("Gerbera Template Engine !"),
			gc.BootstrapCSS(),
		),
		body(),
	))
	log.Fatal(http.ListenAndServe(*addr, mux))
}
```

- `g.Handler(children...)` returns an `http.Handler` that renders the given components as an HTML page
- `gd.Head(children...)` creates a `<head>` element with the given children
- `gc.BootstrapCSS()` adds Bootstrap 5 CSS/JS CDN links and ensures charset/viewport meta tags are present

## Running

```bash
go run example/hello/hello.go
```

Open http://localhost:8800 in your browser to verify.

To change the port:

```bash
go run example/hello/hello.go -addr :3000
```

## Exercises

1. Add sections to the page using `gd.H2` and `gd.Div`
2. Add a link (`gd.A`) with `gp.Attr("href", "...")`
3. Try creating a list using `gd.Ul` and `gd.Li`

## API Reference

| Function | Description |
|----------|-------------|
| `g.Handler(children...)` | Returns an `http.Handler` that renders the given components |
| `gc.BootstrapCSS()` | Adds Bootstrap 5 CSS/JS to `<head>` |
| `gd.Head(children...)` | `<head>` element |
| `gd.Title(text)` | `<title>` element |
| `gd.Body(children...)` | `<body>` element |
| `gd.H1(children...)` | `<h1>` element |
| `gd.P(children...)` | `<p>` element |
| `gp.Class(names...)` | Sets the `class` attribute |
| `gp.Value(v)` | Sets text content |
