# Catalog Tutorial

## Overview

Build a single-page interactive UI component catalog that showcases every widget in the `ui/` and `ui/live/` packages. This tutorial covers the following features:

- `gu.AdminShell` + `gu.MobileHeader` + `gul.Drawer` — Responsive admin layout with mobile navigation
- `gu.Theme` / `gu.ThemeWith` / `gu.ThemeAuto` — Runtime theme switching (light, dark, custom, auto)
- SPA navigation via `gl.Click` + `gl.ClickValue` — Page dispatch with a `switch` statement
- 35+ component demo pages — Static widgets, live widgets, data visualization, theming
- 50+ event handlers — Toggle, input, sort, keyboard navigation, and more

The catalog is a **single-file LiveView application** (~1900 lines) that serves as both a visual reference and a working code example for every `ui/` component. Each page demonstrates both static and LiveView usage of its component.

## Prerequisites

- Go 1.22 or later installed
- This repository cloned locally

## Step 1: Architecture — Single-File Structure

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

The catalog uses six import aliases:

| Alias | Package | Purpose |
|-------|---------|---------|
| `g` | `gerbera` | Core types (`ComponentFunc`, `Element`) |
| `gd` | `dom` | HTML element helpers (`Div`, `Body`, `H1`, etc.) |
| `expr` | `expr` | Control flow (`If`, `Unless`, `Each`) |
| `gl` | `live` | LiveView runtime (`View`, `Handler`, `Click`, etc.) |
| `gp` | `property` | Attribute setters (`Class`, `Attr`, `Value`, etc.) |
| `gu` | `ui` | Static UI widgets (Card, Button, Table, etc.) |
| `gul` | `ui/live` | LiveView UI widgets (Modal, Drawer, DataTable, etc.) |

The entire catalog is a single `CatalogView` struct with approximately 87 fields that hold the state for all component demos:

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

	// ... ~60 more fields for other demos
}
```

Key design decisions:

- **Single struct** — All state for every demo page lives in one struct. Navigation switches the `Page` field, and only the active page renders its widgets.
- **No database or external state** — Everything is in-memory per session. This makes the catalog self-contained.
- **Flat organization** — Each page has its own `page*()` method (e.g., `pageCard()`, `pageModal()`, `pageCalendar()`).

## Step 2: Theme Switching — currentTheme() with 4 Modes

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

The `currentTheme()` method returns a `ComponentFunc` that emits CSS custom properties into the `<head>`. The four modes are:

| Mode | Function | Description |
|------|----------|-------------|
| `light` | `gu.Theme()` | Default light theme |
| `dark` | `gu.ThemeWith(gu.DarkTheme())` | Full dark theme via pre-built config |
| `custom` | `gu.ThemeWith(gu.ThemeConfig{...})` | Custom accent color override |
| `auto` | `gu.ThemeAuto(light, dark)` | Uses `prefers-color-scheme` media query to follow OS preference |

The theme is applied in `Render()` as a child of `gd.Head(...)`:

```go
gd.Head(
	gd.Title("Gerbera UI Catalog"),
	gd.Meta(gp.Attr("name", "viewport"), gp.Attr("content", "width=device-width, initial-scale=1")),
	v.currentTheme(),
),
```

When a user clicks a theme button, the `"setTheme"` event updates `v.ThemeMode`, and the entire page re-renders with the new CSS custom properties. Because Gerbera Live diffs the entire tree, only the changed `<style>` element and any affected attributes are patched.

## Step 3: Responsive Layout — AdminShell + MobileHeader + Drawer

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
			// Mobile nav drawer
			gul.Drawer(v.MobileNavOpen, "closeMobileNav", "left",
				gul.DrawerHeader("UI Catalog", "closeMobileNav"),
				gul.DrawerBody(v.renderNavLinks()),
			),
		),
	}
}
```

The layout uses three complementary components:

- **`gu.AdminShell(sidebar, content)`** — A two-column layout. On desktop (>768px) the sidebar is always visible. On mobile it is hidden.
- **`gu.MobileHeader(title, clickAttr)`** — A fixed top bar shown only on screens <=768px, with a hamburger menu button. The click binding triggers the `"toggleMobileNav"` event.
- **`gul.Drawer(open, closeEvent, side, children...)`** — A LiveView-driven slide-out panel. When `v.MobileNavOpen` is `true`, the drawer slides in from the left with the same navigation links as the sidebar.

This pattern gives the catalog a responsive admin dashboard layout with zero custom CSS.

## Step 4: Navigation — Categorized Sidebar and Page Dispatch

```go
func (v *CatalogView) renderNavLinks() g.ComponentFunc {
	return gd.Div(
		navLink(v, "overview", "Overview"),
		navDividerLabel("Static Widgets"),
		navLink(v, "card", "Card"),
		navLink(v, "button", "Button"),
		navLink(v, "icon", "Icon"),
		// ... more links
		navDividerLabel("Live Widgets"),
		navLink(v, "modal", "Modal"),
		navLink(v, "datatable", "DataTable"),
		// ... more links
		navDividerLabel("Data Visualization"),
		navLink(v, "chart", "Charts"),
		navLink(v, "avatar", "Avatar"),
		navDividerLabel("Theming"),
		navLink(v, "theme", "Theme"),
	)
}
```

Navigation links are organized into four categories: **Static Widgets**, **Live Widgets**, **Data Visualization**, and **Theming**.

Two helper functions keep the code DRY:

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

- `navLink` uses `gu.SidebarLink` with the third argument `v.Page == page` to highlight the active link
- `gl.Click("nav")` + `gl.ClickValue(page)` sends `{e: "nav", p: {value: "card"}}` to the server

Page dispatch is a simple `switch` statement in `renderContent()`:

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
	// ... 30+ more cases
	default:
		body = v.pageOverview()
	}
	return gd.Div(
		gu.PageHeader("Gerbera UI Catalog"),
		gd.Div(gp.Class("g-page-body"), body),
	)
}
```

When the user clicks a nav link, `HandleEvent` receives `"nav"` and updates `v.Page`. The framework re-renders, and the switch dispatches to the new page method. Only the content area is patched because the sidebar remains the same (with the active highlight changing).

## Step 5: Static Component Pages — Card and Button

Static component pages demonstrate `gu.*` widgets that require no LiveView state:

### Card Page

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

- `gu.Card(children...)` wraps content in a styled container with border and shadow
- `gu.CardHeader(title, actions...)` renders a header bar with optional action buttons
- `gu.CardFooter(children...)` renders a footer area

### Button Page

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

- Button variants are applied by passing modifier `ComponentFunc` values: `gu.ButtonPrimary`, `gu.ButtonOutline`, `gu.ButtonDanger`, `gu.ButtonSmall`
- Icons are composed inline with `gu.Icon(name, size)`
- `gu.Stack` provides vertical spacing, `gu.HStack` provides horizontal arrangement

The `section()` helper creates a consistent heading and description for each demo:

```go
func section(title, desc string) g.ComponentFunc {
	return gd.Div(
		gd.H2(gp.Attr("style", "font-size:18px;font-weight:600;margin:0 0 4px 0"), gp.Value(title)),
		gd.P(gp.Attr("style", "color:var(--g-text-secondary);margin:0 0 16px 0;font-size:13px"), gp.Value(desc)),
	)
}
```

## Step 6: Interactive Widget Pages — Calendar and Accordion

Interactive widget pages demonstrate components that use LiveView state and Opts structs for event binding.

### Calendar Page

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

The Calendar component accepts a `CalendarOpts` struct with both state (year, month, selected date) and event names. The component itself emits `gl.Click` bindings internally using the provided event names. This is the standard pattern for `ui/` components with LiveView support: the Opts struct bridges the component and the LiveView event system.

### Accordion Page

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

- `AccordionOpts{Exclusive: true}` means opening one section closes all others (the server handles this logic in `HandleEvent`)
- Without `ToggleEvent`, the Accordion falls back to native `<details>`/`<summary>` elements that work without JavaScript

Both pages also show a **Static** variant without event bindings, demonstrating that every `ui/` component works in both static and LiveView contexts.

## Step 7: LiveView Component Pages — Modal and DataTable

LiveView-specific components live in the `ui/live/` package (aliased as `gul`):

### Modal Page

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

- `gul.Modal(open, closeEvent, children...)` renders a modal overlay. When `v.ModalOpen` is `false`, the modal is hidden via CSS.
- The second argument `"closeModal"` is the event name sent when the overlay backdrop is clicked.
- `gul.ModalHeader`, `gul.ModalBody`, `gul.ModalFooter` structure the modal content with an X close button in the header.

### DataTable Page

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

- `gul.DataTable` takes a `DataTableOpts` struct that configures columns, rows (as `[][]string`), sorting, and pagination
- The component emits sort and pagination events using the event names from the Opts struct
- Server-side slicing (`rows[start:end]`) handles the actual pagination; the component only renders the current page

## Step 8: Event Handling — Representative HandleEvent Snippets

The `HandleEvent` method contains a single `switch` statement with 50+ cases. Here are representative patterns:

### Navigation

```go
case "nav":
	v.Page = payload["value"]
	v.resetOverlays()
```

`resetOverlays()` closes all modals, drawers, dropdowns, and toasts when navigating away:

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

### Toggle Pattern (Modal, Checkbox, Drawer)

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

### Sort Pattern (DataTable)

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

### Keyboard Navigation (SearchSelect)

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

### Calendar Navigation

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

## How It Works

```
Browser                                        Go Server
  |                                               |
  |  1. GET /                                     |
  |  <-- Initial HTML (overview page) + JS        |
  |                                               |
  |  2. WebSocket connect                         |
  |  <-> Session created, CatalogView allocated   |
  |                                               |
  |  3. Click "Card" in sidebar                   |
  |  --> {"e":"nav","p":{"value":"card"}}         |
  |                                               |
  |      HandleEvent:                             |
  |        v.Page = "card"                        |
  |        v.resetOverlays()                      |
  |      Render -> switch "card" -> pageCard()    |
  |      Diff old tree vs new tree                |
  |                                               |
  |  4. <-- Patches (content area only;           |
  |          sidebar active link highlight)        |
  |                                               |
  |  5. JS applies patches to DOM                 |
  |                                               |
  |  6. Click "Open Modal" button                 |
  |  --> {"e":"openModal","p":{}}                 |
  |                                               |
  |      HandleEvent: v.ModalOpen = true          |
  |      Render -> modal overlay appears          |
  |      Diff -> modal visibility patch           |
  |                                               |
  |  7. <-- Patches (modal display attribute)     |
```

All navigation is handled server-side via WebSocket events. The browser never makes a second HTTP request. The URL does not change, but the initial page can be set via query parameter: `/?page=calendar`.

## Running

```bash
go run example/catalog/catalog.go
```

Open http://localhost:8900 in your browser. Use the sidebar to browse all component pages. Resize the browser window to see the responsive mobile layout with the hamburger menu and slide-out drawer.

To change the port:

```bash
go run example/catalog/catalog.go -addr :3000
```

### Debug Mode

```bash
go run example/catalog/catalog.go -debug
```

When debug mode is enabled, a floating "G" button appears in the bottom-right corner. Click it or press `Ctrl+Shift+D` to open the Gerbera Debug Panel, which shows events, DOM patches, state, and session information in real time.

## Exercises

1. **Add a custom component page** — Create a new `pageMyWidget()` method, add a `navLink` entry, and add a case to the `switch` in `renderContent()`. Try composing existing `gu.*` widgets into a new pattern.
2. **Add a new theme** — Extend `currentTheme()` with a fifth mode (e.g., `"ocean"`) using a custom `gu.ThemeConfig` with blue/teal accent colors. Add a button for it on the Theme page.
3. **Add URL-based navigation** — Use `params["page"]` in `Mount()` (already partially implemented) to support deep linking to specific pages via `/?page=modal`.
4. **Add a search/filter** — Add a text input above the sidebar nav links that filters the visible links by label. Use `gl.Input("filterNav")` and `expr.If` to show/hide links.
5. **Add a new chart type** — Extend the Charts page with a new `ChartType` case and a corresponding `gu.ButtonGroupItem` entry.

## API Reference

| Function / Type | Description |
|---|---|
| `gu.AdminShell(sidebar, content)` | Two-column admin layout (responsive) |
| `gu.MobileHeader(title, click)` | Fixed top bar for mobile screens |
| `gu.Sidebar(children...)` | Sidebar container |
| `gu.SidebarHeader(title)` | Sidebar header with title |
| `gu.SidebarLink(href, label, active, attrs...)` | Navigation link with active state |
| `gu.SidebarDivider()` | Horizontal divider in sidebar |
| `gu.PageHeader(title)` | Page title header |
| `gu.Theme()` | Default light theme CSS |
| `gu.ThemeWith(config)` | Custom theme CSS from `ThemeConfig` |
| `gu.ThemeAuto(light, dark)` | OS-preference-based theme (media query) |
| `gu.DarkTheme()` | Pre-built dark `ThemeConfig` |
| `gu.Card(children...)` | Card container |
| `gu.Button(label, modifiers...)` | Styled button |
| `gu.Stack(children...)` | Vertical flex layout with gap |
| `gu.HStack(children...)` | Horizontal flex layout |
| `gu.Grid(modifier, children...)` | CSS Grid layout |
| `gu.Calendar(opts)` | Month-view calendar |
| `gu.Accordion(items, opts)` | Collapsible sections |
| `gu.NumberInput(name, value, opts)` | Number input with +/- buttons |
| `gu.Slider(name, value, opts)` | Range slider |
| `gu.TimePicker(name, h, m, s, opts)` | Time picker |
| `gu.Pagination(opts)` | Page navigation |
| `gu.ButtonGroup(items, opts)` | Segmented button control |
| `gul.Modal(open, closeEvent, children...)` | Modal dialog overlay |
| `gul.Toast(msg, variant, dismissEvent)` | Notification popup |
| `gul.DataTable(opts)` | Sortable, paginated table |
| `gul.Dropdown(open, toggleEvent, trigger, menu)` | Toggle dropdown menu |
| `gul.Tabs(id, index, tabs, switchEvent)` | Accessible tab panel |
| `gul.Drawer(open, closeEvent, side, children...)` | Slide-out panel |
| `gul.SearchSelect(opts)` | Filterable combobox |
| `gul.Confirm(open, title, msg, okEvent, cancelEvent)` | Confirmation dialog |
