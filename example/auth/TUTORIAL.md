# Auth Example Tutorial

This example demonstrates session-based authentication using the `session/` package with CSRF protection, session middleware, and push-based session invalidation.

## Overview

The auth example implements:

- **Login form** with CSRF token protection (server-side rendered)
- **Protected dashboard** as an SSR page that embeds a LiveView via `gl.LiveMount`, guarded by `session.RequireKey` middleware
- **Session invalidation** — logging out in one tab automatically redirects all other tabs via `SessionExpiredHandler`

## Architecture

The dashboard uses a hybrid SSR + LiveView approach:

- **SSR shell** (`dashboardPage()`) — rendered by `g.Handler()`, provides the `<head>` (title, theme CSS) and outer layout (`<body>`, centering, container). It embeds the LiveView with `gl.LiveMount("/live/dashboard")`.
- **LiveView** (`DashboardView`) — mounted inside the SSR shell. Its `Render()` only returns `<body>` content (the welcome card, session-expired warning, logout button). It does **not** render `<head>` or the full page layout.

This separation means the static page structure is served as plain HTML on the initial request, while the LiveView handles the dynamic, WebSocket-driven portion.

## Key Concepts

### MemoryStore

`session.MemoryStore` stores session data in memory with HMAC-SHA256 signed cookies. The cookie only contains the session ID — all data stays on the server.

```go
key := []byte("example-secret-key-change-in-prod")
store := session.NewMemoryStore(key)
defer store.Close()  // stops background GC goroutine
```

Options include `WithMaxAge(duration)`, `WithCookie(config)`, and `WithGCInterval(duration)`.

### Session Middleware

`session.Middleware` wraps an HTTP handler to automatically load the session from the cookie on each request and save it after the handler returns:

```go
sessionMW := session.Middleware(store)
handler := sessionMW(mux)
```

The session is then available via `session.FromContext(r.Context())`.

### CSRF Protection

The `session/` package provides CSRF token helpers using constant-time comparison:

```go
sess := session.FromContext(r.Context())
token := session.GenerateCSRFToken(sess)  // create token, store in session
// Include token in a hidden form field:
// <input type="hidden" name="csrf_token" value="...">

// On form submission, validate:
if !session.ValidCSRFToken(sess, r.FormValue("csrf_token")) {
    http.Error(w, "invalid CSRF token", http.StatusForbidden)
    return
}
```

### RequireKey Auth Guard

`session.RequireKey` creates middleware that redirects unauthenticated users:

```go
authGuard := session.RequireKey("username", "/login")
mux.Handle("GET /{$}", authGuard(protectedHandler))
```

If the session does not contain a `"username"` key, the user is redirected to `/login`.

### SSR Shell with LiveMount

The dashboard page is an SSR handler that embeds a LiveView using `gl.LiveMount`:

```go
func dashboardPage() []g.ComponentFunc {
    return []g.ComponentFunc{
        gd.Head(
            gd.Title("Auth Demo"),
            gu.Theme(),
        ),
        gd.Body(
            gu.Center(
                gp.Attr("style", "min-height: 100vh"),
                gu.ContainerNarrow(
                    gu.Stack(
                        gd.H1(gp.Value("Auth Demo")),
                        gl.LiveMount("/live/dashboard"),
                    ),
                ),
            ),
        ),
    }
}
```

`g.Handler(dashboardPage()...)` serves this as a standard `http.Handler`. The `gl.LiveMount("/live/dashboard")` call inserts a placeholder that the client-side JavaScript upgrades to a WebSocket-connected LiveView.

The corresponding LiveView endpoint is registered separately:

```go
mux.Handle("GET /{$}", authGuard(g.Handler(dashboardPage()...)))
mux.Handle("/live/dashboard", authGuard(gl.Handler(func(_ context.Context) gl.View {
    return &DashboardView{}
}, liveOpts...)))
```

### Push-based Session Invalidation

When `gl.WithSessionStore(store)` is passed to the LiveView handler, WebSocket connections subscribe to session invalidation events via the store's `Broker`. Destroying a session (e.g., on logout) notifies all subscribed connections.

Views implementing `SessionExpiredHandler` receive a callback:

```go
func (v *DashboardView) OnSessionExpired() error {
    v.SessionExpired = true
    v.Navigate("/login")
    return nil
}
```

This enables real-time cross-tab logout: logging out in one browser tab redirects all other tabs showing the dashboard.

## Walkthrough

### main()

1. Creates a `MemoryStore` with a signing key
2. Sets up `session.Middleware` for automatic session loading/saving
3. Sets up `session.RequireKey("username", "/login")` as an auth guard
4. Registers routes:
   - `GET /login` — renders the login form via inline `HandleFunc` (redirects to `/` if already logged in)
   - `POST /login` — validates credentials and CSRF token, uses `g.ExecuteTemplate()` for error responses (needs status code control)
   - `/logout` — destroys session and redirects to `/login`
   - `GET /{$}` — SSR dashboard shell (`g.Handler`) with `gl.LiveMount("/live/dashboard")`
   - `/live/dashboard` — LiveView WebSocket endpoint (`gl.Handler`) with `WithSessionStore`

### Login Flow

1. User visits `/login` — the `GET /login` handler renders `loginPage()` with a CSRF token via `g.ExecuteTemplate()`
2. User submits credentials — the `POST /login` handler validates the CSRF token and password
3. On failure, the handler calls `w.WriteHeader(http.StatusUnauthorized)` then `g.ExecuteTemplate()` to re-render the login form with an error message. Using `g.ExecuteTemplate()` here (instead of `g.Handler()`) allows setting the HTTP status code before writing the response body.
4. On success, `sess.Set("username", username)` stores the user in the session and redirects to `/`
5. User is redirected to `GET /{$}` which serves the SSR shell; the embedded `LiveMount` connects the `DashboardView` LiveView

### DashboardView

- `Render()` returns only `<body>` content — the welcome card, session-expired warning, and logout button. It does **not** include `<head>` or the full page layout; those are provided by the SSR shell.
- `Mount()` reads `params.Conn.Session.GetString("username")` to display the username
- `OnSessionExpired()` sets a warning flag and calls `v.Navigate("/login")` to redirect

## Running

```bash
go run example/auth/auth.go          # http://localhost:8895
go run example/auth/auth.go -debug   # with debug panel
```

Login with any username and the password `"password"`. Open multiple tabs to test cross-tab session invalidation.
