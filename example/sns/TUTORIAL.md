# SNS (Twitter Clone) Tutorial

This example demonstrates a full-featured Twitter clone combining MySQL persistence, session authentication, multi-page LiveView routing, real-time Hub notifications, file uploads, and mobile-first responsive design using Gerbera's `ui/` component library.

## Overview

The SNS example implements:

- **Session auth** — Login/register with CSRF protection and SHA-256 password hashing (static pages)
- **Single LiveView SPA** — Internal page routing via `v.page` state (same pattern as `example/admin/`)
- **MySQL persistence** — Users, posts, likes, retweets, follows, comments, messages, notifications
- **Real-time Hub** — Typed notification messages pushed via `InfoReceiver` + `SendInfo()`
- **File uploads** — Avatar and post image uploads via `UploadHandler`
- **Mobile-first UI** — Drawer navigation on mobile, sidebar on desktop, responsive post cards

## Prerequisites

- Go 1.22 or later installed
- Docker and Docker Compose installed (for MySQL)
- This repository cloned locally

## Key Concepts

### Independent Module

Like `example/mdviewer`, this example has its own `go.mod` with a `replace` directive to reference the local gerbera module. This keeps the MySQL driver dependency (`go-sql-driver/mysql`) isolated from the root module.

```go
module github.com/tomo3110/gerbera/example/sns

require (
    github.com/go-sql-driver/mysql v1.8.1
    github.com/tomo3110/gerbera v0.0.0
)

replace github.com/tomo3110/gerbera => ../..
```

### Password Hashing

The example uses SHA-256 with a random salt (standard library only, no bcrypt dependency):

```go
func hashPassword(password string) string {
    salt := make([]byte, 16)
    rand.Read(salt)
    hash := sha256.Sum256(append(salt, []byte(password)...))
    return hex.EncodeToString(salt) + ":" + hex.EncodeToString(hash[:])
}

func verifyPassword(stored, password string) bool {
    parts := strings.SplitN(stored, ":", 2)
    salt, _ := hex.DecodeString(parts[0])
    hash := sha256.Sum256(append(salt, []byte(password)...))
    return parts[1] == hex.EncodeToString(hash[:])
}
```

### Hub with User Mapping

Unlike the chat example's simple session-based Hub, the SNS Hub maps database user IDs to LiveView sessions, enabling targeted notifications to specific users:

```go
type Hub struct {
    mu      sync.RWMutex
    clients map[string]*live.Session // live session ID → session
    userMap map[int64]string         // user DB ID → live session ID
}

func (h *Hub) Notify(userID int64, msg any) {
    if sessID, ok := h.userMap[userID]; ok {
        if sess, ok := h.clients[sessID]; ok {
            sess.SendInfo(msg)
        }
    }
}
```

Typed notification messages are used so `HandleInfo` can distinguish them:

```go
type NewLikeNotif struct {
    PostID    int64
    ActorName string
}
type NewMessageNotif struct {
    SenderID int64
    Content  string
}
```

### Auto-Migration

Instead of requiring manual SQL execution, `db.go` runs `CREATE TABLE IF NOT EXISTS` statements on startup:

```go
func autoMigrate(db *sql.DB) error {
    queries := []string{
        `CREATE TABLE IF NOT EXISTS users (...)`,
        `CREATE TABLE IF NOT EXISTS posts (...)`,
        // ...
    }
    for _, q := range queries {
        if _, err := db.Exec(q); err != nil {
            return fmt.Errorf("migration: %w", err)
        }
    }
    return nil
}
```

## Walkthrough

### main.go — Entry Point

1. Parses flags (`-addr`, `-dsn`, `-debug`)
2. Opens MySQL connection and runs auto-migration
3. Creates `session.MemoryStore` and session middleware
4. Registers routes:
   - `GET /login` — Renders login form (redirects to `/` if already authenticated)
   - `POST /login` — Validates credentials, sets `user_id` in session
   - `GET /register` — Renders registration form
   - `POST /register` — Creates user, sets `user_id` in session
   - `/logout` — Destroys session, redirects to `/login`
   - `/` — Protected LiveView (`SNSView`) behind `session.RequireKey("user_id", "/login")`

The LiveView handler passes `gl.WithSessionStore(store)` to enable session access in `Mount()`:

```go
mux.Handle("/", authGuard(gl.Handler(func(_ context.Context) gl.View {
    return &SNSView{
        db:  db,
        hub: hub,
    }
}, liveOpts...)))
```

### auth.go — Static Auth Pages

Login and register pages are server-side rendered (not LiveView) using `g.ExecuteTemplate()`. They share a `renderAuthPage()` function that generates the HTML based on `isRegister` flag.

CSRF tokens are generated via `session.GenerateCSRFToken(sess)` and validated on form submission.

### view.go — SNSView

The `SNSView` struct holds all application state:

```go
type SNSView struct {
    gl.CommandQueue
    db      *sql.DB
    hub     *Hub
    session *gl.Session

    userID int64
    user   *User
    page   string  // "home", "profile", "post", "messages", "settings", "search"

    // Page-specific state...
    posts        []PostWithMeta
    composeDraft string
    // ...
}
```

**Mount** reads the user ID from the HTTP session, loads the user from the database, joins the Hub, and loads initial page data:

```go
func (v *SNSView) Mount(params gl.Params) error {
    v.session = params.Conn.LiveSession
    if params.Conn.Session != nil {
        if uid, ok := params.Conn.Session.Get("user_id").(int64); ok {
            v.userID = uid
        }
    }
    // Load user, join hub, load page data...
}
```

**HandleEvent** routes 25+ events grouped by feature:

- **Navigation**: `"nav"`, `"toggleDrawer"`, `"closeDrawer"`
- **Compose**: `"composeInput"`, `"submitPost"`
- **Post actions**: `"toggleLike"`, `"toggleRetweet"`, `"viewPost"`, `"viewProfile"`, `"sharePost"`
- **Follow**: `"toggleFollow"`
- **Comments**: `"commentInput"`, `"submitComment"`
- **Messages**: `"openChat"`, `"backToConversations"`, `"chatInput"`, `"chatSend"`, `"chatKeydown"`
- **Settings**: `"settingsDisplayNameInput"`, `"settingsEmailInput"`, `"settingsBioInput"`, `"saveProfile"`, `"savePassword"`
- **Search**: `"searchInput"` (with `gl.Debounce(300)`)
- **Delete**: `"confirmDeletePost"`, `"doDelete"`, `"cancelDelete"`

**HandleInfo** receives typed notifications from the Hub and updates UI accordingly:

```go
func (v *SNSView) HandleInfo(msg any) error {
    switch m := msg.(type) {
    case NewLikeNotif:
        v.showToast(fmt.Sprintf("%s liked your post", m.ActorName), "info")
        v.unreadNotifications++
        v.refreshCurrentPosts()
    case NewMessageNotif:
        if v.page == "messages" && v.chatPartnerID == m.SenderID {
            v.loadChat(m.SenderID)  // refresh chat in real-time
        } else {
            v.showToast("New message received", "info")
        }
    // ...
    }
}
```

**Unmount** leaves the Hub when the WebSocket closes:

```go
func (v *SNSView) Unmount() {
    if v.session != nil {
        v.hub.Leave(v.session.ID, v.userID)
    }
}
```

### pages.go — Page Rendering

The `Render()` method builds the full page layout:

```go
func (v *SNSView) Render() []g.ComponentFunc {
    return []g.ComponentFunc{
        gd.Head(
            gd.Title("SNS"),
            gd.Meta(gp.Attr("name", "viewport"), gp.Attr("content", "width=device-width, initial-scale=1")),
            gu.Theme(),
            gs.CSS(snsCSS),
        ),
        gd.Body(
            // Mobile header (hidden on desktop via CSS)
            // Mobile drawer (gul.Drawer)
            // Main layout: sidebar + content
            // Toast overlay
            // Confirm dialog
        ),
    }
}
```

`renderContent()` dispatches to page-specific methods via `switch v.page`.

Each page method (e.g., `renderTimeline()`, `renderProfile()`, `renderMessages()`) composes `ui/` and `ui/live/` components:

- **Timeline**: `composeBox()` + list of `postCard()` components
- **Profile**: `gu.LetterAvatar` / `gu.ImageAvatar`, follow button, profile stats, post list
- **Post detail**: `postCard()` + comment list + comment form
- **Messages**: Conversation list or `gu.ChatContainer` / `gu.ChatInput` for active chat
- **Settings**: `gu.FormGroup` / `gu.FormInput` / `gu.FormTextarea` for profile editing
- **Search**: `gu.FormInput` with `gl.Debounce(300)`, split results for users and posts

### components.go — Shared Components

Reusable `ComponentFunc` builders used across multiple pages:

- **`postCard(p PostWithMeta)`** — Renders a post with avatar, author name, content, image, and action buttons (like/retweet/comment/share). Like and retweet buttons change color based on state.
- **`navLink(icon, label, page, active, badgeCount)`** — Sidebar/drawer navigation item with optional unread badge.
- **`composeBox(draft, charCount)`** — Post composition area with character counter and photo upload button.
- **`iconHeart(filled)`**, **`iconRetweet()`**, **`iconComment()`**, **`iconShare()`** — Inline SVG icons via `g.Literal()`.
- **`avatarComponent(u User)`** — Returns `gu.ImageAvatar` or `gu.LetterAvatar` based on whether the user has an avatar path.

### css.go — Mobile-First CSS

All CSS is defined as a Go string constant. Key responsive patterns:

- Base styles target mobile (single column, full-width cards)
- `@media (min-width: 769px)` shows the desktop sidebar and hides the mobile header
- Post action buttons use `inline-flex` with icon + count
- Active like: `color: #ef4444` (red), active retweet: `color: #22c55e` (green)
- Gerbera theme variables (`var(--g-*)`) are used throughout for consistent theming

### db.go — Database Layer

All database operations are plain functions accepting `*sql.DB`:

- **Timeline query** joins posts with users and aggregates like/retweet/comment counts with `EXISTS` subqueries for the current user's interaction state
- **Toggle operations** (like, retweet, follow) check for existing rows and insert/delete accordingly
- **Conversation list** uses a subquery to find the latest message per conversation partner

## Notification Flow

```
User A likes User B's post
  → HandleEvent("toggleLike") on User A's View
  → dbToggleLike(db, A.userID, postID)
  → dbCreateNotification(db, B.userID, A.userID, "like", &postID)
  → hub.Notify(B.userID, NewLikeNotif{PostID, A.DisplayName})
  → User B's HandleInfo(NewLikeNotif{...}) is called
  → User B sees toast: "Alice liked your post"
  → User B's post counts refresh in real-time
```

## Running

```bash
# Start everything (MySQL + app) with Docker Compose
cd example/sns && docker compose up

# Or for local development: MySQL in Docker, app locally
cd example/sns && docker compose up -d mysql
cd example/sns && go run .             # http://localhost:8930
cd example/sns && go run . -debug      # with debug panel
```

1. Open http://localhost:8930 — you will be redirected to the login page
2. Click "Register" to create a new account
3. Create a second account in another browser/incognito window
4. Post, like, follow, and send messages — actions appear in real-time in both tabs

## Exercises

1. **Infinite scroll** — Replace the fixed 50-post limit on the timeline with `gu.Pagination` or implement infinite scrolling by tracking an offset and loading more posts on a `"loadMore"` event.
2. **Image uploads** — Currently `HandleUpload` logs the file but doesn't store it. Save uploaded images to disk and store the path in `posts.image_path` or `users.avatar_path`.
3. **Notification page** — Add a dedicated notifications page that lists all notifications from the `notifications` table with read/unread state.
4. **Bookmark feature** — Add a bookmarks table and a bookmark button on post cards. Create a "Bookmarks" page to view saved posts.
5. **Dark mode** — Add a theme toggle in settings. Store the preference in the user's session and conditionally apply a `dark` class to `<body>` with inverted CSS variables.
