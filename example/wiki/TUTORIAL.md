# Wiki Tutorial

## Overview

Build a Wiki application with CRUD functionality using gerbera. This tutorial covers the following features:

- `HandlerFunc` — Wrapping request-dependent functions as HTTP handlers
- `Form`, `Textarea`, `Button`, `Label` — Form elements
- `If` / `Each` — Conditional branching and list rendering
- `ConvertToMap` — Converting structs to template data
- `Attr` — Setting arbitrary HTML attributes

## Prerequisites

- Go 1.22 or later installed
- This repository cloned locally

## Step 1: Defining the Data Model

```go
type Page struct {
	Title string
	Body  []byte
}

func (p *Page) ToMap() g.Map {
	return g.Map{
		"title": p.Title,
		"body":  p.Body,
	}
}
```

Implementing the `ConvertToMap` interface allows the struct to be used with `expr.Each()`.
`g.Map` is an alias for `map[string]any` and has a `Get(key)` method.

## Step 2: Setting Up HTTP Handlers

```go
func main() {
	mux := http.NewServeMux()
	mux.Handle("GET /view/{title}", g.HandlerFunc(viewHandle))
	mux.Handle("GET /edit/{title}", g.HandlerFunc(editHandle))
	mux.Handle("POST /save/{title}", http.HandlerFunc(saveHandle))
	mux.Handle("GET /", g.HandlerFunc(listHandle))
	log.Fatal(http.ListenAndServe(":8880", mux))
}
```

`g.HandlerFunc()` wraps a function with the signature `func(r *http.Request) []g.ComponentFunc` into an `http.Handler`. The wrapped function receives the request and returns a slice of `ComponentFunc` that gerbera renders as HTML automatically. For handlers that need direct access to `http.ResponseWriter` (e.g., redirects), use the standard `http.HandlerFunc` instead, as shown with `saveHandle`.

Routes use Go 1.22's enhanced `ServeMux` pattern syntax with method prefixes (`GET`, `POST`) and path parameters (`{title}`).

## Step 3: Building Views — Page Functions

```go
func viewPage(page *Page) []g.ComponentFunc {
	return []g.ComponentFunc{
		gd.Head(
			gd.Title(page.Title),
			gc.BootstrapCSS(),
		),
		gd.Body(
			gp.Class("container"),
			gd.H1(gp.Value(page.Title)),
			gd.A(
				gp.Href("/"),
				gp.Value("list"),
			),
			gd.A(
				gp.Href("/edit/"+page.Title),
				gp.Value("Edit"),
			),
			gd.P(
				gp.Value(string(page.Body)),
			),
		),
	}
}
```

Each page function takes the data it needs and returns `[]g.ComponentFunc`. The handler function calls the appropriate page function and returns its result:

```go
func viewHandle(r *http.Request) []g.ComponentFunc {
	title := r.PathValue("title")
	page, err := LoadPage(title)
	if err != nil {
		// Page not found — return edit page for creation
		return editPage(&Page{Title: title})
	}
	return viewPage(page)
}
```

When a page is not found, `viewHandle` returns `editPage` directly instead of issuing an HTTP redirect. This keeps the response in a single round-trip while still presenting the user with the edit form for the new page.

## Step 4: Building Forms

```go
gd.Form(
	gp.Attr("action", "/save/"+page.Title),
	gp.Attr("method", "POST"),
	gd.Div(
		gp.Class("form-group"),
		gd.Label(
			gp.For("body"),
			gp.Value("body"),
		),
		gd.Textarea(
			gp.Class("form-control"),
			gp.Attr("row", 20),
			gp.Name("body"),
			gp.Value(string(page.Body)),
		),
	),
	gd.Button(
		gp.Type("submit"),
		gp.Class("btn", "btn-primary"),
		gp.Value("Save"),
	),
)
```

- `gp.Attr(key, value)` sets any HTML attribute
- `gp.Name(text)` is a shortcut for `gp.Attr("name", text)`
- Form elements return `ComponentFunc` just like everything else, so they compose freely

## Step 5: Conditional Branching and List Rendering

```go
ge.If(len(list) > 0,
	ge.Each(list, func(page g.ConvertToMap) g.ComponentFunc {
		title := page.ToMap().Get("title").(string)
		return gd.Li(
			gd.A(
				gp.Href("/view/"+title),
				gp.Value(title),
			),
		)
	}),
	gd.P(gp.Value("empty pages ...")),
)
```

- `ge.If(cond, trueComponent, falseComponent...)` — Renders the 2nd argument if true, the 3rd if false
- `ge.Each(list, callback)` — Iterates over a `ConvertToMap` slice and calls the callback for each element
- Inside the callback, use `ToMap().Get(key)` to access fields

## Running

```bash
go run example/wiki/wiki.go
```

Open http://localhost:8880 in your browser and try the following:
1. View the page list on the top page
2. Navigate to `/view/test` — since the page does not exist yet, the edit form is displayed automatically
3. Enter content and save, then verify it at `/view/test`

## Exercises

1. Add a delete feature (a `/delete/` handler)
2. Use `ge.Unless` to implement "show when page is not empty" logic
3. Integrate a Markdown parser to render the body as Markdown

## API Reference

| Function | Description |
|----------|-------------|
| `g.HandlerFunc(fn)` | Wraps `func(r *http.Request) []ComponentFunc` as an `http.Handler` |
| `g.ConvertToMap` | Interface with `ToMap() Map` |
| `g.Map` | `map[string]any` type with `Get(key)` method |
| `gd.Form(children...)` | `<form>` element |
| `gd.Textarea(children...)` | `<textarea>` element |
| `gd.Button(children...)` | `<button>` element |
| `gd.Label(children...)` | `<label>` element |
| `gd.A(children...)` | `<a>` element |
| `gd.Li(children...)` | `<li>` element |
| `gp.Attr(key, value)` | Sets any HTML attribute |
| `gp.Name(text)` | Sets the `name` attribute |
| `ge.If(cond, true, false...)` | Conditional branching |
| `ge.Each(list, callback)` | List iteration |
