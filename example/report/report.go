package main

import (
	"fmt"
	"os"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	ge "github.com/tomo3110/gerbera/expr"
	gp "github.com/tomo3110/gerbera/property"
	gs "github.com/tomo3110/gerbera/styles"
)

type Item struct {
	Title       string
	Description string
	Status      string
	Done        bool
}

func (item *Item) ToMap() g.Map {
	return g.Map{
		"title":       item.Title,
		"description": item.Description,
		"status":      item.Status,
		"done":        item.Done,
	}
}

var items = []*Item{
	{"認証機能の実装", "OAuth2 を用いたログイン機能", "完了", true},
	{"ダッシュボード UI", "管理画面のレイアウト作成", "進行中", false},
	{"API ドキュメント", "OpenAPI 仕様書の整備", "未着手", false},
}

func main() {
	tmpl := &g.Template{Lang: "ja"}

	if err := tmpl.Mount(
		gd.Head(
			gd.Title("プロジェクト進捗レポート"),
			gd.Meta(gp.Attr("charset", "UTF-8")),
		),
		gd.Body(
			gd.H1(gp.Value("プロジェクト進捗レポート")),
			summarySection(),
			gd.Hr(),
			taskListSection(),
			gd.Hr(),
			notesSection(),
		),
	); err != nil {
		fmt.Fprintf(os.Stderr, "Mount error: %v\n", err)
		os.Exit(1)
	}

	// String() でHTML文字列として出力
	fmt.Print(tmpl.String())

	// Bytes() の動作確認をログ出力
	fmt.Fprintf(os.Stderr, "\n--- String() output length: %d bytes ---\n", len(tmpl.String()))
	fmt.Fprintf(os.Stderr, "--- Bytes() output length: %d bytes ---\n", len(tmpl.Bytes()))
}

func summarySection() g.ComponentFunc {
	doneCount := 0
	for _, item := range items {
		if item.Done {
			doneCount++
		}
	}
	return gd.Section(
		gd.H2(gp.Value("概要")),
		gd.P(gp.Value(fmt.Sprintf("全 %d タスク中 %d 件完了", len(items), doneCount))),
	)
}

func taskListSection() g.ComponentFunc {
	list := make([]g.ConvertToMap, len(items))
	for i, item := range items {
		list[i] = item
	}
	return gd.Section(
		gd.H2(gp.Value("タスク一覧")),
		gd.Ol(
			ge.Each(list, func(item g.ConvertToMap) g.ComponentFunc {
				m := item.ToMap()
				title := m.Get("title").(string)
				desc := m.Get("description").(string)
				status := m.Get("status").(string)
				done := m.Get("done").(bool)
				return gd.Li(
					gd.H4(gp.Value(title)),
					gd.P(gp.Value(desc)),
					gd.P(
						gs.Style(g.StyleMap{"font-weight": "bold"}),
						gp.Value("ステータス: "+status),
					),
					ge.If(done,
						g.Skip(),
						// details/summary は dom/ にないため Tag() で直接生成
						g.Tag("details",
							g.Tag("summary", gp.Value("残作業")),
							gd.P(gp.Value("このタスクにはまだ作業が残っています。")),
						),
					),
				)
			}),
		),
	)
}

func notesSection() g.ComponentFunc {
	return gd.Section(
		gd.H2(gp.Value("備考")),
		gd.H5(gp.Value("次回マイルストーン")),
		gd.P(gp.Value("2026年3月末までにダッシュボード UI を完成予定。")),
		ge.Unless(false,
			gd.P(
				gs.Style(g.StyleMap{"color": "red"}),
				gp.Value("注意: 期限が近づいています。"),
			),
		),
	)
}
