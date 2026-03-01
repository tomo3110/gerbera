# Markdown ビューアー チュートリアル

## 概要

Gerbera Live の実践的な機能を活用した、リアルタイム Markdown ビューアー/エディタを構築します:

- `gl.View` インターフェース — ファイル I/O を含むステートフルな LiveView
- `gl.Input` — リアルタイムテキスト入力バインディング
- `g.Literal()` — 生 HTML の注入（レンダリング済み Markdown）
- `OpSetHTML` — `innerHTML` を使用する WebSocket 差分パッチ
- 独立モジュール — `replace` ディレクティブと外部依存を持つ `go.mod`
- クライアントサイドポーリング — サーバープッシュなしのファイル変更検知

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
└── viewer.go       # MarkdownView（gl.View 実装）
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
    Source     string    // Markdown ソーステキスト
    HTML       string    // レンダリング済み HTML（キャッシュ）
    FilePath   string    // .md ファイルパス（空 = ファイルなし）
    Preview    bool      // プレビューのみモード
    ModTime    time.Time // ファイルの最終更新時刻
    FileName   string    // 表示用ファイル名
    SaveStatus string    // "", "saved", "error"
}
```

`HTML` フィールドはレンダリング済みの Markdown をキャッシュし、毎回の `Render()` 呼び出しでの再レンダリングを避けます。`Source` が変更された時のみ更新されます。

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
    case "check-file":
        // ModTime を比較、変更があれば再読み込み
    }
    return nil
}
```

3 つのイベント:
- **edit**: テキストエリアの `gl.Input("edit")` で発火。ソースを更新し HTML を再レンダリング。
- **save**: 現在のソースをファイルに書き込み。`FilePath` が設定されている場合のみ有効。
- **check-file**: ファイルの更新時刻をポーリング。外部で変更されていた場合は再読み込み。

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

## Step 7: ポーリングによるファイル監視

Gerbera Live にはサーバープッシュ機能がないため、ファイル監視はクライアントサイドポーリングで実装します:

```go
// DOM 内の隠しボタン
gd.Button(
    gp.ID("file-check-trigger"),
    gl.Click("check-file"),
    gs.Style(g.StyleMap{"display": "none"}),
)
```

```javascript
// 2 秒ごとに自動クリック
setInterval(function() {
    document.getElementById('file-check-trigger').click();
}, 2000);
```

ファイルが変更されていない場合、`HandleEvent` は状態を変更せずに返り、パッチ数は 0 になります。ハンドラは WebSocket メッセージの送信をスキップします（`handler.go` L191 参照）。

## Step 8: スクロール同期

```javascript
ta.addEventListener('scroll', function() {
    var pct = ta.scrollTop / (ta.scrollHeight - ta.clientHeight);
    pv.scrollTop = pct * (pv.scrollHeight - pv.clientHeight);
});
```

エディタのスクロール位置をパーセンテージに変換し、プレビューペインに適用します。`g.Literal()` でインライン `<script>` として注入します。

## Step 9: CLI とサーバー

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
