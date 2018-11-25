package main

import (
	"io/ioutil"
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
	http.HandleFunc("/view/", viewHandle)
	http.HandleFunc("/edit/", editHandle)
	http.HandleFunc("/save/", saveHandle)
	http.HandleFunc("/", listHandle)
	log.Fatal(http.ListenAndServe(":8880", nil))
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
	return ioutil.WriteFile(filePath, p.Body, 0600)
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
	body, err := ioutil.ReadFile(filePath)
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
	files, err := ioutil.ReadDir(wd)
	if err != nil {
		return nil, err
	}
	for _, info := range files {
		name := info.Name()
		if info.IsDir() {
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

func viewHandle(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/view/"):]
	page, err := LoadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	if err := viewLayout(w, page); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func editHandle(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/edit/"):]
	page, err := LoadPage(title)
	if err != nil {
		page = &Page{Title: title}
	}
	if err := editLayout(w, page); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func saveHandle(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/save/"):]
	body := r.FormValue("body")
	page := &Page{Title: title, Body: []byte(body)}
	if err := page.Save(); err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func listHandle(w http.ResponseWriter, _ *http.Request) {
	list, err := FetchPageNameList()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := listLayout(w, list); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

/** view */

func viewLayout(w http.ResponseWriter, page *Page) error {
	return g.ExecuteTemplate(w, "jp",
		gc.BootStrapCDNHead(page.Title),
		gd.Body(
			gp.Class("container"),
			gd.H1(gp.Value(page.Title)),
			gd.A(
				gp.Attr("href", "/"),
				gp.Value("list"),
			),
			gd.A(
				gp.Attr("href", "/edit/"+page.Title),
				gp.Value("Edit"),
			),
			gd.P(
				gp.Value(string(page.Body)),
			),
		),
	)
}

func editLayout(w http.ResponseWriter, page *Page) error {
	return g.ExecuteTemplate(w, "jp",
		gc.BootStrapCDNHead(page.Title),
		gd.Body(
			gp.Class("container"),
			gd.H1(gp.Value(page.Title)),
			gd.A(
				gp.Attr("href", "/view/"+page.Title),
				gp.Value("cancel"),
			),
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
			),
		),
	)
}

func listLayout(w http.ResponseWriter, pages []*Page) error {
	list := make([]g.ConvertToMap, len(pages))
	for i, page := range pages {
		list[i] = page
	}
	return g.ExecuteTemplate(w, "jp",
		gc.BootStrapCDNHead("ページ一覧"),
		gd.Body(
			gp.Class("container"),
			gd.H1(gp.Value("ページ一覧")),
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
			),
		),
	)
}
