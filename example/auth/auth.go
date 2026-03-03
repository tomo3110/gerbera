package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	gp "github.com/tomo3110/gerbera/property"
	"github.com/tomo3110/gerbera/session"
	gu "github.com/tomo3110/gerbera/ui"
)

func dashboardPage(r *http.Request) []g.ComponentFunc {
	sess := session.FromContext(r.Context())
	username := ""
	if sess != nil {
		username = sess.GetString("username")
	}
	return []g.ComponentFunc{
		gd.Head(
			gd.Title("Auth Demo"),
			gu.Theme(),
		),
		gd.Body(
			gu.ContainerNarrow(
				gu.Stack(
					gd.H1(gp.Value("Auth Demo")),
					gu.Card(
						gu.CardHeader(fmt.Sprintf("Welcome, %s!", username)),
						gd.P(gp.Value("You are logged in. This page is protected by session middleware.")),
						gu.CardFooter(
							gd.A(gp.Attr("href", "/logout"),
								gu.Button("Logout", gu.ButtonDanger),
							),
						),
					),
				),
			),
		),
	}
}

func loginFormPage(r *http.Request) []g.ComponentFunc {
	sess := session.FromContext(r.Context())
	var csrfToken string
	if sess != nil {
		csrfToken = session.CSRFToken(sess)
		if csrfToken == "" {
			csrfToken = session.GenerateCSRFToken(sess)
		}
	}
	return renderLoginComponents(csrfToken, "")
}

func renderLoginComponents(csrfToken string, errMsg string) []g.ComponentFunc {
	return []g.ComponentFunc{
		gd.Head(
			gd.Title("Login - Auth Demo"),
			gu.Theme(),
		),
		gd.Body(
			gu.Center(
				gp.Attr("style", "min-height: 100vh"),
				gu.ContainerNarrow(
					gu.Stack(
						gd.H1(gp.Value("Login")),
						expr.If(errMsg != "",
							gu.Card(
								gp.Attr("style", "border-color: var(--g-danger-border); background: var(--g-danger-bg)"),
								gd.P(
									gp.Attr("style", "color: var(--g-danger); margin: 0"),
									gp.Value(errMsg),
								),
							),
						),
						gu.Card(
							gd.Form(
								gp.Attr("method", "POST"),
								gp.Attr("action", "/login"),
								expr.If(csrfToken != "",
									gd.Input(
										gp.Attr("type", "hidden"),
										gp.Name("csrf_token"),
										gp.Attr("value", csrfToken),
									),
								),
								gu.Stack(
									gu.FormGroup(
										gu.FormLabel("Username", "username"),
										gu.FormInput("username",
											gp.ID("username"),
											gp.Attr("type", "text"),
											gp.Attr("required", "required"),
										),
									),
									gu.FormGroup(
										gu.FormLabel("Password", "password"),
										gu.FormInput("password",
											gp.ID("password"),
											gp.Attr("type", "password"),
											gp.Attr("required", "required"),
										),
									),
									gu.Button("Sign In", gu.ButtonPrimary,
										gp.Attr("type", "submit"),
										gp.Attr("style", "width: 100%"),
									),
								),
							),
							gu.CardFooter(
								gd.P(
									gp.Attr("style", "color: var(--g-text-secondary); font-size: 0.9em; margin: 0"),
									gp.Value(`Hint: any username, password is "password"`),
								),
							),
						),
					),
				),
			),
		),
	}
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	sess := session.FromContext(r.Context())

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

	var newCSRFToken string
	if sess != nil {
		newCSRFToken = session.CSRFToken(sess)
		if newCSRFToken == "" {
			newCSRFToken = session.GenerateCSRFToken(sess)
		}
	}
	w.WriteHeader(http.StatusUnauthorized)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	g.ExecuteTemplate(w, "en", renderLoginComponents(newCSRFToken, "Invalid username or password")...)
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
	flag.Parse()

	key := []byte("example-secret-key-change-in-prod")
	store := session.NewMemoryStore(key)
	defer store.Close()

	sessionMW := session.Middleware(store)
	authGuard := session.RequireKey("username", "/login")

	mux := http.NewServeMux()

	// GET /login — render login form using g.HandlerFunc
	mux.Handle("GET /login", g.HandlerFunc(loginFormPage))

	// POST /login — handle login form submission
	mux.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
		loginPostHandler(w, r)
	})

	// GET /logout — destroy session and redirect
	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		logoutHandler(w, r, store)
	})

	// GET / — protected dashboard page using g.HandlerFunc
	mux.Handle("/", authGuard(g.HandlerFunc(dashboardPage)))

	handler := sessionMW(mux)

	log.Printf("auth demo running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, handler))
}
