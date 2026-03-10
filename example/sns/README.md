# SNS (Twitter Clone)

A full-featured Twitter clone built with Gerbera LiveView, MySQL, and real-time notifications.

Features:
- **Timeline** — Post, like, retweet with real-time count updates
- **User profiles** — Follow/unfollow, post history, bio
- **Direct messages** — Real-time 1:1 chat via Hub pub/sub
- **Comments** — Threaded replies on posts
- **Search** — Debounced search for users and posts
- **Settings** — Edit profile, change password, upload avatar
- **Real-time notifications** — Likes, retweets, follows, and messages push instantly via `InfoReceiver`
- **Mobile-first** — Responsive layout with drawer navigation on mobile, sidebar on desktop
- **Session auth** — Login/register with CSRF protection, SHA-256 password hashing

## Usage

```bash
# Start everything (MySQL + app) with Docker Compose
cd example/sns && docker compose up

# Or run in background
cd example/sns && docker compose up -d

# Rebuild after code changes
cd example/sns && docker compose up --build
```

### Local Development (without Docker for the app)

```bash
# Start MySQL only
cd example/sns && docker compose up -d mysql

# Run the app locally
cd example/sns && go run .

# With debug panel
cd example/sns && go run . -debug

# Custom port and DSN
cd example/sns && go run . -addr :9000 -dsn "user:pass@tcp(host:3306)/dbname?parseTime=true"
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-addr` | `:8930` | Listen address |
| `-dsn` | `sns:snspass@tcp(127.0.0.1:3306)/sns?parseTime=true` | MySQL DSN |
| `-debug` | `false` | Enable Gerbera debug panel |

## Dependencies

- [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql) — MySQL driver for `database/sql`
- MySQL 8.0 (via `docker-compose.yml`)

## Architecture

This example uses an independent `go.mod` with `replace` directive (same pattern as `example/mdviewer`).

The app follows a **multi-View SSR shell** architecture: an SSR layout (`layout.go`) renders the sidebar navigation and embeds individual LiveView endpoints via `LiveMount`. Each page has its own View struct (`TimelineView`, `ProfileView`, `PostDetailView`, `MessagesView`, `SearchView`, `SettingsView`) with isolated state, all sharing common logic through an embedded `baseView`.

SSR shell pages handle routing and badge counts, while LiveView endpoints handle real-time interactivity:

```
GET /           → SSR shell → LiveMount("/live/timeline")  → TimelineView
GET /profile    → SSR shell → LiveMount("/live/profile")   → ProfileView
GET /post/{id}  → SSR shell → LiveMount("/live/post?id=…") → PostDetailView
GET /messages   → SSR shell → LiveMount("/live/messages")  → MessagesView
GET /search     → SSR shell → LiveMount("/live/search")    → SearchView
GET /settings   → SSR shell → LiveMount("/live/settings")  → SettingsView
```

Authentication uses the `session/` package with static login/register pages (same pattern as `example/auth/`), while all LiveView endpoints are protected behind `session.RequireKey`.

Real-time notifications use the Hub pub/sub pattern from `example/chat/`: a shared `Hub` maps user IDs to multiple `*live.Session` references (one per View/WebSocket), and `SendInfo()` pushes typed notification messages to the correct user's `HandleInfo()` method across all active Views.

## File Structure

| File | Purpose |
|------|---------|
| `main.go` | Entry point, routes, MySQL init, session store |
| `db.go` | DB connection, auto-migration, all SQL queries |
| `models.go` | Data structs (User, Post, Like, Follow, etc.) |
| `hub.go` | Real-time notification Hub with multi-session user mapping |
| `auth.go` | Login/register pages + POST handlers + password hashing |
| `layout.go` | SSR shell layout with sidebar, drawer, badge counts, LiveMount |
| `view_base.go` | Shared view logic: mountBase, unmountBase, post actions, notifications |
| `view_timeline.go` | TimelineView: home feed, compose, post/delete |
| `view_profile.go` | ProfileView: user profile, follow/unfollow |
| `view_post_detail.go` | PostDetailView: single post, comments, delete |
| `view_messages.go` | MessagesView: conversations, 1:1 chat |
| `view_search.go` | SearchView: debounced user/post search |
| `view_settings.go` | SettingsView: profile edit, password change, avatar upload |
| `components.go` | Shared UI components (post card, compose box, avatars, SVG icons) |
| `css.go` | All CSS (mobile-first, responsive breakpoints) |
| `schema.sql` | Reference SQL schema |
| `docker-compose.yml` | MySQL 8.0 container |
