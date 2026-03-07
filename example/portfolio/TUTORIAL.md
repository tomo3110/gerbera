# Portfolio Tutorial

## Overview

Build a portfolio site styled with Materialize CSS. This tutorial covers the following features:

- `MaterializeCSS` — Adding Materialize CSS/JS to `<head>`
- Semantic HTML5: `Header`, `Footer`, `Nav`, `Main`, `Section`, `Article`, `Aside`
- `Img` — Image elements
- `ID` — Setting the `id` attribute
- `Literal` — Embedding raw text
- `B`, `Code`, `Hr`, `Br`, `Address` — Inline/block elements

## Prerequisites

- Go 1.22 or later installed
- This repository cloned locally

## Step 1: Using Materialize CSS

```go
mux := http.NewServeMux()
mux.Handle("GET /", g.Handler(
	gd.Head(
		gd.Title("Portfolio - Taro Tanaka"),
		gc.MaterializeCSS(),
	),
	body(),
))
```

`gc.MaterializeCSS()` adds Materialize CSS CDN links to the parent `<head>` element and ensures charset/viewport meta tags are present. It can be used as an alternative to Bootstrap.

## Step 2: Semantic HTML5 Structure

```go
func body() g.ComponentFunc {
	return gd.Body(
		navBar(),      // Header + Nav
		gd.Main(       // Main
			aboutSection(),    // Section + Article
			skillsSection(),   // Section
			projectsSection(), // Section + Article + Aside
		),
		footer(),      // Footer
	)
}
```

gerbera supports all HTML5 semantic elements such as `Header`, `Footer`, `Nav`, `Main`, `Section`, `Article`, and `Aside`. Each function has the same interface as `gd.Div`.

## Step 3: Navigation Bar

```go
func navBar() g.ComponentFunc {
	return gd.Header(
		gd.Nav(
			gp.Class("nav-wrapper"),
			gd.A(
				gp.Href("#"),
				gp.Class("brand-logo"),
				gp.Value("Taro Tanaka"),
			),
			gd.Ul(
				gp.ID("nav-mobile"),
				gp.Class("right"),
				gd.Li(gd.A(gp.Href("#about"), gp.Value("About"))),
				gd.Li(gd.A(gp.Href("#skills"), gp.Value("Skills"))),
			),
		),
	)
}
```

- `gp.ID("nav-mobile")` sets `id="nav-mobile"`
- Use `#section-name` for in-page links and set `gp.ID()` on the corresponding sections

## Step 4: Images and Literal

```go
gd.Img("https://via.placeholder.com/150",
	gp.Attr("alt", "Profile image"),
	gp.Class("circle", "responsive-img"),
),
gd.P(
	gp.Value("Go engineer. "),
	g.Literal("Specializing in web application development."),
),
```

- `gd.Img(src, children...)` — The first argument is the image URL. Since `<img>` is a void element, no closing tag is generated
- `g.Literal(lines...)` — Appends raw text directly to the element's `Value`. Can be combined with `gp.Value` to concatenate text

## Step 5: Inline Elements

```go
gd.Li(
	gp.Class("collection-item"),
	gd.B(gp.Value("Go")),
	g.Literal(" — Web / CLI development"),
),
gd.P(
	gp.Value("Favorite editor: "),
	gd.Code(gp.Value("Vim")),
),
```

- `gd.B(...)` — `<b>` bold element
- `gd.Code(...)` — `<code>` code element
- `gd.Hr()` — `<hr>` horizontal rule (void element)
- `gd.Br()` — `<br>` line break (void element)

## Step 6: Footer and Address

```go
func footer() g.ComponentFunc {
	return gd.Footer(
		gp.Class("page-footer"),
		gd.Div(
			gp.Class("container"),
			gd.P(gp.Value("© 2026 Taro Tanaka")),
			gd.Address(
				gp.Value("Contact: taro@example.com"),
				gd.Br(),
				g.Literal("Shibuya, Tokyo"),
			),
		),
	)
}
```

- `gd.Address(...)` — `<address>` element, used for contact information
- `gd.Br()` is a void element and can be used without children

## Running

```bash
go run example/portfolio/portfolio.go
```

Open http://localhost:8810 in your browser to verify.

## Exercises

1. Add a "Career" section using `gd.Section`
2. Set an actual profile image with `gd.Img`
3. Add external links using `gp.Attr("target", "_blank")`

## API Reference

| Function | Description |
|----------|-------------|
| `gc.MaterializeCSS()` | Adds Materialize CSS/JS to `<head>` |
| `gd.Head(children...)` | `<head>` element |
| `gd.Title(text)` | `<title>` element |
| `gd.Header(children...)` | `<header>` element |
| `gd.Footer(children...)` | `<footer>` element |
| `gd.Nav(children...)` | `<nav>` element |
| `gd.Main(children...)` | `<main>` element |
| `gd.Section(children...)` | `<section>` element |
| `gd.Article(children...)` | `<article>` element |
| `gd.Aside(children...)` | `<aside>` element |
| `gd.Img(src, children...)` | `<img>` element |
| `gd.B(children...)` | `<b>` element |
| `gd.Code(children...)` | `<code>` element |
| `gd.Hr(children...)` | `<hr>` element |
| `gd.Br(children...)` | `<br>` element |
| `gd.Address(children...)` | `<address>` element |
| `gp.ID(text)` | Sets the `id` attribute |
| `g.Literal(lines...)` | Appends raw text to Value |
