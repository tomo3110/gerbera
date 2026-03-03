# Counter チュートリアル

## 概要

ページリロードなしにリアルタイムで更新されるインタラクティブなカウンターを作成します。このチュートリアルでは以下の機能を学びます。

- `gl.View` インターフェース — ステートフルな LiveView コンポーネントの定義
- `gl.Handler` — HTTP と WebSocket による LiveView の配信
- `gl.Click` — ブラウザイベントをサーバーサイドハンドラにバインド
- `diff.Diff` — ツリー差分による自動 DOM パッチ（内部で使用）

このサンプルは **Gerbera Live** のアーキテクチャを示します。ユーザーイベントは WebSocket 経由で Go サーバーに送信され、サーバーが状態を更新・再レンダーし、変更された部分のみをパッチとしてブラウザに返します。

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
	gp "github.com/tomo3110/gerbera/property"
	gl "github.com/tomo3110/gerbera/live"
)
```

標準の gerbera エイリアス（`g`, `gd`, `gp`）に加えて、LiveView ランタイムを提供する `live` パッケージをインポートします:
- `live` — LiveView コア: `View` インターフェース、`Handler()`、イベントバインディング

## ステップ 2: View 構造体の定義

```go
type CounterView struct{ Count int }
```

LiveView は `gl.View` インターフェースを実装する任意の構造体です。構造体のフィールドがコンポーネントの状態を保持します。ブラウザセッションごとに独立したインスタンスが生成されるため、ユーザー間で状態は共有されません。

## ステップ 3: Mount の実装

```go
func (v *CounterView) Mount(params gl.Params) error {
	v.Count = 0
	return nil
}
```

`Mount` は新しいセッション作成時（最初の HTTP リクエスト時）に一度だけ呼ばれます。状態の初期化に使用します。`params` には URL クエリパラメータが `map[string]string` として渡されます。

## ステップ 4: Render の実装

```go
func (v *CounterView) Render() []g.ComponentFunc {
	return []g.ComponentFunc{
		gd.Head(gd.Title("カウンター")),
		gd.Body(
			gp.Class("container"),
			gd.H1(gp.Value(fmt.Sprintf("カウント: %d", v.Count))),
			gd.Div(
				gd.Button(gl.Click("dec"), gp.Value("-")),
				gd.Button(gl.Click("inc"), gp.Value("+")),
			),
		),
	}
}
```

`Render` は現在の状態に対応するコンポーネントツリーを返します。イベント発生ごとに呼ばれ、新しいツリーが前回のツリーと差分比較されます。

ポイント:
- `gl.Click("inc")` は `gerbera-click="inc"` 属性を設定し、ボタンクリック時にクライアント JS が `"inc"` イベントを送信するようにします
- 戻り値は `ComponentFunc` のスライスで、`<html>` ルート要素の子になります
- サーバーサイドレンダリングと同じ gerbera の DOM ヘルパー（`gd.Body`, `gd.H1` など）をそのまま使えます

## ステップ 5: HandleEvent の実装

```go
func (v *CounterView) HandleEvent(event string, payload gl.Payload) error {
	switch event {
	case "inc":
		v.Count++
	case "dec":
		v.Count--
	}
	return nil
}
```

`HandleEvent` はブラウザから送信されたイベントを受け取ります。`event` 文字列は `gl.Click()` に渡した値と一致します。`payload` は `map[string]string` で、追加データ（`gl.Input` イベントの入力値など）を含みます。

`HandleEvent` が返った後、フレームワークは自動的に `Render()` を呼び、新旧ツリーを差分比較し、パッチのみをブラウザに送信します。

## ステップ 6: サーバーの起動

```go
func main() {
	addr := flag.String("addr", ":8840", "listen address")
	debug := flag.Bool("debug", false, "enable debug panel")
	flag.Parse()

	var opts []gl.Option
	if *debug {
		opts = append(opts, gl.WithDebug())
	}

	http.Handle("/", gl.Handler(func(_ *http.Request) gl.View { return &CounterView{} }, opts...))
	log.Printf("counter running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
```

- `gl.Handler(factory)` は初期 HTML レンダリングと WebSocket 接続の両方を処理する `http.Handler` を返します
- ファクトリ関数 `func(*http.Request) gl.View` はセッションごとに呼ばれ、新しい `View` インスタンスを生成します — リクエストからコンテキスト値（認証済みユーザーなど）にアクセス可能
- オプション: `gl.WithLang("en")` で HTML の lang 属性を設定、`gl.WithSessionTTL(10*time.Minute)` でセッションタイムアウトを調整
- `gl.WithDebug()` でブラウザ DevPanel とサーバーサイド構造化ログを有効化（後述）

## 動作の仕組み

```
ブラウザ                                  Go サーバー
  |                                          |
  |  1. GET /                                |
  |  <-- 初期 HTML + gerbera.js              |
  |                                          |
  |  2. WebSocket 接続                       |
  |  <-> 双方向接続確立                       |
  |                                          |
  |  3. 「+」ボタンをクリック                  |
  |  --> {"e":"inc","p":{}}                  |
  |                                          |
  |      HandleEvent -> Count++ -> Render    |
  |      -> 旧ツリーと差分比較 -> パッチ生成   |
  |                                          |
  |  4. <-- [{"op":"text",                   |
  |           "path":[1,0],                  |
  |           "val":"カウント: 1"}]            |
  |                                          |
  |  5. JS がパッチを DOM に適用              |
```

## 実行方法

```bash
go run example/counter/counter.go
```

ブラウザで http://localhost:8840 を開きます。**+** と **-** ボタンをクリックすると、ページリロードなしにカウントがリアルタイムで更新されます。

ポートを変更する場合:

```bash
go run example/counter/counter.go -addr :3000
```

ブラウザの開発者ツールの Network タブで「WS」をフィルタすると、WebSocket メッセージの送受信を確認できます。

### デバッグモード

`-debug` フラグを付けて起動すると、組み込みのデバッグパネルが有効になります:

```bash
go run example/counter/counter.go -debug
```

デバッグモードを有効にすると:
- ブラウザ画面の右下にフローティングの **"G" ボタン** が表示されます
- ボタンをクリックするか `Ctrl+Shift+D` を押すと **Gerbera Debug Panel** が開閉します
- パネルには 4 つのタブがあります:
  - **Events** — サーバーに送信されたイベントのリアルタイムログ（イベント名・ペイロード・タイムスタンプ）
  - **Patches** — サーバーから受信した DOM パッチ（パッチ数・サーバー処理時間 ms）
  - **State** — `CounterView` の現在の状態を JSON 表示（各イベント後に更新）
  - **Session** — セッション ID・TTL・WebSocket 接続状態
- サーバーのターミナルにも `log/slog` による構造化ログが出力されます

## 発展課題

1. カウントを 0 に戻す「リセット」ボタンを追加してみましょう
2. `gl.Input("set")` を使って、ユーザーが数値を直接入力できる `<input>` フィールドを追加してみましょう
3. `ge.If` を使って、カウントが負の場合にテキストの色を変えてみましょう
4. 別の View（例: To-Do リスト）を持つ 2 つ目のルートを追加してみましょう

## API リファレンス

| 関数 | 説明 |
|------|------|
| `gl.Handler(factory, opts...)` | LiveView 用の `http.Handler` を返す |
| `gl.WithLang(lang)` | HTML の `lang` 属性を設定（デフォルト `"ja"`） |
| `gl.WithSessionTTL(d)` | セッションタイムアウトを設定（デフォルト 5 分） |
| `gl.WithDebug()` | ブラウザ DevPanel とサーバーサイド構造化ログを有効化 |
| `gl.Click(event)` | クリックイベントをバインド |
| `gl.Input(event)` | 入力イベントをバインド |
| `gl.Change(event)` | 変更イベントをバインド |
| `gl.Submit(event)` | フォーム送信イベントをバインド |
| `gl.Focus(event)` | フォーカスイベントをバインド |
| `gl.Blur(event)` | ブラーイベントをバインド |
| `gl.Keydown(event)` | キーダウンイベントをバインド |
| `gl.Key(key)` | キー名でキーダウンをフィルタ（例: `"Enter"`） |
| `gl.View` | インターフェース: `Mount`, `Render`, `HandleEvent` |
| `gl.Params` | `map[string]string` — URL パラメータ |
| `gl.Payload` | `map[string]string` — イベントデータ |
