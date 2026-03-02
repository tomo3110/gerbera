# Gerbera UI Component Library

Pre-built UI components for Gerbera — build professional interfaces by calling Go functions, without writing CSS.

[日本語版はこちら](README.ja.md)

## Philosophy

Engineers should be able to build polished UIs by writing Go code alone, without wrestling with CSS or design decisions. Every component ships with sensible defaults, ARIA accessibility attributes, and theme-aware styling via CSS custom properties.

## Quick Start

```go
import gu "github.com/tomo3110/gerbera/ui"

// Place theme CSS in <head>
gd.Head(gu.Theme())

// Use components in <body>
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

Import conventions:

```go
gu  "github.com/tomo3110/gerbera/ui"       // Static + interactive components
gul "github.com/tomo3110/gerbera/ui/live"   // LiveView-only components
```

## Theme System

All components use CSS custom properties (`--g-*`) for colors, spacing, and typography. Changing the theme automatically reskins every component.

### Functions

| Function | Description |
|----------|-------------|
| `Theme()` | Default light theme |
| `ThemeWith(ThemeConfig{...})` | Custom theme — zero-value fields fall back to defaults |
| `DarkTheme()` | Dark theme preset (returns `ThemeConfig`) |
| `ThemeAuto(light, dark)` | Switches via `prefers-color-scheme` media query |

### ThemeConfig Fields

**Background**: `Bg`, `BgSurface`, `BgOverlay`, `BgInset`
**Text**: `Text`, `TextSecondary`, `TextTertiary`, `TextInverse`
**Border**: `Border`, `BorderStrong`
**Shadow**: `ShadowSm`, `Shadow`, `ShadowLg`
**Accent**: `Accent`, `AccentLight`, `AccentHover`
**Semantic**: `Danger`/`DangerBg`/`DangerBorder`/`DangerHover`, `Success`/`SuccessBg`/`SuccessBorder`, `Warning`/`WarningBg`/`WarningBorder`, `Info`/`InfoBg`/`InfoBorder`
**Spacing**: `SpaceXs`(4px), `SpaceSm`(8px), `SpaceMd`(16px), `SpaceLg`(28px), `SpaceXl`(44px)
**Typography**: `Font`, `FontMono`, `Radius`

### CSS Variables

All map to `--g-<field>` (e.g., `Bg` → `--g-bg`, `Accent` → `--g-accent`).

### Examples

```go
// Light theme (default)
gu.Theme()

// Dark theme
gu.ThemeWith(gu.DarkTheme())

// Custom accent color
gu.ThemeWith(gu.ThemeConfig{
    Accent:      "#1a73e8",
    AccentLight: "#e8f0fe",
    AccentHover: "#1557b0",
})

// Auto light/dark based on OS preference
gu.ThemeAuto(gu.ThemeConfig{}, gu.ThemeConfig{})
```

## Layout System

### AdminShell

Two-column admin layout with fixed sidebar:

```go
gu.AdminShell(
    gu.Sidebar(
        gu.SidebarHeader("My App"),
        gu.SidebarLink("#", "Dashboard", true),
        gu.SidebarLink("#", "Settings", false),
    ),
    content, // main content area
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
gu.Grid(gu.GridCols3, items...)         // 3-column grid
gu.Grid(gu.GridAutoFit("200px"), items...) // Auto-fit min 200px
gu.Grid(gu.GridAutoFill("200px"), items...) // Auto-fill min 200px
```

Column constants: `GridCols2`, `GridCols3`, `GridCols4`, `GridCols5`, `GridCols6`

Spanning: `GridSpan(2)`, `GridRowSpan(2)`

### Flex Containers

| Function | Description |
|----------|-------------|
| `Row(children...)` | Horizontal flex, gap 8px |
| `Column(children...)` | Vertical flex, gap 8px |
| `Stack(children...)` | Vertical stack, gap 16px |
| `HStack(children...)` | Horizontal, center-aligned |
| `VStack(children...)` | Vertical, center-aligned |
| `Center(children...)` | Center both axes |
| `Container(children...)` | Max-width 960px, centered |
| `ContainerNarrow(children...)` | Max-width 640px |
| `ContainerWide(children...)` | Max-width 1280px |
| `Spacer()` | Flex spacer (fills available space) |
| `SpaceY(size)` | Vertical spacing (xs/sm/md/lg/xl) |

### Flex Modifiers

**Gap**: `GapNone`, `GapXs`, `GapSm`, `GapMd`, `GapLg`, `GapXl`
**Justify**: `JustifyStart`, `JustifyCenter`, `JustifyEnd`, `JustifyBetween`, `JustifyAround`
**Align**: `AlignStart`, `AlignCenter`, `AlignEnd`, `AlignStretch`, `AlignBaseline`
**Other**: `Wrap`, `Grow`, `Shrink0`

## Component Reference

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
gu.Badge("Active", "default")  // variants: "default", "dark", "outline", "light"
```

### alert.go — Alert, AlertSuccess, AlertWarning, AlertDanger, AlertInfo

```go
gu.Alert("Operation succeeded!", "success")
// variants: "info", "success", "warning", "danger"
```

### stat.go — StatCard

```go
gu.StatCard("Total Users", "1,234")
```

### table.go — StyledTable, THead

```go
gu.StyledTable(
    gu.THead("Name", "Email", "Role"),
    gd.Tr(gd.Td(gp.Value("Alice")), gd.Td(gp.Value("alice@example.com")), gd.Td(gp.Value("Admin"))),
)
```

### form.go — FormGroup, FormLabel, FormInput, FormSelect, FormTextarea, Checkbox, Radio

```go
gu.FormGroup(
    gu.FormLabel("Email", "email"),
    gu.FormInput("email", gp.ID("email"), gp.Type("email"), gp.Placeholder("you@example.com")),
    gu.FormError(errMsg),
)

gu.FormSelect("role", []gu.FormOption{
    {Value: "admin", Label: "Admin"},
    {Value: "user", Label: "User"},
})

gu.Checkbox("agree", "I agree to the terms", checked)
gu.Radio("plan", "pro", "Pro plan", selected)
```

### icon.go — Icon

```go
gu.Icon("home")           // default size (md)
gu.Icon("search", "sm")   // small
gu.Icon("settings", "lg") // large
```

30+ built-in icons: `home`, `user`, `users`, `settings`, `search`, `plus`, `minus`, `x`, `check`, `edit`, `trash`, `chevron-right`, `chevron-down`, `chevron-left`, `chevron-up`, `menu`, `folder`, `file`, `mail`, `bell`, `calendar`, `chart`, `download`, `upload`, `link`, `info`, `warning`, `circle`, `lock`, `logout`, `filter`, `sort-asc`, `sort-desc`, `eye`, `copy`.

### spinner.go — Spinner

```go
gu.Spinner("sm")  // sizes: "sm", "md", "lg"
```

### misc.go — Divider, EmptyState, Progress

```go
gu.Divider()
gu.EmptyState("No items found")
gu.Progress(75)  // 0–100 percentage
```

### nav.go — Sidebar, SidebarHeader, SidebarLink, SidebarDivider, Breadcrumb, MobileHeader, PageHeader

See [Layout System](#layout-system) above.

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
    gu.ChatMessageView(gu.ChatMessage{Author: "Alice", Content: "Hello!", Timestamp: "09:00", Avatar: "A"}),
    gu.ChatMessageView(gu.ChatMessage{Content: "Hi!", Timestamp: "09:01", Sent: true}),
)
gu.ChatInput("msg", draft, gu.ChatInputOpts{
    Placeholder:  "Type a message...",
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

Sizes: `"xs"` (24px), `"sm"` (32px), `"md"` (40px), `"lg"` (48px), `"xl"` (64px). Shapes: `"circle"` (default), `"rounded"`.

### chart.go — LineChart, ColumnChart, BarChart, PieChart, ScatterPlot, Histogram, StackedBarChart

```go
series := []gu.Series{
    {Name: "Revenue", Points: []gu.DataPoint{{Label: "Jan", Value: 420}, {Label: "Feb", Value: 380}}},
}
gu.LineChart(series, gu.ChartOpts{Width: 600, Height: 400, Title: "Monthly Data", ShowGrid: true, ShowLegend: true})
gu.PieChart([]gu.DataPoint{{Label: "A", Value: 45}, {Label: "B", Value: 55}}, gu.ChartOpts{})
gu.Histogram(values, gu.HistogramOpts{ChartOpts: gu.ChartOpts{}, BinCount: 10})
```

## Interactive Components (Event Integration)

Interactive components use Opts structs with event fields. When event fields are empty, the component renders as static HTML. When set, `gerbera-*` attributes are added for LiveView event binding.

### Static vs Interactive Example

```go
// Static calendar (no events)
gu.Calendar(gu.CalendarOpts{
    Year:  2026,
    Month: time.March,
    Today: time.Now(),
})

// Interactive calendar (LiveView events)
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

### All Interactive Components

| Component | Opts Struct | Key Event Fields |
|-----------|-----------|-----------------|
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
| Charts | `ChartOpts` | `ClickEvent`, `MouseEnterEvent`, `MouseLeaveEvent` |

## ui/live/ Package

Components that require LiveView WebSocket connection. Import:

```go
gul "github.com/tomo3110/gerbera/ui/live"
```

| Component | Function | Description |
|-----------|----------|-------------|
| `Modal` | `Modal(open, closeEvent, children...)` | Overlay dialog with backdrop |
| `Toast` | `Toast(message, variant, dismissEvent)` | Notification toast (top-right) |
| `DataTable` | `DataTable(DataTableOpts)` | Sortable, paginated table |
| `Dropdown` | `Dropdown(open, toggleEvent, trigger, menu)` | Dropdown menu |
| `Confirm` | `Confirm(open, title, message, confirmEvent, cancelEvent)` | Confirmation dialog |
| `Tabs` | `Tabs(id, activeIndex, tabs, changeEvent)` | Tab panel with ARIA |
| `Drawer` | `Drawer(open, closeEvent, side, children...)` | Slide-out panel |
| `SearchSelect` | `SearchSelect(SearchSelectOpts)` | Filterable combobox with keyboard navigation |

### Modal Example

```go
gul.Modal(v.ModalOpen, "closeModal",
    gul.ModalHeader("User Details", "closeModal"),
    gul.ModalBody(content),
    gul.ModalFooter(
        gu.Button("Close", gu.ButtonOutline, gl.Click("closeModal")),
    ),
)
```

### DataTable Example

```go
gul.DataTable(gul.DataTableOpts{
    Columns: []gul.Column{
        {Key: "name", Label: "Name", Sortable: true},
        {Key: "email", Label: "Email", Sortable: true},
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

## CSS Naming Convention

All CSS classes use the `g-` prefix (e.g., `.g-card`, `.g-btn-primary`, `.g-grid-cols-3`). Components consume CSS custom properties from the theme (`--g-bg`, `--g-text`, `--g-accent`, etc.), so changing the theme automatically adapts every component's appearance.
