# Wiki Live チュートリアル

## 概要

`live/` パッケージを使って、ページリロードなしで動作する SPA 風 Wiki を構築します。このチュートリアルでは以下の機能を学びます。

- `gl.View` インターフェース — 状態フィールドによる複数画面（一覧 / 閲覧 / 編集）の管理
- `gl.Click` / `gl.ClickValue` — データを伴うクリックイベントによるナビゲーション
- `gl.Input` — リアルタイムフォーム入力バインディング
- `expr.If` / `expr.Each` — LiveView 内での条件分岐とリスト描画
- ファイル I/O — Wiki ページの `.txt` ファイルへの永続化

[Counter サンプル](../counter/) と比較して、フォームやデータ永続化を含むマルチ画面アプリケーションの構築方法を示す、より実践的な Gerbera Live のデモンストレーションです。

## 前提条件

- Go 1.22 以上がインストールされていること
- このリポジトリをクローン済みであること

## ステップ 1: パッケージのインポート

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

標準の gerbera エイリアス（`g`, `gd`, `gp`）に加えて、以下をインポートします:
- `ge` — `expr` パッケージ: `If()` と `Each()`
- `live` — LiveView コア: `View` インターフェース、`Handler()`、イベントバインディング

## ステップ 2: ファイル I/O ヘルパー

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

ページは `data/` ディレクトリ内の `.txt` ファイルとして保存されます（起動時に自動作成）。[既存の Wiki サンプル](../wiki/) と同じファイルベースのアプローチで、データ層をシンプルに保っています。

## ステップ 3: View 構造体の定義

```go
type WikiView struct {
	Mode    string   // "list", "view", "edit"
	Title   string   // 現在のページタイトル
	Body    string   // ページ本文
	Pages   []string // ページ一覧
	EditBuf string   // 編集中のテキスト
	IsNew   bool     // 新規作成中かどうか
}
```

Counter サンプル（`Count` フィールド1つ）と異なり、WikiView は `Mode` フィールドで複数の画面モードを管理します。レンダリングロジックはこの値に基づいて分岐します。

`expr.Each` を使うために、`ConvertToMap` を実装するヘルパー構造体も必要です:

```go
type pageItem struct{ Title string }

func (p pageItem) ToMap() g.Map {
	return g.Map{"title": p.Title}
}
```

## ステップ 4: Mount の実装

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

セッション作成時に一覧モードで初期化し、ディスクからページ一覧を読み込みます。

## ステップ 5: HandleEvent の実装

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

ポイント:
- **画面遷移** — `nav-list`, `nav-view`, `nav-edit`, `nav-new` などのイベントが `Mode` フィールドを変更し、`Render()` が異なるコンポーネントツリーを出力します
- **ClickValue によるデータ送信** — `nav-view` は `payload["value"]` からクリックされたページタイトルを取得します（`gl.ClickValue` で設定）
- **入力バインディング** — `edit-input` と `title-input` は `payload["value"]` から状態フィールドを更新します（`gl.Input` で設定）
- **永続化** — `save` でディスクに書き込み、閲覧モードに遷移します

## ステップ 6: Render の実装

`Render` メソッドは共通レイアウトを返し、メインコンテンツは `Mode` に応じて切り替えるヘルパーに委譲します:

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

### 一覧モード — `expr.If` + `expr.Each` + `gl.ClickValue`

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

- `gl.Click("nav-view")` でクリックイベントをバインド
- `gl.ClickValue(title)` で `gerbera-value` 属性を設定し、ページタイトルをペイロードの `"value"` として送信
- `ge.If` でリストまたは空メッセージを表示
- `ge.Each` でページアイテムを反復

### 閲覧モード

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

### 編集モード — `gl.Input`

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

- `gl.Input("edit-input")` で各キーストロークをサーバーに送信し、`EditBuf` を更新
- `ge.If(v.IsNew, ...)` で新規作成時のみタイトル入力フィールドを表示
- `g.Skip()` は何も描画しない — `ge.If` の偽分岐として使用

## ステップ 7: サーバーの起動

```go
func main() {
	addr := flag.String("addr", ":8850", "listen address")
	flag.Parse()
	http.Handle("/", gl.Handler(func() gl.View { return &WikiView{} }))
	log.Printf("wiki_live running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
```

`gl.Handler` 1つで初期 HTML レンダリング、WebSocket アップグレード、イベント処理をすべて処理します。複数ルートは不要です。

## 動作の仕組み

```
ブラウザ                                  Go サーバー
  |                                          |
  |  1. GET /                                |
  |  <-- 初期 HTML（一覧モード）              |
  |       + gerbera.js                       |
  |                                          |
  |  2. WebSocket 接続                       |
  |  <-> 双方向接続確立                       |
  |                                          |
  |  3. ページタイトル「MyPage」をクリック      |
  |  --> {"e":"nav-view",                    |
  |       "p":{"value":"MyPage"}}            |
  |                                          |
  |      HandleEvent -> ファイル読み込み       |
  |      -> Mode="view" -> Render            |
  |      -> 旧ツリーと差分比較 -> パッチ生成   |
  |                                          |
  |  4. <-- DOM パッチ（一覧を                |
  |          閲覧コンテンツに置換）            |
  |                                          |
  |  5. 「編集」をクリック                     |
  |  --> {"e":"nav-edit","p":{}}             |
  |                                          |
  |  6. テキストエリアに入力                   |
  |  --> {"e":"edit-input",                  |
  |       "p":{"value":"新しいテキスト..."}}   |
  |                                          |
  |  7. 「保存」をクリック                     |
  |  --> {"e":"save","p":{}}                 |
  |      -> ファイル保存 -> Mode="view"       |
  |  <-- DOM パッチ                           |
```

## 既存 Wiki との比較

| 観点 | `example/wiki/` | `example/wiki_live/` |
|------|-----------------|---------------------|
| レンダリング | サーバーサイド（フルページ） | LiveView（WebSocket パッチ） |
| ナビゲーション | ページリロード（`<a href>`） | リロードなし（`gl.Click`） |
| ルート数 | 4（`/`, `/view/`, `/edit/`, `/save/`） | 1（`/`） |
| フォーム処理 | POST + リダイレクト | `gl.Input` + `gl.Click` |
| ユーザー体験 | 従来型 Web アプリ | SPA 風 |

## 実行方法

```bash
go run example/wiki_live/wiki_gl.go
```

ブラウザで http://localhost:8850 を開き、以下を試してみましょう:
1. 「新規作成」をクリックして新しいページを作成
2. タイトルと本文を入力し、「保存」をクリック
3. 「ページ一覧」をクリックして一覧に戻る
4. ページタイトルをクリックして閲覧
5. 「編集」をクリックしてページを編集

すべての操作がページリロードなしで行われます。

ポートを変更する場合:

```bash
go run example/wiki_live/wiki_gl.go -addr :3000
```

## 発展課題

1. 閲覧画面にページの削除ボタンを追加してみましょう
2. 一覧画面に `gl.Input` を使った検索・フィルター入力を追加してみましょう
3. 保存前に確認メッセージを表示してみましょう（`Message` 状態フィールドを追加）
4. Markdown パーサーを組み込んで、本文を Markdown として描画してみましょう

## API リファレンス

| 関数 | 説明 |
|------|------|
| `gl.Handler(factory, opts...)` | LiveView 用の `http.Handler` を返す |
| `gl.WithDebug()` | ブラウザ DevPanel とサーバーサイド構造化ログを有効化 |
| `gl.Click(event)` | クリックイベントをバインド |
| `gl.ClickValue(value)` | クリックイベントで送信する値を設定 |
| `gl.Input(event)` | 入力イベントをバインド（ペイロードに `value` を含む） |
| `gl.View` | インターフェース: `Mount`, `Render`, `HandleEvent` |
| `gl.Params` | `map[string]string` — URL パラメータ |
| `gl.Payload` | `map[string]string` — イベントデータ |
| `ge.If(cond, true, false...)` | 条件分岐 |
| `ge.Each(list, callback)` | `ConvertToMap` スライスのリスト反復 |
| `g.Skip()` | 何も描画しない（`If` の偽分岐に便利） |
| `g.ConvertToMap` | `ToMap() Map` を持つインターフェース（`Each` 用） |
