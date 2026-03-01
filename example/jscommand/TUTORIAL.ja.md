# JS Commands チュートリアル

## 概要

サーバーからクライアントサイドの JavaScript コマンドを実行するデモを作成します。このチュートリアルでは以下の機能を学びます。

- `gl.CommandQueue` — View にコマンドキューを埋め込み、JS コマンドを発行
- `JSCommander` インターフェース — `CommandQueue` 埋め込みで自動的に満たされる
- JS コマンド — `Focus`, `Blur`, `ScrollTo`, `AddClass`, `RemoveClass`, `ToggleClass`, `Show`, `Hide`, `Toggle`
- `ge.Each` — リストの反復レンダリング

このサンプルは **サーバーからクライアントへのコマンド実行** を示します。ユーザーがボタンをクリックすると、サーバーがイベントを処理し、DOM パッチと共に JS コマンドをクライアントに送信して実行させます。

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

	g  "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	ge "github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
	gs "github.com/tomo3110/gerbera/styles"
)
```

## ステップ 2: Each 用の ConvertToMap 型の定義

```go
type logEntry struct {
	Index   int
	Message string
}

func (e logEntry) ToMap() g.Map {
	return g.Map{"index": e.Index, "message": e.Message}
}
```

`ge.Each` は `g.ConvertToMap` を実装したアイテムを必要とします。`logEntry` 型はインデックス付きのログメッセージをラップし、`ToMap()` で `Each` コールバック内でフィールドを使えるようにします。

## ステップ 3: View 構造体の定義

```go
type JSCommandView struct {
	gl.CommandQueue
	Log []logEntry
}
```

重要な設計パターン: **`gl.CommandQueue` を View 構造体に直接埋め込み** ます。これにより全ての JS コマンドメソッド（`Focus`, `Blur`, `ScrollTo` 等）にアクセスでき、`DrainCommands()` メソッドにより `JSCommander` インターフェースを自動的に満たします。フレームワークは各レンダーサイクル後にこれを検出し、キューに溜まったコマンドをクライアントに送信します。

`Log` は実行したコマンドの履歴を表示用に保持します。

## ステップ 4: HandleEvent の実装

```go
func (v *JSCommandView) HandleEvent(event string, _ gl.Payload) error {
	switch event {
	case "do_focus":
		v.Focus("#target-input")
		v.appendLog("Focus → #target-input")
	case "do_blur":
		v.Blur("#target-input")
		v.appendLog("Blur → #target-input")
	case "scroll_top":
		v.ScrollTo("#scroll-box", "0", "0")
		v.appendLog("ScrollTo top → #scroll-box")
	case "scroll_bottom":
		v.ScrollTo("#scroll-box", "9999", "0")
		v.appendLog("ScrollTo bottom → #scroll-box")
	case "add_class":
		v.AddClass("#style-box", "highlighted")
		v.appendLog("AddClass 'highlighted' → #style-box")
	case "remove_class":
		v.RemoveClass("#style-box", "highlighted")
		v.appendLog("RemoveClass 'highlighted' → #style-box")
	case "toggle_class":
		v.ToggleClass("#style-box", "highlighted")
		v.appendLog("ToggleClass 'highlighted' → #style-box")
	case "do_show":
		v.Show("#toggle-box")
		v.appendLog("Show → #toggle-box")
	case "do_hide":
		v.Hide("#toggle-box")
		v.appendLog("Hide → #toggle-box")
	case "do_toggle":
		v.Toggle("#toggle-box")
		v.appendLog("Toggle → #toggle-box")
	case "clear_log":
		v.Log = nil
	}
	return nil
}
```

各イベントハンドラは `CommandQueue` のメソッドを 1 つ呼び出し、操作をログに記録します。コマンド一覧:

| メソッド | 効果 |
|----------|------|
| `Focus(selector)` | 対象要素にフォーカス |
| `Blur(selector)` | 対象要素からフォーカスを外す |
| `ScrollTo(selector, top, left)` | 指定位置にスクロール |
| `AddClass(selector, class)` | CSS クラスを追加 |
| `RemoveClass(selector, class)` | CSS クラスを削除 |
| `ToggleClass(selector, class)` | CSS クラスをトグル |
| `Show(selector)` | 非表示の要素を表示 |
| `Hide(selector)` | 要素を非表示 |
| `Toggle(selector)` | 要素の表示/非表示を切り替え |

セレクタはすべて標準の CSS セレクタ（例: `"#my-id"`, `".my-class"`）です。

## ステップ 5: Each を使った Render の実装

```go
func (v *JSCommandView) Render() []g.ComponentFunc {
	logItems := make([]g.ConvertToMap, len(v.Log))
	for i, e := range v.Log {
		logItems[i] = e
	}

	return []g.ComponentFunc{
		gd.Head(...),
		gd.Body(
			...
			// コマンドログ
			ge.If(len(v.Log) > 0,
				gd.Div(
					gp.Class("log"),
					ge.Each(logItems, func(item g.ConvertToMap) g.ComponentFunc {
						m := item.ToMap()
						idx := m.Get("index").(int)
						msg := m.Get("message").(string)
						return gd.Div(
							gp.Class("log-entry"),
							gp.Value(fmt.Sprintf("#%d  %s", idx, msg)),
						)
					}),
				),
				gd.P(gp.Value("No commands executed yet.")),
			),
			...
		),
	}
}
```

ポイント:
- `[]logEntry` スライスを `[]g.ConvertToMap` に変換してから `ge.Each` に渡す
- コールバック内で `item.ToMap()` を呼び、`.Get("key").(型)` で型付き値を取得
- `ge.If` でログがある場合はログを表示し、なければ空状態メッセージを表示

## ステップ 6: UI セクション

UI は 4 つのデモセクションに分かれ、それぞれ特定の要素を操作します:

```go
// Focus / Blur セクション
gd.Input(gp.Type("text"), gp.ID("target-input"), gp.Placeholder("Click Focus to focus this input"))
gd.Button(gl.Click("do_focus"), gp.Value("Focus"))
gd.Button(gl.Click("do_blur"), gp.Value("Blur"))

// ScrollTo セクション
gd.Div(gp.ID("scroll-box"), scrollContent())
gd.Button(gl.Click("scroll_top"), gp.Value("Scroll to Top"))
gd.Button(gl.Click("scroll_bottom"), gp.Value("Scroll to Bottom"))

// クラス操作セクション
gd.Div(gp.ID("style-box"), gp.Value("Target Box"))
gd.Button(gl.Click("add_class"), gp.Value("AddClass"))
gd.Button(gl.Click("remove_class"), gp.Value("RemoveClass"))
gd.Button(gl.Click("toggle_class"), gp.Value("ToggleClass"))

// 表示切替セクション
gd.Div(gp.ID("toggle-box"), gp.Value("Toggle Me"))
gd.Button(gl.Click("do_show"), gp.Value("Show"))
gd.Button(gl.Click("do_hide"), gp.Value("Hide"))
gd.Button(gl.Click("do_toggle"), gp.Value("Toggle"))
```

各セクションにはユニークな `id` を持つターゲット要素と、対応するイベントを発火するボタンがあります。

## 動作の仕組み

```
ブラウザ                                  Go サーバー
  |                                          |
  |  1. 「Focus」ボタンをクリック              |
  |  --> {"e":"do_focus","p":{}}             |
  |                                          |
  |      HandleEvent:                        |
  |        v.Focus("#target-input")          |
  |        v.appendLog(...)                  |
  |      -> Render -> Diff -> パッチ生成      |
  |      -> DrainCommands -> JS コマンド      |
  |                                          |
  |  <-- { patches: [...],                   |
  |        commands: [                        |
  |          {"cmd":"focus",                  |
  |           "target":"#target-input"}       |
  |        ]}                                |
  |                                          |
  |  2. JS が DOM パッチを適用（ログ更新）     |
  |  3. JS がコマンドを実行: focus(...)       |
```

フレームワークは DOM パッチと JS コマンドを同じ WebSocket メッセージで送信します。クライアントはパッチを先に適用し、その後コマンドを実行します。

## 実行方法

```bash
go run example/jscommand/jscommand.go
```

ブラウザで http://localhost:8875 を開きます。各セクションのボタンをクリックして効果を確認してください。画面下部のコマンドログが実行した全コマンドを記録します。

### デバッグモード

```bash
go run example/jscommand/jscommand.go -debug
```

## 発展課題

1. `SetAttribute` ボタンを追加して、入力欄のプレースホルダーテキストを変更してみましょう
2. `Dispatch` ボタンを追加して、要素にカスタム DOM イベントを発行してみましょう
3. 1 つのイベントで複数コマンドを組み合わせてみましょう（例: Focus + AddClass）
4. `ScrollIntoPct` デモを追加して、パーセンテージ位置へのスクロールを試してみましょう

## API リファレンス

| 関数 | 説明 |
|------|------|
| `gl.CommandQueue` | View 構造体に埋め込み、JS コマンドを有効化 |
| `q.Focus(sel)` | 要素にフォーカス |
| `q.Blur(sel)` | フォーカスを外す |
| `q.ScrollTo(sel, top, left)` | 指定位置にスクロール |
| `q.ScrollIntoPct(sel, pct)` | パーセンテージ位置にスクロール |
| `q.AddClass(sel, cls)` | CSS クラスを追加 |
| `q.RemoveClass(sel, cls)` | CSS クラスを削除 |
| `q.ToggleClass(sel, cls)` | CSS クラスをトグル |
| `q.Show(sel)` | 要素を表示 |
| `q.Hide(sel)` | 要素を非表示 |
| `q.Toggle(sel)` | 表示/非表示を切り替え |
| `q.SetAttribute(sel, key, val)` | HTML 属性を設定 |
| `q.RemoveAttribute(sel, key)` | HTML 属性を削除 |
| `q.SetProperty(sel, key, val)` | DOM プロパティを設定 |
| `q.Dispatch(sel, event)` | DOM イベントを発行 |
| `ge.Each(items, callback)` | リストを反復してレンダリング |
