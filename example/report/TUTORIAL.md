# Report Generation Tutorial

## Overview

Generate an HTML report to stdout from the CLI without an HTTP server. This tutorial covers the following features:

- Using `Template` as an `io.Reader`
- `Template.Mount()`, `Template.String()`, `Template.Bytes()`
- `Skip` — Conditionally omitting sections
- `Tag` — Directly generating elements not in `dom/`
- `styles.Style` — Inline CSS
- `Ol`, `Li` — Ordered lists
- `H4`, `H5` — Subheadings
- `Unless` — Negative conditional rendering

## Prerequisites

- Go 1.22 or later installed
- This repository cloned locally

## Step 1: Using Template Directly

```go
tmpl := &g.Template{Lang: "ja"}

if err := tmpl.Mount(
	gd.Head(
		gd.Title("Project Progress Report"),
		gd.Meta(gp.Attr("charset", "UTF-8")),
	),
	gd.Body(...),
); err != nil {
	fmt.Fprintf(os.Stderr, "Mount error: %v\n", err)
	os.Exit(1)
}
```

- `g.Template{Lang: "ja"}` creates a template directly
- `Mount(children...)` internally runs `Parse` then `Render`, accumulating HTML in a buffer
- This is the lowest-level template operation, without using `ExecuteTemplate` or `NewServeMux`

## Step 2: Output with String() / Bytes()

```go
fmt.Print(tmpl.String())     // Output HTML as a string to stdout
data := tmpl.Bytes()          // Get HTML as []byte
```

After `Mount()`, use `String()` to get the buffer contents as a string. `Bytes()` returns `[]byte`.

`Template` also implements the `io.Reader` interface. It is used internally by the `Execute()` method and can be leveraged for writing to HTTP responses and other writers.

## Step 3: Conditional Omission with Skip

```go
ge.If(done,
	g.Skip(),
	g.Tag("details",
		g.Tag("summary", gp.Value("Remaining work")),
		gd.P(gp.Value("There is still work remaining on this task.")),
	),
),
```

- `g.Skip()` returns a `ComponentFunc` that renders nothing
- Use it in conditional branches for "display nothing" cases
- In the example above, if the task is done nothing is shown; if incomplete, details are displayed

## Step 4: Generating Arbitrary Elements with Tag

```go
g.Tag("details",
	g.Tag("summary", gp.Value("Remaining work")),
	gd.P(gp.Value("There is still work remaining on this task.")),
),
```

- `g.Tag(tagName, children...)` generates an element with any tag name
- Useful for elements without wrappers in the `dom/` package (e.g., `<details>`, `<summary>`, `<dialog>`)
- In fact, all `dom/` functions are thin wrappers around `g.Tag`

## Step 5: Inline Styles and Unless

```go
ge.Unless(false,
	gd.P(
		gs.Style(g.StyleMap{"color": "red"}),
		gp.Value("Warning: The deadline is approaching."),
	),
),
```

- `ge.Unless(cond, component)` renders when the condition is false
- Combined with `gs.Style`, you can make warning messages stand out

## Running

```bash
go run example/report/report.go
```

HTML is output to stdout. To save to a file:

```bash
go run example/report/report.go > report.html
```

Open `report.html` in a browser to view.

## Exercises

1. Add a feature to write the report directly to a file (`os.Create` + `os.Stdout.Write`)
2. Try using `g.Tag` for `<dialog>` or `<progress>` elements
3. Make the report title configurable via command-line arguments

## API Reference

| Function | Description |
|----------|-------------|
| `g.Template{Lang: "ja"}` | Direct creation of a Template struct |
| `tmpl.Mount(children...)` | Mount and render components |
| `tmpl.String()` | Get HTML as a string |
| `tmpl.Bytes()` | Get HTML as `[]byte` |
| `tmpl.Read(b)` | `io.Reader` interface implementation |
| `g.Skip()` | A ComponentFunc that renders nothing |
| `g.Tag(name, children...)` | Generate an element with any tag name |
| `gs.Style(styleMap)` | Set inline CSS |
| `ge.Unless(cond, component)` | Renders when condition is false |
| `gd.H4(children...)` | `<h4>` heading |
| `gd.H5(children...)` | `<h5>` heading |
| `gd.Ol(children...)` | `<ol>` ordered list |
