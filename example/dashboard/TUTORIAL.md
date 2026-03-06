# Dashboard Tutorial

## Overview

Build a dashboard that displays employee data in list and detail views. This tutorial covers the following features:

- `styles.Style` — Applying inline CSS
- Table elements: `Table`, `Thead`, `Tbody`, `Tfoot`, `Tr`, `Th`, `Td`
- `Each` + `ConvertToMap` — Data-driven list rendering
- Direct use of the `Map` type
- Multi-route configuration

## Prerequisites

- Go 1.22 or later installed
- This repository cloned locally

## Step 1: Defining the Data Model

```go
type Record struct {
	Name       string
	Department string
	Role       string
	Salary     int
}

func (r *Record) ToMap() g.Map {
	return g.Map{
		"name":       r.Name,
		"department": r.Department,
		"role":       r.Role,
		"salary":     r.Salary,
	}
}
```

Implementing the `ConvertToMap` interface enables data-driven table row generation with `ge.Each()`.

## Step 2: Inline CSS — styles.Style

```go
import gs "github.com/tomo3110/gerbera/styles"

gd.H2(
	gs.Style(g.StyleMap{"margin-top": "20px", "margin-bottom": "20px"}),
	gp.Value("Employee Dashboard"),
),
```

- `gs.Style(styleMap)` returns a `ComponentFunc` that sets the `style` attribute
- `g.StyleMap` is an alias for `map[string]any`
- Use CSS property names as keys with their values

## Step 3: Building Tables

```go
gd.Table(
	gp.Class("table", "table-striped"),
	gd.Thead(
		gd.Tr(
			gd.Th(gp.Value("Name")),
			gd.Th(gp.Value("Department")),
			gd.Th(gp.Value("Role")),
			gd.Th(gp.Value("Salary")),
		),
	),
	gd.Tbody(...),
	gd.Tfoot(...),
)
```

Table elements are structured in the following hierarchy:
- `Table` -> `Thead` / `Tbody` / `Tfoot` -> `Tr` -> `Th` / `Td`
- All follow the standard `ComponentFunc` pattern

## Step 4: Generating Data Rows with Each

```go
list := make([]g.ConvertToMap, len(records))
for i, r := range records {
	list[i] = r
}

gd.Tbody(
	ge.Each(list, func(item g.ConvertToMap) g.ComponentFunc {
		m := item.ToMap()
		name := m.Get("name").(string)
		dept := m.Get("department").(string)
		role := m.Get("role").(string)
		salary := m.Get("salary").(int)
		return gd.Tr(
			gd.Td(gp.Value(name)),
			gd.Td(gp.Value(dept)),
			gd.Td(gp.Value(role)),
			gd.Td(gp.Value(formatYen(salary))),
		)
	}),
),
```

Inside the `ge.Each` callback, use `ToMap().Get(key)` to retrieve each field with a type assertion.

## Step 5: Footer Row and colspan

```go
gd.Tfoot(
	gd.Tr(
		gd.Td(
			gp.Attr("colspan", "3"),
			gp.Value("Total"),
		),
		gd.Td(gp.Value(formatYen(totalSalary()))),
		gd.Td(),
	),
),
```

`gp.Attr("colspan", "3")` merges columns. `gp.Attr` works with any HTML attribute.

## Step 6: Multi-Route Configuration

```go
func main() {
	mux := http.NewServeMux()
	mux.Handle("GET /detail/{name}", g.HandlerFunc(detailHandle))
	mux.Handle("GET /", g.Handler(listPage()...))
	log.Fatal(http.ListenAndServe(":8820", mux))
}
```

`g.Handler()` is the recommended approach for serving static pages, and `g.HandlerFunc()` for dynamic pages that depend on the request.

## Running

```bash
go run example/dashboard/dashboard.go
```

Open the following URLs in your browser:
- http://localhost:8820 — Employee list table
- http://localhost:8820/detail/佐藤花子 — Individual detail page

## Exercises

1. Add a sorting feature to the table (sorting via query parameters)
2. Use `ge.If` to highlight rows where salary exceeds a certain threshold
3. Implement conditional color-coding with `gs.Style`

## API Reference

| Function | Description |
|----------|-------------|
| `gs.Style(styleMap)` | Sets inline CSS |
| `g.StyleMap` | `map[string]any` type (CSS property map) |
| `gd.Table(children...)` | `<table>` element |
| `gd.Thead(children...)` | `<thead>` element |
| `gd.Tbody(children...)` | `<tbody>` element |
| `gd.Tfoot(children...)` | `<tfoot>` element |
| `gd.Tr(children...)` | `<tr>` element |
| `gd.Th(children...)` | `<th>` element |
| `gd.Td(children...)` | `<td>` element |
| `ge.Each(list, callback)` | List iteration |
| `g.ConvertToMap` | `ToMap() Map` interface |
