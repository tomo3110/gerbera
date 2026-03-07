# Admin Dashboard Tutorial

## Overview

Build a multi-page admin dashboard that combines server-side rendered (SSR) pages with embedded LiveView components for real-time interactivity. This example demonstrates the **SSR + LiveMount hybrid architecture pattern**, where each page is a separate URL rendered as static HTML, and interactive sections are powered by independent LiveView components mounted within those pages.

Key concepts and components used:

- `g.Handler()` — Serves static SSR pages as standard HTTP handlers
- `gl.LiveMount()` — Embeds a LiveView component mount point inside an SSR page
- `gl.Handler()` — Registers LiveView WebSocket endpoints for each mount point
- `gl.View` interface — Stateful LiveView with `Mount`, `Render`, `HandleEvent`
- `gu.Theme()` — Built-in design system theme (CSS variables + reset)
- `gu.AdminShell()` — Sidebar + content layout scaffold
- `gu.Sidebar()` / `gu.SidebarLink()` — Navigation sidebar with regular `<a>` links
- `gu.StatCard()` / `gu.Grid()` — Dashboard metric cards in responsive grid
- `gu.Calendar()` — Interactive calendar with month/year navigation
- `gul.DataTable()` — Sortable, paginated data table (LiveView-aware)
- `gul.Modal()` / `gul.Confirm()` — Modal dialog and confirmation dialog
- `gu.ChatContainer()` / `gu.ChatInput()` — Real-time chat interface
- `gu.NumberInput()` / `gu.Slider()` — Form controls with increment/decrement events
- `gul.Toast()` — Dismissible notification toasts

The entire application is a single Go file with no external dependencies beyond Gerbera itself.

## Architecture: SSR + LiveMount Hybrid

Unlike a traditional single-page application (SPA) where one monolithic LiveView handles all pages through WebSocket navigation, this admin dashboard uses a hybrid approach:

```
Browser                                     Go Server
  |                                            |
  | 1. GET /admin                              |
  | <-- Full HTML (AdminShell + LiveMount div) |
  |                                            |
  | 2. LiveMount JS connects WebSocket to      |
  |    /admin/live/dashboard                   |
  |                                            |
  | 3. DashboardView.Mount() + Render()        |
  | <-- LiveView HTML injected into mount div  |
  |                                            |
  | 4. User clicks "Users" <a> link            |
  | --> GET /admin/users (full page load)      |
  |                                            |
  | 5. New SSR page with LiveMount for         |
  |    /admin/live/users                       |
  | <-- Full HTML (AdminShell + LiveMount div) |
  |                                            |
  | 6. LiveMount JS connects WebSocket to      |
  |    /admin/live/users                       |
  |                                            |
  | 7. UsersView.Mount() + Render()            |
  | <-- LiveView HTML injected into mount div  |
  |                                            |
  | 8. User sorts table column                 |
  | --> {"e":"sort","p":{"value":"email"}}     |
  |                                            |
  |     HandleEvent -> Render() -> Diff        |
  | <-- patches for table rows only            |
```

The key differences from a monolithic SPA approach:

- **Page navigation uses regular `<a>` links** -- clicking a sidebar link triggers a full page load (standard HTTP GET), not a WebSocket event
- **Each page has its own small View struct** -- `DashboardView`, `UsersView`, `MessagesView`, `SettingsView` instead of one giant `AdminView`
- **Each View manages only its own state** -- no need to reset modal/toast state on page transitions since each page starts fresh
- **Multiple independent WebSocket connections** -- each `LiveMount` establishes its own connection to its dedicated endpoint

This architecture is simpler to reason about: SSR handles the page shell and navigation, while LiveView handles the interactive parts within each page.

## Prerequisites

- Go 1.22 or later installed
- This repository cloned locally

## Step 1: Importing Packages

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

Seven aliased imports form the vocabulary of this application:

- `g` (gerbera) -- Core types: `ComponentFunc`, `Handler()`
- `gd` (dom) -- HTML element helpers: `Head()`, `Body()`, `Div()`, `Tr()`, `Td()`, etc.
- `expr` -- Control flow: `If()` for conditional rendering
- `gl` (live) -- LiveView runtime: `View` interface, `Handler()`, `LiveMount()`, `Click()`, `Payload`
- `gp` (property) -- Attribute setters: `Class()`, `ID()`, `Attr()`, `Value()`, `Readonly()`
- `gu` (ui) -- Static UI components: `Theme()`, `AdminShell()`, `Grid()`, `StatCard()`, `Calendar()`, `Card()`, `Button()`, `FormGroup()`, `NumberInput()`, `Slider()`, `ChatContainer()`, `ChatInput()`, etc.
- `gul` (ui/live) -- LiveView-aware UI components: `DataTable()`, `Modal()`, `Confirm()`, `Toast()`

The distinction between `gu` and `gul` is important: `gu` components are pure rendering functions with no LiveView dependency, while `gul` components emit LiveView event attributes and require a WebSocket connection to function.

Note the addition of `"context"` compared to the old monolithic approach -- the `gl.Handler()` factory function receives a `context.Context` parameter.

## Step 2: SSR Layout -- The Shared Page Shell

Instead of a single `Render()` method that switches between pages, the new architecture uses a shared layout function that produces SSR pages:

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

This function takes two arguments:

- `page` -- a string identifying the active page (used to highlight the correct sidebar link)
- `liveEndpoint` -- the URL path of the LiveView endpoint to mount (e.g., `"/admin/live/dashboard"`)

The return value is a slice of `ComponentFunc` values (the `<head>` and `<body>` elements) that `g.Handler()` will render as static HTML.

The critical line is `gl.LiveMount(liveEndpoint)`. This renders a placeholder `<div>` with a `data-live-mount` attribute. When the page loads in the browser, the Gerbera client-side JavaScript detects this attribute, opens a WebSocket connection to the specified endpoint, and injects the LiveView-rendered HTML into the placeholder. From that point on, the LiveView operates independently within that div.

**`gu.Theme()`** injects the Gerbera design system CSS (CSS variables for colors, spacing, typography, and a CSS reset). This single call provides the entire visual foundation.

**`gu.AdminShell(sidebar, content)`** creates a two-column layout: a fixed sidebar on the left and a scrollable content area on the right.

## Step 3: Sidebar Navigation with Regular Links

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

Navigation uses standard `<a>` tags with real URL paths -- no LiveView events, no `gl.Click()`, no `gl.ClickValue()`. This is a fundamental difference from the old SPA approach.

- **`gu.SidebarLink(href, label, active)`** renders an anchor tag pointing to a real URL. The third argument (`active == "dashboard"`) controls the highlighted state.
- **`gu.SidebarDivider()`** renders a visual separator between nav groups.

When the user clicks a link, the browser performs a standard HTTP GET request to the new URL, which returns a fresh SSR page with its own `LiveMount`. There is no WebSocket event, no DOM diff -- just a full page load. This keeps navigation simple and makes each page independently bookmarkable.

## Step 4: View Structs -- One Per Page

Instead of a single monolithic `AdminView` struct that holds state for all pages, each page has its own small, focused View struct that implements the `gl.View` interface (`Mount`, `Render`, `HandleEvent`).

### DashboardView

```go
type DashboardView struct {
    CalYear     int
    CalMonth    time.Month
    CalSelected *time.Time
}
```

The `DashboardView` holds only the state it needs: calendar navigation and date selection. No user table state, no chat state, no settings state.

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

The `UsersView` manages the data table, modal, confirmation dialog, and toast -- everything related to user management.

### MessagesView

```go
type MessagesView struct {
    ChatMessages []gu.ChatMessage
    ChatDraft    string
}
```

The `MessagesView` only tracks the chat message history and the current draft input.

### SettingsView

```go
type SettingsView struct {
    ItemsPerPage int
    NotifyFreq   int
}
```

The `SettingsView` holds just the two form control values.

Each browser session that visits a page gets its own instance of that page's View. Since each View is independent, there is no risk of state from one page leaking into another.

## Step 5: Mount -- Per-View Initialization

Each View's `Mount` method initializes only the state relevant to that page:

```go
func (v *DashboardView) Mount(_ gl.Params) error {
    now := time.Now()
    v.CalYear = now.Year()
    v.CalMonth = now.Month()
    return nil
}

func (v *UsersView) Mount(_ gl.Params) error {
    v.SortCol = "name"
    v.SortDir = "asc"
    v.SelectedUser = -1
    v.DeleteTarget = -1
    return nil
}

func (v *MessagesView) Mount(_ gl.Params) error {
    v.ChatMessages = []gu.ChatMessage{
        {Author: "System", Content: "Welcome to Admin Messages.", Timestamp: "09:00", Sent: false},
        {Author: "Alice Johnson", Content: "The new user report is ready for review.", Timestamp: "09:15", Sent: false, Avatar: "A"},
        {Content: "Thanks, I'll take a look now.", Timestamp: "09:16", Sent: true},
    }
    return nil
}

func (v *SettingsView) Mount(_ gl.Params) error {
    v.ItemsPerPage = 10
    v.NotifyFreq = 30
    return nil
}
```

Compare this to the old monolithic `Mount` that initialized 15+ fields for all pages at once. Each `Mount` is now short, focused, and easy to understand.

Note that `SelectedUser` and `DeleteTarget` start at `-1` (no selection) in `UsersView`. The chat messages are pre-populated with sample data in `MessagesView`.

## Step 6: Dashboard Page Render

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
                                activityRow("Bob Smith", "Uploaded document", "15 min ago"),
                                activityRow("Diana Prince", "Created post", "1 hour ago"),
                                activityRow("Frank Castle", "Deleted comment", "3 hours ago"),
                                activityRow("Grace Hopper", "Logged in", "5 hours ago"),
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

The dashboard demonstrates several layout and component patterns:

- **`gu.PageHeader(title, breadcrumb)`** renders a page title with breadcrumb navigation
- **`gu.Stack(children...)`** adds vertical spacing between children
- **`gu.Grid(cols, children...)`** creates a CSS grid; `gu.GridCols4` produces a 4-column grid, `gu.GridCols2` a 2-column grid
- **`gu.StatCard(label, value)`** renders a styled metric card
- **`gu.Card()` / `gu.CardHeader()`** creates a bordered card container with a header
- **`gu.StyledTable()` / `gu.THead()`** renders a styled HTML table with header columns
- **`gu.Calendar(CalendarOpts)`** renders an interactive calendar widget. The `Opts` struct maps each user interaction to a named event: clicking a date fires `"calSelect"`, clicking the previous/next arrows fires `"calPrev"`/`"calNext"`, and changing the month/year dropdowns fires `"calMonthChange"`/`"calYearChange"`.
- **`expr.If(condition, component)`** conditionally renders the selected date badge only when a date has been chosen

Since `DashboardView` is a LiveView, the calendar is fully interactive -- clicking dates, navigating months, and selecting years all happen over the WebSocket connection without reloading the page.

## Step 7: Users Page with DataTable

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

The users page combines four LiveView-aware components:

- **`gul.DataTable(DataTableOpts)`** renders a sortable, paginated table. The `Columns` slice defines which columns appear and whether they are sortable. `SortEvent` and `PageEvent` name the events fired when users click column headers or pagination controls. Pagination is computed server-side by slicing the `rows` array.
- **`gul.Modal()`** renders a modal dialog for user details.
- **`gul.Confirm(open, title, message, confirmEvent, cancelEvent)`** renders a confirmation dialog with OK/Cancel buttons for the delete user confirmation.
- **`gul.Toast()`** renders a dismissible notification that appears after a successful delete.

Note that the breadcrumb uses `Href: "/admin"` -- a real URL pointing back to the dashboard SSR page, not a LiveView event.

### User Detail Modal

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
            // ... email, role fields ...
        ),
        gul.ModalFooter(
            gu.Button("Close", gu.ButtonOutline, gl.Click("closeUserModal")),
        ),
    )
}
```

Note the guard clause: when no user is selected (`idx < 0`), a hidden modal is still rendered. This is important because the LiveView diff engine needs a stable DOM tree -- if the modal element disappeared entirely, the diff would be unable to patch it into existence later.

In the new architecture, `renderUserModal` is a standalone function (not a method on a View) since it only needs the `open` flag and user index as parameters.

## Step 8: Messages Page with Chat

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

The messages page uses three chat-related components:

- **`gu.ChatContainer(msgViews...)`** wraps chat messages in a scrollable container
- **`gu.ChatMessageView(msg)`** renders a single chat message. The `gu.ChatMessage` struct has `Author`, `Content`, `Timestamp`, `Sent` (true = outgoing), and `Avatar` fields. Sent messages are styled differently from received messages.
- **`gu.ChatInput(name, value, opts)`** renders a text input with a send button. Three events are bound:
  - `"chatInput"` fires on every keystroke, updating `ChatDraft`
  - `"chatSend"` fires when the send button is clicked
  - `"chatKeydown"` fires on keydown, allowing Enter key to send

## Step 9: Settings Page

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

The settings page demonstrates interactive form controls:

- **`gu.FormInput()`** / **`gu.FormLabel()`** / **`gu.FormGroup()`** create standard form fields
- **`gu.FormSelect(name, options, attrs...)`** renders a `<select>` dropdown with `FormOption` structs
- **`gu.NumberInput(name, value, opts)`** renders a number input with +/- buttons. `Min` and `Max` are `*int` pointers (nil means unbounded). `IncrementEvent` and `DecrementEvent` fire when the +/- buttons are clicked.
- **`gu.Slider(name, value, opts)`** renders a range slider. `InputEvent` fires whenever the slider value changes.

Note that `Min` and `Max` for `NumberInput` must be passed as pointers. Local variables (`min5`, `max50`) are declared and their addresses taken.

## Step 10: Event Handling -- Per-View Handlers

Instead of a single massive `HandleEvent` with 20+ cases, each View handles only its own events:

### DashboardView Events

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

### UsersView Events

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

### MessagesView Events

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

### SettingsView Events

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

Key patterns across the handlers:

- **Sort toggle** -- Clicking the same column toggles between `"asc"` and `"desc"`; clicking a different column resets to `"asc"`.
- **Payload parsing** -- `fmt.Sscanf(payload["value"], "%d", &v.TablePage)` is used to parse integer values from string payloads.
- **Chat send** -- Both the send button (`"chatSend"`) and Enter key (`"chatKeydown"`) share the same logic: append a new `ChatMessage` with `Sent: true` and clear the draft.
- **Toast flow** -- Delete confirmation triggers a toast by setting `ToastVisible = true` with a message and variant. The toast is dismissed by the `"dismissToast"` event.

Notice that there is no `"nav"` event handler anywhere. Navigation is handled entirely by the browser via `<a>` links and HTTP requests -- the LiveView layer does not need to know about page transitions at all.

## Step 11: Server -- Wiring SSR Pages and LiveView Endpoints

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

    // SSR pages -- each embeds a LiveMount for its interactive section
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

The server setup has two layers of route registration:

### SSR Page Routes

```go
mux.Handle("GET /admin", g.Handler(adminPage("dashboard", "/admin/live/dashboard")...))
```

**`g.Handler(components...)`** takes a variadic list of `ComponentFunc` values and returns an `http.Handler` that renders them as static HTML. The `adminPage()` function produces the full page shell (head, body, sidebar, LiveMount placeholder). These routes use the `"GET"` method prefix, meaning they only respond to GET requests.

Each SSR page is independently addressable -- `/admin`, `/admin/users`, `/admin/messages`, `/admin/settings` are all separate URLs that return complete HTML documents.

### LiveView Endpoint Routes

```go
mux.Handle("/admin/live/dashboard", gl.Handler(func(_ context.Context) gl.View { return &DashboardView{} }, opts...))
```

**`gl.Handler(factory, opts...)`** creates an HTTP handler that serves both the initial LiveView HTML (via HTTP) and the WebSocket upgrade for live updates. The factory function `func(_ context.Context) gl.View { return &DashboardView{} }` creates a fresh View instance for each session. The `context.Context` parameter carries request-scoped values (e.g., an authenticated user set by middleware).

These routes do NOT use a method prefix because they must handle both GET requests (for the initial HTML content injected by `LiveMount`) and WebSocket upgrade requests.

### How SSR and LiveMount Connect

The connection between an SSR page and its LiveView works as follows:

1. The browser requests `GET /admin/users`
2. `g.Handler()` renders the `adminPage("users", "/admin/live/users")` component tree as HTML
3. The HTML includes a `<div data-live-mount="/admin/live/users"></div>` placeholder (produced by `gl.LiveMount()`)
4. The browser loads the page and the Gerbera client JavaScript finds the `data-live-mount` div
5. The JavaScript opens a WebSocket connection to `/admin/live/users`
6. The server creates a new `UsersView`, calls `Mount()`, then `Render()`, diffs the result, and sends the initial HTML
7. The JavaScript injects the LiveView HTML into the placeholder div
8. From this point, all interactions within the LiveView (sorting, modals, toasts) happen over the WebSocket

### Debug Mode

**`gl.WithDebug()`** enables the debug panel, which shows WebSocket messages, event payloads, and render timing in the browser. Pass `-debug` when running:

```bash
go run example/admin/admin.go -debug
```

## Running

```bash
# Standard mode
go run example/admin/admin.go

# With debug panel
go run example/admin/admin.go -debug

# Custom port
go run example/admin/admin.go -addr :9000
```

Open http://localhost:8910 and explore the dashboard, users table, chat, and settings pages. Notice that clicking sidebar links triggers full page loads (watch the browser's loading indicator), while interactions within each page (sorting tables, opening modals, sending chat messages) happen instantly via WebSocket.

## Exercises

1. **Add a new page** -- Create a "Reports" page. Define a `ReportsView` struct, add an SSR route at `/admin/reports` using `g.Handler(adminPage("reports", "/admin/live/reports")...)`, register a LiveView endpoint at `/admin/live/reports` with `gl.Handler()`, and add a new `gu.SidebarLink()` in the `sidebar()` function.
2. **Search filter** -- Add a text input above the DataTable in `UsersView` that filters users by name or email. Add a `SearchQuery string` field to `UsersView`, bind it with a `"userSearch"` event, and filter the `rows` slice before passing it to `gul.DataTable()`.
3. **Theme toggle** -- Add a dark/light mode toggle button in the sidebar footer. Since the sidebar is rendered by SSR, consider using a cookie or query parameter to persist the theme choice across page loads, or embed a small LiveView in the sidebar itself.
4. **Persistent delete** -- Currently `"doDelete"` only shows a toast. Implement actual deletion by removing the user from the `users` slice and adjusting pagination.
5. **Shared state** -- The current architecture gives each LiveView its own state. Experiment with sharing data between Views using a shared in-memory store (e.g., a mutex-protected map) that multiple Views read from in their `Render()` methods.
