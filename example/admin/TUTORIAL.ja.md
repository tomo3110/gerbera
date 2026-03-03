# 管理画面ダッシュボード チュートリアル

## 概要

Gerbera の UI コンポーネントを組み合わせた、リアルタイムなシングルページ管理画面ダッシュボードを構築します:

- `gl.View` インターフェース — マルチページ SPA ナビゲーションを持つステートフルな LiveView
- `gu.Theme()` — 組み込みデザインシステムテーマ（CSS 変数 + リセット）
- `gu.AdminShell()` — サイドバー + コンテンツのレイアウト骨格
- `gu.Sidebar()` / `gu.SidebarLink()` — アクティブ状態付きナビゲーションサイドバー
- `gu.StatCard()` / `gu.Grid()` — レスポンシブグリッド内のダッシュボード指標カード
- `gu.Calendar()` — 月・年ナビゲーション付きインタラクティブカレンダー
- `gul.DataTable()` — ソート・ページネーション対応データテーブル（LiveView 対応）
- `gul.Modal()` / `gul.Confirm()` — モーダルダイアログと確認ダイアログ
- `gu.ChatContainer()` / `gu.ChatInput()` — リアルタイムチャットインターフェース
- `gu.NumberInput()` / `gu.Slider()` — 増減イベント付きフォームコントロール
- `gul.Toast()` — 閉じられる通知トースト
- `gl.Click` / `gl.ClickValue` — SPA ページナビゲーション用イベントバインディング

アプリケーション全体が Gerbera 本体以外の外部依存なしの 529 行の単一 Go ファイルで構成されています。

## 前提条件

- Go 1.22 以上がインストール済み
- このリポジトリがローカルにクローン済み

## Step 1: パッケージのインポート

```go
import (
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

6 つのエイリアスインポートがこのアプリケーションの語彙を構成します:

- `g` (gerbera) — コア型: `ComponentFunc`、`Literal()`
- `gd` (dom) — HTML 要素ヘルパー: `Head()`、`Body()`、`Div()`、`Tr()`、`Td()` など
- `expr` — 制御フロー: `If()` による条件付きレンダリング
- `gl` (live) — LiveView ランタイム: `View` インターフェース、`Handler()`、`Click()`、`ClickValue()`、`Payload`
- `gp` (property) — 属性セッター: `Class()`、`ID()`、`Attr()`、`Value()`、`Readonly()`
- `gu` (ui) — 静的 UI コンポーネント: `Theme()`、`AdminShell()`、`Grid()`、`StatCard()`、`Calendar()`、`Card()`、`Button()`、`FormGroup()`、`NumberInput()`、`Slider()`、`ChatContainer()`、`ChatInput()` など
- `gul` (ui/live) — LiveView 対応 UI コンポーネント: `DataTable()`、`Modal()`、`Confirm()`、`Toast()`

`gu` と `gul` の区別は重要です。`gu` コンポーネントは LiveView に依存しない純粋なレンダリング関数であるのに対し、`gul` コンポーネントは LiveView のイベント属性を出力し、機能するには WebSocket 接続が必要です。

## Step 2: View 構造体

```go
type AdminView struct {
    Page          string
    UserTablePage int
    SortCol       string
    SortDir       string
    ModalOpen     bool
    SelectedUser  int
    ConfirmOpen   bool
    DeleteTarget  int
    ToastVisible  bool
    ToastMessage  string
    ToastVariant  string

    // Settings
    ItemsPerPage int
    NotifyFreq   int

    // Calendar
    CalYear     int
    CalMonth    time.Month
    CalSelected *time.Time

    // Messages
    ChatMessages []gu.ChatMessage
    ChatDraft    string
}
```

`AdminView` 構造体は SPA 全体の状態を保持します。フィールドは関心事ごとにグループ化されています:

- **ナビゲーション** — `Page` がレンダリングするページを決定（`"dashboard"`、`"users"`、`"messages"`、`"settings"`）
- **ユーザーテーブル** — `UserTablePage`、`SortCol`、`SortDir` が DataTable のページネーションとソートを制御。`SelectedUser` と `DeleteTarget` がモーダル表示や削除対象のユーザーを追跡
- **モーダル** — `ModalOpen` と `ConfirmOpen` がユーザー詳細モーダルと削除確認ダイアログの表示を切り替え
- **トースト** — `ToastVisible`、`ToastMessage`、`ToastVariant` が通知トーストを制御
- **設定** — `ItemsPerPage` と `NotifyFreq` が NumberInput と Slider にバインドされたフォーム値
- **カレンダー** — `CalYear`、`CalMonth`、`CalSelected` がカレンダーのナビゲーションと日付選択を追跡
- **メッセージ** — `ChatMessages` がチャット履歴を格納。`ChatDraft` が現在の入力テキストを保持

各ブラウザセッションが独自の `AdminView` インスタンスを持つため、すべての状態はセッションごとに独立しています。

## Step 3: Mount

```go
func (v *AdminView) Mount(params gl.Params) error {
    v.Page = "dashboard"
    v.SortCol = "name"
    v.SortDir = "asc"
    v.SelectedUser = -1
    v.DeleteTarget = -1
    v.ItemsPerPage = 10
    v.NotifyFreq = 30
    now := time.Now()
    v.CalYear = now.Year()
    v.CalMonth = now.Month()
    v.ChatMessages = []gu.ChatMessage{
        {Author: "System", Content: "Welcome to Admin Messages.", Timestamp: "09:00", Sent: false},
        {Author: "Alice Johnson", Content: "The new user report is ready for review.", Timestamp: "09:15", Sent: false, Avatar: "A"},
        {Content: "Thanks, I'll take a look now.", Timestamp: "09:16", Sent: true},
    }
    if p := params.Get("page"); p != "" {
        v.Page = p
    }
    return nil
}
```

`Mount` はすべての状態を適切なデフォルト値で初期化します。主なポイント:

- `SelectedUser` と `DeleteTarget` は `-1`（未選択）で開始
- カレンダーは現在の年月をデフォルトに設定
- チャットメッセージはチャット UI のデモ用にサンプルデータで事前設定
- `params.Get("page")` により URL クエリパラメータ経由で特定ページへのディープリンクが可能（例: `?page=users`）

## Step 4: テーマとレイアウト

```go
func (v *AdminView) Render() []g.ComponentFunc {
    return []g.ComponentFunc{
        gd.Head(
            gd.Title("Admin Panel"),
            gu.Theme(),
        ),
        gd.Body(
            gu.AdminShell(
                v.renderSidebar(),
                v.renderContent(),
            ),
            expr.If(v.ToastVisible,
                gul.Toast(v.ToastMessage, v.ToastVariant, "dismissToast"),
            ),
        ),
    }
}
```

`Render()` メソッドは `<head>` と `<body>` の 2 つのトップレベル要素を返します。

- **`gu.Theme()`** は Gerbera デザインシステム CSS（色・スペーシング・タイポグラフィの CSS 変数と CSS リセット）を注入します。この 1 回の呼び出しで視覚的な基盤全体が提供されます。
- **`gu.AdminShell(sidebar, content)`** は 2 カラムレイアウトを作成します: 左側に固定サイドバー、右側にスクロール可能なコンテンツエリア。
- **`gul.Toast()`** はシェルの外側で条件付きレンダリングされ、ページ全体にオーバーレイ表示されます。第 3 引数の `"dismissToast"` はユーザーが閉じるボタンをクリックした際に発火するイベント名です。

## Step 5: サイドバーナビゲーション

```go
func (v *AdminView) renderSidebar() g.ComponentFunc {
    return gu.Sidebar(
        gu.SidebarHeader("Admin Panel"),
        gu.SidebarLink("#", "Dashboard", v.Page == "dashboard",
            gl.Click("nav"), gl.ClickValue("dashboard")),
        gu.SidebarLink("#", "Users", v.Page == "users",
            gl.Click("nav"), gl.ClickValue("users")),
        gu.SidebarLink("#", "Messages", v.Page == "messages",
            gl.Click("nav"), gl.ClickValue("messages")),
        gu.SidebarDivider(),
        gu.SidebarLink("#", "Settings", v.Page == "settings",
            gl.Click("nav"), gl.ClickValue("settings")),
    )
}
```

SPA ナビゲーションは LiveView イベントのみで実装されており、クライアントサイドルーターは不要です。

- **`gu.SidebarLink(href, label, active, attrs...)`** はスタイル付きナビゲーションリンクをレンダリングします。第 3 引数（`v.Page == "dashboard"`）がアクティブ／ハイライト状態を制御します。
- **`gl.Click("nav")`** はクリックイベントをバインドし、`{"e":"nav","p":{"value":"dashboard"}}` をサーバーに送信します。
- **`gl.ClickValue("dashboard")`** はイベントペイロードの `value` フィールドを設定します。
- **`gu.SidebarDivider()`** はナビゲーショングループ間の視覚的な区切り線をレンダリングします。

ユーザーがリンクをクリックすると、`HandleEvent` が `"nav"` イベントを受信し、`v.Page` を更新。次のレンダリングで対応するページコンテンツが表示されます。WebSocket 差分エンジンは変更された DOM ノードのみをパッチするため、ページ遷移はフルページリロードなしに瞬時に行われます。

## Step 6: ダッシュボードページ

```go
func (v *AdminView) pageDashboard() g.ComponentFunc {
    return gd.Div(
        gu.PageHeader("Dashboard",
            gu.Breadcrumb(gu.BreadcrumbItem{Label: "Dashboard", Href: ""}),
        ),
        gd.Div(gp.Class("g-page-body"),
            gu.Stack(
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
    )
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

## Step 7: ユーザーページと DataTable

```go
func (v *AdminView) pageUsers() g.ComponentFunc {
    rows := userRows()
    pageSize := 8
    start := v.UserTablePage * pageSize
    end := start + pageSize
    if end > len(rows) {
        end = len(rows)
    }
    pageRows := rows[start:end]

    return gd.Div(
        gu.PageHeader("User Management",
            gu.Breadcrumb(
                gu.BreadcrumbItem{Label: "Dashboard", Href: "#"},
                gu.BreadcrumbItem{Label: "Users", Href: ""},
            ),
        ),
        gd.Div(gp.Class("g-page-body"),
            gul.DataTable(gul.DataTableOpts{
                Columns: []gul.Column{
                    {Key: "name", Label: "Name", Sortable: true},
                    {Key: "email", Label: "Email", Sortable: true},
                    {Key: "role", Label: "Role", Sortable: true},
                    {Key: "status", Label: "Status", Sortable: false},
                    {Key: "actions", Label: "Actions", Sortable: false},
                },
                Rows:      pageRows,
                SortCol:   v.SortCol,
                SortDir:   v.SortDir,
                SortEvent: "sort",
                Page:      v.UserTablePage,
                PageSize:  pageSize,
                Total:     len(users),
                PageEvent: "userPage",
            }),
        ),
        v.renderUserModal(),
        gul.Confirm(v.ConfirmOpen, "Delete User",
            "Are you sure you want to delete this user? This action cannot be undone.",
            "doDelete", "cancelDelete"),
    )
}
```

ユーザーページでは 3 つの LiveView 対応コンポーネントを組み合わせています:

- **`gul.DataTable(DataTableOpts)`** はソート・ページネーション対応テーブルをレンダリング。`Columns` スライスは表示する列とソート可否を定義。`SortEvent` と `PageEvent` はユーザーが列ヘッダーやページネーションコントロールをクリックした際に発火するイベント名。ページネーションは `rows` 配列のスライスによりサーバーサイドで計算
- **`gul.Modal(open, closeEvent, children...)`** は `open` が `true` のときに表示されるモーダルダイアログをレンダリング。ユーザー詳細モーダルは名前、メール、ロール、ステータスを読み取り専用フォームフィールドで表示
- **`gul.Confirm(open, title, message, confirmEvent, cancelEvent)`** は OK/キャンセルボタン付きの確認ダイアログをレンダリング。ユーザー削除確認に使用

### ユーザー詳細モーダル

```go
func (v *AdminView) renderUserModal() g.ComponentFunc {
    if v.SelectedUser < 0 || v.SelectedUser >= len(users) {
        return gul.Modal(false, "closeUserModal")
    }
    u := users[v.SelectedUser]
    return gul.Modal(v.ModalOpen, "closeUserModal",
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

ガード節に注目してください: ユーザーが未選択（`SelectedUser < 0`）の場合でも、非表示のモーダルがレンダリングされます。これは重要で、LiveView の差分エンジンは安定した DOM ツリーを必要とするためです。モーダル要素が完全に消えると、差分エンジンは後からそれを出現させるパッチを適用できなくなります。

## Step 8: メッセージページとチャット

```go
func (v *AdminView) pageMessages() g.ComponentFunc {
    var msgViews []g.ComponentFunc
    for _, m := range v.ChatMessages {
        msgViews = append(msgViews, gu.ChatMessageView(m))
    }

    return gd.Div(
        gu.PageHeader("Messages",
            gu.Breadcrumb(
                gu.BreadcrumbItem{Label: "Dashboard", Href: "#"},
                gu.BreadcrumbItem{Label: "Messages", Href: ""},
            ),
        ),
        gd.Div(gp.Class("g-page-body"),
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
    )
}
```

メッセージページでは 3 つのチャット関連コンポーネントを使用しています:

- **`gu.ChatContainer(msgViews...)`** はチャットメッセージをスクロール可能なコンテナで囲む
- **`gu.ChatMessageView(msg)`** は個々のチャットメッセージをレンダリング。`gu.ChatMessage` 構造体には `Author`、`Content`、`Timestamp`、`Sent`（true = 送信済み）、`Avatar` フィールドがある。送信メッセージと受信メッセージは異なるスタイルで表示される
- **`gu.ChatInput(name, value, opts)`** はテキスト入力と送信ボタンをレンダリング。3 つのイベントがバインドされる:
  - `"chatInput"` はキーストロークごとに発火し、`ChatDraft` を更新
  - `"chatSend"` は送信ボタンクリック時に発火
  - `"chatKeydown"` はキーダウンで発火し、Enter キーでの送信を可能にする

## Step 9: 設定ページ

```go
func (v *AdminView) pageSettings() g.ComponentFunc {
    min0, max100 := 0, 100
    min5, max50 := 5, 50

    return gd.Div(
        // ... PageHeader + Breadcrumb ...
        gd.Div(gp.Class("g-page-body"),
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
    )
}
```

設定ページではインタラクティブなフォームコントロールを示しています:

- **`gu.FormInput()`** / **`gu.FormLabel()`** / **`gu.FormGroup()`** は標準的なフォームフィールドを作成
- **`gu.FormSelect(name, options, attrs...)`** は `FormOption` 構造体で `<select>` ドロップダウンをレンダリング
- **`gu.NumberInput(name, value, opts)`** は +/- ボタン付きの数値入力をレンダリング。`Min` と `Max` は `*int` ポインタ（nil は制限なし）。`IncrementEvent` と `DecrementEvent` は +/- ボタンクリック時に発火
- **`gu.Slider(name, value, opts)`** はレンジスライダーをレンダリング。`InputEvent` はスライダーの値が変更されるたびに発火

`NumberInput` の `Min` と `Max` はポインタとして渡す必要があることに注意してください。ローカル変数（`min5`、`max50`）を宣言し、そのアドレスを取得しています。

## Step 10: イベント処理

```go
func (v *AdminView) HandleEvent(event string, payload gl.Payload) error {
    switch event {
    // ナビゲーション
    case "nav":
        v.Page = payload["value"]
        v.ModalOpen = false
        v.ConfirmOpen = false
        v.ToastVisible = false

    // ユーザーテーブルのソートとページネーション
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
        fmt.Sscanf(payload["value"], "%d", &v.UserTablePage)

    // ユーザーモーダル
    case "viewUser":
        fmt.Sscanf(payload["value"], "%d", &v.SelectedUser)
        v.ModalOpen = true
    case "closeUserModal":
        v.ModalOpen = false

    // 削除確認
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

    // カレンダー
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

    // チャット
    case "chatInput":
        v.ChatDraft = payload["value"]
    case "chatSend":
        if strings.TrimSpace(v.ChatDraft) != "" {
            v.ChatMessages = append(v.ChatMessages, gu.ChatMessage{
                Content:   v.ChatDraft,
                Timestamp: time.Now().Format("15:04"),
                Sent:      true,
            })
            v.ChatDraft = ""
        }
    case "chatKeydown":
        if payload["key"] == "Enter" && strings.TrimSpace(v.ChatDraft) != "" {
            v.ChatMessages = append(v.ChatMessages, gu.ChatMessage{
                Content:   v.ChatDraft,
                Timestamp: time.Now().Format("15:04"),
                Sent:      true,
            })
            v.ChatDraft = ""
        }

    // 設定
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

`HandleEvent` は SPA 全体の集中イベントルーターです。20 以上のイベントが機能領域ごとにグループ化された単一の `switch` 文で処理されます。主要なパターン:

- **ナビゲーションのクリーンアップ** — `"nav"` ハンドラは `ModalOpen`、`ConfirmOpen`、`ToastVisible` をリセットし、ページ遷移時に古い UI 状態が残ることを防止
- **ソートのトグル** — 同じ列をクリックすると `"asc"` と `"desc"` を切り替え、異なる列をクリックすると `"asc"` にリセット
- **ペイロードの解析** — `fmt.Sscanf(payload["value"], "%d", &v.UserTablePage)` で文字列ペイロードから整数値を解析
- **チャット送信** — 送信ボタン（`"chatSend"`）と Enter キー（`"chatKeydown"`）は同じロジックを共有: `Sent: true` の新しい `ChatMessage` を追加し、ドラフトをクリア
- **トーストのフロー** — 削除確認が `ToastVisible = true` とメッセージ・バリアントの設定でトーストをトリガー。トーストは `"dismissToast"` イベントで閉じられる

## Step 11: サーバー

```go
func main() {
    addr := flag.String("addr", ":8910", "listen address")
    debug := flag.Bool("debug", false, "enable debug panel")
    flag.Parse()

    var opts []gl.Option
    if *debug {
        opts = append(opts, gl.WithDebug())
    }

    http.Handle("/", gl.Handler(func(_ context.Context) gl.View { return &AdminView{} }, opts...))
    log.Printf("admin running on %s", *addr)
    log.Fatal(http.ListenAndServe(*addr, nil))
}
```

サーバーのセットアップは最小限です:

- **`gl.Handler(factory, opts...)`** は初期 HTML を配信し、ライブ更新のために WebSocket にアップグレードする HTTP ハンドラを作成。ファクトリ関数 `func(_ context.Context) gl.View { return &AdminView{} }` は各セッションに新しい `AdminView` を生成。`context.Context` パラメータにより、ミドルウェアで設定した認証ユーザーなどのリクエストスコープの値にアクセス可能
- **`gl.WithDebug()`** はデバッグパネルを有効化し、ブラウザで WebSocket メッセージ、イベントペイロード、レンダリングタイミングを確認可能

## 動作の仕組み

```
ブラウザ                                     Go サーバー
  |                                            |
  | 1. GET /                                   |
  | <-- 完全な HTML（AdminShell + Dashboard）    |
  |                                            |
  | 2. WebSocket 接続                           |
  |                                            |
  | 3. サイドバーの「Users」をクリック              |
  | --> {"e":"nav","p":{"value":"users"}}      |
  |                                            |
  |     HandleEvent: Page = "users"            |
  |     Render() -> 新旧ツリーを Diff            |
  |                                            |
  | 4. <-- [{"op":"replace",                   |
  |          "path":[1,0,1],                   |
  |          "val":"<div>...users page...</div>"}]
  |                                            |
  | 5. JS が DOM をその場でパッチ                 |
  |    （サイドバーのアクティブ状態 + コンテンツ領域）|
  |                                            |
  | 6. 列ヘッダー「Email」をクリック              |
  | --> {"e":"sort","p":{"value":"email"}}     |
  |                                            |
  |     HandleEvent: SortCol="email",SortDir="asc"
  |     Render() -> Diff -> テーブル本体のパッチ  |
  |                                            |
  | 7. <-- テーブル行のみのパッチ                 |
```

重要なポイントは、SPA 全体が単一の WebSocket 接続で動作することです。ページ遷移、モーダル表示、テーブルソート、チャットメッセージのすべてが同じサイクルに従います: イベント -> 状態変更 -> 再レンダリング -> 差分計算 -> 最小限の DOM パッチ。

## 実行方法

```bash
# 標準モード
go run example/admin/admin.go

# デバッグパネル付き
go run example/admin/admin.go -debug

# カスタムポート
go run example/admin/admin.go -addr :9000
```

http://localhost:8910 を開き、ダッシュボード、ユーザーテーブル、チャット、設定ページを確認してください。

## 演習課題

1. **新しいページの追加** — チャートやデータサマリーを含む「Reports」ページを作成する。サイドバーに新しいリンクを追加し、`renderContent()` に新しい case を追加する。
2. **検索フィルター** — DataTable の上に名前やメールでユーザーをフィルタリングするテキスト入力を追加する。`"userSearch"` イベントでバインドし、`gul.DataTable()` に渡す前に `rows` スライスをフィルタリングする。
3. **テーマ切り替え** — サイドバーフッターにダーク/ライトモード切り替えボタンを追加する。テーマ設定を `AdminView` に格納し、`<body>` に条件付きで CSS クラスを追加する。
4. **永続的な削除** — 現在 `"doDelete"` はトーストを表示するのみ。`users` スライスからユーザーを削除し、ページネーションを調整する実際の削除を実装する。
5. **チャットのタイムスタンプ** — 絶対時刻の代わりに相対タイムスタンプ（「2分前」）を表示し、`TickerView` で更新する。
