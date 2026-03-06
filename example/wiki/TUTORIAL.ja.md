# Wiki チュートリアル

## 概要

gerbera を使って CRUD 機能を持つ Wiki アプリケーションを構築します。このチュートリアルでは以下の機能を学びます。

- `ExecuteTemplate` — テンプレートの直接実行
- `Form`, `Textarea`, `Button`, `Label` — フォーム要素
- `If` / `Each` — 条件分岐とリスト描画
- `ConvertToMap` — 構造体からテンプレートデータへの変換
- `Attr` — 任意の HTML 属性設定

## 前提条件

- Go 1.22 以上がインストールされていること
- このリポジトリをクローン済みであること

## ステップ 1: データモデルの定義

```go
type Page struct {
	Title string
	Body  []byte
}

func (p *Page) ToMap() g.Map {
	return g.Map{
		"title": p.Title,
		"body":  p.Body,
	}
}
```

`ConvertToMap` インターフェースを実装すると、`expr.Each()` で利用できるようになります。
`g.Map` は `map[string]any` のエイリアスで、`Get(key)` メソッドを持ちます。

## ステップ 2: HTTP ハンドラの構成

```go
func main() {
	http.HandleFunc("/view/", viewHandle)
	http.HandleFunc("/edit/", editHandle)
	http.HandleFunc("/save/", saveHandle)
	http.HandleFunc("/", listHandle)
	log.Fatal(http.ListenAndServe(":8880", nil))
}
```

`g.Handler()` は静的ページの配信に推奨されるアプローチで、`g.HandlerFunc()` はリクエストに依存する動的ページ向けです。複数ルートを持つアプリでは、標準の `http.HandleFunc` を直接使用します。

## ステップ 3: ビューの構築 — ExecuteTemplate

```go
func viewLayout(w http.ResponseWriter, page *Page) error {
	return g.ExecuteTemplate(w, "jp",
		gd.Head(
			gd.Title(page.Title),
			gc.BootstrapCSS(),
		),
		gd.Body(
			gp.Class("container"),
			gd.H1(gp.Value(page.Title)),
			gd.P(gp.Value(string(page.Body))),
		),
	)
}
```

`g.ExecuteTemplate(w, lang, children...)` は内部で `Template` を作成し、マウント・レンダリング・書き出しを一括で行います。

## ステップ 4: フォームの構築

```go
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
)
```

- `gp.Attr(key, value)` で任意の HTML 属性を設定できます
- `gp.Name(text)` は `gp.Attr("name", text)` のショートカットです
- フォーム要素も他と同様に `ComponentFunc` を返すため、自由に合成できます

## ステップ 5: 条件分岐とリスト表示

```go
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
)
```

- `ge.If(cond, trueComponent, falseComponent...)` — 条件が真なら第2引数、偽なら第3引数を描画
- `ge.Each(list, callback)` — `ConvertToMap` スライスを反復し、各要素に対してコールバックを呼びます
- コールバック内で `ToMap().Get(key)` を使ってフィールドにアクセスします

## 実行方法

```bash
go run example/wiki/wiki.go
```

ブラウザで http://localhost:8880 を開き、以下を試してみましょう:
1. トップページでページ一覧を確認
2. URL に `/edit/テスト` を入力して新規ページを作成
3. 保存後、`/view/テスト` で表示を確認

## 発展課題

1. ページの削除機能を追加してみましょう（`/delete/` ハンドラ）
2. `ge.Unless` を使って「ページが空でない場合」の表示を実装してみましょう
3. Markdown パーサーを組み込んで、本文を Markdown として描画してみましょう

## API リファレンス

| 関数 | 説明 |
|------|------|
| `g.ExecuteTemplate(w, lang, children...)` | テンプレートを作成・レンダリング・書き出し |
| `g.ConvertToMap` | `ToMap() Map` を持つインターフェース |
| `g.Map` | `map[string]any` 型。`Get(key)` メソッドを持つ |
| `gd.Form(children...)` | `<form>` 要素 |
| `gd.Textarea(children...)` | `<textarea>` 要素 |
| `gd.Button(children...)` | `<button>` 要素 |
| `gd.Label(children...)` | `<label>` 要素 |
| `gd.A(children...)` | `<a>` 要素 |
| `gd.Li(children...)` | `<li>` 要素 |
| `gp.Attr(key, value)` | 任意の HTML 属性を設定 |
| `gp.Name(text)` | `name` 属性を設定 |
| `ge.If(cond, true, false...)` | 条件分岐 |
| `ge.Each(list, callback)` | リスト反復 |
