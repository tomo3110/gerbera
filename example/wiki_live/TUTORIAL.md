# Wiki Live Tutorial

## Overview

Build a SPA-style Wiki that runs entirely without page reloads, using the `live/` package. This tutorial covers the following features:

- `gl.View` interface — Managing multiple screens (list / view / edit) with state fields
- `gl.Click` / `gl.ClickValue` — Navigation with click events that carry data
- `gl.Input` — Real-time form input binding
- `expr.If` / `expr.Each` — Conditional branching and list rendering inside a LiveView
- File I/O — Persisting wiki pages to `.txt` files on disk

Compared to the [Counter example](../counter/), this is a more practical demonstration of Gerbera Live, showing how to build a multi-screen application with forms and data persistence.

## Prerequisites

- Go 1.22 or later installed
- This repository cloned locally

## Step 1: Importing Packages

```go
import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	g  "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	ge "github.com/tomo3110/gerbera/expr"
	gp "github.com/tomo3110/gerbera/property"
	gl "github.com/tomo3110/gerbera/live"
)
```

In addition to the standard gerbera aliases (`g`, `gd`, `gp`), import:
- `ge` — `expr` package for `If()` and `Each()`
- `live` — LiveView core: `View` interface, `Handler()`, event bindings

## Step 2: File I/O Helpers

```go
var dataDir string

func init() {
	wd, _ := os.Getwd()
	dataDir = filepath.Join(wd, "data")
	os.MkdirAll(dataDir, 0755)
}

func savePage(title, body string) error {
	return os.WriteFile(filepath.Join(dataDir, title+".txt"), []byte(body), 0600)
}

func loadPage(title string) (string, error) {
	b, err := os.ReadFile(filepath.Join(dataDir, title+".txt"))
	return string(b), err
}

func listPages() ([]string, error) {
	entries, _ := os.ReadDir(dataDir)
	var pages []string
	for _, e := range entries {
		name := e.Name()
		if filepath.Ext(name) == ".txt" {
			pages = append(pages, name[:len(name)-len(".txt")])
		}
	}
	return pages, nil
}
```

Pages are stored as `.txt` files in a `data/` directory (auto-created on startup). This is the same file-based approach as the [original Wiki example](../wiki/), keeping the data layer simple.

## Step 3: Defining the View Struct

```go
type WikiView struct {
	Mode    string   // "list", "view", "edit"
	Title   string   // current page title
	Body    string   // page body text
	Pages   []string // page list
	EditBuf string   // text being edited
	IsNew   bool     // true when creating a new page
}
```

Unlike the Counter example (which has a single `Count` field), WikiView manages multiple screen modes through the `Mode` field. The rendering logic branches based on this value.

To use `expr.Each`, a helper struct implementing `ConvertToMap` is also needed:

```go
type pageItem struct{ Title string }

func (p pageItem) ToMap() g.Map {
	return g.Map{"title": p.Title}
}
```

## Step 4: Implementing Mount

```go
func (v *WikiView) Mount(_ gl.Params) error {
	v.Mode = "list"
	return v.refreshPages()
}

func (v *WikiView) refreshPages() error {
	pages, err := listPages()
	if err != nil {
		return err
	}
	v.Pages = pages
	return nil
}
```

On session creation, the view initializes in list mode and loads the page list from disk.

## Step 5: Implementing HandleEvent

```go
func (v *WikiView) HandleEvent(event string, payload gl.Payload) error {
	switch event {
	case "nav-list":
		v.Mode = "list"
		return v.refreshPages()
	case "nav-view":
		title := payload["value"]
		body, err := loadPage(title)
		if err != nil {
			return err
		}
		v.Title = title
		v.Body = body
		v.Mode = "view"
	case "nav-edit":
		v.EditBuf = v.Body
		v.IsNew = false
		v.Mode = "edit"
	case "nav-new":
		v.Title = ""
		v.Body = ""
		v.EditBuf = ""
		v.IsNew = true
		v.Mode = "edit"
	case "edit-input":
		v.EditBuf = payload["value"]
	case "title-input":
		v.Title = payload["value"]
	case "save":
		if v.Title == "" {
			return fmt.Errorf("title is required")
		}
		savePage(v.Title, v.EditBuf)
		v.Body = v.EditBuf
		v.IsNew = false
		v.Mode = "view"
	}
	return nil
}
```

Key points:
- **Screen transitions** — Events like `nav-list`, `nav-view`, `nav-edit`, `nav-new` change the `Mode` field, causing `Render()` to output different component trees
- **Data passing via ClickValue** — `nav-view` reads `payload["value"]` to get the clicked page title (set by `gl.ClickValue`)
- **Input binding** — `edit-input` and `title-input` update state fields from `payload["value"]` (set by `gl.Input`)
- **Persistence** — `save` writes to disk, then transitions to view mode

## Step 6: Implementing Render

The `Render` method returns a common layout, and delegates the main content to a helper that switches on `Mode`:

```go
func (v *WikiView) Render() []g.ComponentFunc {
	return []g.ComponentFunc{
		gd.Head(gd.Title("Wiki Live")),
		gd.Body(
			gp.Class("container"),
			gd.H1(gp.Value("Wiki Live")),
			gd.Nav(
				gd.Button(gl.Click("nav-list"), gp.Value("ページ一覧")),
				gd.Button(gl.Click("nav-new"), gp.Value("新規作成")),
			),
			gd.Hr(),
			v.renderContent(),
		),
	}
}

func (v *WikiView) renderContent() g.ComponentFunc {
	switch v.Mode {
	case "view":
		return v.renderView()
	case "edit":
		return v.renderEdit()
	default:
		return v.renderList()
	}
}
```

### List Mode — `expr.If` + `expr.Each` + `gl.ClickValue`

```go
func (v *WikiView) renderList() g.ComponentFunc {
	items := make([]g.ConvertToMap, len(v.Pages))
	for i, p := range v.Pages {
		items[i] = pageItem{Title: p}
	}
	return gd.Div(
		gd.H2(gp.Value("ページ一覧")),
		ge.If(len(items) > 0,
			gd.Ul(
				ge.Each(items, func(item g.ConvertToMap) g.ComponentFunc {
					title := item.ToMap().Get("title").(string)
					return gd.Li(
						gd.Button(
							gl.Click("nav-view"),
							gl.ClickValue(title),
							gp.Value(title),
						),
					)
				}),
			),
			gd.P(gp.Value("ページがありません")),
		),
	)
}
```

- `gl.Click("nav-view")` binds the click event
- `gl.ClickValue(title)` sets the `gerbera-value` attribute so the page title is sent in the payload as `"value"`
- `ge.If` shows the list or an empty message
- `ge.Each` iterates over the page items

### View Mode

```go
func (v *WikiView) renderView() g.ComponentFunc {
	return gd.Div(
		gd.H2(gp.Value(v.Title)),
		gd.Div(
			gd.Button(gl.Click("nav-edit"), gp.Value("編集")),
		),
		gd.P(gp.Value(v.Body)),
	)
}
```

### Edit Mode — `gl.Input`

```go
func (v *WikiView) renderEdit() g.ComponentFunc {
	return gd.Div(
		gd.H2(
			ge.If(v.IsNew,
				gp.Value("新規ページ"),
				gp.Value(v.Title+" を編集"),
			),
		),
		ge.If(v.IsNew,
			gd.Div(
				gd.Label(gp.Attr("for", "title"), gp.Value("タイトル")),
				gd.Input(
					gp.Attr("type", "text"),
					gp.Name("title"),
					gp.Attr("placeholder", "ページタイトル"),
					gl.Input("title-input"),
				),
			),
			g.Skip(),
		),
		gd.Div(
			gd.Label(gp.Attr("for", "body"), gp.Value("本文")),
			gd.Textarea(
				gp.Name("body"),
				gp.Attr("rows", "10"),
				gp.Attr("cols", "60"),
				gp.Value(v.EditBuf),
				gl.Input("edit-input"),
			),
		),
		gd.Div(
			gd.Button(gl.Click("save"), gp.Value("保存")),
			gd.Button(gl.Click("nav-list"), gp.Value("キャンセル")),
		),
	)
}
```

- `gl.Input("edit-input")` sends each keystroke to the server, updating `EditBuf`
- `ge.If(v.IsNew, ...)` shows the title input field only when creating a new page
- `g.Skip()` renders nothing — used as the false branch of `ge.If`

## Step 7: Starting the Server

```go
func main() {
	addr := flag.String("addr", ":8850", "listen address")
	flag.Parse()
	http.Handle("/", gl.Handler(func() gl.View { return &WikiView{} }))
	log.Printf("wiki_live running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
```

A single `gl.Handler` handles everything — initial HTML rendering, WebSocket upgrade, and event processing. No multiple routes needed.

## How It Works

```
Browser                                   Go Server
  |                                          |
  |  1. GET /                                |
  |  <-- Initial HTML (list mode)            |
  |       + gerbera.js                       |
  |                                          |
  |  2. WebSocket connect                    |
  |  <-> Bidirectional connection            |
  |                                          |
  |  3. Click page title "MyPage"            |
  |  --> {"e":"nav-view",                    |
  |       "p":{"value":"MyPage"}}            |
  |                                          |
  |      HandleEvent -> load file            |
  |      -> Mode="view" -> Render            |
  |      -> Diff old tree vs new -> Patches  |
  |                                          |
  |  4. <-- DOM patches (replace list        |
  |          with view content)              |
  |                                          |
  |  5. Click "編集"                          |
  |  --> {"e":"nav-edit","p":{}}             |
  |                                          |
  |  6. Type in textarea                     |
  |  --> {"e":"edit-input",                  |
  |       "p":{"value":"new text..."}}       |
  |                                          |
  |  7. Click "保存"                          |
  |  --> {"e":"save","p":{}}                 |
  |      -> save file -> Mode="view"         |
  |  <-- DOM patches                         |
```

## Comparison with the Original Wiki

| Aspect | `example/wiki/` | `example/wiki_live/` |
|--------|-----------------|---------------------|
| Rendering | Server-side (full page) | LiveView (WebSocket patches) |
| Navigation | Page reloads (`<a href>`) | No reloads (`gl.Click`) |
| Routes | 4 (`/`, `/view/`, `/edit/`, `/save/`) | 1 (`/`) |
| Form handling | POST + redirect | `gl.Input` + `gl.Click` |
| User experience | Traditional web app | SPA-like |

## Running

```bash
go run example/wiki_live/wiki_gl.go
```

Open http://localhost:8850 in your browser and try:
1. Click "新規作成" to create a new page
2. Enter a title and body, then click "保存"
3. Click "ページ一覧" to return to the list
4. Click a page title to view it
5. Click "編集" to edit the page

All interactions happen without page reloads.

To change the port:

```bash
go run example/wiki_live/wiki_gl.go -addr :3000
```

## Exercises

1. Add a delete button to remove a page from the view screen
2. Add a search/filter input on the list screen using `gl.Input`
3. Show a confirmation message before saving (add a `Message` state field)
4. Add Markdown rendering support by integrating a Markdown parser

## API Reference

| Function | Description |
|----------|-------------|
| `gl.Handler(factory, opts...)` | Returns an `http.Handler` for a LiveView |
| `gl.WithDebug()` | Enables browser DevPanel and server-side structured logging |
| `gl.Click(event)` | Binds a click event |
| `gl.ClickValue(value)` | Sets a value to send with click events |
| `gl.Input(event)` | Binds an input event (sends `value` in payload) |
| `gl.View` | Interface: `Mount`, `Render`, `HandleEvent` |
| `gl.Params` | `map[string]string` — URL parameters |
| `gl.Payload` | `map[string]string` — Event data |
| `ge.If(cond, true, false...)` | Conditional branching |
| `ge.Each(list, callback)` | List iteration over `ConvertToMap` slice |
| `g.Skip()` | Renders nothing (useful as `If` false branch) |
| `g.ConvertToMap` | Interface with `ToMap() Map` for `Each` |
