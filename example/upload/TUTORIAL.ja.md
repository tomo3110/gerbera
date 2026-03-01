# ファイルアップロード チュートリアル

## 概要

ファイルを選択・アップロードし、結果を表示するデモを作成します。このチュートリアルでは以下の機能を学びます。

- `gl.UploadHandler` インターフェース — サーバーでのファイルアップロード処理
- `gl.Upload()` — ファイル入力にアップロードイベントをバインド
- `gl.UploadMultiple()` — 複数ファイル選択の許可
- `gl.UploadedFile` — アップロードされたファイルのメタデータとデータへのアクセス
- `ge.Each` / `ge.If` — 条件付き・反復レンダリング

このサンプルは **Gerbera のアップロードフロー** を示します。クライアント JS がファイル選択時に自動で POST し、サーバーが `HandleUpload` で処理、その後クライアントが WebSocket 経由で `gerbera:upload_complete` イベントを送信して UI を再描画します。

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

## ステップ 2: FileInfo 型の定義

```go
type FileInfo struct {
	Name     string
	Size     int64
	MIMEType string
}

func (f FileInfo) ToMap() g.Map {
	return g.Map{
		"name":     f.Name,
		"size":     f.Size,
		"mimetype": f.MIMEType,
	}
}
```

`FileInfo` はアップロードされた各ファイルのメタデータを保持します。`g.ConvertToMap` を実装しているため、ファイルテーブルのレンダリングで `ge.Each` と共に使用できます。

## ステップ 3: View 構造体の定義

```go
type UploadView struct {
	Files []FileInfo
}
```

View はアップロードされたファイルのメタデータのリストを保持するだけです。JSCommandView とは異なり、`CommandQueue` の埋め込みは不要です — アップロード処理はインターフェースベースです。

## ステップ 4: HandleUpload の実装

```go
func (v *UploadView) HandleUpload(_ string, files []gl.UploadedFile) error {
	for _, f := range files {
		v.Files = append(v.Files, FileInfo{
			Name:     f.Name,
			Size:     f.Size,
			MIMEType: f.MIMEType,
		})
	}
	return nil
}
```

これがアップロード機能の中核です。`UploadHandler` インターフェースを実装することで、View はアップロードされたファイルを受け取ります:
- `Name` — 元のファイル名
- `Size` — ファイルサイズ（バイト）
- `MIMEType` — コンテンツタイプ（例: `"image/png"`, `"application/pdf"`）
- `Data` — 生のファイルバイト列（このデモでは未使用だが、処理に利用可能）

第 1 引数の `event` は `gl.Upload("on_upload")` で指定したイベント名と一致します。同じページに複数のアップロード入力がある場合に区別するために使用できます。

## ステップ 5: HandleEvent の実装

```go
func (v *UploadView) HandleEvent(event string, _ gl.Payload) error {
	switch event {
	case "gerbera:upload_complete":
		// UI は自動的に再描画される。ファイルは HandleUpload で既に追加済み。
	case "clear":
		v.Files = nil
	}
	return nil
}
```

2 つのイベントを処理します:
- `gerbera:upload_complete` — アップロード成功後にクライアント JS が自動送信。ファイルは `HandleUpload` で既に `v.Files` に追加済みなので、`HandleEvent` 後の再描画だけが必要です。
- `clear` — ファイルリストをリセット。

## ステップ 6: アップロード入力のバインド

```go
gd.Input(
	gp.Type("file"),
	gl.Upload("on_upload"),
	gl.UploadMultiple(),
)
```

主要な属性:
- `gp.Type("file")` — 標準的な HTML ファイル入力
- `gl.Upload("on_upload")` — `gerbera-upload="on_upload"` を設定し、ファイル選択時に gerbera.js が自動で POST するよう指示
- `gl.UploadMultiple()` — `multiple="multiple"` を設定し、複数ファイルの選択を許可

オプション: `gl.UploadAccept("image/*,.pdf")` でファイルタイプを制限、`gl.UploadMaxSize(5 << 20)` でクライアントサイドのサイズ上限を設定できます。

## ステップ 7: ファイルテーブルのレンダリング

```go
fileItems := make([]g.ConvertToMap, len(v.Files))
for i, f := range v.Files {
	fileItems[i] = f
}

ge.If(len(v.Files) > 0,
	gd.Div(
		gd.Table(
			gd.Thead(
				gd.Tr(
					gd.Th(gp.Value("Name")),
					gd.Th(gp.Value("Size")),
					gd.Th(gp.Value("MIME Type")),
				),
			),
			gd.Tbody(
				ge.Each(fileItems, func(item g.ConvertToMap) g.ComponentFunc {
					m := item.ToMap()
					name := m.Get("name").(string)
					size := m.Get("size").(int64)
					mime := m.Get("mimetype").(string)
					return gd.Tr(
						gd.Td(gp.Value(name)),
						gd.Td(gp.Value(formatSize(size))),
						gd.Td(gp.Value(mime)),
					)
				}),
			),
		),
		gd.Button(gl.Click("clear"), gp.Value("Clear")),
	),
	gd.P(gp.Class("empty"), gp.Value("No files uploaded yet.")),
)
```

ファイル一覧は `ge.Each` を使って HTML テーブルとしてレンダリングされます。ヘルパー関数 `formatSize` がバイト数を人間が読みやすい文字列（B, KB, MB）に変換します。

## アップロードフロー

```
ブラウザ                                  Go サーバー
  |                                          |
  |  1. ユーザーが <input> でファイルを選択    |
  |                                          |
  |  2. gerbera.js が [gerbera-upload]       |
  |     入力の change イベントを検知          |
  |                                          |
  |  3. POST /?gerbera-upload=1              |
  |     &session=xxx&event=on_upload         |
  |     Content-Type: multipart/form-data    |
  |     [ファイルデータ]                      |
  |                                          |
  |      Handler が handleUpload() にルート   |
  |      -> ParseMultipartForm              |
  |      -> UploadHandler.HandleUpload()     |
  |      <- 200 OK {"ok":true}              |
  |                                          |
  |  4. gerbera.js が 200 OK を受信          |
  |  --> WebSocket: {"e":                    |
  |       "gerbera:upload_complete",         |
  |       "p":{"event":"on_upload"}}         |
  |                                          |
  |      HandleEvent -> Render -> Diff       |
  |                                          |
  |  <-- DOM パッチ（ファイルテーブル更新）    |
  |                                          |
  |  5. ブラウザに更新されたファイル一覧を表示  |
```

アップロードは 2 段階のプロセスです:
1. **HTTP POST** — ファイルデータがマルチパートフォームで送信され、`HandleUpload` で処理
2. **WebSocket イベント** — クライアントがアップロード完了をサーバーに通知し、再描画をトリガー

## 実行方法

```bash
go run example/upload/upload.go
```

ブラウザで http://localhost:8885 を開きます。1 つまたは複数のファイルを選択すると、即座にアップロードされ、メタデータがテーブルに表示されます。**Clear** をクリックするとリセットされます。

### デバッグモード

```bash
go run example/upload/upload.go -debug
```

デバッグパネルの **Events** タブを観察すると、各アップロード後に `gerbera:upload_complete` イベントが到着する様子を確認できます。

## 発展課題

1. `gl.UploadAccept("image/*")` を追加して、画像ファイルのみに制限してみましょう
2. `gl.UploadMaxSize(1 << 20)` を追加して、ファイルサイズを 1MB に制限してみましょう
3. アップロードされた `Data` バイト列を保存し、データ URL で画像プレビューを表示してみましょう
4. 個別のファイルを削除できるボタンを各行に追加してみましょう

## API リファレンス

| 関数 | 説明 |
|------|------|
| `gl.UploadHandler` | インターフェース: `View` を拡張し `HandleUpload(event, files)` を追加 |
| `gl.UploadedFile` | 構造体: `Name`, `Size`, `MIMEType`, `Data` |
| `gl.Upload(event)` | ファイル入力にアップロードイベントをバインド |
| `gl.UploadMultiple()` | 複数ファイル選択を許可 |
| `gl.UploadAccept(types)` | 受け入れるファイルタイプを制限 |
| `gl.UploadMaxSize(bytes)` | クライアントサイドの最大ファイルサイズを設定 |
| `ge.Each(items, cb)` | リストを反復してレンダリング |
| `ge.If(cond, t, f)` | 条件付きレンダリング |
