# Gerbera Frontend Development Best Practices

A practical guide for building frontends with Gerbera. Covers patterns from basic composition to LiveView implementation techniques, performance optimization, and testing.

> **Audience**: Developers who understand the basics of Gerbera (`ComponentFunc`, `Tag()`, `dom/` package) and are building real applications.

---

## Table of Contents

1. [Project Structure](#1-project-structure)
2. [Component Design](#2-component-design)
3. [LiveView Basics](#3-liveview-basics)
4. [Event Handling Patterns](#4-event-handling-patterns)
5. [URL Management (PushPatch / HandleParams)](#5-url-management-pushpatch--handleparams)
6. [Forms and Validation](#6-forms-and-validation)
7. [Real-time Communication](#7-real-time-communication)
8. [JS Commands (CommandQueue)](#8-js-commands-commandqueue)
9. [Styling](#9-styling)
10. [Performance Optimization](#10-performance-optimization)
11. [Testing](#11-testing)
12. [Security](#12-security)
13. [Debugging](#13-debugging)
14. [Anti-patterns](#14-anti-patterns)

---

## 1. Project Structure

### Recommended Package Aliases

```go
import (
    g   "github.com/tomo3110/gerbera"            // Core: Tag, Literal, ComponentFunc
    gd  "github.com/tomo3110/gerbera/dom"         // HTML elements: Div, H1, Body, ...
    gp  "github.com/tomo3110/gerbera/property"    // Attributes: Class, Attr, ID, Value, ...
    ge  "github.com/tomo3110/gerbera/expr"        // Control flow: If, Unless, Each
    gs  "github.com/tomo3110/gerbera/styles"      // CSS: Style, CSS, ScopedCSS
    gl  "github.com/tomo3110/gerbera/live"        // LiveView: View, Handler, Click, ...
    gu  "github.com/tomo3110/gerbera/ui"          // UI components: Card, Badge, ...
    gul "github.com/tomo3110/gerbera/ui/live"     // LiveView-only UI: Modal, Toast, ...
)
```

Using consistent aliases across the project significantly improves code readability.

### File Organization Guidelines

For large-scale LiveView applications (see the SNS example), organize files as follows:

```
myapp/
├── main.go          # Entry point, routing, middleware
├── view.go          # View struct, Mount, HandleEvent, HandleInfo
├── pages.go         # Render() and per-page rendering functions
├── components.go    # Reusable ComponentFunc builders
├── css.go           # CSS string constants
├── db.go            # Database access functions
└── models.go        # Data model structs
```

**Tip**: Separating `HandleEvent` logic from `Render` UI construction into different files keeps both manageable as they grow.

---

## 2. Component Design

### Basics: Compose with Functions

All Gerbera components are `ComponentFunc` (`func(Node)`). Define them as regular Go functions and compose at the call site:

```go
// Good: Reusable component function
func userCard(name, email string) g.ComponentFunc {
    return gd.Div(
        gp.Class("user-card"),
        gd.H3(gp.Value(name)),
        gd.P(gp.Value(email)),
    )
}

// Usage
gd.Body(
    userCard("Alice", "alice@example.com"),
    userCard("Bob", "bob@example.com"),
)
```

### Conditional Rendering

Use `expr.If` / `expr.Unless`:

```go
ge.If(user.IsAdmin,
    gd.Button(gp.Value("Admin Panel"), gl.Click("openAdmin")),
)

ge.Unless(isLoggedIn,
    gd.A(gp.Href("/login"), gp.Value("Login")),
)
```

The third argument onwards serves as the `else` branch:

```go
ge.If(len(items) > 0,
    renderItemList(items),    // when true
    renderEmptyState(),       // when false
)
```

### List Rendering

Use `expr.Each`. For performance, **always include `property.Key()`**:

```go
ge.Each(posts, func(i int, post Post) g.ComponentFunc {
    return gd.Div(
        gp.Key(fmt.Sprintf("post-%d", post.ID)),  // Important: key for diff detection
        gp.Class("post-card"),
        gd.P(gp.Value(post.Content)),
    )
})
```

> **Note**: Without `Key`, adding/removing elements in a list may cause the entire tree to be rebuilt. Use a unique identifier (e.g., database ID).

### Skip() to Render Nothing

Use `g.Skip()` when a branch should produce no output:

```go
func maybeAlert(msg string) g.ComponentFunc {
    if msg == "" {
        return g.Skip()  // produces no output
    }
    return gd.Div(gp.Class("alert"), gp.Value(msg))
}
```

---

## 3. LiveView Basics

### Implementing the View Interface

A minimal LiveView implements 3 methods:

```go
type MyView struct {
    gl.CommandQueue  // Embed for JS command support
    count int
}

func (v *MyView) Mount(params gl.Params) error {
    // Initialize state. Access URL path via params.Path,
    // query parameters via params.Query,
    // and HTTP session via params.Conn.Session.
    v.count = 0
    return nil
}

func (v *MyView) Render() []g.ComponentFunc {
    // Return children of the <html> root.
    // Always include Head and Body.
    return []g.ComponentFunc{
        gd.Head(gd.Title("My App")),
        gd.Body(
            gd.H1(gp.Value(fmt.Sprintf("Count: %d", v.count))),
            gd.Button(gl.Click("inc"), gp.Value("+")),
        ),
    }
}

func (v *MyView) HandleEvent(event string, payload gl.Payload) error {
    // payload is map[string]string.
    // Branch on event name and update View state.
    // After return, Render -> Diff -> patch send happens automatically.
    switch event {
    case "inc":
        v.count++
    }
    return nil
}
```

### Information Available in Mount

The `Params` struct contains:

```go
func (v *MyView) Mount(params gl.Params) error {
    // URL path (e.g., "/profile")
    path := params.Path

    // URL query parameters (e.g., "?id=123" -> params.Get("id") == "123")
    id := params.Get("id")

    // HTTP session (authentication data, etc.)
    if params.Conn.Session != nil {
        userID, _ := params.Conn.Session.Get("user_id").(int64)
    }

    // LiveView session (for SendInfo)
    v.liveSession = params.Conn.LiveSession

    // Connection info
    _ = params.Conn.RemoteAddr  // client IP
    _ = params.Conn.UserAgent   // browser UA string
    return nil
}
```

### Handler Registration

```go
// Basic
http.Handle("/", gl.Handler(func(ctx context.Context) gl.View {
    return &MyView{}
}, gl.WithDebug()))

// Options
gl.WithLang("en")                   // HTML lang attribute (default "ja")
gl.WithSessionTTL(10 * time.Minute) // Session TTL (default 5 min)
gl.WithDebug()                      // Enable debug panel
gl.WithSessionStore(store)          // HTTP session integration
gl.WithMiddleware(authMW)           // Apply middleware
gl.WithCheckOrigin(fn)              // WebSocket Origin check
```

---

## 4. Event Handling Patterns

### Event Binding Reference

| Binding | HTML Attribute | Values in payload |
|---------|---------------|-------------------|
| `gl.Click("evt")` | `gerbera-click="evt"` | `"value"`: value set via `gl.ClickValue()` |
| `gl.Input("evt")` | `gerbera-input="evt"` | `"value"`: current input field value |
| `gl.Change("evt")` | `gerbera-change="evt"` | `"value"`: changed value |
| `gl.Submit("evt")` | `gerbera-submit="evt"` | All form fields (name -> value) |
| `gl.Keydown("evt")` | `gerbera-keydown="evt"` | `"key"`: pressed key name |
| `gl.Focus("evt")` | `gerbera-focus="evt"` | none |
| `gl.Blur("evt")` | `gerbera-blur="evt"` | none |
| `gl.Scroll("evt")` | `gerbera-scroll="evt"` | `scrollTop`, `scrollHeight`, `clientHeight`, etc. |

### Passing Data with ClickValue

Use `gl.ClickValue()` to attach data to click events on list items:

```go
ge.Each(posts, func(i int, post Post) g.ComponentFunc {
    id := strconv.FormatInt(post.ID, 10)
    return gd.Div(
        gp.Key(id),
        gd.P(gp.Value(post.Content)),
        gd.Button(
            gl.Click("toggleLike"),
            gl.ClickValue(id),   // set in payload["value"]
            gp.Value("Like"),
        ),
    )
})
```

HandleEvent side:

```go
case "toggleLike":
    postID, _ := strconv.ParseInt(payload["value"], 10, 64)
    if postID == 0 {
        return nil  // early return for invalid values
    }
    // ...
```

### Controlling Send Frequency with Debounce

For search inputs and text fields, use `gl.Debounce()` to throttle server sends:

```go
gd.Input(
    gp.Type("text"),
    gl.Input("searchInput"),
    gl.Debounce(300),           // wait 300ms before sending
    gp.Placeholder("Search..."),
)
```

> **Note**: Without Debounce, every keystroke sends an event to the server. **Always add Debounce to text input events**.

### Keydown Filtering

Use `gl.Key()` to react only to specific keys:

```go
gd.Input(
    gl.Keydown("chatSend"),
    gl.Key("Enter"),            // only send on Enter
    gl.Input("chatInput"),      // capture input value via separate event
)
```

### Form Submission

`gl.Submit()` automatically collects all form fields:

```go
gd.Form(
    gl.Submit("saveProfile"),
    gd.Input(gp.Name("email"), gp.Type("email"), gp.Value(v.email)),
    gd.Input(gp.Name("name"), gp.Value(v.name)),
    gd.Button(gp.Type("submit"), gp.Value("Save")),
)
```

```go
case "saveProfile":
    email := payload["email"]
    name := payload["name"]
    // ...
```

---

## 5. URL Management (PushPatch / HandleParams)

### Overview

LiveView page transitions do not change the browser URL by default. Using `PushPatch`, you can sync the URL with page transitions, enabling browser back/forward and bookmarks.

### Three Commands

| Command | Use Case | Browser History |
|---------|----------|----------------|
| `v.PushPatch(path)` | In-page navigation (URL change + history entry) | Added |
| `v.ReplacePatch(path)` | URL replacement (e.g., sort changes, no back entry) | Replaced |
| `v.PushNavigate(path)` | Full page navigation (WebSocket disconnect) | Added |

### Patcher Interface

Implement the `Patcher` interface to handle browser back/forward:

```go
// Implementing Patcher causes HandleParams to be called on popstate events
func (v *MyView) HandleParams(path string, params url.Values) error {
    // path: URL path (e.g., "/profile")
    // params: query parameters (e.g., "id=123")
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

### Calling PushPatch in HandleEvent

For events that trigger page transitions, call `PushPatch` after updating state:

```go
case "nav":
    v.page = payload["value"]
    // ... update state ...
    v.PushPatch(v.buildPath())  // sync URL

case "viewPost":
    postID, _ := strconv.ParseInt(payload["value"], 10, 64)
    v.page = "post"
    v.loadPostDetail(postID)
    v.PushPatch(v.buildPath())  // call after state change
```

### buildPath Helper Pattern

Create a function that builds URLs from View state for consistency:

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

### Initial URL Path Handling in Mount

Use `params.Path` in `Mount` to determine the initial page:

```go
func (v *MyView) Mount(params gl.Params) error {
    // Determine initial page from URL path
    switch params.Path {
    case "/profile":
        v.page = "profile"
    case "/settings":
        v.page = "settings"
    default:
        v.page = "home"
    }
    // Query parameters are also available
    if id := params.Get("id"); id != "" {
        // ...
    }
    return v.loadPageData()
}
```

> **Note**: Ensure your server-side routing matches the URLs changed by `PushPatch` to the LiveView handler. In Go's `http.ServeMux`, `mux.Handle("/", handler)` acts as a catch-all.

### Path Parameters (MatchPath)

Use `live.MatchPath` to handle path-based parameters like `/users/:id`:

```go
// MatchPath compares a pattern against a path and extracts ":param" segments
params, ok := gl.MatchPath("/users/:id", "/users/42")
// params.Get("id") == "42", ok == true

// Multiple parameters
params, ok := gl.MatchPath("/users/:uid/posts/:pid", "/users/1/posts/99")
// params.Get("uid") == "1", params.Get("pid") == "99"

// No match
_, ok := gl.MatchPath("/users/:id", "/posts/42")
// ok == false
```

#### Building a Route Table

For applications with multiple patterns, structure route definitions:

```go
type route struct {
    page   string
    params gl.PathParams
}

func matchRoute(path string) route {
    // Match parameterized routes first
    patterns := []struct {
        pattern string
        page    string
    }{
        {"/profile/:id", "profile"},
        {"/post/:id", "post"},
        {"/messages/:id", "messages"},
    }
    for _, p := range patterns {
        if params, ok := gl.MatchPath(p.pattern, path); ok {
            return route{page: p.page, params: params}
        }
    }
    // Fallback for static routes
    // ...
}
```

#### Usage in HandleParams

```go
func (v *MyView) HandleParams(path string, params url.Values) error {
    r := matchRoute(path)
    v.page = r.page
    switch v.page {
    case "profile":
        if idStr := r.params.Get("id"); idStr != "" {
            id, _ := strconv.ParseInt(idStr, 10, 64)
            return v.loadProfile(id)
        }
        return v.loadProfile(v.userID)
    case "search":
        // Query parameters are still accessed via url.Values
        v.keyword = params.Get("keyword")
    }
    return nil
}
```

#### Path Generation in buildPath

```go
func (v *MyView) buildPath() string {
    switch v.page {
    case "profile":
        if v.profileUser != nil && v.profileUser.ID != v.userID {
            return "/profile/" + strconv.FormatInt(v.profileUser.ID, 10)
        }
        return "/profile"
    case "search":
        if v.keyword != "" {
            return "/search?keyword=" + url.QueryEscape(v.keyword)
        }
        return "/search"
    }
    return "/"
}
```

> **Tip**: `MatchPath` only handles path segment parameter extraction. Query strings (`?keyword=...`) are still processed via `url.Values`.

---

## 6. Forms and Validation

### live.Field / live.SelectField

The `live/` package includes accessible form field helpers:

```go
gl.Field("email", gl.FieldOpts{
    Type:        "email",
    Label:       "Email Address",
    Placeholder: "you@example.com",
    Value:       v.email,
    Required:    true,
    Event:       "validateEmail",  // LiveView input event name
    Error:       v.errors.ErrorFor("email"),
})
```

### Managing Validation Errors

Use the `live.Errors` type to manage validation errors:

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
            v.errors.AddError("email", "Please enter a valid email address")
        }
    case "submit":
        if v.errors.HasErrors() {
            return nil  // don't submit if there are errors
        }
        // save...
    }
    return nil
}
```

### Real-time Validation

For validation during input, combine `gl.Input` with `gl.Debounce`:

```go
gd.Input(
    gp.Name("email"),
    gp.Type("email"),
    gp.Value(v.email),
    gl.Input("validateEmail"),
    gl.Debounce(500),  // validate 500ms after input stops
)
```

---

## 7. Real-time Communication

### TickerView — Periodic Updates

Use for periodic server-side data polling (e.g., clocks, dashboards):

```go
type DashboardView struct {
    stats Stats
}

func (v *DashboardView) TickInterval() time.Duration {
    return 5 * time.Second  // update every 5 seconds
}

func (v *DashboardView) HandleTick() error {
    v.stats = fetchLatestStats()  // fetch from DB or API
    return nil
    // after return, Render -> Diff -> patch send happens automatically
}
```

> **Note**: Returning `0` from `TickInterval()` disables ticking. Too-short intervals increase CPU load, so choose an appropriate interval for your use case.

### InfoReceiver — Inter-user Communication

Use `InfoReceiver` + `Session.SendInfo()` to broadcast real-time notifications:

```go
// 1. Define typed messages
type NewMessageNotif struct {
    SenderName string
    Content    string
}

// 2. Implement InfoReceiver on the View
func (v *ChatView) HandleInfo(msg any) error {
    switch m := msg.(type) {
    case NewMessageNotif:
        v.messages = append(v.messages, m)
        v.ScrollIntoPct("#messages", "1.0")  // scroll to bottom
    }
    return nil
}

// 3. Call SendInfo from the sender side
func (v *ChatView) HandleEvent(event string, payload gl.Payload) error {
    if event == "send" {
        // Send to the partner's LiveView session
        partnerSession.SendInfo(NewMessageNotif{
            SenderName: v.user.Name,
            Content:    payload["value"],
        })
    }
    return nil
}
```

### Unmounter — Cleanup on Disconnect

Release resources when the WebSocket closes:

```go
func (v *ChatView) Unmount() {
    hub.Leave(v.sessionID, v.userID)  // leave chat room
}
```

### SessionExpiredHandler — Session Invalidation

Handle session invalidation from another tab (e.g., logout):

```go
func (v *MyView) OnSessionExpired() error {
    v.showExpiredMessage = true
    v.Navigate("/login")  // redirect client
    return nil
}
```

---

## 8. JS Commands (CommandQueue)

### Usage

Embed `gl.CommandQueue` in your View to manipulate client DOM from the server without writing JavaScript:

```go
type MyView struct {
    gl.CommandQueue
    // ...
}
```

### Command Reference

| Method | Description |
|--------|-------------|
| `v.ScrollTo(selector, top, left)` | Scroll to position |
| `v.ScrollIntoPct(selector, pct)` | Scroll to 0.0–1.0 percentage |
| `v.Focus(selector)` | Focus element |
| `v.Blur(selector)` | Remove focus |
| `v.SetAttribute(sel, key, val)` | Set attribute |
| `v.RemoveAttribute(sel, key)` | Remove attribute |
| `v.AddClass(sel, class)` | Add CSS class |
| `v.RemoveClass(sel, class)` | Remove CSS class |
| `v.ToggleClass(sel, class)` | Toggle CSS class |
| `v.SetProperty(sel, key, val)` | Set DOM property |
| `v.Dispatch(sel, event)` | Dispatch custom event |
| `v.Show(selector)` | Show (display="") |
| `v.Hide(selector)` | Hide (display="none") |
| `v.Toggle(selector)` | Toggle visibility |
| `v.Navigate(url)` | Full page navigation |
| `v.PushPatch(path)` | URL change (add history) |
| `v.ReplacePatch(path)` | URL replace (overwrite history) |
| `v.PushNavigate(path)` | Full page navigation (WS disconnect) |

### Examples

```go
func (v *MyView) HandleEvent(event string, payload gl.Payload) error {
    switch event {
    case "sendMessage":
        // After sending a message, scroll chat area to bottom
        v.messages = append(v.messages, newMsg)
        v.ScrollIntoPct("#chat-messages", "1.0")

    case "openModal":
        v.modalOpen = true
        v.Focus("#modal-input")  // focus input in modal

    case "search":
        v.doSearch(payload["value"])
        v.PushPatch("/search?keyword=" + url.QueryEscape(payload["value"]))
    }
    return nil
}
```

> **Note**: Commands are sent to the client together with patches after `HandleEvent` / `HandleTick` / `HandleInfo` returns. Multiple commands can be queued within a single event.

---

## 9. Styling

### CSS Management

In Gerbera, CSS is managed as Go string constants:

```go
// css.go
const appCSS = `
.container { max-width: 800px; margin: 0 auto; padding: 16px; }
.card { border: 1px solid var(--g-border); border-radius: 8px; padding: 16px; }
`

// In Render
gd.Head(
    gu.Theme(),              // inject Gerbera theme variables
    gs.CSS(appCSS),          // insert CSS as <style>
)
```

### Theme System

The `ui/` package theme system defines CSS custom properties:

```go
gu.Theme()                                // default light theme
gu.ThemeWith(gu.DarkTheme())              // dark theme
gu.ThemeWith(gu.ThemeConfig{              // custom theme
    Primary: "#6366f1",
    Radius:  "12px",
})
gu.ThemeAuto(gu.ThemeConfig{}, gu.DarkTheme())  // auto-switch with OS setting
```

Use theme variables in your CSS for consistent design:

```css
.my-card {
    background: var(--g-surface);
    color: var(--g-text);
    border: 1px solid var(--g-border);
    border-radius: var(--g-radius);
}
```

### ScopedCSS

Scope component-specific styles with `ScopedCSS`:

```go
gs.ScopedCSS(".my-widget", map[string]string{
    "padding":    "16px",
    "background": "var(--g-surface)",
})
// Output: .my-widget { padding: 16px; background: var(--g-surface); }
```

### Inline Styles

Use `styles.Style` for dynamic styles:

```go
gs.Style(g.StyleMap{
    "width":  fmt.Sprintf("%d%%", progress),
    "height": "4px",
})

// Conditional
gs.StyleIf(isHighlighted, g.StyleMap{
    "background": "#fef3c7",
    "border-left": "4px solid #f59e0b",
})
```

### Responsive Design

Use `styles.MediaQuery` for media queries:

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

## 10. Performance Optimization

### Use Key (Most Important)

Always add `property.Key()` to elements in lists. It is used by the diff algorithm to identify element identity:

```go
// Good
gd.Div(
    gp.Key(fmt.Sprintf("user-%d", user.ID)),
    gp.Value(user.Name),
)

// Bad — without Key, all elements may be rebuilt
gd.Div(gp.Value(user.Name))
```

### Keep Render Output Stable

The Diff algorithm returns empty patches when nothing has changed. If Render returns the same tree, no unnecessary DOM updates occur:

```go
// Good: same tree returned when state hasn't changed
func (v *MyView) Render() []g.ComponentFunc {
    return []g.ComponentFunc{
        gd.Head(gd.Title("App")),
        gd.Body(gd.H1(gp.Value(fmt.Sprintf("Count: %d", v.count)))),
    }
}

// Bad: random values or current time in Render causes patches every time
func (v *MyView) Render() []g.ComponentFunc {
    return []g.ComponentFunc{
        gd.Body(gd.P(gp.Value(time.Now().String()))),  // changes every time!
    }
}
```

### Keep HandleEvent Lightweight

HandleEvent is called for every client event. Avoid heavy processing:

```go
// Good: fetch only necessary data
case "toggleLike":
    dbToggleLike(v.db, v.userID, postID)
    v.refreshCurrentPosts()  // only reload current page

// Bad: reload all data every time
case "toggleLike":
    dbToggleLike(v.db, v.userID, postID)
    v.loadAllData()  // fetches data for pages not being viewed
```

### Don't Skip Debounce

Always add `gl.Debounce()` to input events (`gl.Input`):

```go
// Good
gd.Input(gl.Input("search"), gl.Debounce(300))

// Bad: sends to server on every keystroke
gd.Input(gl.Input("search"))
```

---

## 11. Testing

### Unit Testing with TestView

Use `live.TestView` to test LiveViews without WebSocket:

```go
func TestCounterIncrement(t *testing.T) {
    tv := gl.NewTestView(&CounterView{})
    if err := tv.Mount(gl.Params{}); err != nil {
        t.Fatal(err)
    }

    // Simulate an event
    patches, err := tv.SimulateEvent("inc", gl.Payload{})
    if err != nil {
        t.Fatal(err)
    }

    // Verify patches were generated
    if tv.PatchCount() == 0 {
        t.Error("no patches generated")
    }

    // Verify text update patch is included
    if !tv.HasPatch(diff.OpSetText) {
        t.Error("missing text update patch")
    }

    // Verify HTML output
    html, _ := tv.RenderHTML()
    if !strings.Contains(html, "Count: 1") {
        t.Errorf("expected 'Count: 1', got: %s", html)
    }
}
```

### Testing URL Navigation with SimulateParams

Test Views that implement `Patcher`:

```go
func TestHandleParams(t *testing.T) {
    tv := gl.NewTestView(&MyView{})
    tv.Mount(gl.Params{})

    // Simulate browser back button
    params := url.Values{"id": []string{"42"}}
    patches, err := tv.SimulateParams("/profile", params)
    if err != nil {
        t.Fatal(err)
    }

    // Verify view state
    view := tv.View.(*MyView)
    if view.page != "profile" {
        t.Errorf("expected page=profile, got %s", view.page)
    }
}
```

### SimulateTick / SimulateInfo

```go
// Test TickerView
patches, err := tv.SimulateTick()

// Test InfoReceiver
patches, err := tv.SimulateInfo(NewMessageNotif{Content: "hello"})
```

### Testing ViewLoop with TestTransport

For more integrated testing, use `TestTransport`:

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

    // Inject an event
    tr.PushEvent("inc", gl.Payload{})
    time.Sleep(50 * time.Millisecond)

    // Verify sent messages
    msg := tr.LastMessage()
    if len(msg.Patches) == 0 {
        t.Error("no patches sent")
    }

    tr.Close()
    wg.Wait()
}
```

---

## 12. Security

### XSS Prevention

`property.Attr()` automatically applies `html.EscapeString`. **Never pass user input directly to `Literal()`**:

```go
// Good: auto-escaped
gp.Value(userInput)
gp.Attr("title", userInput)

// DANGEROUS: not escaped!
g.Literal(userInput)
g.Literal(fmt.Sprintf("<div>%s</div>", userInput))
```

Use `Literal()` only for trusted HTML (e.g., SVG icons).

### CSRF Protection

For static forms (non-LiveView), use the `session` package CSRF helpers:

```go
token := session.GenerateCSRFToken(sess)
// Include as hidden field in form

// Verify on submission
if !session.ValidCSRFToken(sess, submittedToken) {
    http.Error(w, "CSRF token invalid", http.StatusForbidden)
    return
}
```

LiveView WebSocket connections automatically verify the CSRF token at connect time, so no additional check is needed in `HandleEvent`.

### Session Management

```go
// Session auth guard — redirect unauthenticated users
authGuard := session.RequireKey("user_id", "/login")
mux.Handle("/", authGuard(gl.Handler(factory, gl.WithSessionStore(store))))
```

### WebSocket Origin Check

Origin header validation is enabled by default. You can customize with `WithCheckOrigin` for special environments, but **never disable it in production**.

---

## 13. Debugging

### Debug Panel

Enable `gl.WithDebug()` to show a DevPanel in the browser:

```go
gl.Handler(factory, gl.WithDebug())
```

- `Ctrl+Shift+D` or the **"G" button** in the bottom-right to toggle
- **Events tab**: sent event names, payloads, and timestamps
- **Patches tab**: received DOM patch counts and server processing times
- **State tab**: current View struct state as JSON
- **Session tab**: session ID, TTL, and connection status

> **Note**: Use `WithDebug()` only during development. Disabling it in production removes all overhead — no debug JS injection or timing measurement.

### Server-side Logging

In debug mode, structured logs are also output via `log/slog`:

```
[INFO] session_created  session_id=abc123
[INFO] event_received   session_id=abc123 event=inc payload=map[]
[INFO] patches_generated session_id=abc123 count=1 duration=0.2ms
```

---

## 14. Anti-patterns

### Side Effects in Render

```go
// Bad: executing DB queries in Render
func (v *MyView) Render() []g.ComponentFunc {
    posts, _ := dbGetPosts(v.db)  // Don't fetch here!
    return []g.ComponentFunc{ ... }
}

// Good: fetch data in HandleEvent / Mount / HandleTick and store in fields
func (v *MyView) HandleEvent(event string, _ gl.Payload) error {
    v.posts, _ = dbGetPosts(v.db)
    return nil
}
```

Render should only build views — perform side effects (DB access, file I/O) during state update phases (Mount, HandleEvent, HandleTick, HandleInfo).

### Unnecessary State Duplication

```go
// Bad: storing computable values as fields
type MyView struct {
    items    []Item
    count    int      // computable via len(items)
    hasItems bool     // computable via len(items) > 0
}

// Good: compute in Render
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

### Ignoring Errors in HandleEvent

```go
// Bad: swallowing errors
case "save":
    dbSave(v.db, data)  // no error check

// Good: notify the user of errors
case "save":
    if err := dbSave(v.db, data); err != nil {
        v.showToast("Failed to save", "danger")
        return nil  // returning err logs it as HandleEvent error
    }
    v.showToast("Saved successfully", "success")
```

When `HandleEvent` returns an `error`, processing is aborted and **patch sending to the client is also skipped**. If you want to show the user an error, reflect it in the View state and return `nil`.

### Overusing Literal()

```go
// Bad: writing entire HTML with Literal
g.Literal(fmt.Sprintf(`<div class="card"><h2>%s</h2><p>%s</p></div>`, title, body))

// Good: use Gerbera's type-safe API
gd.Div(
    gp.Class("card"),
    gd.H2(gp.Value(title)),
    gd.P(gp.Value(body)),
)
```

`Literal()` carries XSS risk and bypasses Diff algorithm optimization. Use it only for trusted static strings like inline SVG icons.

### Massive HandleEvent

```go
// Bad: all events in one giant switch (hundreds of lines)
func (v *MyView) HandleEvent(event string, payload gl.Payload) error {
    switch event {
    case "nav": // 20 lines
    case "save": // 30 lines
    case "delete": // 25 lines
    // ... 30 more cases
    }
}

// Good: delegate events to category-specific methods
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

## Reference Links

- [README.md](README.md) — Project overview and Quick Start
- [ui/README.md](ui/README.md) — UI component library API reference
- [example/counter/TUTORIAL.md](example/counter/TUTORIAL.md) — LiveView introductory tutorial
- [example/sns/TUTORIAL.md](example/sns/TUTORIAL.md) — Full-featured SNS application tutorial
- [example/chat/TUTORIAL.md](example/chat/TUTORIAL.md) — Real-time chat (InfoReceiver)
- [example/auth/TUTORIAL.md](example/auth/TUTORIAL.md) — Session authentication and CSRF
- [example/jscommand/TUTORIAL.md](example/jscommand/TUTORIAL.md) — JS command examples
