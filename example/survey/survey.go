package main

import (
	"flag"
	"log"
	"net/http"
	"sync"

	g "github.com/tomo3110/gerbera"
	gc "github.com/tomo3110/gerbera/components"
	gd "github.com/tomo3110/gerbera/dom"
	ge "github.com/tomo3110/gerbera/expr"
	gp "github.com/tomo3110/gerbera/property"
)

type SurveyResponse struct {
	Name       string
	Email      string
	Language   string
	Interests  []string
}

func (s *SurveyResponse) ToMap() g.Map {
	return g.Map{
		"name":      s.Name,
		"email":     s.Email,
		"language":  s.Language,
		"interests": s.Interests,
	}
}

var (
	responses []*SurveyResponse
	mu        sync.Mutex
)

func main() {
	addr := flag.String("addr", ":8830", "running address")
	flag.Parse()
	http.HandleFunc("/submit", submitHandle)
	http.HandleFunc("/results", resultsHandle)
	http.HandleFunc("/", formHandle)
	log.Printf("survey server running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func formHandle(w http.ResponseWriter, _ *http.Request) {
	if err := g.ExecuteTemplate(w, "ja",
		gc.BootStrapCDNHead("アンケートフォーム"),
		gd.Body(
			gp.Class("container"),
			gd.H1(gp.Value("アンケートフォーム")),
			gd.A(gp.Attr("href", "/results"), gp.Value("結果を見る")),
			gd.Form(
				gp.Attr("action", "/submit"),
				gp.Attr("method", "POST"),

				gd.Div(
					gp.Class("form-group"),
					gd.Label(
						gp.Attr("for", "name"),
						gp.Value("お名前"),
					),
					gd.Input(
						gp.Attr("type", "text"),
						gp.Class("form-control"),
						gp.ID("name"),
						gp.Name("name"),
						gp.Attr("required", "required"),
					),
				),

				gd.Div(
					gp.Class("form-group"),
					gd.Label(
						gp.Attr("for", "email"),
						gp.Value("メールアドレス"),
					),
					gd.Input(
						gp.Attr("type", "email"),
						gp.Class("form-control"),
						gp.ID("email"),
						gp.Name("email"),
					),
				),

				gd.Div(
					gp.Class("form-group"),
					gd.Label(
						gp.Attr("for", "language"),
						gp.Value("好きなプログラミング言語"),
					),
					gd.Select(
						gp.Class("form-control"),
						gp.ID("language"),
						gp.Name("language"),
						gd.Option(gp.Attr("value", ""), gp.Value("選択してください")),
						gd.Option(gp.Attr("value", "Go"), gp.Value("Go")),
						gd.Option(gp.Attr("value", "TypeScript"), gp.Value("TypeScript")),
						gd.Option(gp.Attr("value", "Python"), gp.Value("Python")),
						gd.Option(gp.Attr("value", "Rust"), gp.Value("Rust")),
					),
				),

				gd.Div(
					gp.Class("form-group"),
					gd.Label(gp.Value("興味のある分野")),
					gd.Div(
						gd.InputCheckbox(false, "web",
							gp.Name("interests"),
							gp.ID("interest-web"),
						),
						gd.Label(gp.Attr("for", "interest-web"), gp.Value("Web 開発")),
					),
					gd.Div(
						gd.InputCheckbox(false, "cli",
							gp.Name("interests"),
							gp.ID("interest-cli"),
						),
						gd.Label(gp.Attr("for", "interest-cli"), gp.Value("CLI ツール")),
					),
					gd.Div(
						gd.InputCheckbox(false, "data",
							gp.Name("interests"),
							gp.ID("interest-data"),
						),
						gd.Label(gp.Attr("for", "interest-data"), gp.Value("データ分析")),
					),
					gd.Div(
						gd.InputCheckbox(false, "infra",
							gp.Name("interests"),
							gp.ID("interest-infra"),
						),
						gd.Label(gp.Attr("for", "interest-infra"), gp.Value("インフラ")),
					),
				),

				gd.Submit(
					gp.Class("btn", "btn-primary"),
					gp.Value("送信"),
				),
			),
		),
	); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func submitHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp := &SurveyResponse{
		Name:      r.FormValue("name"),
		Email:     r.FormValue("email"),
		Language:  r.FormValue("language"),
		Interests: r.Form["interests"],
	}
	mu.Lock()
	responses = append(responses, resp)
	mu.Unlock()
	http.Redirect(w, r, "/results", http.StatusSeeOther)
}

func resultsHandle(w http.ResponseWriter, _ *http.Request) {
	mu.Lock()
	list := make([]g.ConvertToMap, len(responses))
	for i, r := range responses {
		list[i] = r
	}
	mu.Unlock()

	if err := g.ExecuteTemplate(w, "ja",
		gc.BootStrapCDNHead("アンケート結果"),
		gd.Body(
			gp.Class("container"),
			gd.H1(gp.Value("アンケート結果")),
			gd.A(gp.Attr("href", "/"), gp.Value("フォームに戻る")),
			ge.Unless(len(list) > 0,
				gd.P(gp.Class("text-muted"), gp.Value("まだ回答がありません。")),
			),
			ge.If(len(list) > 0,
				gd.Ol(
					ge.Each(list, func(item g.ConvertToMap) g.ComponentFunc {
						m := item.ToMap()
						name := m.Get("name").(string)
						lang := m.Get("language").(string)
						return gd.Li(
							gd.P(
								gd.B(gp.Value(name)),
								g.Literal(" — 好きな言語: "+lang),
							),
						)
					}),
				),
			),
		),
	); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
