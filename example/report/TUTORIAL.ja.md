# レポート生成 チュートリアル

## 概要

HTTP サーバーを使わず、CLI から HTML レポートを標準出力に生成します。このチュートリアルでは以下の機能を学びます。

- `Template` を `io.Reader` として使用
- `Template.Mount()`, `Template.String()`, `Template.Bytes()`
- `Skip` — 条件付きセクション省略
- `Tag` — `dom/` にない要素の直接生成
- `styles.Style` — インライン CSS
- `Ol`, `Li` — 番号付きリスト
- `H4`, `H5` — 小見出し
- `Unless` — 否定条件の表示

## 前提条件

- Go 1.22 以上がインストールされていること
- このリポジトリをクローン済みであること

## ステップ 1: Template の直接利用

```go
tmpl := &g.Template{Lang: "ja"}

if err := tmpl.Mount(
	gd.Head(
		gd.Title("プロジェクト進捗レポート"),
		gd.Meta(gp.Attr("charset", "UTF-8")),
	),
	gd.Body(...),
); err != nil {
	fmt.Fprintf(os.Stderr, "Mount error: %v\n", err)
	os.Exit(1)
}
```

- `g.Template{Lang: "ja"}` で直接テンプレートを作成します
- `Mount(children...)` は内部で `Parse` → `Render` を実行し、バッファに HTML を蓄積します
- `ExecuteTemplate` や `Handler` を使わない、最も低レベルなテンプレート操作です

## ステップ 2: String() / Bytes() で出力

```go
fmt.Print(tmpl.String())     // HTML を文字列として標準出力
data := tmpl.Bytes()          // HTML を []byte として取得
```

`Mount()` 後に `String()` でバッファの内容を文字列として取得できます。`Bytes()` は `[]byte` を返します。

`Template` は `io.Reader` インターフェースも実装しています。`Execute()` メソッド内部で利用されており、HTTP レスポンスへの書き込みなどに活用されます。

## ステップ 3: Skip で条件付き省略

```go
ge.If(done,
	g.Skip(),
	gd.Details(
		gd.Summary(gp.Value("残作業")),
		gd.P(gp.Value("このタスクにはまだ作業が残っています。")),
	),
),
```

- `g.Skip()` は何も描画しない `ComponentFunc` を返します
- 条件分岐で「何も表示しない」ケースに使います
- 上の例では、タスクが完了していれば何も表示せず、未完了なら詳細を表示します
- `gd.Details()` と `gd.Summary()` は `dom/` パッケージの専用ヘルパーとして利用可能です

## ステップ 4: Tag で任意の要素を生成

```go
gd.Details(
	gd.Summary(gp.Value("残作業")),
	gd.P(gp.Value("このタスクにはまだ作業が残っています。")),
),
```

- `g.Tag(tagName, children...)` は任意のタグ名で要素を生成でき、任意の要素に対して引き続き利用可能です
- `<details>` と `<summary>` は `dom/` パッケージに専用の `gd.Details()` / `gd.Summary()` ラッパーが追加されました
- 実は `dom/` の関数はすべて `g.Tag` の薄いラッパーです

## ステップ 5: インラインスタイルと Unless

```go
ge.Unless(false,
	gd.P(
		gs.Style(g.StyleMap{"color": "red"}),
		gp.Value("注意: 期限が近づいています。"),
	),
),
```

- `ge.Unless(cond, component)` は条件が偽のときに描画します
- `gs.Style` と組み合わせて、警告メッセージを目立たせることができます

## 実行方法

```bash
go run example/report/report.go
```

HTML が標準出力に出力されます。ファイルに保存する場合:

```bash
go run example/report/report.go > report.html
```

ブラウザで `report.html` を開いて確認できます。

## 発展課題

1. レポートをファイルに直接書き出す機能を追加してみましょう（`os.Create` + `io.Copy`）
2. `g.Tag` で `<dialog>` や `<progress>` 要素を使ってみましょう
3. コマンドライン引数でレポートのタイトルを変更できるようにしてみましょう

## API リファレンス

| 関数 | 説明 |
|------|------|
| `g.Template{Lang: "ja"}` | テンプレート構造体の直接作成 |
| `tmpl.Mount(children...)` | コンポーネントをマウント・レンダリング |
| `tmpl.String()` | HTML を文字列として取得 |
| `tmpl.Bytes()` | HTML を `[]byte` として取得 |
| `tmpl.Read(b)` | `io.Reader` インターフェースの実装 |
| `g.Skip()` | 何も描画しない ComponentFunc |
| `g.Tag(name, children...)` | 任意のタグ名で要素を生成 |
| `gs.Style(styleMap)` | インライン CSS を設定 |
| `ge.Unless(cond, component)` | 条件が偽のとき描画 |
| `gd.H4(children...)` | `<h4>` 見出し |
| `gd.H5(children...)` | `<h5>` 見出し |
| `gd.Ol(children...)` | `<ol>` 番号付きリスト |
