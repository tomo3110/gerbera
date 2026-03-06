package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	g "github.com/tomo3110/gerbera"
	gc "github.com/tomo3110/gerbera/components"
	gd "github.com/tomo3110/gerbera/dom"
	ge "github.com/tomo3110/gerbera/expr"
	gp "github.com/tomo3110/gerbera/property"
)

func init() {
	cachePages = make(map[string]*Page)
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("GET /view/{title}", g.HandlerFunc(viewHandle))
	mux.Handle("GET /edit/{title}", g.HandlerFunc(editHandle))
	mux.Handle("POST /save/{title}", http.HandlerFunc(saveHandle))
	mux.Handle("GET /", g.HandlerFunc(listHandle))
	log.Fatal(http.ListenAndServe(":8880", mux))
}

/** model */

var cachePages map[string]*Page

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) Save() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	filePath := filepath.Join(wd, p.Title+".txt")
	cachePages[p.Title] = p
	return os.WriteFile(filePath, p.Body, 0600)
}

func (p *Page) ToMap() g.Map {
	return g.Map{
		"title": p.Title,
		"body":  p.Body,
	}
}

func LoadPage(title string) (*Page, error) {
	if page, ok := cachePages[title]; ok {
		return page, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	filePath := filepath.Join(wd, title+".txt")
	if _, err := os.Stat(filePath); err != nil {
		return nil, err
	}
	body, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func FetchPageNameList() ([]*Page, error) {
	pages := make([]*Page, 0)
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(wd)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(name) != ".txt" {
			continue
		}
		title := name[:len(name)-len(".txt")]
		pages = append(pages, &Page{Title: title})
	}
	return pages, nil
}

/** controller */

func viewHandle(r *http.Request) []g.ComponentFunc {
	title := r.PathValue("title")
	page, err := LoadPage(title)
	if err != nil {
		// Page not found — return edit page for creation
		return editPage(&Page{Title: title})
	}
	return viewPage(page)
}

func editHandle(r *http.Request) []g.ComponentFunc {
	title := r.PathValue("title")
	page, err := LoadPage(title)
	if err != nil {
		page = &Page{Title: title}
	}
	return editPage(page)
}

func saveHandle(w http.ResponseWriter, r *http.Request) {
	title := r.PathValue("title")
	body := r.FormValue("body")
	page := &Page{Title: title, Body: []byte(body)}
	if err := page.Save(); err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func listHandle(r *http.Request) []g.ComponentFunc {
	list, err := FetchPageNameList()
	if err != nil {
		return nil
	}
	return listPage(list)
}

/** view */

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

func editPage(page *Page) []g.ComponentFunc {
	return []g.ComponentFunc{
		gd.Head(
			gd.Title(page.Title),
			gc.BootstrapCSS(),
		),
		gd.Body(
			gp.Class("container"),
			gd.H1(gp.Value(page.Title)),
			gd.A(
				gp.Href("/view/"+page.Title),
				gp.Value("cancel"),
			),
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
			),
		),
	}
}

func listPage(pages []*Page) []g.ComponentFunc {
	list := make([]g.ConvertToMap, len(pages))
	for i, page := range pages {
		list[i] = page
	}
	return []g.ComponentFunc{
		gd.Head(
			gd.Title("ページ一覧"),
			gc.BootstrapCSS(),
		),
		gd.Body(
			gp.Class("container"),
			gd.H1(gp.Value("ページ一覧")),
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
			),
		),
	}
}
