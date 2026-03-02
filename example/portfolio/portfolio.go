package main

import (
	"flag"
	"log"
	"net/http"

	g "github.com/tomo3110/gerbera"
	gc "github.com/tomo3110/gerbera/components"
	gd "github.com/tomo3110/gerbera/dom"
	gp "github.com/tomo3110/gerbera/property"
)

func main() {
	addr := flag.String("addr", ":8810", "running address")
	flag.Parse()
	mux := g.NewServeMux(
		gd.Head(
			gd.Title("ポートフォリオ - 田中太郎"),
			gc.MaterializeCSS(),
		),
		body(),
	)
	log.Printf("portfolio server running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func body() g.ComponentFunc {
	return gd.Body(
		navBar(),
		gd.Main(
			gp.Class("container"),
			aboutSection(),
			gd.Hr(),
			skillsSection(),
			gd.Hr(),
			projectsSection(),
		),
		footer(),
	)
}

func navBar() g.ComponentFunc {
	return gd.Header(
		gd.Nav(
			gp.Class("nav-wrapper"),
			gd.A(
				gp.Href("#"),
				gp.Class("brand-logo"),
				gp.Value("田中太郎"),
			),
			gd.Ul(
				gp.ID("nav-mobile"),
				gp.Class("right"),
				gd.Li(gd.A(gp.Href("#about"), gp.Value("自己紹介"))),
				gd.Li(gd.A(gp.Href("#skills"), gp.Value("スキル"))),
				gd.Li(gd.A(gp.Href("#projects"), gp.Value("プロジェクト"))),
			),
		),
	)
}

func aboutSection() g.ComponentFunc {
	return gd.Section(
		gp.ID("about"),
		gd.H2(gp.Value("自己紹介")),
		gd.Article(
			gd.Img("https://via.placeholder.com/150",
				gp.Attr("alt", "プロフィール画像"),
				gp.Class("circle", "responsive-img"),
			),
			gd.P(
				gp.Value("Go エンジニア。"),
				g.Literal("Web アプリケーション開発が専門です。"),
			),
			gd.P(
				gp.Value("好きなエディタは "),
				gd.Code(gp.Value("Vim")),
				g.Literal(" です。"),
			),
		),
	)
}

func skillsSection() g.ComponentFunc {
	return gd.Section(
		gp.ID("skills"),
		gd.H2(gp.Value("スキル")),
		gd.Ul(
			gp.Class("collection"),
			gd.Li(gp.Class("collection-item"), gd.B(gp.Value("Go")), g.Literal(" — Web / CLI 開発")),
			gd.Li(gp.Class("collection-item"), gd.B(gp.Value("TypeScript")), g.Literal(" — フロントエンド開発")),
			gd.Li(gp.Class("collection-item"), gd.B(gp.Value("Python")), g.Literal(" — データ分析")),
			gd.Li(gp.Class("collection-item"), gd.B(gp.Value("Docker")), g.Literal(" — コンテナ運用")),
		),
	)
}

func projectsSection() g.ComponentFunc {
	return gd.Section(
		gp.ID("projects"),
		gd.H2(gp.Value("プロジェクト")),
		gd.Article(
			gd.H3(gp.Value("Gerbera Template Engine")),
			gd.P(gp.Value("Go 向け HTML テンプレートエンジン。関数合成でビューを構築します。")),
			gd.Aside(
				gp.Class("card-panel"),
				gd.P(
					gp.Value("技術スタック: "),
					gd.Code(gp.Value("Go")),
					g.Literal(", "),
					gd.Code(gp.Value("net/http")),
				),
			),
		),
		gd.Article(
			gd.H3(gp.Value("CLI ツール集")),
			gd.P(gp.Value("業務効率化のためのコマンドラインツール。")),
		),
	)
}

func footer() g.ComponentFunc {
	return gd.Footer(
		gp.Class("page-footer"),
		gd.Div(
			gp.Class("container"),
			gd.P(gp.Value("© 2026 田中太郎")),
			gd.Address(
				gp.Value("連絡先: taro@example.com"),
				gd.Br(),
				g.Literal("東京都渋谷区"),
			),
		),
	)
}
