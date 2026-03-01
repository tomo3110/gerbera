# Gerbera

関数合成によって HTML を構築する Go テンプレートエンジンです。

従来のテンプレートファイルの代わりに、`ComponentFunc` 関数（`func(*Element) error`）をプログラム的に合成して HTML を組み立てます。型安全性と Go の全機能をテンプレート構築時に活用できます。

[English README](README.md)

## 特徴

- **関数合成** — テンプレート文字列の代わりに、小さく再利用可能な関数を合成して HTML を構築
- **型安全** — コンパイル時の保証により、実行時のテンプレートパースエラーが発生しない
- **LiveView** — Phoenix LiveView 風の WebSocket によるリアルタイム DOM 更新（`live/` パッケージ）
- **DOM 差分** — 要素ツリーの差分比較による自動的な最小限の DOM パッチ適用（`diff/` パッケージ）
- **XSS 対策** — 属性値は `html.EscapeString` により自動エスケープ
- **コア部分は外部依存なし** — `github.com/gorilla/websocket` は `live/` パッケージのみが使用

## クイックスタート

### インストール

```bash
go get github.com/tomo3110/gerbera
```

### Hello World

```go
package main

import (
	"log"
	"net/http"

	g  "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	gp "github.com/tomo3110/gerbera/property"
)

func main() {
	mux := g.NewServeMux(
		gd.Head(gd.Title("Hello Gerbera")),
		gd.Body(
			gd.H1(gp.Value("Hello, World!")),
			gd.P(gp.Value("Built with Gerbera")),
		),
	)
	log.Fatal(http.ListenAndServe(":8800", mux))
}
```

### LiveView カウンター

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	g  "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	gp "github.com/tomo3110/gerbera/property"
	gl "github.com/tomo3110/gerbera/live"
)

type CounterView struct{ Count int }

func (v *CounterView) Mount(_ gl.Params) error { v.Count = 0; return nil }

func (v *CounterView) Render() []g.ComponentFunc {
	return []g.ComponentFunc{
		gd.Head(gd.Title("カウンター")),
		gd.Body(
			gd.H1(gp.Value(fmt.Sprintf("カウント: %d", v.Count))),
			gd.Button(gl.Click("dec"), gp.Value("-")),
			gd.Button(gl.Click("inc"), gp.Value("+")),
		),
	}
}

func (v *CounterView) HandleEvent(event string, _ gl.Payload) error {
	switch event {
	case "inc":
		v.Count++
	case "dec":
		v.Count--
	}
	return nil
}

func main() {
	http.Handle("/", gl.Handler(func() gl.View { return &CounterView{} }))
	log.Fatal(http.ListenAndServe(":8840", nil))
}
```

http://localhost:8840 を開くと、ページリロードなしにカウンターがリアルタイムで更新されます。

## アーキテクチャ

### レンダリングパイプライン

```
ComponentFunc 合成 → Template.Mount() → Parse() → Render() → HTML 出力
```

### LiveView パイプライン

```
ブラウザクリック → WebSocket → HandleEvent() → Render() → Diff() → DOM パッチ → ブラウザ
```

### パッケージ構成

```
gerbera/
├── common.go          # ComponentFunc 型、Tag()、Literal()、ExecuteTemplate()
├── element.go         # Element 構造体 (TagName, ClassNames, Attr, Children, Value)
├── template.go        # Template: マウント/レンダリング制御、io.Reader
├── render.go          # 再帰的 HTML レンダリング
├── server.go          # NewServeMux() HTTP ヘルパー
├── map.go             # ConvertToMap インターフェース、Map 型
│
├── dom/               # HTML 要素ラッパー (H1, Div, Body, Span, Details など)
├── property/          # 属性セッター (Class, Attr, ID, Href, Type, For など)
├── expr/              # 制御フロー: If()、Unless()、Each()
├── styles/            # Style(), CSS(), ScopedCSS() による CSS 処理
├── components/        # ビルド済みヘッドコンポーネント (Bootstrap CDN など)
├── diff/              # 要素ツリー差分比較とフラグメントレンダリング
├── live/              # LiveView: View インターフェース、WebSocket ハンドラ、クライアント JS
│
└── example/           # サンプルアプリケーション
    ├── hello/         # 最小限の Hello World
    ├── portfolio/     # ポートフォリオページ
    ├── dashboard/     # テーブル付きダッシュボード
    ├── survey/        # アンケートフォーム
    ├── report/        # レポート出力 (標準出力)
    ├── wiki/          # マルチルート Wiki (サーバーサイドレンダリング)
    ├── counter/       # LiveView カウンター
    ├── wiki_live/     # LiveView Wiki (SPA 風)
    └── mdviewer/      # LiveView Markdown ビューアー/エディタ
```

### コアコンセプト

**ComponentFunc** — 基本的な構成要素。すべての DOM ヘルパー、属性セッター、制御フロー関数が `ComponentFunc` を返します:

```go
type ComponentFunc func(*Element) error
```

**Tag()** — 汎用要素ファクトリ。`dom/` の関数はすべて `Tag()` の薄いラッパーです:

```go
gd.Div(children...)  // g.Tag("div", children...) と同等
```

**関数合成** — コンポーネントは可変長引数 `children ...ComponentFunc` で合成します:

```go
gd.Body(
    gp.Class("container"),           // クラス属性を設定
    gd.H1(gp.Value("タイトル")),      // テキスト付き子要素
    ge.If(showDetails,               // 条件付きレンダリング
        gd.P(gp.Value("詳細")),
    ),
)
```

### LiveView（`live/` パッケージ）

Gerbera Live は Phoenix LiveView 風のリアルタイム更新を提供します。`gl.View` インターフェースを実装してください:

| メソッド | 説明 |
|---------|------|
| `Mount(params)` | セッション作成時に一度呼ばれ、状態を初期化 |
| `Render()` | 現在の状態に対応するコンポーネントツリーを返す |
| `HandleEvent(event, payload)` | WebSocket 経由のブラウザイベントを処理 |

イベントバインディング:

| 関数 | 説明 |
|------|------|
| `gl.Click(event)` | クリックイベントをバインド |
| `gl.ClickValue(value)` | クリックイベントで送信する値を設定 |
| `gl.Input(event)` | 入力イベントをバインド（ペイロードに `value` を含む） |
| `gl.Change(event)` | 変更イベントをバインド |
| `gl.Submit(event)` | フォーム送信イベントをバインド |
| `gl.Focus(event)` / `gl.Blur(event)` | フォーカス/ブラーイベント |
| `gl.Keydown(event)` | キーダウンイベントをバインド |
| `gl.Key(key)` | キー名でキーダウンをフィルタ |
| `gl.Scroll(event)` | スクロールイベントをバインド（ペイロードにスクロール位置を含む） |
| `gl.Throttle(ms)` | スクロールイベントのスロットル間隔をミリ秒で設定（デフォルト 100ms） |
| `gl.Debounce(ms)` | イベントをデバウンス — 最後のイベントから指定時間待機後に送信 |
| `gl.DblClick(event)` | ダブルクリックイベントをバインド |
| `gl.MouseEnter(event)` / `gl.MouseLeave(event)` | マウスエンター/リーブイベント |
| `gl.TouchStart(event)` / `gl.TouchEnd(event)` | タッチイベント |

### TickerView

`TickerView` インターフェースを実装すると、定期的なサーバーサイド更新が可能です（例: ファイルポーリング）:

```go
type MyView struct { /* ... */ }

func (v *MyView) TickInterval() time.Duration { return 2 * time.Second }
func (v *MyView) HandleTick() error {
    // 外部状態をチェック、フィールドを更新 — ビューは自動的に再レンダリングされます
    return nil
}
```

### CommandQueue（サーバー→クライアントコマンド）

`gl.CommandQueue` を View に埋め込むと、JavaScript を書かずに JS コマンドを発行できます:

```go
type MyView struct {
    gl.CommandQueue
    // ...
}

func (v *MyView) HandleEvent(event string, payload gl.Payload) error {
    v.ScrollTo(".target", "100", "0")   // スクロール位置を移動
    v.Focus("#input")                    // 要素にフォーカス
    v.AddClass(".item", "active")        // CSS クラスを追加
    v.Hide(".modal")                     // 要素を非表示
    return nil
}
```

利用可能なコマンド: `ScrollTo`, `ScrollIntoPct`, `Focus`, `Blur`, `SetAttribute`, `RemoveAttribute`, `AddClass`, `RemoveClass`, `ToggleClass`, `SetProperty`, `Dispatch`, `Show`, `Hide`, `Toggle`

### プロパティヘルパー

よく使う HTML 属性には `property/` パッケージに専用ヘルパーがあります:

| 関数 | 同等の記述 |
|------|-----------|
| `gp.Href(url)` | `gp.Attr("href", url)` |
| `gp.For(id)` | `gp.Attr("for", id)` |
| `gp.Type(t)` | `gp.Attr("type", t)` |
| `gp.Placeholder(text)` | `gp.Attr("placeholder", text)` |
| `gp.Required(true)` | `gp.Attr("required", "required")` |
| `gp.Disabled(true)` | `gp.Attr("disabled", "disabled")` |
| `gp.Src(url)` | `gp.Attr("src", url)` |
| `gp.ClassIf(cond, name)` | 条件付き CSS クラス |
| `gp.ClassMap(map)` | 複数の条件付き CSS クラス |

### スタイルヘルパー

`styles/` パッケージには追加の CSS ユーティリティがあります:

| 関数 | 説明 |
|------|------|
| `gs.Style(styleMap)` | マップからインライン CSS |
| `gs.CSS(cssText)` | 生 CSS を含む `<style>` 要素 |
| `gs.ScopedCSS(scope, rules)` | セレクタにスコープされた CSS ルール |
| `gs.StyleIf(cond, styleMap)` | 条件付きインラインスタイル |

### 追加 DOM 要素

`dom/` パッケージは基本要素に加えて多数の HTML 要素ヘルパーを提供します:

`Span`, `Details`, `Summary`, `Dialog`, `Progress`, `Meter`, `Time`, `Mark`, `Figure`, `Figcaption`, `Video`, `Audio`, `Source`, `Canvas`, `Pre`, `Blockquote`, `Em`, `Strong`, `Small`, `Dl`, `Dt`, `Dd`, `Fieldset`, `Legend`, `Iframe` など

ハンドラオプション:

| 関数 | 説明 |
|------|------|
| `gl.WithLang(lang)` | HTML の `lang` 属性を設定（デフォルト `"ja"`） |
| `gl.WithSessionTTL(d)` | セッションタイムアウトを設定（デフォルト 5 分） |
| `gl.WithDebug()` | ブラウザ DevPanel とサーバーサイド構造化ログを有効化 |

### デバッグモード

ハンドラに `gl.WithDebug()` を渡すことでデバッグモードを有効化できます:

```go
http.Handle("/", gl.Handler(factory, gl.WithDebug()))
```

有効化すると:
- **ブラウザ DevPanel** — フローティングオーバーレイ（`Ctrl+Shift+D` または "G" ボタンで切替）に 4 つのタブ:
  - **Events** — 送信イベントの名前・ペイロード・タイムスタンプをリアルタイム表示
  - **Patches** — 受信 DOM パッチの件数とサーバー処理時間
  - **State** — View 構造体の現在の状態を JSON 表示（各イベント後に更新）
  - **Session** — セッション ID・TTL・接続状態
- **サーバーサイドログ** — `log/slog` によるイベント処理、パッチ生成、セッションライフサイクルの構造化ログ

`WithDebug()` 未設定時はオーバーヘッドゼロ — デバッグ JS の注入もタイミング計測も行われません。

## サンプル

各サンプルにはバイリンガルチュートリアル（`TUTORIAL.md` / `TUTORIAL.ja.md`）が含まれています。

| サンプル | 説明 | ポート | コマンド |
|---------|------|--------|---------|
| [hello](example/hello/) | 最小限の Hello World | :8800 | `go run example/hello/hello.go` |
| [portfolio](example/portfolio/) | ポートフォリオページ | :8810 | `go run example/portfolio/portfolio.go` |
| [dashboard](example/dashboard/) | テーブル付きダッシュボード | :8820 | `go run example/dashboard/dashboard.go` |
| [survey](example/survey/) | アンケートフォーム | :8830 | `go run example/survey/survey.go` |
| [report](example/report/) | レポート出力 | 標準出力 | `go run example/report/report.go` |
| [wiki](example/wiki/) | マルチルート Wiki（SSR） | :8880 | `go run example/wiki/wiki.go` |
| [counter](example/counter/) | LiveView カウンター | :8840 | `go run example/counter/counter.go` |
| [wiki_live](example/wiki_live/) | LiveView Wiki（SPA） | :8850 | `go run example/wiki_live/wiki_gl.go` |
| [mdviewer](example/mdviewer/) | LiveView Markdown ビューアー | :8860 | `cd example/mdviewer && go run .` |

## 開発

```bash
go build ./...    # 全パッケージをビルド
go test ./...     # 全テストを実行
go test -v ./...  # 詳細なテスト出力
```

Go 1.22 以上が必要です。

## ライセンス

MIT
