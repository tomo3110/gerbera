# Gerbera フロントエンド開発 ベストプラクティス集

Gerbera でフロントエンドを書く際の手引書です。基本的なパターンから LiveView の実装テクニック、パフォーマンス最適化、テストまで、実際の開発で役立つ知識をまとめています。

> **対象読者**: Gerbera の基本（`ComponentFunc`、`Tag()`、`dom/` パッケージ）を理解しており、実際のアプリケーションを構築する開発者。

---

## 目次

1. [プロジェクト構成](#1-プロジェクト構成)
2. [コンポーネント設計](#2-コンポーネント設計)
3. [LiveView の基本](#3-liveview-の基本)
4. [イベント処理のパターン](#4-イベント処理のパターン)
5. [URL 管理（PushPatch / HandleParams）](#5-url-管理pushpatch--handleparams)
6. [フォームとバリデーション](#6-フォームとバリデーション)
7. [リアルタイム通信](#7-リアルタイム通信)
8. [JS コマンド（CommandQueue）](#8-js-コマンドcommandqueue)
9. [スタイリング](#9-スタイリング)
10. [パフォーマンス最適化](#10-パフォーマンス最適化)
11. [テスト](#11-テスト)
12. [セキュリティ](#12-セキュリティ)
13. [デバッグ](#13-デバッグ)
14. [アンチパターン集](#14-アンチパターン集)

---

## 1. プロジェクト構成

### 推奨パッケージエイリアス

```go
import (
    g   "github.com/tomo3110/gerbera"            // コア: Tag, Literal, ComponentFunc
    gd  "github.com/tomo3110/gerbera/dom"         // HTML 要素: Div, H1, Body, ...
    gp  "github.com/tomo3110/gerbera/property"    // 属性: Class, Attr, ID, Value, ...
    ge  "github.com/tomo3110/gerbera/expr"        // 制御フロー: If, Unless, Each
    gs  "github.com/tomo3110/gerbera/styles"      // CSS: Style, CSS, ScopedCSS
    gl  "github.com/tomo3110/gerbera/live"        // LiveView: View, Handler, Click, ...
    gu  "github.com/tomo3110/gerbera/ui"          // UI コンポーネント: Card, Badge, ...
    gul "github.com/tomo3110/gerbera/ui/live"     // LiveView 専用 UI: Modal, Toast, ...
)
```

プロジェクト全体で統一したエイリアスを使うことで、コードの可読性が大幅に向上します。

### ファイル分割の指針

規模の大きな LiveView アプリケーション（SNS サンプル参照）では、以下のように分割します:

```
myapp/
├── main.go          # エントリポイント、ルーティング、ミドルウェア
├── view.go          # View 構造体、Mount, HandleEvent, HandleInfo
├── pages.go         # Render() と各ページのレンダリング関数
├── components.go    # 再利用可能な ComponentFunc ビルダー
├── css.go           # CSS 文字列定数
├── db.go            # データベースアクセス関数
└── models.go        # データモデル構造体
```

**ポイント**: `HandleEvent` のロジックと `Render` の UI 構築は別ファイルに分けると、それぞれが肥大化しても見通しが良くなります。

---

## 2. コンポーネント設計

### 基本: 関数で合成する

Gerbera のコンポーネントはすべて `ComponentFunc`（`func(*Element) error`）です。通常のGo関数として定義し、呼び出し側で合成します:

```go
// Good: 再利用可能なコンポーネント関数
func userCard(name, email string) g.ComponentFunc {
    return gd.Div(
        gp.Class("user-card"),
        gd.H3(gp.Value(name)),
        gd.P(gp.Value(email)),
    )
}

// 使用側
gd.Body(
    userCard("Alice", "alice@example.com"),
    userCard("Bob", "bob@example.com"),
)
```

### 条件付きレンダリング

`expr.If` / `expr.Unless` を使います:

```go
ge.If(user.IsAdmin,
    gd.Button(gp.Value("管理画面"), gl.Click("openAdmin")),
)

ge.Unless(isLoggedIn,
    gd.A(gp.Href("/login"), gp.Value("ログイン")),
)
```

`If` の第3引数以降は `else` 節です:

```go
ge.If(len(items) > 0,
    renderItemList(items),    // true の場合
    renderEmptyState(),       // false の場合
)
```

### リストレンダリング

`expr.Each` を使います。パフォーマンス上、**必ず `property.Key()` を付与してください**:

```go
ge.Each(posts, func(i int, post Post) g.ComponentFunc {
    return gd.Div(
        gp.Key(fmt.Sprintf("post-%d", post.ID)),  // 重要: 差分検出のキー
        gp.Class("post-card"),
        gd.P(gp.Value(post.Content)),
    )
})
```

> **注意**: `Key` がないと、リスト内の要素の追加・削除時にツリー全体が再構築されます。一意な識別子（データベースの ID など）を使いましょう。

### Skip() で何もレンダリングしない

条件分岐で「何も表示しない」場合は `g.Skip()` を使います:

```go
func maybeAlert(msg string) g.ComponentFunc {
    if msg == "" {
        return g.Skip()  // 何も出力しない
    }
    return gd.Div(gp.Class("alert"), gp.Value(msg))
}
```

---

## 3. LiveView の基本

### View インターフェースの実装

最小限の LiveView は 3 つのメソッドを実装します:

```go
type MyView struct {
    gl.CommandQueue  // JS コマンド機能を使う場合に埋め込む
    count int
}

func (v *MyView) Mount(params gl.Params) error {
    // 初期状態の設定。params.Path で URL パス、
    // params.Query で URL クエリパラメータにアクセスできる。
    // params.Conn.Session で HTTP セッションにアクセスできる。
    v.count = 0
    return nil
}

func (v *MyView) Render() []g.ComponentFunc {
    // <html> 直下の子要素を返す。
    // 必ず Head と Body を含めること。
    return []g.ComponentFunc{
        gd.Head(gd.Title("My App")),
        gd.Body(
            gd.H1(gp.Value(fmt.Sprintf("Count: %d", v.count))),
            gd.Button(gl.Click("inc"), gp.Value("+")),
        ),
    }
}

func (v *MyView) HandleEvent(event string, payload gl.Payload) error {
    // payload は map[string]string 型。
    // イベント名で分岐し、View の状態を更新する。
    // return 後、自動で Render → Diff → パッチ送信が行われる。
    switch event {
    case "inc":
        v.count++
    }
    return nil
}
```

### Mount で使える情報

`Params` 構造体には以下の情報が含まれます:

```go
func (v *MyView) Mount(params gl.Params) error {
    // URL パス（例: "/profile"）
    path := params.Path

    // URL クエリパラメータ（例: "?id=123" → params.Get("id") == "123"）
    id := params.Get("id")

    // HTTP セッション（認証情報など）
    if params.Conn.Session != nil {
        userID, _ := params.Conn.Session.Get("user_id").(int64)
    }

    // LiveView セッション（SendInfo 用）
    v.liveSession = params.Conn.LiveSession

    // 接続情報
    _ = params.Conn.RemoteAddr  // クライアント IP
    _ = params.Conn.UserAgent   // ブラウザ UA 文字列
    return nil
}
```

### ハンドラの登録

```go
// 基本
http.Handle("/", gl.Handler(func(ctx context.Context) gl.View {
    return &MyView{}
}, gl.WithDebug()))

// オプション
gl.WithLang("ja")                   // HTML lang 属性（デフォルト "ja"）
gl.WithSessionTTL(10 * time.Minute) // セッション TTL（デフォルト 5 分）
gl.WithDebug()                      // デバッグパネル有効化
gl.WithSessionStore(store)          // HTTP セッション連携
gl.WithMiddleware(authMW)           // ミドルウェア適用
gl.WithCheckOrigin(fn)              // WebSocket Origin チェック
```

---

## 4. イベント処理のパターン

### イベントバインディング一覧

| バインディング | HTML 属性 | payload に含まれる値 |
|---------------|----------|---------------------|
| `gl.Click("evt")` | `gerbera-click="evt"` | `"value"`: `gl.ClickValue()` で設定した値 |
| `gl.Input("evt")` | `gerbera-input="evt"` | `"value"`: 入力フィールドの現在の値 |
| `gl.Change("evt")` | `gerbera-change="evt"` | `"value"`: 変更後の値 |
| `gl.Submit("evt")` | `gerbera-submit="evt"` | フォーム内の全フィールド（name → value） |
| `gl.Keydown("evt")` | `gerbera-keydown="evt"` | `"key"`: 押されたキー名 |
| `gl.Focus("evt")` | `gerbera-focus="evt"` | なし |
| `gl.Blur("evt")` | `gerbera-blur="evt"` | なし |
| `gl.Scroll("evt")` | `gerbera-scroll="evt"` | `scrollTop`, `scrollHeight`, `clientHeight` 等 |

### ClickValue でデータを渡す

リスト内の各アイテムにイベントを付ける場合は `gl.ClickValue()` を使います:

```go
ge.Each(posts, func(i int, post Post) g.ComponentFunc {
    id := strconv.FormatInt(post.ID, 10)
    return gd.Div(
        gp.Key(id),
        gd.P(gp.Value(post.Content)),
        gd.Button(
            gl.Click("toggleLike"),
            gl.ClickValue(id),   // payload["value"] にセットされる
            gp.Value("Like"),
        ),
    )
})
```

HandleEvent 側:

```go
case "toggleLike":
    postID, _ := strconv.ParseInt(payload["value"], 10, 64)
    if postID == 0 {
        return nil  // 不正な値は早期 return
    }
    // ...
```

### Debounce で送信頻度を制御する

検索入力やテキスト入力では `gl.Debounce()` でサーバーへの送信を間引きます:

```go
gd.Input(
    gp.Type("text"),
    gl.Input("searchInput"),
    gl.Debounce(300),           // 300ms 待ってから送信
    gp.Placeholder("検索..."),
)
```

> **注意**: Debounce を設定しないと、1 文字入力するたびにサーバーにイベントが送信されます。テキスト入力系のイベントには **必ず Debounce を付けましょう**。

### Keydown のフィルタリング

特定のキーのみに反応するには `gl.Key()` を併用します:

```go
gd.Input(
    gl.Keydown("chatSend"),
    gl.Key("Enter"),            // Enter キーのみ送信
    gl.Input("chatInput"),      // 入力値は別イベントで取得
)
```

### フォーム送信

`gl.Submit()` はフォーム内の全フィールドを自動収集します:

```go
gd.Form(
    gl.Submit("saveProfile"),
    gd.Input(gp.Name("email"), gp.Type("email"), gp.Value(v.email)),
    gd.Input(gp.Name("name"), gp.Value(v.name)),
    gd.Button(gp.Type("submit"), gp.Value("保存")),
)
```

```go
case "saveProfile":
    email := payload["email"]
    name := payload["name"]
    // ...
```

---

## 5. URL 管理（PushPatch / HandleParams）

### 概要

LiveView のページ遷移はデフォルトではブラウザの URL を変更しません。`PushPatch` を使うと、ページ遷移時に URL を同期でき、ブラウザの戻る/進む操作やブックマークに対応できます。

### 3 つのコマンド

| コマンド | 用途 | ブラウザ履歴 |
|---------|------|-------------|
| `v.PushPatch(path)` | ページ内遷移（URL 変更 + 履歴追加） | 追加される |
| `v.ReplacePatch(path)` | URL 置換（ソート変更など、戻るに残さない場合） | 置換される |
| `v.PushNavigate(path)` | 別ページへのフルナビゲーション（WebSocket 切断） | 追加される |

### Patcher インターフェース

ブラウザの戻る/進むボタンに対応するには `Patcher` インターフェースを実装します:

```go
// Patcher を実装すると、popstate イベント時に HandleParams が呼ばれる
func (v *MyView) HandleParams(path string, params url.Values) error {
    // path: URL パス（例: "/profile"）
    // params: クエリパラメータ（例: "id=123"）
    switch path {
    case "/":
        v.page = "home"
        return v.loadTimeline()
    case "/profile":
        v.page = "profile"
        id, _ := strconv.ParseInt(params.Get("id"), 10, 64)
        return v.loadProfile(id)
    case "/search":
        v.page = "search"
        v.keyword = params.Get("keyword")
        return v.doSearch(v.keyword)
    }
    return nil
}
```

### HandleEvent 内での PushPatch 呼び出し

ページ遷移を行うイベントでは、状態を変更した後に `PushPatch` を呼びます:

```go
case "nav":
    v.page = payload["value"]
    // ... 状態を変更 ...
    v.PushPatch(v.buildPath())  // URL を同期

case "viewPost":
    postID, _ := strconv.ParseInt(payload["value"], 10, 64)
    v.page = "post"
    v.loadPostDetail(postID)
    v.PushPatch(v.buildPath())  // 状態変更後に呼ぶ
```

### buildPath ヘルパーのパターン

View の状態から URL を構築する関数を用意すると一貫性を保てます:

```go
func (v *MyView) buildPath() string {
    switch v.page {
    case "home":
        return "/"
    case "profile":
        q := url.Values{}
        if v.profileUser != nil && v.profileUser.ID != v.currentUserID {
            q.Set("id", strconv.FormatInt(v.profileUser.ID, 10))
        }
        if len(q) == 0 {
            return "/profile"
        }
        return "/profile?" + q.Encode()
    case "search":
        if v.keyword != "" {
            return "/search?keyword=" + url.QueryEscape(v.keyword)
        }
        return "/search"
    default:
        return "/" + v.page
    }
}
```

### Mount での初回 URL パス処理

`Mount` では `params.Path` を使って初期ページを判定します:

```go
func (v *MyView) Mount(params gl.Params) error {
    // URL パスから初期ページを決定
    switch params.Path {
    case "/profile":
        v.page = "profile"
    case "/settings":
        v.page = "settings"
    default:
        v.page = "home"
    }
    // クエリパラメータも読み取れる
    if id := params.Get("id"); id != "" {
        // ...
    }
    return v.loadPageData()
}
```

> **注意**: `PushPatch` で変更した URL に直接アクセスできるよう、サーバー側のルーティングでそのパスを LiveView ハンドラにマッチさせる必要があります。Go の `http.ServeMux` では `mux.Handle("/", handler)` がキャッチオールとして機能します。

---

## 6. フォームとバリデーション

### live.Field / live.SelectField

`live/` パッケージにはアクセシブルなフォームフィールドのヘルパーがあります:

```go
gl.Field("email", gl.FieldOpts{
    Type:        "email",
    Label:       "メールアドレス",
    Placeholder: "you@example.com",
    Value:       v.email,
    Required:    true,
    Event:       "validateEmail",  // LiveView 入力イベント名
    Error:       v.errors.ErrorFor("email"),
})
```

### バリデーションエラーの管理

`live.Errors` 型でバリデーションエラーを管理できます:

```go
type MyView struct {
    gl.CommandQueue
    errors gl.Errors
    email  string
}

func (v *MyView) HandleEvent(event string, payload gl.Payload) error {
    switch event {
    case "validateEmail":
        v.email = payload["value"]
        v.errors.Clear()
        if !strings.Contains(v.email, "@") {
            v.errors.AddError("email", "有効なメールアドレスを入力してください")
        }
    case "submit":
        if v.errors.HasErrors() {
            return nil  // エラーがあれば送信しない
        }
        // 保存処理...
    }
    return nil
}
```

### リアルタイムバリデーション

入力中にバリデーションを行う場合は、`gl.Input` + `gl.Debounce` を組み合わせます:

```go
gd.Input(
    gp.Name("email"),
    gp.Type("email"),
    gp.Value(v.email),
    gl.Input("validateEmail"),
    gl.Debounce(500),  // 入力が止まって 500ms 後にバリデーション
)
```

---

## 7. リアルタイム通信

### TickerView — 定期更新

サーバーサイドで定期的にデータをポーリングする場合に使います（例: 時計、ダッシュボード）:

```go
type DashboardView struct {
    stats Stats
}

func (v *DashboardView) TickInterval() time.Duration {
    return 5 * time.Second  // 5 秒ごとに更新
}

func (v *DashboardView) HandleTick() error {
    v.stats = fetchLatestStats()  // DB やAPIからデータ取得
    return nil
    // return 後、自動で Render → Diff → パッチ送信
}
```

> **注意**: `TickInterval()` が `0` を返すとティック無効。間隔が短すぎると CPU 負荷が上がるため、用途に応じて適切な間隔を設定してください。

### InfoReceiver — ユーザー間通信

他のユーザーのアクションをリアルタイムに通知する場合は `InfoReceiver` + `Session.SendInfo()` を使います:

```go
// 1. 型付きメッセージを定義
type NewMessageNotif struct {
    SenderName string
    Content    string
}

// 2. View に InfoReceiver を実装
func (v *ChatView) HandleInfo(msg any) error {
    switch m := msg.(type) {
    case NewMessageNotif:
        v.messages = append(v.messages, m)
        v.ScrollIntoPct("#messages", "1.0")  // 最下部にスクロール
    }
    return nil
}

// 3. 送信側で SendInfo を呼ぶ
func (v *ChatView) HandleEvent(event string, payload gl.Payload) error {
    if event == "send" {
        // 相手の LiveView セッションに送信
        partnerSession.SendInfo(NewMessageNotif{
            SenderName: v.user.Name,
            Content:    payload["value"],
        })
    }
    return nil
}
```

### Unmounter — 切断時のクリーンアップ

WebSocket が閉じた時にリソースを解放します:

```go
func (v *ChatView) Unmount() {
    hub.Leave(v.sessionID, v.userID)  // チャットルームから退出
}
```

### SessionExpiredHandler — セッション無効化対応

別タブからのログアウトなどでセッションが無効化された場合:

```go
func (v *MyView) OnSessionExpired() error {
    v.showExpiredMessage = true
    v.Navigate("/login")  // クライアントをリダイレクト
    return nil
}
```

---

## 8. JS コマンド（CommandQueue）

### 使い方

`gl.CommandQueue` を View に埋め込むと、サーバーから JavaScript を書かずにクライアントの DOM を操作できます:

```go
type MyView struct {
    gl.CommandQueue
    // ...
}
```

### コマンド一覧

| メソッド | 説明 |
|---------|------|
| `v.ScrollTo(selector, top, left)` | 指定位置にスクロール |
| `v.ScrollIntoPct(selector, pct)` | 0.0〜1.0 の割合でスクロール |
| `v.Focus(selector)` | 要素にフォーカス |
| `v.Blur(selector)` | フォーカスを外す |
| `v.SetAttribute(sel, key, val)` | 属性を設定 |
| `v.RemoveAttribute(sel, key)` | 属性を削除 |
| `v.AddClass(sel, class)` | CSS クラスを追加 |
| `v.RemoveClass(sel, class)` | CSS クラスを削除 |
| `v.ToggleClass(sel, class)` | CSS クラスをトグル |
| `v.SetProperty(sel, key, val)` | DOM プロパティを設定 |
| `v.Dispatch(sel, event)` | カスタムイベントを発火 |
| `v.Show(selector)` | 表示（display=""） |
| `v.Hide(selector)` | 非表示（display="none"） |
| `v.Toggle(selector)` | 表示/非表示をトグル |
| `v.Navigate(url)` | フルページ遷移 |
| `v.PushPatch(path)` | URL 変更（履歴追加） |
| `v.ReplacePatch(path)` | URL 置換（履歴上書き） |
| `v.PushNavigate(path)` | フルページ遷移（WS 切断） |

### 使用例

```go
func (v *MyView) HandleEvent(event string, payload gl.Payload) error {
    switch event {
    case "sendMessage":
        // メッセージ送信後、チャットエリアを最下部にスクロール
        v.messages = append(v.messages, newMsg)
        v.ScrollIntoPct("#chat-messages", "1.0")

    case "openModal":
        v.modalOpen = true
        v.Focus("#modal-input")  // モーダル内の入力にフォーカス

    case "search":
        v.doSearch(payload["value"])
        v.PushPatch("/search?keyword=" + url.QueryEscape(payload["value"]))
    }
    return nil
}
```

> **注意**: コマンドは `HandleEvent` / `HandleTick` / `HandleInfo` の return 後に、パッチと一緒にクライアントへ送信されます。複数のコマンドを同一イベント内でキューに入れることが可能です。

---

## 9. スタイリング

### CSS の管理方法

Gerbera では CSS を Go の文字列定数として管理します:

```go
// css.go
const appCSS = `
.container { max-width: 800px; margin: 0 auto; padding: 16px; }
.card { border: 1px solid var(--g-border); border-radius: 8px; padding: 16px; }
`

// pages.go の Render 内
gd.Head(
    gu.Theme(),              // Gerbera テーマ変数を注入
    gs.CSS(appCSS),          // CSS を <style> として挿入
)
```

### テーマシステム

`ui/` パッケージのテーマシステムでは CSS カスタムプロパティが定義されます:

```go
gu.Theme()                                // デフォルトライトテーマ
gu.ThemeWith(gu.DarkTheme())              // ダークテーマ
gu.ThemeWith(gu.ThemeConfig{              // カスタムテーマ
    Primary: "#6366f1",
    Radius:  "12px",
})
gu.ThemeAuto(gu.ThemeConfig{}, gu.DarkTheme())  // OS設定に連動
```

自分のCSS内でテーマ変数を使うと一貫したデザインになります:

```css
.my-card {
    background: var(--g-surface);
    color: var(--g-text);
    border: 1px solid var(--g-border);
    border-radius: var(--g-radius);
}
```

### ScopedCSS

コンポーネント固有のスタイルは `ScopedCSS` でスコープを限定できます:

```go
gs.ScopedCSS(".my-widget", map[string]string{
    "padding":    "16px",
    "background": "var(--g-surface)",
})
// 出力: .my-widget { padding: 16px; background: var(--g-surface); }
```

### インラインスタイル

動的なスタイルは `styles.Style` を使います:

```go
gs.Style(g.StyleMap{
    "width":  fmt.Sprintf("%d%%", progress),
    "height": "4px",
})

// 条件付き
gs.StyleIf(isHighlighted, g.StyleMap{
    "background": "#fef3c7",
    "border-left": "4px solid #f59e0b",
})
```

### レスポンシブデザイン

メディアクエリは `styles.MediaQuery` で:

```go
gs.CSS(`
    .sidebar { display: none; }
    .mobile-header { display: flex; }
`)
gs.MediaQuery("(min-width: 769px)", `
    .sidebar { display: flex; }
    .mobile-header { display: none; }
`)
```

---

## 10. パフォーマンス最適化

### Key を付ける（最重要）

リスト内の要素には必ず `property.Key()` を付けてください。差分比較アルゴリズムが要素の同一性を判定するために使われます:

```go
// Good
gd.Div(
    gp.Key(fmt.Sprintf("user-%d", user.ID)),
    gp.Value(user.Name),
)

// Bad — Key がないと全要素が再構築される可能性がある
gd.Div(gp.Value(user.Name))
```

### Render の出力を安定させる

Diff アルゴリズムは「何も変わっていない」場合に空のパッチを返します。Render が毎回同じツリーを返せば、不要な DOM 更新は発生しません:

```go
// Good: 状態が変わらなければ同じツリーが返る
func (v *MyView) Render() []g.ComponentFunc {
    return []g.ComponentFunc{
        gd.Head(gd.Title("App")),
        gd.Body(gd.H1(gp.Value(fmt.Sprintf("Count: %d", v.count)))),
    }
}

// Bad: ランダムな値や現在時刻を Render 内で使うと毎回パッチが発生
func (v *MyView) Render() []g.ComponentFunc {
    return []g.ComponentFunc{
        gd.Body(gd.P(gp.Value(time.Now().String()))),  // 毎回変わる!
    }
}
```

### HandleEvent は軽く保つ

HandleEvent はクライアントからのイベントごとに呼ばれます。重い処理は避けてください:

```go
// Good: 必要最小限のデータ取得
case "toggleLike":
    dbToggleLike(v.db, v.userID, postID)
    v.refreshCurrentPosts()  // 現在表示中のページだけ再取得

// Bad: 全データを毎回再取得
case "toggleLike":
    dbToggleLike(v.db, v.userID, postID)
    v.loadAllData()  // 不要なページのデータまで取得
```

### 不要な Debounce の省略を避ける

入力系イベント（`gl.Input`）には `gl.Debounce()` を付けるのが原則です:

```go
// Good
gd.Input(gl.Input("search"), gl.Debounce(300))

// Bad: 1文字ごとにサーバーへ送信
gd.Input(gl.Input("search"))
```

---

## 11. テスト

### TestView を使ったユニットテスト

`live.TestView` を使えば WebSocket なしで LiveView をテストできます:

```go
func TestCounterIncrement(t *testing.T) {
    tv := gl.NewTestView(&CounterView{})
    if err := tv.Mount(gl.Params{}); err != nil {
        t.Fatal(err)
    }

    // イベントをシミュレート
    patches, err := tv.SimulateEvent("inc", gl.Payload{})
    if err != nil {
        t.Fatal(err)
    }

    // パッチが生成されたことを確認
    if tv.PatchCount() == 0 {
        t.Error("パッチが生成されていない")
    }

    // テキスト更新パッチが含まれることを確認
    if !tv.HasPatch(diff.OpSetText) {
        t.Error("テキスト更新パッチがない")
    }

    // HTML 出力を確認
    html, _ := tv.RenderHTML()
    if !strings.Contains(html, "Count: 1") {
        t.Errorf("期待: 'Count: 1', 実際: %s", html)
    }
}
```

### SimulateParams で URL ナビゲーションをテスト

`Patcher` を実装した View のテスト:

```go
func TestHandleParams(t *testing.T) {
    tv := gl.NewTestView(&MyView{})
    tv.Mount(gl.Params{})

    // ブラウザの戻るボタンをシミュレート
    params := url.Values{"id": []string{"42"}}
    patches, err := tv.SimulateParams("/profile", params)
    if err != nil {
        t.Fatal(err)
    }

    // ビューの状態を検証
    view := tv.View.(*MyView)
    if view.page != "profile" {
        t.Errorf("expected page=profile, got %s", view.page)
    }
}
```

### SimulateTick / SimulateInfo

```go
// TickerView のテスト
patches, err := tv.SimulateTick()

// InfoReceiver のテスト
patches, err := tv.SimulateInfo(NewMessageNotif{Content: "hello"})
```

### TestTransport で ViewLoop をテスト

より統合的なテストには `TestTransport` を使います:

```go
func TestViewLoop(t *testing.T) {
    view := &MyView{}
    view.Mount(gl.Params{})

    tr := gl.NewTestTransport()
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        gl.ViewLoop(view, tr, gl.ViewLoopConfig{
            SessionID: "test", Lang: "en",
        })
    }()

    // イベントを注入
    tr.PushEvent("inc", gl.Payload{})
    time.Sleep(50 * time.Millisecond)

    // 送信されたメッセージを確認
    msg := tr.LastMessage()
    if len(msg.Patches) == 0 {
        t.Error("パッチが送信されていない")
    }

    tr.Close()
    wg.Wait()
}
```

---

## 12. セキュリティ

### XSS 対策

`property.Attr()` は自動で `html.EscapeString` を適用します。**ユーザー入力を `Literal()` に直接渡さないでください**:

```go
// Good: 自動エスケープ
gp.Value(userInput)
gp.Attr("title", userInput)

// DANGEROUS: エスケープされない！
g.Literal(userInput)
g.Literal(fmt.Sprintf("<div>%s</div>", userInput))
```

`Literal()` は信頼できる HTML（SVG アイコンなど）にのみ使用してください。

### CSRF 対策

静的フォーム（非 LiveView）では `session` パッケージの CSRF ヘルパーを使います:

```go
token := session.GenerateCSRFToken(sess)
// フォームに hidden フィールドとして含める

// 送信時に検証
if !session.ValidCSRFToken(sess, submittedToken) {
    http.Error(w, "CSRF token invalid", http.StatusForbidden)
    return
}
```

LiveView の WebSocket は接続時に自動で CSRF トークンを検証するため、`HandleEvent` 内での追加チェックは不要です。

### セッション管理

```go
// セッション認証ガード — 未認証ユーザーをリダイレクト
authGuard := session.RequireKey("user_id", "/login")
mux.Handle("/", authGuard(gl.Handler(factory, gl.WithSessionStore(store))))
```

### WebSocket Origin チェック

デフォルトで Origin ヘッダーの検証が行われます。特殊な環境では `WithCheckOrigin` でカスタマイズ可能ですが、**本番環境でチェックを無効にしないでください**。

---

## 13. デバッグ

### デバッグパネル

`gl.WithDebug()` を有効にすると、ブラウザに DevPanel が表示されます:

```go
gl.Handler(factory, gl.WithDebug())
```

- `Ctrl+Shift+D` または右下の **"G" ボタン** で開閉
- **Events タブ**: 送信イベントの名前・ペイロード・タイムスタンプ
- **Patches タブ**: 受信 DOM パッチの件数とサーバー処理時間
- **State タブ**: View 構造体の現在の状態を JSON 表示
- **Session タブ**: セッション ID・TTL・接続状態

> **注意**: `WithDebug()` は開発時のみ使用してください。本番環境では無効にすることで、デバッグ JS の注入やタイミング計測のオーバーヘッドがゼロになります。

### サーバーサイドログ

デバッグモード時は `log/slog` による構造化ログも出力されます:

```
[INFO] session_created  session_id=abc123
[INFO] event_received   session_id=abc123 event=inc payload=map[]
[INFO] patches_generated session_id=abc123 count=1 duration=0.2ms
```

---

## 14. アンチパターン集

### Render 内での副作用

```go
// Bad: Render 内で DB クエリを実行
func (v *MyView) Render() []g.ComponentFunc {
    posts, _ := dbGetPosts(v.db)  // ここで取得しない！
    return []g.ComponentFunc{ ... }
}

// Good: HandleEvent / Mount / HandleTick でデータを取得し、フィールドに保持
func (v *MyView) HandleEvent(event string, _ gl.Payload) error {
    v.posts, _ = dbGetPosts(v.db)
    return nil
}
```

Render は純粋にビューの構築のみを行い、副作用（DB アクセス、ファイル I/O など）は状態更新のタイミング（Mount, HandleEvent, HandleTick, HandleInfo）で実行してください。

### 不必要な状態の重複

```go
// Bad: 計算可能な値をフィールドに持つ
type MyView struct {
    items    []Item
    count    int      // len(items) で計算可能
    hasItems bool     // len(items) > 0 で計算可能
}

// Good: Render 内で計算
func (v *MyView) Render() []g.ComponentFunc {
    count := len(v.items)
    return []g.ComponentFunc{
        gd.Body(
            gd.P(gp.Value(fmt.Sprintf("%d items", count))),
            ge.If(count > 0, renderItems(v.items)),
        ),
    }
}
```

### HandleEvent での error 無視

```go
// Bad: エラーを握りつぶす
case "save":
    dbSave(v.db, data)  // エラーチェックなし

// Good: エラーをユーザーに通知
case "save":
    if err := dbSave(v.db, data); err != nil {
        v.showToast("保存に失敗しました", "danger")
        return nil  // err を返すと HandleEvent のエラーとして記録される
    }
    v.showToast("保存しました", "success")
```

`HandleEvent` が `error` を返すと処理は中断されますが、**クライアントへのパッチ送信もスキップされます**。ユーザーにエラーを見せたい場合は、View の状態にエラーを反映して `nil` を返してください。

### Literal() の濫用

```go
// Bad: HTML 全体を Literal で書く
g.Literal(fmt.Sprintf(`<div class="card"><h2>%s</h2><p>%s</p></div>`, title, body))

// Good: Gerbera の型安全な API を使う
gd.Div(
    gp.Class("card"),
    gd.H2(gp.Value(title)),
    gd.P(gp.Value(body)),
)
```

`Literal()` は XSS の危険があり、Diff アルゴリズムによる最適化も効きません。SVG インラインアイコンなど、信頼できる固定文字列にのみ使ってください。

### 巨大な HandleEvent

```go
// Bad: 1 つの switch 文に全イベントを書く（数百行）
func (v *MyView) HandleEvent(event string, payload gl.Payload) error {
    switch event {
    case "nav": // 20 行
    case "save": // 30 行
    case "delete": // 25 行
    // ... 30 個の case が続く
    }
}

// Good: イベントをカテゴリ別のメソッドに委譲
func (v *MyView) HandleEvent(event string, payload gl.Payload) error {
    switch {
    case strings.HasPrefix(event, "nav"):
        return v.handleNavEvent(event, payload)
    case strings.HasPrefix(event, "chat"):
        return v.handleChatEvent(event, payload)
    default:
        return v.handleGeneralEvent(event, payload)
    }
}
```

---

## 参考リンク

- [README.ja.md](README.ja.md) — プロジェクト概要とクイックスタート
- [ui/README.ja.md](ui/README.ja.md) — UI コンポーネントライブラリ API リファレンス
- [example/counter/TUTORIAL.ja.md](example/counter/TUTORIAL.ja.md) — LiveView 入門チュートリアル
- [example/sns/TUTORIAL.ja.md](example/sns/TUTORIAL.ja.md) — フル機能 SNS アプリケーションチュートリアル
- [example/chat/TUTORIAL.ja.md](example/chat/TUTORIAL.ja.md) — リアルタイムチャット（InfoReceiver）
- [example/auth/TUTORIAL.ja.md](example/auth/TUTORIAL.ja.md) — セッション認証と CSRF
- [example/jscommand/TUTORIAL.ja.md](example/jscommand/TUTORIAL.ja.md) — JS コマンド活用例
