# ポートフォリオ チュートリアル

## 概要

Materialize CSS を使ったポートフォリオサイトを構築します。このチュートリアルでは以下の機能を学びます。

- `MaterilalizecssCDNHead` — Materialize CSS 付きの `<head>` 生成
- セマンティック HTML5: `Header`, `Footer`, `Nav`, `Main`, `Section`, `Article`, `Aside`
- `Img` — 画像要素
- `ID` — `id` 属性の設定
- `Literal` — 生テキストの埋め込み
- `B`, `Code`, `Hr`, `Br`, `Address` — インライン/ブロック要素

## 前提条件

- Go 1.22 以上がインストールされていること
- このリポジトリをクローン済みであること

## ステップ 1: Materialize CSS の利用

```go
mux := g.NewServeMux(
	gc.MaterilalizecssCDNHead("ポートフォリオ - 田中太郎"),
	body(),
)
```

`gc.MaterilalizecssCDNHead(title)` は Materialize CSS の CDN リンクを含む `<head>` を生成します。Bootstrap の代わりとして利用できます。

## ステップ 2: セマンティック HTML5 構造

```go
func body() g.ComponentFunc {
	return gd.Body(
		navBar(),      // Header + Nav
		gd.Main(       // Main
			aboutSection(),    // Section + Article
			skillsSection(),   // Section
			projectsSection(), // Section + Article + Aside
		),
		footer(),      // Footer
	)
}
```

gerbera は `Header`, `Footer`, `Nav`, `Main`, `Section`, `Article`, `Aside` といった HTML5 セマンティック要素をすべてサポートしています。各関数は `gd.Div` と同じインターフェースで使えます。

## ステップ 3: ナビゲーションバー

```go
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
			),
		),
	)
}
```

- `gp.ID("nav-mobile")` は `id="nav-mobile"` を設定します
- ページ内リンクには `#セクション名` を使い、対応するセクションに `gp.ID()` を付けます

## ステップ 4: 画像と Literal の使用

```go
gd.Img("https://via.placeholder.com/150",
	gp.Attr("alt", "プロフィール画像"),
	gp.Class("circle", "responsive-img"),
),
gd.P(
	gp.Value("Go エンジニア。"),
	g.Literal("Web アプリケーション開発が専門です。"),
),
```

- `gd.Img(src, children...)` — 第1引数に画像 URL を指定。`<img>` はvoid要素なので閉じタグは出力されません
- `g.Literal(lines...)` — 生テキストを要素の `Value` に直接追加します。`gp.Value` と組み合わせてテキストを連結できます

## ステップ 5: インライン要素

```go
gd.Li(
	gp.Class("collection-item"),
	gd.B(gp.Value("Go")),
	g.Literal(" — Web / CLI 開発"),
),
gd.P(
	gp.Value("好きなエディタは "),
	gd.Code(gp.Value("Vim")),
	g.Literal(" です。"),
),
```

- `gd.B(...)` — `<b>` 太字要素
- `gd.Code(...)` — `<code>` コード要素
- `gd.Hr()` — `<hr>` 水平線（void要素）
- `gd.Br()` — `<br>` 改行（void要素）

## ステップ 6: フッターと Address

```go
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
```

- `gd.Address(...)` — `<address>` 要素。連絡先情報に使います
- `gd.Br()` は void 要素のため、子要素なしで使用できます

## 実行方法

```bash
go run example/portfolio/portfolio.go
```

ブラウザで http://localhost:8810 を開いて確認します。

## 発展課題

1. `gd.Section` を追加して「経歴」セクションを作ってみましょう
2. `gd.Img` で実際のプロフィール画像を設定してみましょう
3. `gp.Attr("target", "_blank")` を使って外部リンクを追加してみましょう

## API リファレンス

| 関数 | 説明 |
|------|------|
| `gc.MaterilalizecssCDNHead(title)` | Materialize CSS CDN 付き `<head>` |
| `gd.Header(children...)` | `<header>` 要素 |
| `gd.Footer(children...)` | `<footer>` 要素 |
| `gd.Nav(children...)` | `<nav>` 要素 |
| `gd.Main(children...)` | `<main>` 要素 |
| `gd.Section(children...)` | `<section>` 要素 |
| `gd.Article(children...)` | `<article>` 要素 |
| `gd.Aside(children...)` | `<aside>` 要素 |
| `gd.Img(src, children...)` | `<img>` 要素 |
| `gd.B(children...)` | `<b>` 要素 |
| `gd.Code(children...)` | `<code>` 要素 |
| `gd.Hr(children...)` | `<hr>` 要素 |
| `gd.Br(children...)` | `<br>` 要素 |
| `gd.Address(children...)` | `<address>` 要素 |
| `gp.ID(text)` | `id` 属性を設定 |
| `g.Literal(lines...)` | 生テキストを Value に追加 |
