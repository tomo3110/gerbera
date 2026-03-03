# Tab コンポーネント チュートリアル

## 概要

ページリロードなしでリアルタイムにパネルを切り替えるインタラクティブなタブ UI を構築します。このチュートリアルでは以下の機能を扱います:

- `gc.Tab` 構造体 — ラベル、コンテンツ、イベントバインディングを持つタブの定義
- `gc.Tabs()` — ARIA 属性付きのアクセシブルなタブ UI の描画
- `gc.TabsDefaultCSS()` — タブコンポーネント用の最小限デフォルト CSS
- `gl.Click` / `gl.ClickValue` — `ButtonAttrs` 経由でのイベント注入

`components.Tabs` 関数は `components` パッケージに属し、**`live/` パッケージに依存しません**。LiveView のイベントバインディングは `ButtonAttrs` フィールドを通じて呼び出し側が注入する設計のため、静的レンダリングと LiveView の両方で利用できます。

## 前提条件

- Go 1.22 以降がインストール済み
- このリポジトリをローカルにクローン済み

## ステップ 1: パッケージのインポート

```go
import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	g  "github.com/tomo3110/gerbera"
	gc "github.com/tomo3110/gerbera/components"
	gd "github.com/tomo3110/gerbera/dom"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
)
```

- `gc` (components) — `Tab`、`Tabs()`、`TabsDefaultCSS()` を提供
- `gl` (live) — `Click()`、`ClickValue()`、LiveView ランタイムを提供
- `gd` (dom) — HTML 要素ヘルパー
- `gp` (property) — 属性・値セッター

## ステップ 2: View 構造体の定義

```go
type TabDemoView struct {
	ActiveTab int
}
```

`ActiveTab` フィールドで現在選択中のタブを追跡します（0始まりのインデックス）。ブラウザセッションごとに独立したインスタンスが生成されます。

## ステップ 3: Mount の実装

```go
func (v *TabDemoView) Mount(params gl.Params) error {
	v.ActiveTab = 0
	return nil
}
```

`Mount` でアクティブタブを最初のタブに初期化します。

## ステップ 4: タブ定義の構築

```go
tabs := []gc.Tab{
	{
		Label:       "Home",
		Content:     homeContent(),
		ButtonAttrs: tabClickAttrs("0"),
	},
	{
		Label:       "Profile",
		Content:     profileContent(),
		ButtonAttrs: tabClickAttrs("1"),
	},
	{
		Label:       "Settings",
		Content:     settingsContent(),
		ButtonAttrs: tabClickAttrs("2"),
	},
}
```

各 `gc.Tab` は以下を持ちます:
- **Label** — タブボタンに表示するテキスト
- **Content** — パネル本体を描画する `ComponentFunc`
- **ButtonAttrs** — `<button>` 要素に注入する `ComponentFunc` のスライス

`tabClickAttrs` ヘルパーがクリックイベントをバインドします:

```go
func tabClickAttrs(index string) []g.ComponentFunc {
	return []g.ComponentFunc{
		gl.Click("switch-tab"),
		gl.ClickValue(index),
	}
}
```

これにより `gerbera-click="switch-tab"` と `gerbera-value="0"`（または `"1"`、`"2"`）がボタンに設定され、クリック時にクライアントが `{e: "switch-tab", p: {value: "0"}}` を送信します。

## ステップ 5: Tabs() によるレンダリング

```go
func (v *TabDemoView) Render() []g.ComponentFunc {
	return []g.ComponentFunc{
		gd.Head(
			gd.Title("Tab Demo"),
			gc.TabsDefaultCSS(),
		),
		gd.Body(
			gc.Tabs("demo", v.ActiveTab, tabs),
		),
	}
}
```

- `gc.TabsDefaultCSS()` — タブの最小限スタイルを含む `<style>` 要素を出力
- `gc.Tabs("demo", v.ActiveTab, tabs)` — ARIA 属性付きの完全なタブ構造を描画
- 第1引数 `"demo"` は `aria-controls` / `aria-labelledby` のリンクに使う ID プレフィックス

## ステップ 6: イベントハンドリング

```go
func (v *TabDemoView) HandleEvent(event string, payload gl.Payload) error {
	switch event {
	case "switch-tab":
		if idx, err := strconv.Atoi(payload["value"]); err == nil {
			v.ActiveTab = idx
		}
	}
	return nil
}
```

ユーザーがタブボタンをクリックすると、サーバーは `"switch-tab"` イベントを `payload["value"]` にタブインデックスとともに受け取ります。View が `ActiveTab` を更新すると、フレームワークが再レンダリングして DOM パッチを送信します。

## ステップ 7: サーバーの起動

```go
func main() {
	addr := flag.String("addr", ":8890", "listen address")
	debug := flag.Bool("debug", false, "enable debug panel")
	flag.Parse()

	var opts []gl.Option
	if *debug {
		opts = append(opts, gl.WithDebug())
	}

	http.Handle("/", gl.Handler(func(_ context.Context) gl.View { return &TabDemoView{} }, opts...))
	log.Printf("tabview running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
```

## 動作の仕組み

```
ブラウザ                                    Go サーバー
  |                                            |
  |  1. GET /                                  |
  |  <-- Tabs (active=0) の HTML + JS          |
  |                                            |
  |  2. WebSocket 接続                          |
  |                                            |
  |  3.「Profile」タブをクリック                    |
  |  --> {"e":"switch-tab","p":{"value":"1"}}  |
  |                                            |
  |      HandleEvent -> ActiveTab=1 -> Render  |
  |      -> Diff -> Patches                    |
  |                                            |
  |  4. <-- [aria-selected, hidden, class      |
  |          のパッチ（タブ＋パネル用）]              |
  |                                            |
  |  5. JS が DOM にパッチを適用                   |
```

非アクティブパネルの `hidden` 属性やボタンの `aria-selected` はサーバー側で切り替えられます。diff エンジンは変更された属性のみをパッチとして送信します。

## 実行方法

```bash
go run example/tabview/tabview.go
```

ブラウザで http://localhost:8890 を開き、タブボタンをクリックしてパネルを切り替えてください。

### デバッグモード

```bash
go run example/tabview/tabview.go -debug
```

## 静的利用（LiveView なし）

`Tabs` コンポーネントは LiveView なしのサーバーレンダリングページでも利用できます:

```go
func renderStaticTabs(w io.Writer) error {
	tabs := []gc.Tab{
		{Label: "タブ1", Content: gp.Value("コンテンツ1")},
		{Label: "タブ2", Content: gp.Value("コンテンツ2")},
	}
	return g.ExecuteTemplate(w, "ja",
		gd.Head(gc.TabsDefaultCSS()),
		gd.Body(gc.Tabs("static", 0, tabs)),
	)
}
```

静的モードでは、すべてのパネルが `hidden` 属性付きで HTML に出力されます。必要に応じてクライアント側の JavaScript で表示切り替えを実装できます。

## 発展課題

1. 動的コンテンツ（例: 現在時刻）を持つ4つ目のタブを追加する
2. `Mount` の `params` を使って URL でアクティブタブを永続化する
3. `Tabs()` に `wrapperAttrs` を渡してカスタム CSS を追加する
4. ネストされたタブコンポーネント（タブパネル内にタブ）を作成する

## API リファレンス

| 関数 / 型 | 説明 |
|---|---|
| `gc.Tab` | 構造体: `Label`、`Content`、`ButtonAttrs` |
| `gc.Tabs(id, activeIndex, tabs, wrapperAttrs...)` | アクセシブルなタブ UI を描画 |
| `gc.TabsDefaultCSS()` | デフォルト CSS を含む `<style>` 要素を返す |
| `gl.Click(event)` | クリックイベントをバインド |
| `gl.ClickValue(value)` | クリックイベントと共に送信する値を設定 |
