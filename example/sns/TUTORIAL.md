# SNS (Twitter Clone) Tutorial

This example demonstrates a full-featured Twitter clone combining MySQL persistence, session authentication, multi-view LiveView architecture, real-time Hub notifications, file uploads, and mobile-first responsive design using Gerbera's `ui/` component library.

## Overview

The SNS example implements:

- **Session auth** — Login/register with CSRF protection and SHA-256 password hashing (static pages)
- **Multi-View SSR shell** — Each page has its own View struct, mounted inside an SSR shell layout via `LiveMount`
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

### SSR Shell + LiveMount Architecture

Unlike the `example/admin/` single-SPA pattern, this example uses SSR shell pages that embed individual LiveView endpoints via `LiveMount`. Each page has its own View struct with isolated state:

```
GET /           → SSR shell (layout.go) → LiveMount("/live/timeline")  → TimelineView
GET /profile    → SSR shell             → LiveMount("/live/profile")   → ProfileView
GET /post/{id}  → SSR shell             → LiveMount("/live/post?id=…") → PostDetailView
GET /messages   → SSR shell             → LiveMount("/live/messages")  → MessagesView
GET /search     → SSR shell             → LiveMount("/live/search")    → SearchView
GET /settings   → SSR shell             → LiveMount("/live/settings")  → SettingsView
```

The SSR shell (`layout.go`) renders the sidebar navigation, mobile drawer, and badge counts. The `LiveMount` component connects to the corresponding LiveView endpoint via WebSocket.

### baseView — Shared Logic

All view structs embed `baseView`, which provides common fields and helpers:

```go
type baseView struct {
    gl.CommandQueue
    db   *sql.DB
    hub  *Hub
    sess *gl.Session

    userID int64
    user   *User

    toastMessage string
    toastVariant string
    toastVisible bool
}
```

- `mountBase(params)` — Reads user ID from HTTP session, loads user from DB, joins Hub
- `unmountBase()` — Leaves Hub when WebSocket closes
- `handlePostAction(event, payload)` — Shared like/retweet/share/dismiss-toast event handling
- `handleBaseInfo(msg)` — Shared notification handling (like, retweet, follow, comment toasts)

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
    if len(parts) != 2 {
        return false
    }
    salt, err := hex.DecodeString(parts[0])
    if err != nil {
        return false
    }
    hash := sha256.Sum256(append(salt, []byte(password)...))
    return parts[1] == hex.EncodeToString(hash[:])
}
```

### Hub with Multi-Session User Mapping

The SNS Hub maps database user IDs to multiple LiveView sessions (one per View/WebSocket), enabling targeted notifications:

```go
type Hub struct {
    mu       sync.RWMutex
    clients  map[string]*live.Session         // session ID → Session
    userSess map[int64]map[string]struct{} // user DB ID → set of session IDs
}

func (h *Hub) Notify(userID int64, msg any) {
    if sids, ok := h.userSess[userID]; ok {
        for sid := range sids {
            if sess, ok := h.clients[sid]; ok {
                sess.SendInfo(msg)
            }
        }
    }
}

func (h *Hub) Broadcast(msg any, senderSessionID string) {
    for id, sess := range h.clients {
        if id == senderSessionID {
            continue
        }
        sess.SendInfo(msg)
    }
}
```

Seven typed notification messages are used so `HandleInfo` can distinguish them:

```go
type NewLikeNotif struct{ PostID int64; ActorName string }
type NewRetweetNotif struct{ PostID int64; ActorName string }
type NewFollowNotif struct{ ActorName string }
type NewMessageNotif struct{ SenderID int64; Content string }
type NewCommentNotif struct{ PostID int64; ActorName string }
type NewPostNotif struct{ AuthorName string }
type NewCommentOnViewedPostNotif struct{ PostID int64; ActorName string }
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

1. Parses flags (`-addr`, `-dsn`, `-debug`) with env var overrides (`DATABASE_DSN`, `LISTEN_ADDR`)
2. Opens MySQL connection and runs auto-migration
3. Creates `session.MemoryStore` and session middleware
4. Creates upload directory (`uploads/avatars/`)
5. Registers routes:
   - `GET /avatars/` — Static file serving for uploaded avatars
   - `GET /login` — Renders login form (redirects to `/` if already authenticated)
   - `POST /login` — Validates credentials, sets `user_id` in session
   - `GET /register` — Renders registration form
   - `POST /register` — Creates user, sets `user_id` in session
   - `/logout` — Destroys session, redirects to `/login`
   - SSR shell pages (`/`, `/profile`, `/profile/{id}`, `/post/{id}`, `/messages`, `/messages/{id}`, `/search`, `/settings`) — Each protected by `session.RequireKey("user_id", "/login")`
   - LiveView endpoints (`/live/timeline`, `/live/profile`, `/live/post`, `/live/messages`, `/live/search`, `/live/settings`) — Each creates its own View struct

```go
mux.Handle("/live/timeline", authGuard(gl.Handler(func(_ context.Context) gl.View {
    return NewTimelineView(db, hub)
}, liveOpts...)))
```

### auth.go — Static Auth Pages

Login and register pages are server-side rendered (not LiveView) using `g.Handler()`. They share a `renderAuthPage()` function that generates the HTML based on `isRegister` flag, with `loginPage()` and `registerPage()` as convenience wrappers. POST error responses use `g.ExecuteTemplate()` directly for status code control.

CSRF tokens are generated via `session.GenerateCSRFToken(sess)` and validated on form submission. Username validation uses a regex pattern: `^[a-zA-Z0-9_]{1,30}$`.

### layout.go — SSR Shell Layout

The `snsPage()` function renders the full HTML shell with sidebar navigation:

```go
func snsPage(title, activePage, liveEndpoint string, badges badgeCounts) g.Components {
    return g.Components{
        gd.Head(/* title, viewport, theme, CSS */),
        gd.Body(
            // Mobile header with drawer toggle button
            // Main layout: desktop sidebar + LiveMount content area
            // Mobile drawer overlay with navigation links
        ),
    }
}
```

`sidebarLinks()` renders 6 navigation items (Home, Search, Notifications, Messages, Profile, Settings) plus a Logout button. Each link uses `sidebarLink()` which supports an optional badge count.

`fetchBadgeCounts()` queries notification and message counts from the database for the current user's session.

### view_base.go — Shared View Logic

The `baseView` struct holds common fields and provides helper methods used by all views:

- `mountBase(params)` — Reads user ID from HTTP session, loads user, joins Hub
- `unmountBase()` — Leaves Hub when WebSocket disconnects
- `showToast(msg, variant)` — Sets toast state for display
- `handlePostAction(event, payload)` — Handles `toggleLike`, `toggleRetweet`, `sharePost`, `dismissToast` events shared across views
- `handleBaseInfo(msg)` — Handles common notifications (like, retweet, follow, comment) with toast display

### view_timeline.go — TimelineView

Displays the home timeline with compose box and post cards.

**State**: `posts`, `composeDraft`, `composeChars`, `confirmOpen`, `confirmAction`

**Events**: `composeInput`, `submitPost`, `confirmDeletePost`, `doDelete`, `cancelDelete`, plus shared post actions

**HandleUpload**: Receives `postPhoto` uploads (currently logs only)

**HandleInfo**: Refreshes timeline on `NewPostNotif`, `NewLikeNotif`, etc.

### view_profile.go — ProfileView

Displays a user's profile with avatar, stats, follow button, and post list.

**Mount**: Reads `id` query param; defaults to own profile if absent

**Events**: `toggleFollow`, plus shared post actions

### view_post_detail.go — PostDetailView

Displays a single post with comments and comment form. Owner can delete.

**Events**: `commentInput`, `submitComment`, `confirmDeletePost`, `doDelete`, `cancelDelete`, plus shared post actions

**HandleInfo**: Refreshes comments on `NewCommentOnViewedPostNotif`

### view_messages.go — MessagesView

Conversation list or active chat view.

**HandleParams**: Switches between conversation list and chat based on URL path

**Events**: `openChat`, `backToConversations`, `chatInput`, `chatSend`, `chatKeydown`

**HandleInfo**: Refreshes chat in real-time on `NewMessageNotif` if viewing that conversation

### view_search.go — SearchView

Debounced search for users and posts.

**Events**: `searchInput` with `gl.Debounce(300)`, plus shared post actions

**HandleParams**: Loads results from `keyword` query parameter

### view_settings.go — SettingsView

Profile editing, password change, and avatar upload.

**Events**: `settingsDisplayNameInput`, `settingsEmailInput`, `settingsBioInput`, `saveProfile`, `savePassword`

**HandleUpload**: Saves `avatarUpload` to `uploads/avatars/` and updates DB

### components.go — Shared Components

Reusable `ComponentFunc` builders used across multiple views:

- **`postCard(p PostWithMeta)`** — Renders a post with avatar, author link, content link, image, and action buttons (like/retweet/comment/share). Like and retweet buttons change color based on state. Navigation uses `<a href>` links.
- **`composeBox(draft, charCount)`** — Post composition area with character counter (250 max) and photo upload button.
- **`renderSearchUserItem(u User)`** — Clickable search result for a user with avatar and name.
- **`avatarComponent(u User)`** — Returns `gu.ImageAvatar` or `gu.LetterAvatar` based on whether the user has an avatar path.
- **`iconHeart(filled)`**, **`iconRetweet()`**, **`iconComment()`**, **`iconShare()`** — Inline SVG icons via `g.Literal()`.
- **`formatTimeAgo(t)`** — Relative time string ("now", "5m", "2h", "3d", "Jan 2").
- **`truncate(s, maxLen)`** — Truncates string with "..." suffix.

### css.go — Mobile-First CSS

All CSS is defined as a Go string constant. Key responsive patterns:

- Base styles target mobile (single column, full-width cards)
- `@media (min-width: 769px)` shows the desktop sidebar and hides the mobile header
- Post action buttons use `inline-flex` with icon + count
- Active like: `color: #ef4444` (red), active retweet: `color: #22c55e` (green)
- Gerbera theme variables (`var(--g-*)`) are used throughout for consistent theming
- Drawer CSS is defined in `layout.go` alongside the drawer markup

### db.go — Database Layer

All database operations are plain functions accepting `*sql.DB`:

- **Timeline query** joins posts with users and aggregates like/retweet/comment counts with `EXISTS` subqueries for the current user's interaction state
- **Toggle operations** (like, retweet, follow) check for existing rows and insert/delete accordingly
- **Conversation list** uses a subquery to find the latest message per conversation partner

## Notification Flow

```
User A likes User B's post
  → HandleEvent("toggleLike") on User A's TimelineView (or ProfileView, etc.)
  → baseView.doToggleLike(): dbToggleLike(db, A.userID, postID)
  → dbCreateNotification(db, B.userID, A.userID, "like", &postID)
  → hub.Notify(B.userID, NewLikeNotif{PostID, A.DisplayName})
  → User B's HandleInfo(NewLikeNotif{...}) called on all active Views
  → baseView.handleBaseInfo() shows toast: "Alice liked your post"
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
2. **Image uploads for posts** — Currently `HandleUpload` in `TimelineView` logs the file but doesn't store it. Save uploaded images to disk (like `SettingsView` does for avatars) and store the path in `posts.image_path`.
3. **Notification page** — Add a `NotificationsView` and `/live/notifications` endpoint that lists all notifications from the `notifications` table with read/unread state.
4. **Bookmark feature** — Add a bookmarks table and a bookmark button on post cards. Create a `BookmarksView` to view saved posts.
5. **Dark mode** — Add a theme toggle in settings. Store the preference in the user's session and conditionally apply a `dark` class to `<body>` with inverted CSS variables.
