# Gerbera UI コンポーネントライブラリ

Gerbera 用のビルド済み UI コンポーネント — CSS を書かずに Go 関数の呼び出しだけでプロフェッショナルな UI を構築できます。

[English version](README.md)

## 設計哲学

エンジニアがデザインに悩まされずに、極力気持ちよく Go 言語のコードを書くだけで開発の局面を乗り越えられるようにする。すべてのコンポーネントがセンスの良いデフォルト値、ARIA アクセシビリティ属性、CSS カスタムプロパティによるテーマ対応スタイリングを備えています。

## クイックスタート

```go
import gu "github.com/tomo3110/gerbera/ui"

// <head> にテーマ CSS を配置
gd.Head(gu.Theme())

// <body> でコンポーネントを使用
gd.Body(
    gu.Card(
        gu.CardHeader("Dashboard"),
        gu.Grid(gu.GridCols3,
            gu.StatCard("Users", "1,234"),
            gu.StatCard("Revenue", "$5,678"),
            gu.StatCard("Orders", "890"),
        ),
    ),
)
```

インポート規約:

```go
gu  "github.com/tomo3110/gerbera/ui"       // 静的 + インタラクティブコンポーネント
gul "github.com/tomo3110/gerbera/ui/live"   // LiveView 専用コンポーネント
```

## テーマシステム

すべてのコンポーネントは CSS カスタムプロパティ（`--g-*`）を使用して色、スペーシング、タイポグラフィを管理します。テーマを変更するだけで全コンポーネントの外観が自動的に適応します。

### 関数

| 関数 | 説明 |
|------|------|
| `Theme()` | デフォルトライトテーマ |
| `ThemeWith(ThemeConfig{...})` | カスタムテーマ — ゼロ値フィールドはデフォルトにフォールバック |
| `DarkTheme()` | ダークテーマプリセット（`ThemeConfig` を返す） |
| `ThemeAuto(light, dark)` | `prefers-color-scheme` メディアクエリで自動切替 |

### ThemeConfig フィールド

**背景**: `Bg`, `BgSurface`, `BgOverlay`, `BgInset`
**テキスト**: `Text`, `TextSecondary`, `TextTertiary`, `TextInverse`
**ボーダー**: `Border`, `BorderStrong`
**シャドウ**: `ShadowSm`, `Shadow`, `ShadowLg`
**アクセント**: `Accent`, `AccentLight`, `AccentHover`
**セマンティック**: `Danger`/`DangerBg`/`DangerBorder`/`DangerHover`, `Success`/`SuccessBg`/`SuccessBorder`, `Warning`/`WarningBg`/`WarningBorder`, `Info`/`InfoBg`/`InfoBorder`
**スペーシング**: `SpaceXs`(4px), `SpaceSm`(8px), `SpaceMd`(16px), `SpaceLg`(28px), `SpaceXl`(44px)
**タイポグラフィ**: `Font`, `FontMono`, `Radius`

### CSS 変数

すべて `--g-<フィールド名>` にマッピング（例: `Bg` → `--g-bg`, `Accent` → `--g-accent`）。

### 使用例

```go
// ライトテーマ（デフォルト）
gu.Theme()

// ダークテーマ
gu.ThemeWith(gu.DarkTheme())

// カスタムアクセントカラー
gu.ThemeWith(gu.ThemeConfig{
    Accent:      "#1a73e8",
    AccentLight: "#e8f0fe",
    AccentHover: "#1557b0",
})

// OS 設定に応じた自動ライト/ダーク切替
gu.ThemeAuto(gu.ThemeConfig{}, gu.ThemeConfig{})
```

## レイアウトシステム

### AdminShell

固定サイドバー付きの 2 カラム管理画面レイアウト:

```go
gu.AdminShell(
    gu.Sidebar(
        gu.SidebarHeader("My App"),
        gu.SidebarLink("#", "Dashboard", true),
        gu.SidebarLink("#", "Settings", false),
    ),
    content, // メインコンテンツ領域
)
```

### PageHeader + Breadcrumb

```go
gu.PageHeader("Users",
    gu.Breadcrumb(
        gu.BreadcrumbItem{Label: "Home", Href: "/"},
        gu.BreadcrumbItem{Label: "Users"},
    ),
)
```

### Grid

```go
gu.Grid(gu.GridCols3, items...)            // 3 カラムグリッド
gu.Grid(gu.GridAutoFit("200px"), items...) // 最小 200px で自動フィット
gu.Grid(gu.GridAutoFill("200px"), items...) // 最小 200px で自動フィル
```

カラム定数: `GridCols2`, `GridCols3`, `GridCols4`, `GridCols5`, `GridCols6`

スパン: `GridSpan(2)`, `GridRowSpan(2)`

### フレックスコンテナ

| 関数 | 説明 |
|------|------|
| `Row(children...)` | 水平フレックス、gap 8px |
| `Column(children...)` | 垂直フレックス、gap 8px |
| `Stack(children...)` | 垂直スタック、gap 16px |
| `HStack(children...)` | 水平、中央揃え |
| `VStack(children...)` | 垂直、中央揃え |
| `Center(children...)` | 両軸中央揃え |
| `Container(children...)` | 最大幅 960px、中央配置 |
| `ContainerNarrow(children...)` | 最大幅 640px |
| `ContainerWide(children...)` | 最大幅 1280px |
| `Spacer()` | フレックススペーサー（空きスペースを埋める） |
| `SpaceY(size)` | 垂直スペーシング (xs/sm/md/lg/xl) |

### フレックス修飾子

**Gap**: `GapNone`, `GapXs`, `GapSm`, `GapMd`, `GapLg`, `GapXl`
**Justify**: `JustifyStart`, `JustifyCenter`, `JustifyEnd`, `JustifyBetween`, `JustifyAround`
**Align**: `AlignStart`, `AlignCenter`, `AlignEnd`, `AlignStretch`, `AlignBaseline`
**その他**: `Wrap`, `Grow`, `Shrink0`

## コンポーネントリファレンス

### card.go — Card, CardHeader, CardFooter

```go
gu.Card(
    gu.CardHeader("Title", actionButton),
    content,
    gu.CardFooter(footerContent),
)
```

### button.go — Button, ButtonPrimary, ButtonOutline, ButtonDanger, ButtonSmall

```go
gu.Button("Default")
gu.Button("Primary", gu.ButtonPrimary)
gu.Button("Outline Small", gu.ButtonOutline, gu.ButtonSmall)
gu.Button("Delete", gu.ButtonDanger, gl.Click("delete"))
```

### badge.go — Badge

```go
gu.Badge("Active", "default")  // バリアント: "default", "dark", "outline", "light"
```

### alert.go — Alert, AlertSuccess, AlertWarning, AlertDanger, AlertInfo

```go
gu.Alert("操作が成功しました！", "success")
// バリアント: "info", "success", "warning", "danger"
```

### stat.go — StatCard

```go
gu.StatCard("総ユーザー数", "1,234")
```

### table.go — StyledTable, THead

```go
gu.StyledTable(
    gu.THead("名前", "メール", "役割"),
    gd.Tr(gd.Td(gp.Value("Alice")), gd.Td(gp.Value("alice@example.com")), gd.Td(gp.Value("Admin"))),
)
```

### form.go — FormGroup, FormLabel, FormInput, FormSelect, FormTextarea, Checkbox, Radio

```go
gu.FormGroup(
    gu.FormLabel("メール", "email"),
    gu.FormInput("email", gp.ID("email"), gp.Type("email"), gp.Placeholder("you@example.com")),
    gu.FormError(errMsg),
)

gu.FormSelect("role", []gu.FormOption{
    {Value: "admin", Label: "管理者"},
    {Value: "user", Label: "ユーザー"},
})

gu.Checkbox("agree", "利用規約に同意する", checked)
gu.Radio("plan", "pro", "Pro プラン", selected)
```

### icon.go — Icon

```go
gu.Icon("home")           // デフォルトサイズ (md)
gu.Icon("search", "sm")   // 小
gu.Icon("settings", "lg") // 大
```

30 以上のビルトインアイコン: `home`, `user`, `users`, `settings`, `search`, `plus`, `minus`, `x`, `check`, `edit`, `trash`, `chevron-right`, `chevron-down`, `chevron-left`, `chevron-up`, `menu`, `folder`, `file`, `mail`, `bell`, `calendar`, `chart`, `download`, `upload`, `link`, `info`, `warning`, `circle`, `lock`, `logout`, `filter`, `sort-asc`, `sort-desc`, `eye`, `copy`

### spinner.go — Spinner

```go
gu.Spinner("sm")  // サイズ: "sm", "md", "lg"
```

### misc.go — Divider, EmptyState, Progress

```go
gu.Divider()
gu.EmptyState("アイテムが見つかりません")
gu.Progress(75)  // 0〜100 のパーセンテージ
```

### nav.go — Sidebar, SidebarHeader, SidebarLink, SidebarDivider, Breadcrumb, MobileHeader, PageHeader

[レイアウトシステム](#レイアウトシステム) を参照。

### tree.go — Tree

```go
gu.Tree([]gu.TreeNode{
    {Label: "src", Icon: "folder", Open: true, Children: []gu.TreeNode{
        {Label: "main.go", Icon: "file", Active: true},
        {Label: "utils.go", Icon: "file"},
    }},
})
```

### chat.go — ChatContainer, ChatMessageView, ChatInput

```go
gu.ChatContainer(
    gu.ChatMessageView(gu.ChatMessage{Author: "Alice", Content: "こんにちは！", Timestamp: "09:00", Avatar: "A"}),
    gu.ChatMessageView(gu.ChatMessage{Content: "やあ！", Timestamp: "09:01", Sent: true}),
)
gu.ChatInput("msg", draft, gu.ChatInputOpts{
    Placeholder:  "メッセージを入力...",
    SendEvent:    "chatSend",
    InputEvent:   "chatInput",
    KeydownEvent: "chatKeydown",
})
```

### avatar.go — ImageAvatar, LetterAvatar, AvatarGroup

```go
gu.ImageAvatar("https://example.com/photo.jpg", gu.AvatarOpts{Size: "lg"})
gu.LetterAvatar("Alice", gu.AvatarOpts{Size: "md", Shape: "rounded"})
gu.AvatarGroup(avatars, gu.AvatarGroupOpts{Size: "sm", Max: 3})
```

サイズ: `"xs"` (24px), `"sm"` (32px), `"md"` (40px), `"lg"` (48px), `"xl"` (64px)。形状: `"circle"`（デフォルト）, `"rounded"`

### chart.go — LineChart, ColumnChart, BarChart, PieChart, ScatterPlot, Histogram, StackedBarChart

```go
series := []gu.Series{
    {Name: "売上", Points: []gu.DataPoint{{Label: "1月", Value: 420}, {Label: "2月", Value: 380}}},
}
gu.LineChart(series, gu.ChartOpts{Width: 600, Height: 400, Title: "月次データ", ShowGrid: true, ShowLegend: true})
gu.PieChart([]gu.DataPoint{{Label: "A", Value: 45}, {Label: "B", Value: 55}}, gu.ChartOpts{})
gu.Histogram(values, gu.HistogramOpts{ChartOpts: gu.ChartOpts{}, BinCount: 10})
```

## インタラクティブコンポーネント（イベント統合）

インタラクティブコンポーネントはイベントフィールドを持つ Opts 構造体を使用します。イベントフィールドが空の場合、静的 HTML としてレンダリングされます。値が設定されると、LiveView イベントバインディング用の `gerbera-*` 属性が追加されます。

### 静的 vs インタラクティブの比較例

```go
// 静的カレンダー（イベントなし）
gu.Calendar(gu.CalendarOpts{
    Year:  2026,
    Month: time.March,
    Today: time.Now(),
})

// インタラクティブカレンダー（LiveView イベント付き）
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
})
```

### 全インタラクティブコンポーネント

| コンポーネント | Opts 構造体 | 主要イベントフィールド |
|--------------|-----------|---------------------|
| `Calendar` | `CalendarOpts` | `SelectEvent`, `PrevMonthEvent`, `NextMonthEvent`, `MonthChangeEvent`, `YearChangeEvent` |
| `NumberInput` | `NumberInputOpts` | `IncrementEvent`, `DecrementEvent`, `ChangeEvent` |
| `Slider` | `SliderOpts` | `InputEvent` |
| `Accordion` | `AccordionOpts` | `ToggleEvent` |
| `ButtonGroup` | `ButtonGroupOpts` | `ClickEvent` |
| `Pagination` | `PaginationOpts` | `PageEvent` |
| `Stepper` | `StepperOpts` | `ClickEvent` |
| `InfiniteScroll` | `InfiniteScrollOpts` | `LoadMoreEvent`, `ToggleEvent` |
| `TimePicker` | `TimePickerOpts` | `ChangeEvent` |
| `ChatInput` | `ChatInputOpts` | `SendEvent`, `InputEvent`, `KeydownEvent` |
| `ImageAvatar` / `LetterAvatar` | `AvatarOpts` | `ClickEvent` |
| `AvatarGroup` | `AvatarGroupOpts` | `ClickEvent` |
| チャート系 | `ChartOpts` | `ClickEvent`, `MouseEnterEvent`, `MouseLeaveEvent` |

## ui/live/ パッケージ

LiveView WebSocket 接続が必要なコンポーネント。インポート:

```go
gul "github.com/tomo3110/gerbera/ui/live"
```

| コンポーネント | 関数 | 説明 |
|--------------|------|------|
| `Modal` | `Modal(open, closeEvent, children...)` | バックドロップ付きオーバーレイダイアログ |
| `Toast` | `Toast(message, variant, dismissEvent)` | 通知トースト（右上） |
| `DataTable` | `DataTable(DataTableOpts)` | ソート・ページネーション付きテーブル |
| `Dropdown` | `Dropdown(open, toggleEvent, trigger, menu)` | ドロップダウンメニュー |
| `Confirm` | `Confirm(open, title, message, confirmEvent, cancelEvent)` | 確認ダイアログ |
| `Tabs` | `Tabs(id, activeIndex, tabs, changeEvent)` | ARIA 対応タブパネル |
| `Drawer` | `Drawer(open, closeEvent, side, children...)` | スライドアウトパネル |
| `SearchSelect` | `SearchSelect(SearchSelectOpts)` | キーボードナビゲーション付きフィルタリングコンボボックス |

### Modal の例

```go
gul.Modal(v.ModalOpen, "closeModal",
    gul.ModalHeader("ユーザー詳細", "closeModal"),
    gul.ModalBody(content),
    gul.ModalFooter(
        gu.Button("閉じる", gu.ButtonOutline, gl.Click("closeModal")),
    ),
)
```

### DataTable の例

```go
gul.DataTable(gul.DataTableOpts{
    Columns: []gul.Column{
        {Key: "name", Label: "名前", Sortable: true},
        {Key: "email", Label: "メール", Sortable: true},
    },
    Rows:      rows,
    SortCol:   v.SortCol,
    SortDir:   v.SortDir,
    SortEvent: "sort",
    Page:      v.TablePage,
    PageSize:  10,
    Total:     totalRows,
    PageEvent: "tablePage",
})
```

## CSS 命名規則

すべての CSS クラスに `g-` プレフィックスを使用（例: `.g-card`, `.g-btn-primary`, `.g-grid-cols-3`）。コンポーネントはテーマの CSS カスタムプロパティ（`--g-bg`, `--g-text`, `--g-accent` 等）を使用するため、テーマ変更で全コンポーネントの外観が自動的に適応します。
