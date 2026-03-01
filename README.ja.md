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
	"github.com/tomo3110/gerbera/live"
)

type CounterView struct{ Count int }

func (v *CounterView) Mount(_ live.Params) error { v.Count = 0; return nil }

func (v *CounterView) Render() []g.ComponentFunc {
	return []g.ComponentFunc{
		gd.Head(gd.Title("カウンター")),
		gd.Body(
			gd.H1(gp.Value(fmt.Sprintf("カウント: %d", v.Count))),
			gd.Button(live.Click("dec"), gp.Value("-")),
			gd.Button(live.Click("inc"), gp.Value("+")),
		),
	}
}

func (v *CounterView) HandleEvent(event string, _ live.Payload) error {
	switch event {
	case "inc":
		v.Count++
	case "dec":
		v.Count--
	}
	return nil
}

func main() {
	http.Handle("/", live.Handler(func() live.View { return &CounterView{} }))
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
├── dom/               # HTML 要素ラッパー (H1, Div, Body, Form など)
├── property/          # 属性セッター (Class, Attr, ID, Name, Value)
├── expr/              # 制御フロー: If()、Unless()、Each()
├── styles/            # Style(): マップからインライン CSS を生成
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
    └── wiki_live/     # LiveView Wiki (SPA 風)
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

Gerbera Live は Phoenix LiveView 風のリアルタイム更新を提供します。`live.View` インターフェースを実装してください:

| メソッド | 説明 |
|---------|------|
| `Mount(params)` | セッション作成時に一度呼ばれ、状態を初期化 |
| `Render()` | 現在の状態に対応するコンポーネントツリーを返す |
| `HandleEvent(event, payload)` | WebSocket 経由のブラウザイベントを処理 |

イベントバインディング:

| 関数 | 説明 |
|------|------|
| `live.Click(event)` | クリックイベントをバインド |
| `live.ClickValue(value)` | クリックイベントで送信する値を設定 |
| `live.Input(event)` | 入力イベントをバインド（ペイロードに `value` を含む） |
| `live.Change(event)` | 変更イベントをバインド |
| `live.Submit(event)` | フォーム送信イベントをバインド |
| `live.Focus(event)` / `live.Blur(event)` | フォーカス/ブラーイベント |
| `live.Keydown(event)` | キーダウンイベントをバインド |
| `live.Key(key)` | キー名でキーダウンをフィルタ |

ハンドラオプション:

| 関数 | 説明 |
|------|------|
| `live.WithLang(lang)` | HTML の `lang` 属性を設定（デフォルト `"ja"`） |
| `live.WithSessionTTL(d)` | セッションタイムアウトを設定（デフォルト 5 分） |

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
| [wiki_live](example/wiki_live/) | LiveView Wiki（SPA） | :8850 | `go run example/wiki_live/wiki_live.go` |

## 開発

```bash
go build ./...    # 全パッケージをビルド
go test ./...     # 全テストを実行
go test -v ./...  # 詳細なテスト出力
```

Go 1.22 以上が必要です。

## ライセンス

MIT
