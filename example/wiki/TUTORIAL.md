# Wiki Tutorial

## Overview

Build a Wiki application with CRUD functionality using gerbera. This tutorial covers the following features:

- `ExecuteTemplate` — Direct template execution
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
	http.HandleFunc("/view/", viewHandle)
	http.HandleFunc("/edit/", editHandle)
	http.HandleFunc("/save/", saveHandle)
	http.HandleFunc("/", listHandle)
	log.Fatal(http.ListenAndServe(":8880", nil))
}
```

`NewServeMux` is a helper optimized for single-page apps. For applications with multiple routes, use the standard `http.HandleFunc` directly.

## Step 3: Building Views — ExecuteTemplate

```go
func viewLayout(w http.ResponseWriter, page *Page) error {
	return g.ExecuteTemplate(w, "jp",
		gc.BootStrapCDNHead(page.Title),
		gd.Body(
			gp.Class("container"),
			gd.H1(gp.Value(page.Title)),
			gd.P(gp.Value(string(page.Body))),
		),
	)
}
```

`g.ExecuteTemplate(w, lang, children...)` internally creates a `Template`, mounts components, renders HTML, and writes it to the writer in a single call.

## Step 4: Building Forms

```go
gd.Form(
	gp.Attr("action", "/save/"+page.Title),
	gp.Attr("method", "POST"),
	gd.Div(
		gp.Class("form-group"),
		gd.Label(
			gp.Attr("for", "body"),
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
		gp.Attr("type", "submit"),
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
				gp.Attr("href", "/view/"+title),
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
2. Navigate to `/edit/test` to create a new page
3. After saving, verify it at `/view/test`

## Exercises

1. Add a delete feature (a `/delete/` handler)
2. Use `ge.Unless` to implement "show when page is not empty" logic
3. Integrate a Markdown parser to render the body as Markdown

## API Reference

| Function | Description |
|----------|-------------|
| `g.ExecuteTemplate(w, lang, children...)` | Creates, renders, and writes a template |
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
