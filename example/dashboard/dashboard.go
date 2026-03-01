package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	g "github.com/tomo3110/gerbera"
	gc "github.com/tomo3110/gerbera/components"
	gd "github.com/tomo3110/gerbera/dom"
	ge "github.com/tomo3110/gerbera/expr"
	gp "github.com/tomo3110/gerbera/property"
	gs "github.com/tomo3110/gerbera/styles"
)

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

var records = []*Record{
	{"佐藤花子", "エンジニアリング", "リードエンジニア", 8000000},
	{"鈴木一郎", "エンジニアリング", "シニアエンジニア", 7000000},
	{"高橋美咲", "デザイン", "UI デザイナー", 6500000},
	{"田中翔", "マーケティング", "マネージャー", 7500000},
	{"伊藤さくら", "人事", "採用担当", 5500000},
}

func main() {
	addr := flag.String("addr", ":8820", "running address")
	flag.Parse()
	http.HandleFunc("/detail/", detailHandle)
	http.HandleFunc("/", listHandle)
	log.Printf("dashboard server running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func listHandle(w http.ResponseWriter, _ *http.Request) {
	list := make([]g.ConvertToMap, len(records))
	for i, r := range records {
		list[i] = r
	}
	if err := g.ExecuteTemplate(w, "ja",
		gc.BootStrapCDNHead("社員ダッシュボード"),
		gd.Body(
			gp.Class("container"),
			gd.H2(
				gs.Style(g.StyleMap{"margin-top": "20px", "margin-bottom": "20px"}),
				gp.Value("社員ダッシュボード"),
			),
			gd.H3(gp.Value("社員一覧")),
			gd.Table(
				gp.Class("table", "table-striped"),
				gd.Thead(
					gd.Tr(
						gd.Th(gp.Value("名前")),
						gd.Th(gp.Value("部署")),
						gd.Th(gp.Value("役職")),
						gd.Th(gp.Value("年収")),
						gd.Th(gp.Value("操作")),
					),
				),
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
							gd.Td(gd.A(
								gp.Href("/detail/"+name),
								gp.Value("詳細"),
							)),
						)
					}),
				),
				gd.Tfoot(
					gd.Tr(
						gd.Td(
							gp.Attr("colspan", "3"),
							gp.Value("合計"),
						),
						gd.Td(
							gp.Value(formatYen(totalSalary())),
						),
						gd.Td(),
					),
				),
			),
		),
	); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func detailHandle(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/detail/"):]
	var found *Record
	for _, rec := range records {
		if rec.Name == name {
			found = rec
			break
		}
	}
	if found == nil {
		http.NotFound(w, r)
		return
	}
	if err := g.ExecuteTemplate(w, "ja",
		gc.BootStrapCDNHead(found.Name+" - 詳細"),
		gd.Body(
			gp.Class("container"),
			gd.H2(
				gs.Style(g.StyleMap{"margin-top": "20px"}),
				gp.Value(found.Name),
			),
			gd.Table(
				gp.Class("table"),
				gd.Tbody(
					gd.Tr(gd.Th(gp.Value("名前")), gd.Td(gp.Value(found.Name))),
					gd.Tr(gd.Th(gp.Value("部署")), gd.Td(gp.Value(found.Department))),
					gd.Tr(gd.Th(gp.Value("役職")), gd.Td(gp.Value(found.Role))),
					gd.Tr(gd.Th(gp.Value("年収")), gd.Td(gp.Value(formatYen(found.Salary)))),
				),
			),
			gd.A(
				gp.Href("/"),
				gp.Class("btn", "btn-primary"),
				gp.Value("← 一覧に戻る"),
			),
		),
	); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func totalSalary() int {
	total := 0
	for _, r := range records {
		total += r.Salary
	}
	return total
}

func formatYen(n int) string {
	s := fmt.Sprintf("%d", n)
	result := make([]byte, 0, len(s)+(len(s)-1)/3)
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return "¥" + string(result)
}
