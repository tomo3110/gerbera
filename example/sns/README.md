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

The app follows the **single LiveView with internal routing** pattern from `example/admin/`: a single `SNSView` struct manages all pages via `v.page` state and a `switch` in `renderContent()`. Navigation events update the page state and the diff engine patches only the changed DOM.

Authentication uses the `session/` package with static login/register pages (same pattern as `example/auth/`), while the main app is a protected LiveView behind `session.RequireKey`.

Real-time notifications use the Hub pub/sub pattern from `example/chat/`: a shared `Hub` maps user IDs to `*live.Session` references, and `SendInfo()` pushes typed notification messages to the correct user's `HandleInfo()` method.

## File Structure

| File | Purpose |
|------|---------|
| `main.go` | Entry point, routes, MySQL init, session store |
| `db.go` | DB connection, auto-migration, all SQL queries |
| `models.go` | Data structs (User, Post, Like, Follow, etc.) |
| `hub.go` | Real-time notification Hub |
| `auth.go` | Login/register pages + POST handlers + password hashing |
| `view.go` | SNSView struct, Mount, HandleEvent, HandleUpload, HandleInfo, Unmount |
| `pages.go` | Render method + page renderers (timeline, profile, post, messages, settings, search) |
| `components.go` | Shared UI components (post card, nav links, compose box, SVG icons) |
| `css.go` | All CSS (mobile-first, responsive breakpoints) |
| `schema.sql` | Reference SQL schema |
| `docker-compose.yml` | MySQL 8.0 container |
