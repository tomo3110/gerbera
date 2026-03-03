# Admin Dashboard Tutorial

## Overview

Build a single-page admin dashboard with real-time interactivity, demonstrating how to compose multiple Gerbera UI components into a full application:

- `gl.View` interface — Stateful LiveView with multi-page SPA navigation
- `gu.Theme()` — Built-in design system theme (CSS variables + reset)
- `gu.AdminShell()` — Sidebar + content layout scaffold
- `gu.Sidebar()` / `gu.SidebarLink()` — Navigation sidebar with active state
- `gu.StatCard()` / `gu.Grid()` — Dashboard metric cards in responsive grid
- `gu.Calendar()` — Interactive calendar with month/year navigation
- `gul.DataTable()` — Sortable, paginated data table (LiveView-aware)
- `gul.Modal()` / `gul.Confirm()` — Modal dialog and confirmation dialog
- `gu.ChatContainer()` / `gu.ChatInput()` — Real-time chat interface
- `gu.NumberInput()` / `gu.Slider()` — Form controls with increment/decrement events
- `gul.Toast()` — Dismissible notification toasts
- `gl.Click` / `gl.ClickValue` — Event binding for SPA page navigation

The entire application is a single 529-line Go file with no external dependencies beyond Gerbera itself.

## Prerequisites

- Go 1.22 or later installed
- This repository cloned locally

## Step 1: Importing Packages

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

Six aliased imports form the vocabulary of this application:

- `g` (gerbera) — Core types: `ComponentFunc`, `Literal()`
- `gd` (dom) — HTML element helpers: `Head()`, `Body()`, `Div()`, `Tr()`, `Td()`, etc.
- `expr` — Control flow: `If()` for conditional rendering
- `gl` (live) — LiveView runtime: `View` interface, `Handler()`, `Click()`, `ClickValue()`, `Payload`
- `gp` (property) — Attribute setters: `Class()`, `ID()`, `Attr()`, `Value()`, `Readonly()`
- `gu` (ui) — Static UI components: `Theme()`, `AdminShell()`, `Grid()`, `StatCard()`, `Calendar()`, `Card()`, `Button()`, `FormGroup()`, `NumberInput()`, `Slider()`, `ChatContainer()`, `ChatInput()`, etc.
- `gul` (ui/live) — LiveView-aware UI components: `DataTable()`, `Modal()`, `Confirm()`, `Toast()`

The distinction between `gu` and `gul` is important: `gu` components are pure rendering functions with no LiveView dependency, while `gul` components emit LiveView event attributes and require a WebSocket connection to function.

## Step 2: View Struct

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

The `AdminView` struct holds all state for the entire SPA. Fields are grouped by concern:

- **Navigation** — `Page` determines which page to render (`"dashboard"`, `"users"`, `"messages"`, `"settings"`)
- **User table** — `UserTablePage`, `SortCol`, `SortDir` control DataTable pagination and sorting; `SelectedUser` and `DeleteTarget` track which user is shown in the modal or targeted for deletion
- **Modals** — `ModalOpen` and `ConfirmOpen` toggle the user detail modal and delete confirmation dialog
- **Toast** — `ToastVisible`, `ToastMessage`, `ToastVariant` control the notification toast
- **Settings** — `ItemsPerPage` and `NotifyFreq` are form values bound to NumberInput and Slider
- **Calendar** — `CalYear`, `CalMonth`, `CalSelected` track calendar navigation and date selection
- **Messages** — `ChatMessages` stores chat history; `ChatDraft` holds the current input text

Each browser session gets its own `AdminView` instance, so all state is per-session.

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

`Mount` initializes all state to sensible defaults. Key details:

- `SelectedUser` and `DeleteTarget` start at `-1` (no selection)
- The calendar defaults to the current year and month
- Chat messages are pre-populated with sample data to demonstrate the chat UI
- `params.Get("page")` allows deep-linking to a specific page via URL query parameters (e.g., `?page=users`)

## Step 4: Theme and Layout

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

The `Render()` method returns two top-level elements: `<head>` and `<body>`.

- **`gu.Theme()`** injects the Gerbera design system CSS (CSS variables for colors, spacing, typography, and a CSS reset). This single call provides the entire visual foundation.
- **`gu.AdminShell(sidebar, content)`** creates a two-column layout: a fixed sidebar on the left and a scrollable content area on the right.
- **`gul.Toast()`** is conditionally rendered outside the shell so it overlays the entire page. The third argument `"dismissToast"` is the event name fired when the user clicks the close button.

## Step 5: Sidebar Navigation

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

SPA navigation is implemented entirely through LiveView events -- no client-side router is needed.

- **`gu.SidebarLink(href, label, active, attrs...)`** renders a styled navigation link. The third argument (`v.Page == "dashboard"`) controls the active/highlighted state.
- **`gl.Click("nav")`** binds a click event that sends `{"e":"nav","p":{"value":"dashboard"}}` to the server.
- **`gl.ClickValue("dashboard")`** sets the `value` field in the event payload.
- **`gu.SidebarDivider()`** renders a visual separator between nav groups.

When the user clicks a link, `HandleEvent` receives the `"nav"` event, updates `v.Page`, and the next render shows the corresponding page content. The WebSocket diff engine patches only the changed DOM nodes, so page transitions are instantaneous with no full-page reload.

## Step 6: Dashboard Page

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

The dashboard page demonstrates several layout and component patterns:

- **`gu.PageHeader(title, breadcrumb)`** renders a page title with breadcrumb navigation
- **`gu.Stack(children...)`** adds vertical spacing between children
- **`gu.Grid(cols, children...)`** creates a CSS grid; `gu.GridCols4` produces a 4-column grid, `gu.GridCols2` a 2-column grid
- **`gu.StatCard(label, value)`** renders a styled metric card
- **`gu.Card()` / `gu.CardHeader()`** creates a bordered card container with a header
- **`gu.StyledTable()` / `gu.THead()`** renders a styled HTML table with header columns
- **`gu.Calendar(CalendarOpts)`** renders an interactive calendar widget. The `Opts` struct maps each user interaction to a named event: clicking a date fires `"calSelect"`, clicking the previous/next arrows fires `"calPrev"`/`"calNext"`, and changing the month/year dropdowns fires `"calMonthChange"`/`"calYearChange"`.
- **`expr.If(condition, component)`** conditionally renders the selected date badge only when a date has been chosen

## Step 7: Users Page with DataTable

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

The users page combines three LiveView-aware components:

- **`gul.DataTable(DataTableOpts)`** renders a sortable, paginated table. The `Columns` slice defines which columns appear and whether they are sortable. `SortEvent` and `PageEvent` name the events fired when users click column headers or pagination controls. Pagination is computed server-side by slicing the `rows` array.
- **`gul.Modal(open, closeEvent, children...)`** renders a modal dialog that appears when `open` is `true`. The user detail modal shows name, email, role, and status in read-only form fields.
- **`gul.Confirm(open, title, message, confirmEvent, cancelEvent)`** renders a confirmation dialog with OK/Cancel buttons. This is used for the delete user confirmation.

### User Detail Modal

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
            // ... email, role fields ...
        ),
        gul.ModalFooter(
            gu.Button("Close", gu.ButtonOutline, gl.Click("closeUserModal")),
        ),
    )
}
```

Note the guard clause: when no user is selected (`SelectedUser < 0`), a hidden modal is still rendered. This is important because the LiveView diff engine needs a stable DOM tree -- if the modal element disappeared entirely, the diff would be unable to patch it into existence later.

## Step 8: Messages Page with Chat

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

The messages page uses three chat-related components:

- **`gu.ChatContainer(msgViews...)`** wraps chat messages in a scrollable container
- **`gu.ChatMessageView(msg)`** renders a single chat message. The `gu.ChatMessage` struct has `Author`, `Content`, `Timestamp`, `Sent` (true = outgoing), and `Avatar` fields. Sent messages are styled differently from received messages.
- **`gu.ChatInput(name, value, opts)`** renders a text input with a send button. Three events are bound:
  - `"chatInput"` fires on every keystroke, updating `ChatDraft`
  - `"chatSend"` fires when the send button is clicked
  - `"chatKeydown"` fires on keydown, allowing Enter key to send

## Step 9: Settings Page

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

The settings page demonstrates interactive form controls:

- **`gu.FormInput()`** / **`gu.FormLabel()`** / **`gu.FormGroup()`** create standard form fields
- **`gu.FormSelect(name, options, attrs...)`** renders a `<select>` dropdown with `FormOption` structs
- **`gu.NumberInput(name, value, opts)`** renders a number input with +/- buttons. `Min` and `Max` are `*int` pointers (nil means unbounded). `IncrementEvent` and `DecrementEvent` fire when the +/- buttons are clicked.
- **`gu.Slider(name, value, opts)`** renders a range slider. `InputEvent` fires whenever the slider value changes.

Note that `Min` and `Max` for `NumberInput` must be passed as pointers. Local variables (`min5`, `max50`) are declared and their addresses taken.

## Step 10: Event Handling

```go
func (v *AdminView) HandleEvent(event string, payload gl.Payload) error {
    switch event {
    // Navigation
    case "nav":
        v.Page = payload["value"]
        v.ModalOpen = false
        v.ConfirmOpen = false
        v.ToastVisible = false

    // User table sorting and pagination
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

    // User modal
    case "viewUser":
        fmt.Sscanf(payload["value"], "%d", &v.SelectedUser)
        v.ModalOpen = true
    case "closeUserModal":
        v.ModalOpen = false

    // Delete confirmation
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

    // Calendar
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

    // Chat
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

    // Settings
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

`HandleEvent` is the central event router for the entire SPA. All 20+ events are handled in a single `switch` statement, grouped by feature area. Key patterns:

- **Navigation cleanup** — The `"nav"` handler resets `ModalOpen`, `ConfirmOpen`, and `ToastVisible` to prevent stale UI state from leaking across page transitions.
- **Sort toggle** — Clicking the same column toggles between `"asc"` and `"desc"`; clicking a different column resets to `"asc"`.
- **Payload parsing** — `fmt.Sscanf(payload["value"], "%d", &v.UserTablePage)` is used to parse integer values from string payloads.
- **Chat send** — Both the send button (`"chatSend"`) and Enter key (`"chatKeydown"`) share the same logic: append a new `ChatMessage` with `Sent: true` and clear the draft.
- **Toast flow** — Delete confirmation triggers a toast by setting `ToastVisible = true` with a message and variant. The toast is dismissed by the `"dismissToast"` event.

## Step 11: Server

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

The server setup is minimal:

- **`gl.Handler(factory, opts...)`** creates an HTTP handler that serves the initial HTML and upgrades to WebSocket for live updates. The factory function `func(_ context.Context) gl.View { return &AdminView{} }` creates a fresh `AdminView` for each session. The `context.Context` parameter carries request-scoped values (e.g. authenticated user set by middleware).
- **`gl.WithDebug()`** enables the debug panel, which shows WebSocket messages, event payloads, and render timing in the browser.

## How It Works

```
Browser                                     Go Server
  |                                            |
  | 1. GET /                                   |
  | <-- Full HTML (AdminShell + Dashboard)     |
  |                                            |
  | 2. WebSocket connect                       |
  |                                            |
  | 3. Click "Users" in sidebar                |
  | --> {"e":"nav","p":{"value":"users"}}      |
  |                                            |
  |     HandleEvent: Page = "users"            |
  |     Render() -> Diff old vs new tree       |
  |                                            |
  | 4. <-- [{"op":"replace",                   |
  |          "path":[1,0,1],                   |
  |          "val":"<div>...users page...</div>"}]
  |                                            |
  | 5. JS patches DOM in place                |
  |    (sidebar active state + content area)   |
  |                                            |
  | 6. Click column header "Email"             |
  | --> {"e":"sort","p":{"value":"email"}}     |
  |                                            |
  |     HandleEvent: SortCol="email",SortDir="asc"
  |     Render() -> Diff -> table body patches |
  |                                            |
  | 7. <-- patches for table rows only        |
```

The key insight is that the entire SPA runs through a single WebSocket connection. Page transitions, modal opens, table sorts, and chat messages all follow the same cycle: event -> state mutation -> re-render -> diff -> minimal DOM patches.

## Running

```bash
# Standard mode
go run example/admin/admin.go

# With debug panel
go run example/admin/admin.go -debug

# Custom port
go run example/admin/admin.go -addr :9000
```

Open http://localhost:8910 and explore the dashboard, users table, chat, and settings pages.

## Exercises

1. **Add a new page** — Create a "Reports" page with charts or a data summary. Add a new sidebar link and a new case in `renderContent()`.
2. **Search filter** — Add a text input above the DataTable that filters users by name or email. Bind it with a `"userSearch"` event and filter the `rows` slice before passing it to `gul.DataTable()`.
3. **Theme toggle** — Add a dark/light mode toggle button in the sidebar footer. Store the theme preference in `AdminView` and conditionally add a CSS class to `<body>`.
4. **Persistent delete** — Currently `"doDelete"` only shows a toast. Implement actual deletion by removing the user from the `users` slice and adjusting pagination.
5. **Chat timestamps** — Display relative timestamps ("2 min ago") instead of absolute times, updating them via `TickerView`.
