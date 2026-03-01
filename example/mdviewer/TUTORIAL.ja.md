# Markdown ビューアー チュートリアル

## 概要

Gerbera Live の実践的な機能を活用した、リアルタイム Markdown ビューアー/エディタを構築します:

- `gl.View` インターフェース — ファイル I/O を含むステートフルな LiveView
- `gl.TickerView` インターフェース — サーバーサイドの定期ファイルポーリング（クライアント JS 不要）
- `gl.CommandQueue` — `ScrollIntoPct` によるサーバー→クライアントスクロール命令
- `gl.Input` — リアルタイムテキスト入力バインディング
- `g.Literal()` — 生 HTML の注入（レンダリング済み Markdown）
- `OpSetHTML` — `innerHTML` を使用する WebSocket 差分パッチ
- `gs.CSS()` — `<style>` 要素によるコンポーネントスコープ CSS の埋め込み
- 独立モジュール — `replace` ディレクティブと外部依存を持つ `go.mod`
- `gl.Scroll` / `gl.Throttle` — `CommandQueue` によるサーバー経由のスクロール同期

## 前提条件

- Go 1.22 以上がインストール済み
- このリポジトリがローカルにクローン済み

## Step 1: プロジェクト構成

このサンプルは外部依存（goldmark）を持つ独立した Go モジュールです:

```
example/mdviewer/
├── go.mod          # replace ディレクティブ付きの独立モジュール
├── main.go         # CLI フラグとサーバー起動
├── markdown.go     # goldmark ラッパー
└── viewer.go       # MarkdownView（gl.View + gl.TickerView 実装）
```

`go.mod` は `replace` ディレクティブでローカルの gerbera を参照します:

```
module github.com/tomo3110/gerbera/example/mdviewer

require (
    github.com/tomo3110/gerbera v0.0.0
    github.com/yuin/goldmark v1.7.8
)

replace github.com/tomo3110/gerbera => ../..
```

## Step 2: Markdown レンダリング

```go
var md = goldmark.New(
    goldmark.WithExtensions(extension.GFM),
    goldmark.WithRendererOptions(html.WithUnsafe()),
)

func renderMarkdown(source string) string {
    var buf bytes.Buffer
    if err := md.Convert([]byte(source), &buf); err != nil {
        return "<p>Markdown render error</p>"
    }
    return buf.String()
}
```

goldmark パーサーは GFM（GitHub Flavored Markdown）拡張を有効にし、テーブル・取り消し線・タスクリストに対応します。`html.WithUnsafe()` により Markdown ソース中の生 HTML がそのまま出力されます。

## Step 3: View 構造体

```go
type MarkdownView struct {
    gl.CommandQueue              // サーバー→クライアント JS コマンド用に埋め込み
    Source     string            // Markdown ソーステキスト
    HTML       string            // レンダリング済み HTML（キャッシュ）
    FilePath   string            // .md ファイルパス（空 = ファイルなし）
    Preview    bool              // プレビューのみモード
    ModTime    time.Time         // ファイルの最終更新時刻
    FileName   string            // 表示用ファイル名
    SaveStatus string            // "", "saved", "error"
    ScrollPct  float64           // スクロール位置（0.0〜1.0）
}
```

設計上のポイント:
- **`gl.CommandQueue` の埋め込み** — `JSCommander` インターフェースを実装し、カスタム JavaScript なしでクライアントにスクロール命令を発行可能
- `HTML` フィールドはレンダリング済み Markdown をキャッシュし、毎回の `Render()` 呼び出しでの再レンダリングを回避

## Step 4: Mount

```go
func (v *MarkdownView) Mount(_ gl.Params) error {
    if v.FilePath != "" {
        data, err := os.ReadFile(v.FilePath)
        // ... ファイル読み込み、ModTime と FileName を設定
    }
    if v.Source == "" && !v.Preview {
        v.Source = "# Hello Markdown\n\nStart typing here..."
    }
    v.HTML = renderMarkdown(v.Source)
    return nil
}
```

カウンターの例では `Mount` は単純な値の初期化のみでしたが、ここではファイル I/O を実行します。`FilePath` と `Preview` は `main.go` のファクトリ関数で `Mount` 呼び出し前に設定されます。

## Step 5: イベント処理

```go
func (v *MarkdownView) HandleEvent(event string, payload gl.Payload) error {
    switch event {
    case "edit":
        v.Source = payload["value"]
        v.HTML = renderMarkdown(v.Source)
    case "save":
        os.WriteFile(v.FilePath, []byte(v.Source), 0644)
        v.SaveStatus = "saved"
    case "editor-scroll":
        // スクロール率を計算しスクロールコマンドを送信
    }
    return nil
}
```

主要な 2 つのイベント:
- **edit**: テキストエリアの `gl.Input("edit")` で発火。ソースを更新し HTML を再レンダリング。
- **save**: 現在のソースをファイルに書き込み。`FilePath` が設定されている場合のみ有効。

`check-file` イベント処理は `TickerView` インターフェースに置き換えられました（Step 7 参照）。

## Step 6: `g.Literal()` と `OpSetHTML` による生 HTML

```go
gd.Div(
    gp.ID("md-preview"),
    gp.Class("md-preview"),
    g.Literal(v.HTML),  // goldmark からの生 HTML
)
```

`g.Literal()` は生 HTML を要素の `Value` フィールドに注入します。初回 HTTP レスポンスでは正常にレンダリングされます。WebSocket 更新では、`diff` パッケージが値に `<` が含まれることを検知し、`OpSetText` の代わりに `OpSetHTML` を使用します。これによりクライアント JS は `textContent` の代わりに `innerHTML` を使用します。

これが単純なテキスト更新との重要な違いです: `OpSetHTML` がなければ、WebSocket 更新時に HTML タグがエスケープされ、Markdown レンダリングが壊れてしまいます。

## Step 7: `TickerView` によるファイル監視

クライアントサイドの JavaScript ポーリングの代わりに、`TickerView` インターフェースがサーバーサイドの定期コールバックを提供します:

```go
// TickInterval はファイル変更ポーリングの間隔を返します。
func (v *MarkdownView) TickInterval() time.Duration {
    if v.FilePath == "" {
        return 0  // ファイルなしの場合はティック無効
    }
    return 2 * time.Second
}

// HandleTick は各ティックで外部ファイルの変更をチェックします。
func (v *MarkdownView) HandleTick() error {
    if v.FilePath == "" {
        return nil
    }
    info, err := os.Stat(v.FilePath)
    if err != nil {
        return nil
    }
    if info.ModTime().After(v.ModTime) {
        data, _ := os.ReadFile(v.FilePath)
        v.Source = string(data)
        v.HTML = renderMarkdown(v.Source)
        v.ModTime = info.ModTime()
    }
    return nil
}
```

従来のクライアントサイドポーリングに対する利点:
- **JavaScript 不要** — 隠しボタン、`setInterval`、クライアントサイドのハック不要
- **アイドル時のオーバーヘッドゼロ** — `HandleTick` が状態を変更しなければパッチは送信されない
- **動的な無効化** — `TickInterval()` で `0` を返すとティックを無効化可能（ファイル未読み込み時など）

## Step 8: `CommandQueue` によるスクロール同期

スクロール同期は `CommandQueue` を使ってサーバーからクライアントに `ScrollIntoPct` コマンドを送信し、`MutationObserver` や `data-scroll-pct` 属性ハックを不要にします。

### スクロールイベントのバインディング

```go
gd.Textarea(
    gp.ID("md-editor"),
    gp.Class("md-textarea"),
    gl.Input("edit"),
    gl.Scroll("editor-scroll"),  // スクロールイベントをサーバーに送信
    gl.Throttle(50),             // 50ms に 1 回に制限
    gp.Value(v.Source),
),
```

`gl.Scroll("editor-scroll")` は `gerbera-scroll` 属性を追加します。クライアント JS（`gerbera.js`）が `scroll` イベントを監視し、`scrollTop`・`scrollHeight`・`clientHeight` をペイロードとして送信します。`gl.Throttle(50)` でイベント頻度を制限します。

### CommandQueue を使ったサーバー側ハンドラ

```go
case "editor-scroll":
    scrollTop, _ := strconv.ParseFloat(payload["scrollTop"], 64)
    scrollHeight, _ := strconv.ParseFloat(payload["scrollHeight"], 64)
    clientHeight, _ := strconv.ParseFloat(payload["clientHeight"], 64)
    max := scrollHeight - clientHeight
    if max > 0 {
        v.ScrollPct = scrollTop / max
        v.ScrollIntoPct(".md-preview-pane", fmt.Sprintf("%.6f", v.ScrollPct))
    }
```

ハンドラはスクロール位置を 0.0〜1.0 の範囲に正規化し、`ScrollIntoPct` コマンドを発行します。View に埋め込まれた `CommandQueue` がこのコマンドをキューイングし、次のレンダリング後にフレームワークがクライアントに送信します。

### データフロー

```
ブラウザ                                   Go サーバー
  |                                           |
  | ユーザーがテキストエリアをスクロール          |
  | gerbera.js: scroll イベント（50ms throttle）|
  | --> {"e":"editor-scroll","p":{            |
  |       "scrollTop":"150",                  |
  |       "scrollHeight":"1200",              |
  |       "clientHeight":"400"}}              |
  |                                           |
  |     HandleEvent: ScrollPct = 150/800      |
  |     CommandQueue: ScrollIntoPct(0.1875)   |
  |     Render + Diff + コマンド排出            |
  |                                           |
  | <-- {"patches":[...],                     |
  |      "js_commands":[{"cmd":"scroll_into_pct",
  |        "target":".md-preview-pane",       |
  |        "args":{"pct":"0.187500"}}]}       |
  |                                           |
  | gerbera.js がコマンドを実行:               |
  | el.scrollTop = 0.1875 * maxScroll         |
```

主な利点: `MutationObserver` 不要、`data-scroll-pct` 属性不要、カスタム JavaScript 不要 — フレームワークがすべてを処理します。

## Step 9: `gs.CSS()` による CSS

```go
gd.Head(
    gd.Title(title),
    gs.CSS(cssStyles),  // <style> 要素で CSS を埋め込み
),
```

`gs.CSS()` は生 CSS コンテンツを含む `<style>` 要素を生成します。以前は `g.Tag("style", gp.Value(css))` で行っていましたが、専用ヘルパーの方がより簡潔で意味的に明確です。

## Step 10: CLI とサーバー

```go
func main() {
    addr := flag.String("addr", ":8860", "listen address")
    preview := flag.Bool("preview", false, "preview-only mode")
    debug := flag.Bool("debug", false, "enable debug panel")
    flag.Parse()

    factory := func() gl.View {
        return &MarkdownView{
            FilePath: filePath,
            Preview:  *preview,
        }
    }

    http.Handle("/", gl.Handler(factory, opts...))
    log.Fatal(http.ListenAndServe(*addr, nil))
}
```

ファクトリ関数は CLI 引数からファイルパスとプレビューフラグをキャプチャし、各セッションの `MarkdownView` に渡します。

## 動作の仕組み

```
ブラウザ                                Go サーバー
  |                                         |
  | 1. GET /                                |
  | <-- HTML（エディタ + レンダリング済み MD）  |
  |                                         |
  | 2. WebSocket 接続                       |
  |                                         |
  | 3. テキストエリアに入力                    |
  | --> {"e":"edit","p":{"value":"# Hi"}}   |
  |                                         |
  |     HandleEvent -> renderMarkdown       |
  |     -> Render -> Diff -> Patches        |
  |                                         |
  | 4. <-- [{"op":"html",                   |
  |          "path":[1,1,1,0],              |
  |          "val":"<h1>Hi</h1>"}]          |
  |                                         |
  | 5. JS: node.innerHTML = val             |
  |                                         |
  | 6. （2秒ごと）サーバーティック             |
  |     HandleTick -> ファイル ModTime チェック |
  |     -> 変更あれば再レンダリング -> パッチ   |
```

## 実行方法

```bash
# エディタモード（ファイルなし）
cd example/mdviewer && go run .

# ファイル指定エディタモード
cd example/mdviewer && go run . README.md

# プレビューのみモード（ファイル変更を自動検知）
cd example/mdviewer && go run . -preview README.md

# デバッグパネル付き
cd example/mdviewer && go run . -debug
```

http://localhost:8860 を開き、Markdown を入力するとリアルタイムでレンダリングされます。

## 演習課題

1. 入力に応じて更新される「単語数」表示を追加する
2. `gl.Keydown` と `gl.Key` を使って Ctrl+S のキーボードショートカットで保存できるようにする
3. クライアントサイドライブラリを使ってコードブロックにシンタックスハイライトを追加する
4. エディタのみ・プレビューのみ・分割表示を切り替える「表示モード」トグルボタンを実装する
