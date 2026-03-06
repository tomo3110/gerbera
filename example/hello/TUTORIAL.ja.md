# Hello チュートリアル

## 概要

gerbera を使った最小限の Web ページを作成します。このチュートリアルでは以下の機能を学びます。

- `Handler` — 簡易 HTTP サーバーの作成
- `BootstrapCSS` — `<head>` に Bootstrap 5 CSS/JS を追加
- `Body`, `H1`, `P` — 基本的な DOM 要素の構築
- `Class`, `Value` — 属性・テキストの設定

## 前提条件

- Go 1.22 以上がインストールされていること
- このリポジトリをクローン済みであること

## ステップ 1: パッケージのインポート

```go
import (
	"flag"
	"log"
	"net/http"

	g "github.com/tomo3110/gerbera"
	gc "github.com/tomo3110/gerbera/components"
	gd "github.com/tomo3110/gerbera/dom"
	gp "github.com/tomo3110/gerbera/property"
)
```

gerbera では慣例として以下のエイリアスを使います:
- `g` — コアパッケージ
- `gc` — 事前構築コンポーネント（Bootstrap CDN など）
- `gd` — DOM 要素（`H1`, `P`, `Body` など）
- `gp` — プロパティ（`Class`, `Value` など）

## ステップ 2: body 関数の定義

```go
func body() g.ComponentFunc {
	return gd.Body(
		gp.Class("container"),
		gd.H1(gp.Value("Gerbera Template Engine !")),
		gd.P(gp.Value("view html template Test")),
	)
}
```

`ComponentFunc` は `func(Node)` 型で、gerbera の基本単位です。`Node` は `Element` を抽象化するインターフェースです。すべての DOM ヘルパーは `ComponentFunc` を返します。

- `gd.Body(...)` — `<body>` 要素を作成し、引数の子要素を追加
- `gp.Class("container")` — `class="container"` 属性を設定
- `gp.Value("...")` — テキストコンテンツを設定

## ステップ 3: サーバーの起動

```go
func main() {
	addr := flag.String("addr", ":8800", "running address")
	flag.Parse()
	mux := http.NewServeMux()
	mux.Handle("GET /", g.Handler(
		gd.Head(
			gd.Title("Gerbera Template Engine !"),
			gc.BootstrapCSS(),
		),
		body(),
	))
	log.Fatal(http.ListenAndServe(*addr, mux))
}
```

- `g.Handler(children...)` は `http.Handler` を返し、指定されたコンポーネントを HTML ページとしてレンダリングします
- `gd.Head(children...)` は `<head>` 要素を作成し、子要素を追加します
- `gc.BootstrapCSS()` は Bootstrap 5 の CSS/JS CDN リンクを追加し、charset/viewport メタタグが存在しなければ自動追加します

## 実行方法

```bash
go run example/hello/hello.go
```

ブラウザで http://localhost:8800 を開いて確認します。

ポートを変更する場合:

```bash
go run example/hello/hello.go -addr :3000
```

## 発展課題

1. `gd.H2` や `gd.Div` を使って、ページにセクションを追加してみましょう
2. `gp.Attr("href", "...")` を使ったリンク（`gd.A`）を追加してみましょう
3. `gd.Ul` と `gd.Li` を使ってリスト表示を試してみましょう

## API リファレンス

| 関数 | 説明 |
|------|------|
| `g.Handler(children...)` | 指定コンポーネントをレンダリングする `http.Handler` を返す |
| `gc.BootstrapCSS()` | `<head>` に Bootstrap 5 CSS/JS を追加 |
| `gd.Head(children...)` | `<head>` 要素 |
| `gd.Title(text)` | `<title>` 要素 |
| `gd.Body(children...)` | `<body>` 要素 |
| `gd.H1(children...)` | `<h1>` 要素 |
| `gd.P(children...)` | `<p>` 要素 |
| `gp.Class(names...)` | `class` 属性を設定 |
| `gp.Value(v)` | テキストコンテンツを設定 |
