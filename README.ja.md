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
- **最適化されたレンダリング** — `sync.Pool` による再利用、ゼロアロケーションのレンダリング層、事前計算されたインデント文字列、`bufio.Writer` 直接書き込み
- **アクセシビリティ** — `property/` パッケージに包括的な ARIA 属性ヘルパーを搭載
- **テストユーティリティ** — `live.TestView` によりブラウザなしで LiveView 実装をユニットテスト可能

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
	http.Handle("/", gl.Handler(func(_ context.Context) gl.View { return &CounterView{} }))
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
├── property/          # 属性セッター (Class, Attr, ID, ARIA ヘルパーなど)
├── expr/              # 制御フロー: If()、Unless()、Each()
├── styles/            # Style(), CSS(), ScopedCSS() による CSS 処理
├── components/        # ビルド済みコンポーネント (Bootstrap CDN, TailwindCSS, Tabs など)
├── diff/              # 要素ツリー差分比較とフラグメントレンダリング
├── live/              # LiveView: View インターフェース、WebSocket ハンドラ、クライアント JS
├── session/           # HTTP セッション管理 (MemoryStore, CSRF, ミドルウェア)
├── ui/               # テーマシステム付きビルド済み UI コンポーネントライブラリ
│   └── live/         # LiveView 専用コンポーネント (Modal, Toast, DataTable 等)
│
└── example/           # サンプルアプリケーション
    ├── hello/         # 最小限の Hello World
    ├── portfolio/     # ポートフォリオページ
    ├── dashboard/     # テーブル付きダッシュボード
    ├── survey/        # アンケートフォーム
    ├── report/        # レポート出力 (標準出力)
    ├── wiki/          # マルチルート Wiki (サーバーサイドレンダリング)
    ├── counter/       # LiveView カウンター
    ├── clock/         # LiveView 時計 (TickerView)
    ├── jscommand/     # LiveView JS コマンド例
    ├── upload/        # LiveView ファイルアップロード
    ├── tabview/       # LiveView タブコンポーネントデモ
    ├── wiki_live/     # LiveView Wiki (SPA 風)
    ├── mdviewer/      # LiveView Markdown ビューアー/エディタ
    ├── admin/         # LiveView 管理画面ダッシュボード
    ├── catalog/       # インタラクティブ UI コンポーネントカタログ
    ├── auth/          # CSRF 付きセッション認証
    └── chat/          # マルチユーザーリアルタイムチャット (InfoReceiver)
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

**Skip()** — 何もレンダリングしない no-op の `ComponentFunc` を返します。プレースホルダーやデフォルト値として便利です:

```go
g.Skip()  // 何もレンダリングしない
```

**HandlerFunc()** — リクエストに基づいて動的にページを生成する `http.HandlerFunc` を作成します:

```go
http.HandleFunc("/", g.HandlerFunc(func(r *http.Request) []g.ComponentFunc {
    return []g.ComponentFunc{
        gd.Head(gd.Title("動的ページ")),
        gd.Body(gd.H1(gp.Value("こんにちは、" + r.URL.Query().Get("name")))),
    }
}))
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
| `gl.TouchStart(event)` / `gl.TouchEnd(event)` / `gl.TouchMove(event)` | タッチイベント |
| `gl.Hook(name)` | クライアントサイド JS フックをアタッチ |
| `gl.LiveLink(href)` | フルページリロードなしのクライアントサイドナビゲーション |
| `gl.Loading(cssClass)` | イベント処理中に CSS クラスを設定 |

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

利用可能なコマンド: `ScrollTo`, `ScrollIntoPct`, `Focus`, `Blur`, `SetAttribute`, `RemoveAttribute`, `AddClass`, `RemoveClass`, `ToggleClass`, `SetProperty`, `Dispatch`, `Show`, `Hide`, `Toggle`, `Navigate`

### アップロードサポート

`UploadHandler` インターフェースを実装してファイルアップロードを処理できます:

```go
type MyView struct {
    gl.CommandQueue
    // ...
}

func (v *MyView) HandleUpload(event string, files []gl.UploadedFile) error {
    for _, f := range files {
        // f.Name, f.Size, f.ContentType, f.Data
    }
    return nil
}
```

イベントバインディング: `gl.Upload(event)`, `gl.UploadAccept(accept)`, `gl.UploadMultiple()`, `gl.UploadMaxSize(bytes)`

### フォームユーティリティ

`live/` パッケージにはアクセシブルなフォーム構築用のヘルパーが含まれています:

```go
gl.Field("email", gl.FieldOpts{
    Label: "メール", Type: "email", Placeholder: "you@example.com",
})
gl.TextareaField("bio", gl.FieldOpts{Label: "自己紹介"})
gl.SelectField("role", []gl.SelectOption{
    {Value: "admin", Label: "管理者"},
    {Value: "user", Label: "ユーザー"},
}, gl.FieldOpts{Label: "役割"})
```

### セッション管理（`session/` パッケージ）

`session/` パッケージは HMAC-SHA256 署名付きクッキーによる HTTP セッション管理を提供します:

```go
import "github.com/tomo3110/gerbera/session"

// 署名キーを指定してストアを作成
store := session.NewMemoryStore([]byte("secret-key"),
    session.WithMaxAge(30 * time.Minute),
)
defer store.Close()

// セッションミドルウェアを適用してセッションを自動的に読み込み/保存
handler := session.Middleware(store)(mux)
```

**CSRF 保護:**

```go
sess := session.FromContext(r.Context())
token := session.GenerateCSRFToken(sess)  // トークンを生成しセッションに保存
// ... フォームの hidden フィールドにトークンを含める ...
valid := session.ValidCSRFToken(sess, submittedToken)  // 定数時間比較
```

**認証ガードミドルウェア:**

```go
// セッションに "username" キーがない場合は /login にリダイレクト
authGuard := session.RequireKey("username", "/login")
mux.Handle("/dashboard", authGuard(dashboardHandler))
```

**プッシュベースのセッション無効化:**

`MemoryStore` を `gl.WithSessionStore(store)` と組み合わせると、LiveView 接続はセッション無効化イベントを自動購読します。セッション破棄（例: ログアウト時）により、関連するすべての WebSocket 接続が閉じられます。`SessionExpiredHandler` を実装した View は切断前にコールバックを受け取ります。

### LiveView 上級インターフェース

コアの `View` インターフェースに加えて、高度なシナリオ用のオプショナルインターフェースを提供します:

**InfoReceiver** — `Session.SendInfo()` 経由でプッシュされたサーバーサイドメッセージを受信します。ユーザー間の更新ブロードキャストに便利です:

```go
type InfoReceiver interface {
    View
    HandleInfo(msg any) error
}
```

```go
func (v *ChatView) HandleInfo(msg any) error {
    if cm, ok := msg.(ChatMessage); ok {
        v.Messages = append(v.Messages, cm)
    }
    return nil
}

// 別のゴルーチンや別ユーザーのセッションから送信:
session.SendInfo(chatMessage)
```

**Unmounter** — WebSocket 接続が閉じられた時に呼ばれます。チャットルームからの退出などのクリーンアップに使用します:

```go
type Unmounter interface {
    Unmount()
}
```

**SessionExpiredHandler** — HTTP セッションが無効化された時（例: 別タブからのログアウト）に呼ばれます。切断前に最終状態をレンダリングできます:

```go
type SessionExpiredHandler interface {
    OnSessionExpired() error
}
```

```go
func (v *DashboardView) OnSessionExpired() error {
    v.SessionExpired = true
    v.Navigate("/login")  // クライアントをリダイレクト
    return nil
}
```

### Transport インターフェース

`Transport` インターフェースは View のライフサイクルを WebSocket 接続から分離し、カスタムトランスポートやテストを容易にします:

```go
type Transport interface {
    Send(msg Message) error
    Receive() (event string, payload Payload, err error)
    Close() error
}
```

`ViewLoop()` は任意の `Transport` 実装上で View を駆動します:

```go
gl.ViewLoop(view, transport, gl.ViewLoopConfig{...})
```

**TestTransport** — WebSocket なしで LiveView をテストするためのインメモリトランスポート:

```go
transport := gl.NewTestTransport()
go gl.ViewLoop(view, transport, cfg)
transport.PushEvent("inc", gl.Payload{})
msg := transport.LastMessage()
```

### LiveView のテスト

`live.TestView` を使えば、実際の WebSocket 接続なしに LiveView 実装をユニットテストできます:

```go
tv := gl.NewTestView(&MyView{})
html, _ := tv.RenderHTML()
patches, _ := tv.SimulateEvent("inc", gl.Payload{})
fmt.Println(tv.PatchCount(), tv.HasPatch(diff.OpSetText))
```

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
| `gp.DataAttr(key, val)` | `gp.Attr("data-"+key, val)` |
| `gp.Tabindex(n)` | `gp.Attr("tabindex", n)` |
| `gp.Readonly(true)` | `gp.Attr("readonly", "readonly")` |
| `gp.Title(text)` | `gp.Attr("title", text)` |

ARIA アクセシビリティヘルパー:

| 関数 | 同等の記述 |
|------|-----------|
| `gp.Role(role)` | `gp.Attr("role", role)` |
| `gp.AriaLabel(label)` | `gp.Attr("aria-label", label)` |
| `gp.AriaHidden(bool)` | `gp.Attr("aria-hidden", "true"/"false")` |
| `gp.AriaExpanded(bool)` | `gp.Attr("aria-expanded", "true"/"false")` |
| `gp.AriaSelected(bool)` | `gp.Attr("aria-selected", "true"/"false")` |
| `gp.AriaControls(id)` | `gp.Attr("aria-controls", id)` |
| `gp.AriaLive(mode)` | `gp.Attr("aria-live", mode)` |

その他: `AriaDescribedBy`, `AriaLabelledBy`, `AriaDisabled`, `AriaAtomic`, `AriaCurrent`, `AriaHasPopup`, `AriaInvalid`, `AriaRequired`, `AriaValueNow`, `AriaValueMin`, `AriaValueMax`

### スタイルヘルパー

`styles/` パッケージには追加の CSS ユーティリティがあります:

| 関数 | 説明 |
|------|------|
| `gs.Style(styleMap)` | マップからインライン CSS |
| `gs.CSS(cssText)` | 生 CSS を含む `<style>` 要素 |
| `gs.ScopedCSS(scope, rules)` | セレクタにスコープされた CSS ルール |
| `gs.StyleIf(cond, styleMap)` | 条件付きインラインスタイル |
| `gs.MediaQuery(query, css)` | CSS を `@media` ルールでラップ |
| `gs.Keyframes(name, frames)` | `@keyframes` ルールを生成 |

### 追加 DOM 要素

`dom/` パッケージは基本要素に加えて多数の HTML 要素ヘルパーを提供します:

`Span`, `Details`, `Summary`, `Dialog`, `Progress`, `Meter`, `Time`, `Mark`, `Figure`, `Figcaption`, `Video`, `Audio`, `Source`, `Canvas`, `Pre`, `Blockquote`, `Em`, `Strong`, `Small`, `Dl`, `Dt`, `Dd`, `Fieldset`, `Legend`, `Iframe` など

ハンドラオプション:

| 関数 | 説明 |
|------|------|
| `gl.WithLang(lang)` | HTML の `lang` 属性を設定（デフォルト `"ja"`） |
| `gl.WithSessionTTL(d)` | セッションタイムアウトを設定（デフォルト 5 分） |
| `gl.WithDebug()` | ブラウザ DevPanel とサーバーサイド構造化ログを有効化 |
| `gl.WithMiddleware(mw)` | HTTP ミドルウェアをハンドラに追加 |
| `gl.WithCheckOrigin(fn)` | WebSocket アップグレード用のカスタム Origin ヘッダーチェックを設定 |
| `gl.WithSessionStore(store)` | `session.Store` によるプッシュベースのセッション無効化を有効化 |

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

## UI コンポーネント（`ui/` パッケージ）

`ui/` パッケージは、CSS を書かずに Go コードだけでプロフェッショナルな UI を構築できるビルド済みコンポーネントを提供します。すべてのコンポーネントが ARIA アクセシビリティ属性を含み、CSS カスタムプロパティによるテーマに自動適応します。

```go
import (
    gu  "github.com/tomo3110/gerbera/ui"       // 静的 + インタラクティブコンポーネント
    gul "github.com/tomo3110/gerbera/ui/live"   // LiveView 専用コンポーネント
)
```

### テーマシステム

```go
gu.Theme()                          // デフォルトライトテーマ
gu.ThemeWith(gu.DarkTheme())        // ダークテーマ
gu.ThemeWith(gu.ThemeConfig{...})   // カスタム上書き
gu.ThemeAuto(light, dark)           // prefers-color-scheme による自動切替
```

### コンポーネント一覧

| カテゴリ | コンポーネント |
|---------|-------------|
| レイアウト | `AdminShell`, `Sidebar`, `PageHeader`, `Breadcrumb`, `MobileHeader` |
| グリッド & スペーシング | `Grid`, `Row`, `Column`, `Stack`, `HStack`, `VStack`, `Center`, `Container`, `Spacer` |
| データ表示 | `Card`, `Badge`, `Alert`, `StatCard`, `StyledTable`, `THead`, `Tree` |
| フォーム | `FormGroup`, `FormLabel`, `FormInput`, `FormSelect`, `FormTextarea`, `Checkbox`, `Radio` |
| フィードバック | `Spinner`, `Progress`, `EmptyState`, `Divider` |
| メディア | `Icon`, `Avatar`, `AvatarGroup`, `Chart` (Line, Column, Bar, Pie, Scatter, Histogram, StackedBar) |
| インタラクティブ | `Calendar`, `NumberInput`, `Slider`, `Accordion`, `ButtonGroup`, `Pagination`, `Stepper`, `InfiniteScroll`, `TimePicker`, `ChatContainer`, `ChatInput` |
| LiveView 専用 (`ui/live/`) | `Modal`, `Toast`, `DataTable`, `Dropdown`, `Confirm`, `Tabs`, `Drawer`, `SearchSelect` |

### イベント統合

インタラクティブコンポーネントはイベントフィールドを持つ Opts 構造体を使用します。フィールドが空なら静的 HTML、値があれば LiveView バインディング用の `gerbera-*` 属性が付与されます:

```go
// 静的
gu.Calendar(gu.CalendarOpts{Year: 2026, Month: time.March, Today: time.Now()})

// インタラクティブ（LiveView）
gu.Calendar(gu.CalendarOpts{Year: v.CalYear, Month: v.CalMonth, SelectEvent: "calSelect", ...})
```

詳細な API リファレンスは [`ui/README.ja.md`](ui/README.ja.md) を参照してください。

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
| [clock](example/clock/) | LiveView 時計（TickerView） | :8870 | `go run example/clock/clock.go` |
| [jscommand](example/jscommand/) | LiveView JS コマンド | :8875 | `go run example/jscommand/jscommand.go` |
| [upload](example/upload/) | LiveView ファイルアップロード | :8885 | `go run example/upload/upload.go` |
| [tabview](example/tabview/) | LiveView タブコンポーネント | :8890 | `go run example/tabview/tabview.go` |
| [wiki_live](example/wiki_live/) | LiveView Wiki（SPA） | :8850 | `go run example/wiki_live/wiki_live.go` |
| [mdviewer](example/mdviewer/) | LiveView Markdown ビューアー | :8860 | `cd example/mdviewer && go run .` |
| [admin](example/admin/) | LiveView 管理画面ダッシュボード | :8910 | `go run example/admin/admin.go` |
| [catalog](example/catalog/) | インタラクティブ UI コンポーネントカタログ | :8900 | `go run example/catalog/catalog.go` |
| [auth](example/auth/) | CSRF 付きセッション認証 | :8895 | `go run example/auth/auth.go` |
| [chat](example/chat/) | マルチユーザーリアルタイムチャット | :8920 | `go run example/chat/chat.go` |

## 開発

```bash
go build ./...    # 全パッケージをビルド
go test ./...     # 全テストを実行
go test -v ./...  # 詳細なテスト出力
```

Go 1.22 以上が必要です。

### ベンチマーク

```bash
go test -bench=. -benchmem ./...   # 全ベンチマークをメモリ統計付きで実行
```

レンダリングパイプラインは `sync.Pool` によるバッファ再利用、ゼロアロケーションのコアレンダリングパス、事前計算されたインデント文字列で最適化されています。

## ライセンス

[MIT](LICENSE)
