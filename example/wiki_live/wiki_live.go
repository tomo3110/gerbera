package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	ge "github.com/tomo3110/gerbera/expr"
	gp "github.com/tomo3110/gerbera/property"
	gl "github.com/tomo3110/gerbera/live"
)

// dataDir is the directory where wiki page .txt files are stored.
var dataDir string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	dataDir = filepath.Join(wd, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatal(err)
	}
}

// --- file I/O helpers ---

func savePage(title, body string) error {
	return os.WriteFile(filepath.Join(dataDir, title+".txt"), []byte(body), 0600)
}

func loadPage(title string) (string, error) {
	b, err := os.ReadFile(filepath.Join(dataDir, title+".txt"))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func listPages() ([]string, error) {
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}
	var pages []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if filepath.Ext(name) == ".txt" {
			pages = append(pages, name[:len(name)-len(".txt")])
		}
	}
	return pages, nil
}

// --- pageItem implements g.ConvertToMap for expr.Each ---

type pageItem struct {
	Title string
}

func (p pageItem) ToMap() g.Map {
	return g.Map{"title": p.Title}
}

// --- WikiView ---

type WikiView struct {
	Mode    string   // "list", "view", "edit"
	Title   string   // current page title
	Body    string   // page body text
	Pages   []string // page list
	EditBuf string   // text being edited
	IsNew   bool     // true when creating a new page
}

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
		if err := savePage(v.Title, v.EditBuf); err != nil {
			return err
		}
		v.Body = v.EditBuf
		v.IsNew = false
		v.Mode = "view"
	}
	return nil
}

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

func (v *WikiView) renderView() g.ComponentFunc {
	return gd.Div(
		gd.H2(gp.Value(v.Title)),
		gd.Div(
			gd.Button(gl.Click("nav-edit"), gp.Value("編集")),
		),
		gd.P(gp.Value(v.Body)),
	)
}

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
				gd.Label(gp.For("title"), gp.Value("タイトル")),
				gd.Input(
					gp.Type("text"),
					gp.Name("title"),
					gp.Placeholder("ページタイトル"),
					gl.Input("title-input"),
				),
			),
			g.Skip(),
		),
		gd.Div(
			gd.Label(gp.For("body"), gp.Value("本文")),
			gd.Textarea(
				gp.Name("body"),
				gp.Attr("rows", "10"),
				gp.Attr("cols", "60"),
				gp.Value(v.EditBuf),
				gl.Input("edit-input"),
				gl.Debounce(300),
			),
		),
		gd.Div(
			gd.Button(gl.Click("save"), gp.Value("保存")),
			gd.Button(gl.Click("nav-list"), gp.Value("キャンセル")),
		),
	)
}

func main() {
	addr := flag.String("addr", ":8850", "listen address")
	flag.Parse()
	http.Handle("/", gl.Handler(func(_ *http.Request) gl.View { return &WikiView{} }))
	log.Printf("wiki_live running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
