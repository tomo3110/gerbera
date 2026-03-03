package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
	"github.com/tomo3110/gerbera/session"
)

// AuthView is a LiveView that shows different content based on session state.
type AuthView struct {
	Username string
}

func (v *AuthView) Mount(params gl.Params) error {
	if sess := params.Conn.Session; sess != nil {
		v.Username = sess.GetString("username")
	}
	return nil
}

func (v *AuthView) Render() []g.ComponentFunc {
	return []g.ComponentFunc{
		gd.Head(
			gd.Title("Auth Demo"),
			gd.Style(gp.Value(`
				body { font-family: sans-serif; max-width: 600px; margin: 40px auto; padding: 0 20px; }
				.card { border: 1px solid #ddd; border-radius: 8px; padding: 24px; margin-top: 20px; }
				.btn { padding: 8px 16px; border: none; border-radius: 4px; cursor: pointer; text-decoration: none; display: inline-block; }
				.btn-danger { background: #dc3545; color: white; }
			`)),
		),
		gd.Body(
			gd.H1(gp.Value("Auth Demo")),
			gd.Div(
				gp.Class("card"),
				gd.H2(gp.Value(fmt.Sprintf("Welcome, %s!", v.Username))),
				gd.P(gp.Value("You are logged in. This page is protected by session middleware.")),
				gd.A(gp.Attr("href", "/logout"), gp.Class("btn btn-danger"), gp.Value("Logout")),
			),
		),
	}
}

func (v *AuthView) HandleEvent(_ string, _ gl.Payload) error {
	return nil
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	sess := session.FromContext(r.Context())

	if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")
		csrfToken := r.FormValue("csrf_token")

		if sess != nil && !session.ValidCSRFToken(sess, csrfToken) {
			http.Error(w, "invalid CSRF token", http.StatusForbidden)
			return
		}

		// Simple demo authentication: any username with password "password"
		if username != "" && password == "password" {
			sess.Set("username", username)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		renderLoginHTML(w, sess, "Invalid username or password")
		return
	}

	renderLoginHTML(w, sess, "")
}

func renderLoginHTML(w http.ResponseWriter, sess *session.Session, errMsg string) {
	var csrfField string
	if sess != nil {
		token := session.GenerateCSRFToken(sess)
		csrfField = fmt.Sprintf(`<input type="hidden" name="csrf_token" value="%s">`, token)
	}

	var errHTML string
	if errMsg != "" {
		errHTML = fmt.Sprintf(`<p style="color:red;">%s</p>`, errMsg)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en">
<head>
	<title>Login - Auth Demo</title>
	<style>
		body { font-family: sans-serif; max-width: 400px; margin: 80px auto; padding: 0 20px; }
		.card { border: 1px solid #ddd; border-radius: 8px; padding: 24px; }
		input[type="text"], input[type="password"] { width: 100%%; padding: 8px; margin: 4px 0 16px; border: 1px solid #ccc; border-radius: 4px; box-sizing: border-box; }
		label { font-weight: bold; }
		.btn { padding: 10px 20px; background: #0d6efd; color: white; border: none; border-radius: 4px; cursor: pointer; width: 100%%; }
		.hint { color: #666; font-size: 0.9em; margin-top: 12px; }
	</style>
</head>
<body>
	<h1>Login</h1>
	%s
	<div class="card">
		<form method="POST" action="/login">
			%s
			<label>Username</label>
			<input type="text" name="username" required>
			<label>Password</label>
			<input type="password" name="password" required>
			<button type="submit" class="btn">Sign In</button>
		</form>
		<p class="hint">Hint: any username, password is "password"</p>
	</div>
</body>
</html>`, errHTML, csrfField)
}

func logoutHandler(w http.ResponseWriter, r *http.Request, store session.Store) {
	sess := session.FromContext(r.Context())
	if sess != nil {
		store.Destroy(w, r, sess)
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func main() {
	addr := flag.String("addr", ":8895", "listen address")
	debug := flag.Bool("debug", false, "enable debug panel")
	flag.Parse()

	key := []byte("example-secret-key-change-in-prod")
	store := session.NewMemoryStore(key)
	defer store.Close()

	sessionMW := session.Middleware(store)
	authGuard := session.RequireKey("username", "/login")

	var opts []gl.Option
	if *debug {
		opts = append(opts, gl.WithDebug())
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		loginPage(w, r)
	})
	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		logoutHandler(w, r, store)
	})
	mux.Handle("/", authGuard(gl.Handler(func(_ context.Context) gl.View {
		return &AuthView{}
	}, opts...)))

	handler := sessionMW(mux)

	log.Printf("auth demo running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, handler))
}
