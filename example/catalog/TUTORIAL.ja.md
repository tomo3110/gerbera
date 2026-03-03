# カタログ チュートリアル

## 概要

`ui/` および `ui/live/` パッケージの全ウィジェットを紹介する、シングルページのインタラクティブ UI コンポーネントカタログを構築します。このチュートリアルでは以下の機能を学びます。

- `gu.AdminShell` + `gu.MobileHeader` + `gul.Drawer` — モバイルナビゲーション付きレスポンシブ管理画面レイアウト
- `gu.Theme` / `gu.ThemeWith` / `gu.ThemeAuto` — ランタイムでのテーマ切り替え（ライト、ダーク、カスタム、自動）
- `gl.Click` + `gl.ClickValue` による SPA ナビゲーション — `switch` 文によるページディスパッチ
- 35 以上のコンポーネントデモページ — 静的ウィジェット、ライブウィジェット、データ可視化、テーマ設定
- 50 以上のイベントハンドラ — トグル、入力、ソート、キーボードナビゲーションなど

カタログは **単一ファイルの LiveView アプリケーション**（約 1900 行）であり、すべての `ui/` コンポーネントのビジュアルリファレンスと動作するコード例の両方を兼ねています。各ページでは、コンポーネントの静的版と LiveView 版の両方の使い方を示します。

## 前提条件

- Go 1.22 以上がインストールされていること
- このリポジトリをクローン済みであること

## ステップ 1: アーキテクチャ — 単一ファイル構成

```go
import (
	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
	gu "github.com/tomo3110/gerbera/ui"
	gul "github.com/tomo3110/gerbera/ui/live"
)
```

カタログでは 6 つのインポートエイリアスを使用します:

| エイリアス | パッケージ | 用途 |
|-----------|-----------|------|
| `g` | `gerbera` | コア型（`ComponentFunc`、`Element`） |
| `gd` | `dom` | HTML 要素ヘルパー（`Div`、`Body`、`H1` など） |
| `expr` | `expr` | 制御フロー（`If`、`Unless`、`Each`） |
| `gl` | `live` | LiveView ランタイム（`View`、`Handler`、`Click` など） |
| `gp` | `property` | 属性セッター（`Class`、`Attr`、`Value` など） |
| `gu` | `ui` | 静的 UI ウィジェット（Card、Button、Table など） |
| `gul` | `ui/live` | LiveView UI ウィジェット（Modal、Drawer、DataTable など） |

カタログ全体は、すべてのコンポーネントデモの状態を保持する約 87 個のフィールドを持つ単一の `CatalogView` 構造体です:

```go
type CatalogView struct {
	Page         string
	ModalOpen    bool
	ConfirmOpen  bool
	ToastVisible bool
	ToastMessage string
	ToastVariant string
	DropdownOpen  bool
	DropdownValue string
	TablePage    int
	SortCol      string
	SortDir      string
	TabIndex     int

	// Drawer
	DrawerOpen bool
	DrawerSide string

	// SearchSelect
	SSQuery     string
	SSOpen      bool
	SSValue     string
	SSHighlight int

	// Calendar demo
	CalYear     int
	CalMonth    time.Month
	CalSelected *time.Time

	// Chat demo
	ChatMessages []gu.ChatMessage
	ChatDraft    string

	// Theme demo
	ThemeMode string // "light", "dark", "custom", "auto"

	// ... 他のデモ用フィールド約 60 個
}
```

主要な設計方針:

- **単一構造体** — すべてのデモページの状態が 1 つの構造体に格納されます。ナビゲーションは `Page` フィールドを切り替え、アクティブなページのみがウィジェットをレンダリングします。
- **データベースや外部状態なし** — すべてセッションごとのインメモリ管理です。これによりカタログは自己完結型になります。
- **フラットな構成** — 各ページは独自の `page*()` メソッドを持ちます（例: `pageCard()`、`pageModal()`、`pageCalendar()`）。

## ステップ 2: テーマ切り替え — 4 モードの currentTheme()

```go
func (v *CatalogView) currentTheme() g.ComponentFunc {
	switch v.ThemeMode {
	case "dark":
		return gu.ThemeWith(gu.DarkTheme())
	case "custom":
		return gu.ThemeWith(gu.ThemeConfig{
			Accent:      "#1a73e8",
			AccentLight: "#e8f0fe",
			AccentHover: "#1557b0",
		})
	case "auto":
		return gu.ThemeAuto(gu.ThemeConfig{}, gu.ThemeConfig{})
	default:
		return gu.Theme()
	}
}
```

`currentTheme()` メソッドは `<head>` 内に CSS カスタムプロパティを出力する `ComponentFunc` を返します。4 つのモードは以下の通りです:

| モード | 関数 | 説明 |
|--------|------|------|
| `light` | `gu.Theme()` | デフォルトのライトテーマ |
| `dark` | `gu.ThemeWith(gu.DarkTheme())` | ビルトイン設定によるフルダークテーマ |
| `custom` | `gu.ThemeWith(gu.ThemeConfig{...})` | カスタムアクセントカラーのオーバーライド |
| `auto` | `gu.ThemeAuto(light, dark)` | `prefers-color-scheme` メディアクエリで OS 設定に追従 |

テーマは `Render()` 内で `gd.Head(...)` の子として適用されます:

```go
gd.Head(
	gd.Title("Gerbera UI Catalog"),
	gd.Meta(gp.Attr("name", "viewport"), gp.Attr("content", "width=device-width, initial-scale=1")),
	v.currentTheme(),
),
```

ユーザーがテーマボタンをクリックすると `"setTheme"` イベントが `v.ThemeMode` を更新し、ページ全体が新しい CSS カスタムプロパティで再レンダリングされます。Gerbera Live がツリー全体を差分比較するため、変更された `<style>` 要素と影響を受ける属性のみがパッチされます。

## ステップ 3: レスポンシブレイアウト — AdminShell + MobileHeader + Drawer

```go
func (v *CatalogView) Render() []g.ComponentFunc {
	return []g.ComponentFunc{
		gd.Head(
			gd.Title("Gerbera UI Catalog"),
			gd.Meta(gp.Attr("name", "viewport"), gp.Attr("content", "width=device-width, initial-scale=1")),
			v.currentTheme(),
		),
		gd.Body(
			gu.MobileHeader("UI Catalog", gl.Click("toggleMobileNav")),
			gu.AdminShell(
				v.renderSidebar(),
				v.renderContent(),
			),
			// モバイルナビゲーション用ドロワー
			gul.Drawer(v.MobileNavOpen, "closeMobileNav", "left",
				gul.DrawerHeader("UI Catalog", "closeMobileNav"),
				gul.DrawerBody(v.renderNavLinks()),
			),
		),
	}
}
```

レイアウトは 3 つの補完的なコンポーネントで構成されます:

- **`gu.AdminShell(sidebar, content)`** — 2 カラムレイアウト。デスクトップ（768px 超）ではサイドバーが常に表示され、モバイルでは非表示になります。
- **`gu.MobileHeader(title, clickAttr)`** — 768px 以下の画面でのみ表示される固定トップバーで、ハンバーガーメニューボタンを備えます。クリックバインディングは `"toggleMobileNav"` イベントをトリガーします。
- **`gul.Drawer(open, closeEvent, side, children...)`** — LiveView 駆動のスライドアウトパネル。`v.MobileNavOpen` が `true` のとき、サイドバーと同じナビゲーションリンクを含むドロワーが左からスライドインします。

このパターンにより、カスタム CSS なしでレスポンシブな管理画面ダッシュボードレイアウトが実現できます。

## ステップ 4: ナビゲーション — カテゴリ別サイドバーとページディスパッチ

```go
func (v *CatalogView) renderNavLinks() g.ComponentFunc {
	return gd.Div(
		navLink(v, "overview", "Overview"),
		navDividerLabel("Static Widgets"),
		navLink(v, "card", "Card"),
		navLink(v, "button", "Button"),
		navLink(v, "icon", "Icon"),
		// ... さらにリンクが続く
		navDividerLabel("Live Widgets"),
		navLink(v, "modal", "Modal"),
		navLink(v, "datatable", "DataTable"),
		// ... さらにリンクが続く
		navDividerLabel("Data Visualization"),
		navLink(v, "chart", "Charts"),
		navLink(v, "avatar", "Avatar"),
		navDividerLabel("Theming"),
		navLink(v, "theme", "Theme"),
	)
}
```

ナビゲーションリンクは 4 つのカテゴリに分類されます: **Static Widgets**、**Live Widgets**、**Data Visualization**、**Theming**。

2 つのヘルパー関数でコードの重複を排除しています:

```go
func navLink(v *CatalogView, page, label string) g.ComponentFunc {
	return gu.SidebarLink("#", label, v.Page == page,
		gl.Click("nav"), gl.ClickValue(page))
}

func navDividerLabel(label string) g.ComponentFunc {
	return gd.Div(
		gu.SidebarDivider(),
		gd.Div(gp.Attr("style", "padding:4px 24px;font-size:11px;font-weight:600;color:var(--g-text-tertiary);text-transform:uppercase;letter-spacing:0.04em"), gp.Value(label)),
	)
}
```

- `navLink` は `gu.SidebarLink` を使い、第 3 引数 `v.Page == page` でアクティブなリンクをハイライトします
- `gl.Click("nav")` + `gl.ClickValue(page)` は `{e: "nav", p: {value: "card"}}` をサーバーに送信します

ページディスパッチは `renderContent()` 内のシンプルな `switch` 文です:

```go
func (v *CatalogView) renderContent() g.ComponentFunc {
	var body g.ComponentFunc
	switch v.Page {
	case "card":
		body = v.pageCard()
	case "button":
		body = v.pageButton()
	case "modal":
		body = v.pageModal()
	case "calendar":
		body = v.pageCalendar()
	// ... 30 以上のケースが続く
	default:
		body = v.pageOverview()
	}
	return gd.Div(
		gu.PageHeader("Gerbera UI Catalog"),
		gd.Div(gp.Class("g-page-body"), body),
	)
}
```

ユーザーがナビゲーションリンクをクリックすると、`HandleEvent` が `"nav"` を受け取り `v.Page` を更新します。フレームワークが再レンダリングし、switch が新しいページメソッドにディスパッチします。サイドバーは同じまま（アクティブなハイライトのみ変更）なので、コンテンツ領域のみがパッチされます。

## ステップ 5: 静的コンポーネントページ — Card と Button

静的コンポーネントページは、LiveView の状態を必要としない `gu.*` ウィジェットを紹介します:

### Card ページ

```go
func (v *CatalogView) pageCard() g.ComponentFunc {
	return gd.Div(
		section("Card", "Container with optional header and footer."),
		gu.Card(
			gu.CardHeader("Card Title", gu.Button("Action", gu.ButtonSmall, gu.ButtonOutline)),
			gd.Div(gp.Class("g-page-body"),
				gd.P(gp.Value("Card content goes here.")),
			),
			gu.CardFooter(gd.Span(gp.Value("Footer text"))),
		),
	)
}
```

- `gu.Card(children...)` はボーダーとシャドウ付きのスタイル済みコンテナでコンテンツを囲みます
- `gu.CardHeader(title, actions...)` はオプションのアクションボタン付きヘッダーバーをレンダリングします
- `gu.CardFooter(children...)` はフッター領域をレンダリングします

### Button ページ

```go
func (v *CatalogView) pageButton() g.ComponentFunc {
	return gu.Stack(
		section("Button", "Button variants."),
		gu.HStack(
			gu.Button("Default"),
			gu.Button("Primary", gu.ButtonPrimary),
			gu.Button("Outline", gu.ButtonOutline),
			gu.Button("Danger", gu.ButtonDanger),
			gu.Button("Small", gu.ButtonSmall),
			gu.Button("Small Primary", gu.ButtonSmall, gu.ButtonPrimary),
		),
		section("Button with Icon", "Combine Icon() with Button."),
		gu.HStack(
			gu.Button("Add", gu.ButtonPrimary, gu.Icon("plus", "sm")),
			gu.Button("Delete", gu.ButtonDanger, gu.Icon("trash", "sm")),
		),
	)
}
```

- ボタンバリアントは修飾子 `ComponentFunc` を渡して適用します: `gu.ButtonPrimary`、`gu.ButtonOutline`、`gu.ButtonDanger`、`gu.ButtonSmall`
- アイコンは `gu.Icon(name, size)` でインライン合成します
- `gu.Stack` は垂直方向のスペース確保、`gu.HStack` は水平方向の配置を提供します

`section()` ヘルパーは各デモに統一されたヘッダーと説明文を生成します:

```go
func section(title, desc string) g.ComponentFunc {
	return gd.Div(
		gd.H2(gp.Attr("style", "font-size:18px;font-weight:600;margin:0 0 4px 0"), gp.Value(title)),
		gd.P(gp.Attr("style", "color:var(--g-text-secondary);margin:0 0 16px 0;font-size:13px"), gp.Value(desc)),
	)
}
```

## ステップ 6: インタラクティブウィジェットページ — Calendar と Accordion

インタラクティブウィジェットページは、LiveView の状態とイベントバインディング用 Opts 構造体を使うコンポーネントを紹介します。

### Calendar ページ

```go
func (v *CatalogView) pageCalendar() g.ComponentFunc {
	return gu.Stack(
		section("Calendar", "Month-view calendar with navigation and date selection (LiveView)."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
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
				gd.Div(gp.Attr("style", "margin-top:12px"),
					gu.Alert(fmt.Sprintf("Selected: %s", v.calSelectedStr()), "info"),
				),
			),
		)),
	)
}
```

Calendar コンポーネントは、状態（年、月、選択された日付）とイベント名の両方を持つ `CalendarOpts` 構造体を受け取ります。コンポーネント自体が、提供されたイベント名を使って内部で `gl.Click` バインディングを発行します。これは LiveView 対応の `ui/` コンポーネントの標準パターンです。Opts 構造体がコンポーネントと LiveView イベントシステムを橋渡しします。

### Accordion ページ

```go
func (v *CatalogView) pageAccordion() g.ComponentFunc {
	return gu.Stack(
		section("Accordion", "Collapsible sections with server-controlled open/close (LiveView)."),
		gu.Card(gd.Div(gp.Class("g-page-body"),
			gu.Accordion([]gu.AccordionItem{
				{Title: "What is Gerbera?", Content: gd.P(gp.Value("...")), Open: v.AccordionOpen[0]},
				{Title: "How does it work?", Content: gd.P(gp.Value("...")), Open: v.AccordionOpen[1]},
				{Title: "Is it production ready?", Content: gd.P(gp.Value("...")), Open: v.AccordionOpen[2]},
			}, gu.AccordionOpts{Exclusive: true, ToggleEvent: "accordionToggle"}),
		)),
	)
}
```

- `AccordionOpts{Exclusive: true}` は、1 つのセクションを開くと他のすべてが閉じることを意味します（このロジックはサーバー側の `HandleEvent` で処理されます）
- `ToggleEvent` を指定しない場合、Accordion はネイティブの `<details>`/`<summary>` 要素にフォールバックし、JavaScript なしで動作します

両方のページでは、イベントバインディングなしの **Static** バリアントも表示しており、すべての `ui/` コンポーネントが静的版と LiveView 版の両方で動作することを示しています。

## ステップ 7: LiveView コンポーネントページ — Modal と DataTable

LiveView 専用コンポーネントは `ui/live/` パッケージ（エイリアス `gul`）にあります:

### Modal ページ

```go
func (v *CatalogView) pageModal() g.ComponentFunc {
	return gd.Div(
		section("Modal", "Dialog overlay (LiveView)."),
		gu.Button("Open Modal", gu.ButtonPrimary, gl.Click("openModal")),
		gul.Modal(v.ModalOpen, "closeModal",
			gul.ModalHeader("Sample Modal", "closeModal"),
			gul.ModalBody(
				gd.P(gp.Value("This is a modal dialog built with gerbera/ui/live.")),
				gu.Alert("Modal content can include any widget.", "info"),
			),
			gul.ModalFooter(
				gu.Button("Close", gu.ButtonOutline, gl.Click("closeModal")),
				gu.Button("Save", gu.ButtonPrimary, gl.Click("closeModal")),
			),
		),
	)
}
```

- `gul.Modal(open, closeEvent, children...)` はモーダルオーバーレイをレンダリングします。`v.ModalOpen` が `false` の場合、モーダルは CSS で非表示になります。
- 第 2 引数 `"closeModal"` はオーバーレイの背景がクリックされたときに送信されるイベント名です。
- `gul.ModalHeader`、`gul.ModalBody`、`gul.ModalFooter` はヘッダーに X 閉じるボタンを含むモーダルコンテンツを構造化します。

### DataTable ページ

```go
func (v *CatalogView) pageDataTable() g.ComponentFunc {
	rows := sampleRows()
	pageSize := 5
	start := v.TablePage * pageSize
	end := start + pageSize
	if end > len(rows) {
		end = len(rows)
	}
	pageRows := rows[start:end]

	return gd.Div(
		section("DataTable", "Sortable, paginated table (LiveView)."),
		gul.DataTable(gul.DataTableOpts{
			Columns: []gul.Column{
				{Key: "name", Label: "Name", Sortable: true},
				{Key: "email", Label: "Email", Sortable: true},
				{Key: "role", Label: "Role", Sortable: true},
				{Key: "status", Label: "Status", Sortable: false},
			},
			Rows:      pageRows,
			SortCol:   v.SortCol,
			SortDir:   v.SortDir,
			SortEvent: "sort",
			Page:      v.TablePage,
			PageSize:  pageSize,
			Total:     len(rows),
			PageEvent: "tablePage",
		}),
	)
}
```

- `gul.DataTable` はカラム、行（`[][]string`）、ソート、ページネーションを設定する `DataTableOpts` 構造体を受け取ります
- コンポーネントは Opts 構造体のイベント名を使ってソートおよびページネーションイベントを発行します
- サーバー側のスライシング（`rows[start:end]`）が実際のページネーションを処理し、コンポーネントは現在のページのみをレンダリングします

## ステップ 8: イベント処理 — 代表的な HandleEvent スニペット

`HandleEvent` メソッドは 50 以上のケースを持つ単一の `switch` 文を含みます。以下は代表的なパターンです:

### ナビゲーション

```go
case "nav":
	v.Page = payload["value"]
	v.resetOverlays()
```

`resetOverlays()` はページ遷移時にすべてのモーダル、ドロワー、ドロップダウン、トーストを閉じます:

```go
func (v *CatalogView) resetOverlays() {
	v.ModalOpen = false
	v.ConfirmOpen = false
	v.ToastVisible = false
	v.DropdownOpen = false
	v.DrawerOpen = false
	v.SSOpen = false
	v.MobileNavOpen = false
}
```

### トグルパターン（Modal、Checkbox、Drawer）

```go
case "openModal":
	v.ModalOpen = true
case "closeModal":
	v.ModalOpen = false
case "toggleCheckA":
	v.CheckA = !v.CheckA
case "openDrawer":
	v.DrawerOpen = true
	v.DrawerSide = payload["value"]
```

### ソートパターン（DataTable）

```go
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
```

### キーボードナビゲーション（SearchSelect）

```go
case "ssKeydown":
	key := payload["key"]
	filtered := v.ssFilteredOptions()
	n := len(filtered)
	switch key {
	case "Escape":
		v.SSOpen = false
		v.SSHighlight = -1
	case "ArrowDown":
		if !v.SSOpen {
			v.SSOpen = true
			v.SSHighlight = 0
		} else if n > 0 {
			v.SSHighlight = (v.SSHighlight + 1) % n
		}
	case "ArrowUp":
		if v.SSOpen && n > 0 {
			if v.SSHighlight <= 0 {
				v.SSHighlight = n - 1
			} else {
				v.SSHighlight--
			}
		}
	case "Enter":
		if v.SSOpen && v.SSHighlight >= 0 && v.SSHighlight < n {
			selected := filtered[v.SSHighlight]
			v.SSValue = selected.Value
			v.SSQuery = selected.Label
			v.SSOpen = false
			v.SSHighlight = -1
		}
	}
```

### カレンダーナビゲーション

```go
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
```

## 動作の仕組み

```
ブラウザ                                       Go サーバー
  |                                               |
  |  1. GET /                                     |
  |  <-- 初期 HTML（概要ページ）+ JS              |
  |                                               |
  |  2. WebSocket 接続                            |
  |  <-> セッション作成、CatalogView 割り当て      |
  |                                               |
  |  3. サイドバーの「Card」をクリック              |
  |  --> {"e":"nav","p":{"value":"card"}}         |
  |                                               |
  |      HandleEvent:                             |
  |        v.Page = "card"                        |
  |        v.resetOverlays()                      |
  |      Render -> switch "card" -> pageCard()    |
  |      旧ツリーと新ツリーを差分比較              |
  |                                               |
  |  4. <-- パッチ（コンテンツ領域のみ;            |
  |          サイドバーのアクティブハイライト）      |
  |                                               |
  |  5. JS がパッチを DOM に適用                   |
  |                                               |
  |  6. 「Open Modal」ボタンをクリック             |
  |  --> {"e":"openModal","p":{}}                 |
  |                                               |
  |      HandleEvent: v.ModalOpen = true          |
  |      Render -> モーダルオーバーレイ表示        |
  |      差分比較 -> モーダル表示パッチ            |
  |                                               |
  |  7. <-- パッチ（モーダル display 属性）        |
```

すべてのナビゲーションは WebSocket イベントを介してサーバー側で処理されます。ブラウザは 2 回目の HTTP リクエストを行いません。URL は変更されませんが、初期ページはクエリパラメータで設定できます: `/?page=calendar`。

## 実行方法

```bash
go run example/catalog/catalog.go
```

ブラウザで http://localhost:8900 を開きます。サイドバーを使ってすべてのコンポーネントページを閲覧できます。ブラウザウィンドウのサイズを変更すると、ハンバーガーメニューとスライドアウトドロワーによるレスポンシブモバイルレイアウトが確認できます。

ポートを変更する場合:

```bash
go run example/catalog/catalog.go -addr :3000
```

### デバッグモード

```bash
go run example/catalog/catalog.go -debug
```

デバッグモードを有効にすると、画面右下にフローティングの「G」ボタンが表示されます。ボタンをクリックするか `Ctrl+Shift+D` を押すと Gerbera Debug Panel が開き、イベント、DOM パッチ、状態、セッション情報をリアルタイムで確認できます。

## 発展課題

1. **カスタムコンポーネントページの追加** — 新しい `pageMyWidget()` メソッドを作成し、`navLink` エントリを追加して、`renderContent()` の `switch` にケースを追加します。既存の `gu.*` ウィジェットを組み合わせて新しいパターンを作ってみましょう。
2. **新しいテーマの追加** — `currentTheme()` に 5 番目のモード（例: `"ocean"`）を追加し、ブルー/ティール系アクセントカラーのカスタム `gu.ThemeConfig` を使います。Theme ページにボタンを追加してみましょう。
3. **URL ベースのナビゲーション** — `Mount()` の `params.Get("page")` を使って（部分的に実装済み）、`/?page=modal` のような特定ページへのディープリンクをサポートしてみましょう。
4. **検索/フィルターの追加** — サイドバーのナビゲーションリンクの上にテキスト入力を追加し、ラベルで表示リンクをフィルタします。`gl.Input("filterNav")` と `expr.If` を使ってリンクの表示/非表示を切り替えてみましょう。
5. **新しいチャートタイプの追加** — Charts ページに新しい `ChartType` ケースと対応する `gu.ButtonGroupItem` エントリを追加してみましょう。

## API リファレンス

| 関数 / 型 | 説明 |
|---|---|
| `gu.AdminShell(sidebar, content)` | 2 カラム管理画面レイアウト（レスポンシブ） |
| `gu.MobileHeader(title, click)` | モバイル画面用固定トップバー |
| `gu.Sidebar(children...)` | サイドバーコンテナ |
| `gu.SidebarHeader(title)` | タイトル付きサイドバーヘッダー |
| `gu.SidebarLink(href, label, active, attrs...)` | アクティブ状態付きナビゲーションリンク |
| `gu.SidebarDivider()` | サイドバー内の水平区切り線 |
| `gu.PageHeader(title)` | ページタイトルヘッダー |
| `gu.Theme()` | デフォルトのライトテーマ CSS |
| `gu.ThemeWith(config)` | `ThemeConfig` によるカスタムテーマ CSS |
| `gu.ThemeAuto(light, dark)` | OS 設定連動テーマ（メディアクエリ） |
| `gu.DarkTheme()` | ビルトインのダーク `ThemeConfig` |
| `gu.Card(children...)` | カードコンテナ |
| `gu.Button(label, modifiers...)` | スタイル付きボタン |
| `gu.Stack(children...)` | ギャップ付き垂直フレックスレイアウト |
| `gu.HStack(children...)` | 水平フレックスレイアウト |
| `gu.Grid(modifier, children...)` | CSS Grid レイアウト |
| `gu.Calendar(opts)` | 月表示カレンダー |
| `gu.Accordion(items, opts)` | 折りたたみセクション |
| `gu.NumberInput(name, value, opts)` | +/- ボタン付き数値入力 |
| `gu.Slider(name, value, opts)` | レンジスライダー |
| `gu.TimePicker(name, h, m, s, opts)` | タイムピッカー |
| `gu.Pagination(opts)` | ページナビゲーション |
| `gu.ButtonGroup(items, opts)` | セグメントボタンコントロール |
| `gul.Modal(open, closeEvent, children...)` | モーダルダイアログオーバーレイ |
| `gul.Toast(msg, variant, dismissEvent)` | 通知ポップアップ |
| `gul.DataTable(opts)` | ソート・ページネーション付きテーブル |
| `gul.Dropdown(open, toggleEvent, trigger, menu)` | トグルドロップダウンメニュー |
| `gul.Tabs(id, index, tabs, switchEvent)` | アクセシブルタブパネル |
| `gul.Drawer(open, closeEvent, side, children...)` | スライドアウトパネル |
| `gul.SearchSelect(opts)` | フィルタ付きコンボボックス |
| `gul.Confirm(open, title, msg, okEvent, cancelEvent)` | 確認ダイアログ |
