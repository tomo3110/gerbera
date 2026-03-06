# 管理画面ダッシュボード チュートリアル

## 概要

Gerbera の UI コンポーネントを組み合わせた、管理画面ダッシュボードを構築します。本アプリケーションは **SSR + LiveMount ハイブリッドアーキテクチャ** を採用しています:

- **SSR レイアウトページ** — 各ページは個別の URL で `g.Handler()` による静的 HTML レンダリング
- **`gl.LiveMount()`** — SSR ページ内に LiveView コンポーネントを埋め込むマウントポイント
- **複数の小さな View 構造体** — `DashboardView`, `UsersView`, `MessagesView`, `SettingsView` がそれぞれ独立
- `gu.Theme()` — 組み込みデザインシステムテーマ（CSS 変数 + リセット）
- `gu.AdminShell()` — サイドバー + コンテンツのレイアウト骨格
- `gu.Sidebar()` / `gu.SidebarLink()` — 通常の `<a>` タグによるナビゲーションサイドバー
- `gu.StatCard()` / `gu.Grid()` — レスポンシブグリッド内のダッシュボード指標カード
- `gu.Calendar()` — 月・年ナビゲーション付きインタラクティブカレンダー
- `gul.DataTable()` — ソート・ページネーション対応データテーブル（LiveView 対応）
- `gul.Modal()` / `gul.Confirm()` — モーダルダイアログと確認ダイアログ
- `gu.ChatContainer()` / `gu.ChatInput()` — リアルタイムチャットインターフェース
- `gu.NumberInput()` / `gu.Slider()` — 増減イベント付きフォームコントロール
- `gul.Toast()` — 閉じられる通知トースト

旧アーキテクチャでは単一の巨大な `AdminView` が全ページを WebSocket SPA ナビゲーション（`gl.Click("nav")`）で管理していました。新アーキテクチャでは、ページ遷移は通常の `<a>` リンクによるフルページロード（SSR）で行い、各ページ内のインタラクティブ部分のみを LiveView で処理します。

## 前提条件

- Go 1.22 以上がインストール済み
- このリポジトリがローカルにクローン済み

## Step 1: パッケージのインポート

```go
import (
    "context"
    "flag"
    "fmt"
    "log"
    "net/http"
    "strings"
    "time"

    g "github.com/tomo3110/gerbera"
    gd "github.com/tomo3110/gerbera/dom"
    "github.com/tomo3110/gerbera/expr"
    gl "github.com/tomo3110/gerbera/live"
    gp "github.com/tomo3110/gerbera/property"
    gu "github.com/tomo3110/gerbera/ui"
    gul "github.com/tomo3110/gerbera/ui/live"
)
```

7 つのエイリアスインポートがこのアプリケーションの語彙を構成します:

- `g` (gerbera) — コア型: `ComponentFunc`、`Handler()`
- `gd` (dom) — HTML 要素ヘルパー: `Head()`、`Body()`、`Div()`、`Tr()`、`Td()` など
- `expr` — 制御フロー: `If()` による条件付きレンダリング
- `gl` (live) — LiveView ランタイム: `View` インターフェース、`Handler()`、`LiveMount()`、`Click()`、`Payload`
- `gp` (property) — 属性セッター: `Class()`、`ID()`、`Attr()`、`Value()`、`Readonly()`
- `gu` (ui) — 静的 UI コンポーネント: `Theme()`、`AdminShell()`、`Sidebar()`、`Grid()`、`StatCard()`、`Calendar()`、`Card()`、`Button()`、`FormGroup()`、`NumberInput()`、`Slider()`、`ChatContainer()`、`ChatInput()` など
- `gul` (ui/live) — LiveView 対応 UI コンポーネント: `DataTable()`、`Modal()`、`Confirm()`、`Toast()`

`gu` と `gul` の区別は重要です。`gu` コンポーネントは LiveView に依存しない純粋なレンダリング関数であるのに対し、`gul` コンポーネントは LiveView のイベント属性を出力し、機能するには WebSocket 接続が必要です。

なお、新アーキテクチャでは `context` パッケージが追加されています。これは `gl.Handler()` のファクトリ関数が `context.Context` を受け取るためです。

## Step 2: SSR レイアウトページ

```go
func adminPage(page string, liveEndpoint string) []g.ComponentFunc {
    return []g.ComponentFunc{
        gd.Head(
            gd.Title("Admin Panel"),
            gu.Theme(),
        ),
        gd.Body(
            gu.AdminShell(
                sidebar(page),
                gd.Div(gp.Class("g-page-body"),
                    gl.LiveMount(liveEndpoint),
                ),
            ),
        ),
    }
}
```

これが新アーキテクチャの中核です。`adminPage` は SSR で配信される静的なレイアウトページを生成する関数です。

- **`page`** パラメータは現在のページ名で、サイドバーのアクティブ状態制御に使用
- **`liveEndpoint`** パラメータは埋め込む LiveView のエンドポイントパス（例: `"/admin/live/dashboard"`）
- **`gu.Theme()`** は Gerbera デザインシステム CSS（色・スペーシング・タイポグラフィの CSS 変数と CSS リセット）を注入
- **`gu.AdminShell(sidebar, content)`** は 2 カラムレイアウトを作成: 左側に固定サイドバー、右側にスクロール可能なコンテンツエリア
- **`gl.LiveMount(liveEndpoint)`** がポイントです。これは SSR ページ内に LiveView コンポーネントのマウントポイントを作成します。ブラウザがこのページを読み込むと、指定されたエンドポイントへ自動的に WebSocket 接続を確立し、LiveView を起動します

旧アーキテクチャでは `Render()` メソッド内で全ページのレイアウトとコンテンツを一体で返していましたが、新アーキテクチャではレイアウト（SSR）とインタラクティブコンテンツ（LiveView）を明確に分離しています。

## Step 3: サイドバーナビゲーション

```go
func sidebar(active string) g.ComponentFunc {
    return gu.Sidebar(
        gu.SidebarHeader("Admin Panel"),
        gu.SidebarLink("/admin", "Dashboard", active == "dashboard"),
        gu.SidebarLink("/admin/users", "Users", active == "users"),
        gu.SidebarLink("/admin/messages", "Messages", active == "messages"),
        gu.SidebarDivider(),
        gu.SidebarLink("/admin/settings", "Settings", active == "settings"),
    )
}
```

旧アーキテクチャとの最大の違いがここにあります。

- **旧**: `gu.SidebarLink("#", "Dashboard", v.Page == "dashboard", gl.Click("nav"), gl.ClickValue("dashboard"))` — `gl.Click("nav")` イベントで WebSocket 経由の SPA ナビゲーション
- **新**: `gu.SidebarLink("/admin", "Dashboard", active == "dashboard")` — 通常の `<a>` タグによるリンク

サイドバーリンクはもはや LiveView イベントではなく、標準的な HTML リンクです。クリックするとブラウザはフルページロードを行い、サーバーが SSR でページを返します。これにより:

- **`gu.SidebarLink(href, label, active)`** はスタイル付きナビゲーションリンクをレンダリング。第 3 引数がアクティブ／ハイライト状態を制御
- **`gu.SidebarDivider()`** はナビゲーショングループ間の視覚的な区切り線をレンダリング
- ページ遷移時に JavaScript やイベントバインディングは不要
- 各ページは独立した URL を持ち、ブックマークやリンク共有が可能

## Step 4: 複数の View 構造体

旧アーキテクチャでは、1 つの巨大な `AdminView` 構造体がすべてのページの状態を保持していました。新アーキテクチャでは、ページごとに独立した View 構造体を定義します。

### DashboardView

```go
type DashboardView struct {
    CalYear     int
    CalMonth    time.Month
    CalSelected *time.Time
}
```

### UsersView

```go
type UsersView struct {
    TablePage    int
    SortCol      string
    SortDir      string
    ModalOpen    bool
    SelectedUser int
    ConfirmOpen  bool
    DeleteTarget int
    ToastVisible bool
    ToastMessage string
    ToastVariant string
}
```

### MessagesView

```go
type MessagesView struct {
    ChatMessages []gu.ChatMessage
    ChatDraft    string
}
```

### SettingsView

```go
type SettingsView struct {
    ItemsPerPage int
    NotifyFreq   int
}
```

各 View 構造体は自分が担当するページの状態のみを保持します。これにより:

- 各構造体が小さく理解しやすい
- ページ間の状態の干渉がない
- 新しいページの追加は新しい View 構造体の追加だけで完了
- 各 View は独自の `Mount`、`Render`、`HandleEvent` を持ち、責務が明確

各ブラウザが SSR ページを読み込むたびに、`gl.Handler` のファクトリ関数が新しい View インスタンスを生成するため、状態はセッションごとに独立しています。

## Step 5: Mount — 各 View の初期化

各 View は独自の `Mount` メソッドで、そのページに必要な状態のみを初期化します。

### DashboardView の Mount

```go
func (v *DashboardView) Mount(_ gl.Params) error {
    now := time.Now()
    v.CalYear = now.Year()
    v.CalMonth = now.Month()
    return nil
}
```

カレンダーの年月を現在の日付で初期化するだけです。

### UsersView の Mount

```go
func (v *UsersView) Mount(_ gl.Params) error {
    v.SortCol = "name"
    v.SortDir = "asc"
    v.SelectedUser = -1
    v.DeleteTarget = -1
    return nil
}
```

テーブルのソート設定とモーダル用のインデックスを初期化します。`SelectedUser` と `DeleteTarget` は `-1`（未選択）で開始します。

### MessagesView の Mount

```go
func (v *MessagesView) Mount(_ gl.Params) error {
    v.ChatMessages = []gu.ChatMessage{
        {Author: "System", Content: "Welcome to Admin Messages.", Timestamp: "09:00", Sent: false},
        {Author: "Alice Johnson", Content: "The new user report is ready for review.", Timestamp: "09:15", Sent: false, Avatar: "A"},
        {Content: "Thanks, I'll take a look now.", Timestamp: "09:16", Sent: true},
    }
    return nil
}
```

チャット UI のデモ用にサンプルデータを事前設定します。

### SettingsView の Mount

```go
func (v *SettingsView) Mount(_ gl.Params) error {
    v.ItemsPerPage = 10
    v.NotifyFreq = 30
    return nil
}
```

フォームコントロールのデフォルト値を設定します。

旧アーキテクチャでは単一の `Mount` メソッドで 20 以上のフィールドを初期化していましたが、新アーキテクチャでは各 View が 2〜5 フィールドの初期化だけで済みます。

## Step 6: DashboardView の Render

```go
func (v *DashboardView) Render() []g.ComponentFunc {
    return []g.ComponentFunc{
        gd.Body(
            gu.Stack(
                gu.PageHeader("Dashboard",
                    gu.Breadcrumb(gu.BreadcrumbItem{Label: "Dashboard"}),
                ),
                gu.Grid(gu.GridCols4,
                    gu.StatCard("Total Users", fmt.Sprintf("%d", len(users))),
                    gu.StatCard("Active Users", fmt.Sprintf("%d", countActive())),
                    gu.StatCard("Admins", fmt.Sprintf("%d", countByRole("Admin"))),
                    gu.StatCard("Editors", fmt.Sprintf("%d", countByRole("Editor"))),
                ),
                gu.Grid(gu.GridCols2,
                    gu.Card(
                        gu.CardHeader("Recent Activity"),
                        gu.StyledTable(
                            gu.THead("User", "Action", "Time"),
                            gd.Tbody(
                                activityRow("Alice Johnson", "Updated profile", "2 min ago"),
                                // ...
                            ),
                        ),
                    ),
                    gu.Card(
                        gu.CardHeader("Calendar"),
                        gd.Div(gp.Class("g-page-body"),
                            gu.Calendar(gu.CalendarOpts{
                                Year:             v.CalYear,
                                Month:            v.CalMonth,
                                Selected:         v.CalSelected,
                                Today:            time.Now(),
                                SelectEvent:      "calSelect",
                                PrevMonthEvent:   "calPrev",
                                NextMonthEvent:   "calNext",
                                MonthChangeEvent: "calMonthChange",
                                YearChangeEvent:  "calYearChange",
                            }),
                            expr.If(v.CalSelected != nil,
                                gd.Div(gp.Attr("style", "margin-top:8px"),
                                    gu.Badge(v.CalSelected.Format("2006-01-02"), "dark"),
                                ),
                            ),
                        ),
                    ),
                ),
            ),
        ),
    }
}
```

ダッシュボードページでは、いくつかのレイアウトとコンポーネントのパターンを示しています:

- **`gu.PageHeader(title, breadcrumb)`** はパンくずナビゲーション付きのページタイトルをレンダリング
- **`gu.Stack(children...)`** は子要素間に垂直方向のスペーシングを追加
- **`gu.Grid(cols, children...)`** は CSS グリッドを作成。`gu.GridCols4` は 4 列グリッド、`gu.GridCols2` は 2 列グリッド
- **`gu.StatCard(label, value)`** はスタイル付きの指標カードをレンダリング
- **`gu.Card()` / `gu.CardHeader()`** はヘッダー付きのボーダー付きカードコンテナを作成
- **`gu.StyledTable()` / `gu.THead()`** はヘッダー列付きのスタイル付き HTML テーブルをレンダリング
- **`gu.Calendar(CalendarOpts)`** はインタラクティブなカレンダーウィジェットをレンダリング。`Opts` 構造体は各ユーザー操作を名前付きイベントにマッピング: 日付クリックで `"calSelect"` 発火、前後矢印クリックで `"calPrev"`/`"calNext"` 発火、月・年ドロップダウン変更で `"calMonthChange"`/`"calYearChange"` 発火
- **`expr.If(condition, component)`** は日付が選択済みの場合のみ、選択日バッジを条件付きレンダリング

この `Render` メソッドはダッシュボードに関するコンポーネントだけを返します。カレンダーイベントも `DashboardView.HandleEvent` でのみ処理されるため、他のページの状態と混在しません。

## Step 7: UsersView の Render と DataTable

```go
func (v *UsersView) Render() []g.ComponentFunc {
    rows := userRows()
    pageSize := 8
    start := v.TablePage * pageSize
    end := start + pageSize
    if end > len(rows) {
        end = len(rows)
    }
    pageRows := rows[start:end]

    return []g.ComponentFunc{
        gd.Body(
            gu.PageHeader("User Management",
                gu.Breadcrumb(
                    gu.BreadcrumbItem{Label: "Dashboard", Href: "/admin"},
                    gu.BreadcrumbItem{Label: "Users"},
                ),
            ),
            gul.DataTable(gul.DataTableOpts{
                Columns: []gul.Column{
                    {Key: "name", Label: "Name", Sortable: true},
                    {Key: "email", Label: "Email", Sortable: true},
                    {Key: "role", Label: "Role", Sortable: true},
                    {Key: "status", Label: "Status"},
                    {Key: "actions", Label: "Actions"},
                },
                Rows:      pageRows,
                SortCol:   v.SortCol,
                SortDir:   v.SortDir,
                SortEvent: "sort",
                Page:      v.TablePage,
                PageSize:  pageSize,
                Total:     len(users),
                PageEvent: "userPage",
            }),
            renderUserModal(v.ModalOpen, v.SelectedUser),
            gul.Confirm(v.ConfirmOpen, "Delete User",
                "Are you sure you want to delete this user? This action cannot be undone.",
                "doDelete", "cancelDelete"),
            gul.Toast(v.ToastVisible, v.ToastMessage, v.ToastVariant, "dismissToast"),
        ),
    }
}
```

ユーザーページでは複数の LiveView 対応コンポーネントを組み合わせています:

- **`gul.DataTable(DataTableOpts)`** はソート・ページネーション対応テーブルをレンダリング。`Columns` スライスは表示する列とソート可否を定義。`SortEvent` と `PageEvent` はユーザーが列ヘッダーやページネーションコントロールをクリックした際に発火するイベント名。ページネーションは `rows` 配列のスライスによりサーバーサイドで計算
- **`gul.Confirm(open, title, message, confirmEvent, cancelEvent)`** は OK/キャンセルボタン付きの確認ダイアログをレンダリング。ユーザー削除確認に使用
- **`gul.Toast(visible, message, variant, dismissEvent)`** は通知トーストをレンダリング。旧アーキテクチャでは `AdminShell` の外側に配置していましたが、新アーキテクチャでは各 View の `Render` 内に配置

パンくずリストの「Dashboard」リンクが `Href: "/admin"` と実際の URL を指していることに注目してください。旧アーキテクチャでは `Href: "#"` でした。

### ユーザー詳細モーダル

```go
func renderUserModal(open bool, idx int) g.ComponentFunc {
    if idx < 0 || idx >= len(users) {
        return gul.Modal(false, "closeUserModal")
    }
    u := users[idx]
    return gul.Modal(open, "closeUserModal",
        gul.ModalHeader("User Details", "closeUserModal"),
        gul.ModalBody(
            gu.FormGroup(
                gu.FormLabel("Name", "detail-name"),
                gu.FormInput("name", gp.ID("detail-name"), gp.Attr("value", u.Name), gp.Readonly(true)),
            ),
            // ... メール、ロールフィールド ...
        ),
        gul.ModalFooter(
            gu.Button("Close", gu.ButtonOutline, gl.Click("closeUserModal")),
        ),
    )
}
```

新アーキテクチャでは `renderUserModal` は `UsersView` のメソッドではなく、パッケージレベルの関数です。`open` と `idx` をパラメータとして受け取ります。

ガード節に注目してください: ユーザーが未選択（`idx < 0`）の場合でも、非表示のモーダルがレンダリングされます。これは重要で、LiveView の差分エンジンは安定した DOM ツリーを必要とするためです。モーダル要素が完全に消えると、差分エンジンは後からそれを出現させるパッチを適用できなくなります。

## Step 8: MessagesView の Render とチャット

```go
func (v *MessagesView) Render() []g.ComponentFunc {
    var msgViews []g.ComponentFunc
    for _, m := range v.ChatMessages {
        msgViews = append(msgViews, gu.ChatMessageView(m))
    }

    return []g.ComponentFunc{
        gd.Body(
            gu.PageHeader("Messages",
                gu.Breadcrumb(
                    gu.BreadcrumbItem{Label: "Dashboard", Href: "/admin"},
                    gu.BreadcrumbItem{Label: "Messages"},
                ),
            ),
            gu.Card(
                gu.CardHeader("Team Chat"),
                gd.Div(gp.Attr("style", "height:400px;display:flex;flex-direction:column"),
                    gd.Div(gp.Attr("style", "flex:1;overflow-y:auto"),
                        gu.ChatContainer(msgViews...),
                    ),
                    gu.ChatInput("chatMsg", v.ChatDraft, gu.ChatInputOpts{
                        Placeholder:  "Send a message to the team...",
                        SendEvent:    "chatSend",
                        InputEvent:   "chatInput",
                        KeydownEvent: "chatKeydown",
                    }),
                ),
            ),
        ),
    }
}
```

メッセージページでは 3 つのチャット関連コンポーネントを使用しています:

- **`gu.ChatContainer(msgViews...)`** はチャットメッセージをスクロール可能なコンテナで囲む
- **`gu.ChatMessageView(msg)`** は個々のチャットメッセージをレンダリング。`gu.ChatMessage` 構造体には `Author`、`Content`、`Timestamp`、`Sent`（true = 送信済み）、`Avatar` フィールドがある。送信メッセージと受信メッセージは異なるスタイルで表示される
- **`gu.ChatInput(name, value, opts)`** はテキスト入力と送信ボタンをレンダリング。3 つのイベントがバインドされる:
  - `"chatInput"` はキーストロークごとに発火し、`ChatDraft` を更新
  - `"chatSend"` は送信ボタンクリック時に発火
  - `"chatKeydown"` はキーダウンで発火し、Enter キーでの送信を可能にする

## Step 9: SettingsView の Render

```go
func (v *SettingsView) Render() []g.ComponentFunc {
    min0, max100 := 0, 100
    min5, max50 := 5, 50

    return []g.ComponentFunc{
        gd.Body(
            gu.PageHeader("Settings",
                gu.Breadcrumb(
                    gu.BreadcrumbItem{Label: "Dashboard", Href: "/admin"},
                    gu.BreadcrumbItem{Label: "Settings"},
                ),
            ),
            gu.Stack(
                gu.Card(
                    gu.CardHeader("General Settings"),
                    gd.Div(gp.Class("g-page-body"),
                        gu.FormGroup(
                            gu.FormLabel("Site Name", "site-name"),
                            gu.FormInput("site-name", gp.ID("site-name"), gp.Attr("value", "My Admin Panel")),
                        ),
                        gu.FormGroup(
                            gu.FormLabel("Language", "lang"),
                            gu.FormSelect("lang", []gu.FormOption{
                                {Value: "ja", Label: "Japanese"},
                                {Value: "en", Label: "English"},
                            }, gp.ID("lang")),
                        ),
                        gu.Button("Save", gu.ButtonPrimary),
                    ),
                ),
                gu.Card(
                    gu.CardHeader("Display Settings"),
                    gd.Div(gp.Class("g-page-body"),
                        gu.FormGroup(
                            gu.FormLabel("Items per page", "items-per-page"),
                            gu.NumberInput("items-per-page", v.ItemsPerPage, gu.NumberInputOpts{
                                Min:            &min5,
                                Max:            &max50,
                                Step:           5,
                                IncrementEvent: "settingsItemsInc",
                                DecrementEvent: "settingsItemsDec",
                            }),
                        ),
                        gu.FormGroup(
                            gu.FormLabel("Notification frequency (minutes)", "notify-freq"),
                            gu.Slider("notify-freq", v.NotifyFreq, gu.SliderOpts{
                                Min:        min0,
                                Max:        max100,
                                Step:       5,
                                Label:      "Frequency",
                                InputEvent: "settingsNotifyFreq",
                            }),
                        ),
                    ),
                ),
            ),
        ),
    }
}
```

設定ページではインタラクティブなフォームコントロールを示しています:

- **`gu.FormInput()`** / **`gu.FormLabel()`** / **`gu.FormGroup()`** は標準的なフォームフィールドを作成
- **`gu.FormSelect(name, options, attrs...)`** は `FormOption` 構造体で `<select>` ドロップダウンをレンダリング
- **`gu.NumberInput(name, value, opts)`** は +/- ボタン付きの数値入力をレンダリング。`Min` と `Max` は `*int` ポインタ（nil は制限なし）。`IncrementEvent` と `DecrementEvent` は +/- ボタンクリック時に発火
- **`gu.Slider(name, value, opts)`** はレンジスライダーをレンダリング。`InputEvent` はスライダーの値が変更されるたびに発火

`NumberInput` の `Min` と `Max` はポインタとして渡す必要があることに注意してください。ローカル変数（`min5`、`max50`）を宣言し、そのアドレスを取得しています。

## Step 10: イベント処理 — 各 View の HandleEvent

新アーキテクチャでは、旧アーキテクチャの 20 以上のイベントを処理する巨大な単一 `HandleEvent` の代わりに、各 View が自分のイベントだけを処理します。

### DashboardView の HandleEvent

```go
func (v *DashboardView) HandleEvent(event string, payload gl.Payload) error {
    switch event {
    case "calSelect":
        if dateStr := payload["value"]; dateStr != "" {
            if t, err := time.Parse("2006-01-02", dateStr); err == nil {
                v.CalSelected = &t
            }
        }
    case "calPrev":
        v.CalMonth--
        if v.CalMonth < time.January {
            v.CalMonth = time.December
            v.CalYear--
        }
    case "calNext":
        v.CalMonth++
        if v.CalMonth > time.December {
            v.CalMonth = time.January
            v.CalYear++
        }
    case "calMonthChange":
        var m int
        fmt.Sscanf(payload["value"], "%d", &m)
        if m >= 1 && m <= 12 {
            v.CalMonth = time.Month(m)
        }
    case "calYearChange":
        var y int
        fmt.Sscanf(payload["value"], "%d", &y)
        if y > 0 {
            v.CalYear = y
        }
    }
    return nil
}
```

カレンダー関連のイベントのみを処理します。

### UsersView の HandleEvent

```go
func (v *UsersView) HandleEvent(event string, payload gl.Payload) error {
    switch event {
    case "sort":
        col := payload["value"]
        if v.SortCol == col {
            if v.SortDir == "asc" {
                v.SortDir = "desc"
            } else {
                v.SortDir = "asc"
            }
        } else {
            v.SortCol = col
            v.SortDir = "asc"
        }
    case "userPage":
        fmt.Sscanf(payload["value"], "%d", &v.TablePage)
    case "viewUser":
        fmt.Sscanf(payload["value"], "%d", &v.SelectedUser)
        v.ModalOpen = true
    case "closeUserModal":
        v.ModalOpen = false
    case "confirmDeleteUser":
        fmt.Sscanf(payload["value"], "%d", &v.DeleteTarget)
        v.ConfirmOpen = true
    case "doDelete":
        v.ConfirmOpen = false
        v.ToastVisible = true
        v.ToastMessage = "User deleted successfully."
        v.ToastVariant = "success"
    case "cancelDelete":
        v.ConfirmOpen = false
    case "dismissToast":
        v.ToastVisible = false
    }
    return nil
}
```

テーブル操作、モーダル、確認ダイアログ、トーストに関するイベントを処理します。主要なパターン:

- **ソートのトグル** — 同じ列をクリックすると `"asc"` と `"desc"` を切り替え、異なる列をクリックすると `"asc"` にリセット
- **ペイロードの解析** — `fmt.Sscanf(payload["value"], "%d", &v.TablePage)` で文字列ペイロードから整数値を解析
- **トーストのフロー** — 削除確認が `ToastVisible = true` とメッセージ・バリアントの設定でトーストをトリガー

旧アーキテクチャでは `"nav"` イベントハンドラで `ModalOpen`、`ConfirmOpen`、`ToastVisible` をリセットする必要がありましたが、新アーキテクチャではページ遷移がフルページロードのため、このクリーンアップは不要です。ページ遷移時に新しい View インスタンスが生成されるため、古い UI 状態が残る心配がありません。

### MessagesView の HandleEvent

```go
func (v *MessagesView) HandleEvent(event string, payload gl.Payload) error {
    switch event {
    case "chatInput":
        v.ChatDraft = payload["value"]
    case "chatSend", "chatKeydown":
        if event == "chatKeydown" && payload["key"] != "Enter" {
            return nil
        }
        if strings.TrimSpace(v.ChatDraft) != "" {
            v.ChatMessages = append(v.ChatMessages, gu.ChatMessage{
                Content:   v.ChatDraft,
                Timestamp: time.Now().Format("15:04"),
                Sent:      true,
            })
            v.ChatDraft = ""
        }
    }
    return nil
}
```

チャット関連のイベントのみを処理します。送信ボタン（`"chatSend"`）と Enter キー（`"chatKeydown"`）は同じロジックを共有: `Sent: true` の新しい `ChatMessage` を追加し、ドラフトをクリアします。

### SettingsView の HandleEvent

```go
func (v *SettingsView) HandleEvent(event string, payload gl.Payload) error {
    switch event {
    case "settingsItemsInc":
        v.ItemsPerPage += 5
        if v.ItemsPerPage > 50 {
            v.ItemsPerPage = 50
        }
    case "settingsItemsDec":
        v.ItemsPerPage -= 5
        if v.ItemsPerPage < 5 {
            v.ItemsPerPage = 5
        }
    case "settingsNotifyFreq":
        fmt.Sscanf(payload["value"], "%d", &v.NotifyFreq)
    }
    return nil
}
```

NumberInput と Slider のイベントのみを処理します。

## Step 11: サーバー — SSR ページと LiveView エンドポイントの登録

```go
func main() {
    addr := flag.String("addr", ":8910", "listen address")
    debug := flag.Bool("debug", false, "enable debug panel")
    flag.Parse()

    var opts []gl.Option
    if *debug {
        opts = append(opts, gl.WithDebug())
    }

    mux := http.NewServeMux()

    // SSR pages — each embeds a LiveMount for its interactive section
    mux.Handle("GET /admin", g.Handler(adminPage("dashboard", "/admin/live/dashboard")...))
    mux.Handle("GET /admin/users", g.Handler(adminPage("users", "/admin/live/users")...))
    mux.Handle("GET /admin/messages", g.Handler(adminPage("messages", "/admin/live/messages")...))
    mux.Handle("GET /admin/settings", g.Handler(adminPage("settings", "/admin/live/settings")...))

    // LiveView endpoints for LiveMount
    mux.Handle("/admin/live/dashboard", gl.Handler(func(_ context.Context) gl.View { return &DashboardView{} }, opts...))
    mux.Handle("/admin/live/users", gl.Handler(func(_ context.Context) gl.View { return &UsersView{} }, opts...))
    mux.Handle("/admin/live/messages", gl.Handler(func(_ context.Context) gl.View { return &MessagesView{} }, opts...))
    mux.Handle("/admin/live/settings", gl.Handler(func(_ context.Context) gl.View { return &SettingsView{} }, opts...))

    // Redirect root to /admin
    mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "/admin", http.StatusFound)
    })

    log.Printf("admin running on %s", *addr)
    log.Fatal(http.ListenAndServe(*addr, mux))
}
```

これが新アーキテクチャのサーバー設定です。旧アーキテクチャとの違いを順に説明します。

### SSR ページの登録

```go
mux.Handle("GET /admin", g.Handler(adminPage("dashboard", "/admin/live/dashboard")...))
```

- **`g.Handler(components...)`** は `ComponentFunc` のリストから静的 HTML を生成する HTTP ハンドラを作成。`gl.Handler` ではなく `g.Handler` であることに注意 -- LiveView ではなく通常の SSR レンダリングです
- 各ページは `"GET /admin"`, `"GET /admin/users"` など個別の URL パターンで登録
- `adminPage()` 関数が共通レイアウト（Head, Sidebar, LiveMount）を生成

### LiveView エンドポイントの登録

```go
mux.Handle("/admin/live/dashboard", gl.Handler(func(_ context.Context) gl.View { return &DashboardView{} }, opts...))
```

- **`gl.Handler(factory, opts...)`** は LiveView の HTTP ハンドラを作成。WebSocket 接続のアップグレードを処理
- ファクトリ関数 `func(_ context.Context) gl.View { return &DashboardView{} }` は各セッションに新しい View インスタンスを生成
- LiveView エンドポイントのパスは `/admin/live/dashboard` のように、SSR ページの URL（`/admin`）とは別の名前空間に配置
- このエンドポイントは SSR ページ内の `gl.LiveMount("/admin/live/dashboard")` から参照される
- **`gl.WithDebug()`** はデバッグパネルを有効化し、ブラウザで WebSocket メッセージ、イベントペイロード、レンダリングタイミングを確認可能

### ルートリダイレクト

```go
mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/admin", http.StatusFound)
})
```

ルート URL へのアクセスを `/admin` にリダイレクトします。

## 動作の仕組み

```
ブラウザ                                     Go サーバー
  |                                            |
  | 1. GET /admin                              |
  | <-- 静的 HTML（AdminShell + Sidebar        |
  |     + LiveMount プレースホルダ）             |
  |                                            |
  | 2. LiveMount が WebSocket 接続を確立        |
  |    GET /admin/live/dashboard               |
  |                                            |
  |     DashboardView.Mount() 実行             |
  |     DashboardView.Render() 実行            |
  | <-- LiveView コンテンツ（Dashboard）         |
  |                                            |
  | 3. カレンダーの日付をクリック                  |
  | --> {"e":"calSelect","p":{"value":"..."}}  |
  |                                            |
  |     HandleEvent: CalSelected を更新         |
  |     Render() -> 新旧ツリーを Diff            |
  | <-- カレンダー部分のパッチのみ                |
  |                                            |
  | 4. サイドバーの「Users」をクリック             |
  |    （通常の <a> リンク）                      |
  |                                            |
  | 5. GET /admin/users                        |
  | <-- 新しい静的 HTML（AdminShell + Sidebar   |
  |     + LiveMount プレースホルダ）             |
  |                                            |
  | 6. LiveMount が新しい WebSocket 接続を確立   |
  |    GET /admin/live/users                   |
  |                                            |
  |     UsersView.Mount() 実行                 |
  |     UsersView.Render() 実行                |
  | <-- LiveView コンテンツ（Users テーブル）     |
  |                                            |
  | 7. 列ヘッダー「Email」をクリック              |
  | --> {"e":"sort","p":{"value":"email"}}     |
  |                                            |
  |     HandleEvent: SortCol="email"           |
  |     Render() -> Diff -> テーブル行のパッチ    |
  | <-- テーブル行のみのパッチ                    |
```

旧アーキテクチャとの重要な違い:

- **旧**: 単一の WebSocket 接続で SPA 全体が動作。ページ遷移もイベント経由
- **新**: ページ遷移はフルページロード（SSR）。各ページで独立した WebSocket 接続が確立される

このハイブリッドアーキテクチャの利点:

1. **初期表示が高速** — SSR により HTML が即座に配信される。LiveView 部分は後から WebSocket で追加される
2. **URL によるナビゲーション** — 各ページが独自の URL を持ち、ブックマーク、リンク共有、ブラウザの戻る/進むが正常に動作
3. **状態の分離** — 各 View が独立した状態を持つため、ページ遷移時の状態クリーンアップが不要
4. **コードの保守性** — 小さな View 構造体は理解・テスト・拡張が容易
5. **独立したスケーリング** — 各 LiveView エンドポイントは独立した WebSocket 接続を持つため、重いページが他のページに影響しない

## 実行方法

```bash
# 標準モード
go run example/admin/admin.go

# デバッグパネル付き
go run example/admin/admin.go -debug

# カスタムポート
go run example/admin/admin.go -addr :9000
```

http://localhost:8910/admin を開き、ダッシュボード、ユーザーテーブル、チャット、設定ページを確認してください。

## 演習課題

1. **新しいページの追加** — チャートやデータサマリーを含む「Reports」ページを作成する。新しい `ReportsView` 構造体を定義し、`adminPage("reports", "/admin/live/reports")` で SSR ページを登録、`gl.Handler` で LiveView エンドポイントを登録する。サイドバーに `/admin/reports` へのリンクを追加する。
2. **検索フィルター** — DataTable の上に名前やメールでユーザーをフィルタリングするテキスト入力を追加する。`UsersView` に `SearchQuery` フィールドを追加し、`"userSearch"` イベントでバインドし、`Render` 内で `rows` スライスをフィルタリングする。
3. **テーマ切り替え** — サイドバーフッターにダーク/ライトモード切り替えボタンを追加する。テーマ設定は SSR レイアウト側ではなく、各 LiveView の状態として管理するか、Cookie で永続化する方法を検討する。
4. **永続的な削除** — 現在 `"doDelete"` はトーストを表示するのみ。`users` スライスからユーザーを削除し、ページネーションを調整する実際の削除を実装する。
5. **チャットのタイムスタンプ** — 絶対時刻の代わりに相対タイムスタンプ（「2分前」）を表示し、`TickerView` で更新する。
6. **SSR と LiveView の責務分担** — 現在はすべてのコンテンツを LiveView で描画しているが、ダッシュボードの統計カードなど更新頻度の低い部分を SSR 側に移動し、カレンダーのようなインタラクティブ部分のみを LiveView にする構成を試す。
