# Auth Example Tutorial

This example demonstrates session-based authentication using the `session/` package with CSRF protection, session middleware, and push-based session invalidation.

## Overview

The auth example implements:

- **Login form** with CSRF token protection (server-side rendered)
- **Protected dashboard** as a LiveView, guarded by `session.RequireKey` middleware
- **Session invalidation** — logging out in one tab automatically redirects all other tabs via `SessionExpiredHandler`

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
mux.Handle("/", authGuard(protectedHandler))
```

If the session does not contain a `"username"` key, the user is redirected to `/login`.

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
   - `GET /login` — renders the login form (redirects to `/` if already logged in)
   - `POST /login` — validates credentials and CSRF token
   - `/logout` — destroys session and redirects to `/login`
   - `/` — protected LiveView dashboard with `WithSessionStore`

### Login Flow

1. User visits `/login` → `loginFormPage()` renders a form with a CSRF token
2. User submits credentials → `loginPostHandler()` validates CSRF token and password
3. On success, `sess.Set("username", username)` stores the user in the session
4. User is redirected to `/` which renders `DashboardView`

### DashboardView

- `Mount()` reads `params.Conn.Session.GetString("username")` to display the username
- `OnSessionExpired()` sets a warning flag and calls `v.Navigate("/login")` to redirect

## Running

```bash
go run example/auth/auth.go          # http://localhost:8895
go run example/auth/auth.go -debug   # with debug panel
```

Login with any username and the password `"password"`. Open multiple tabs to test cross-tab session invalidation.
