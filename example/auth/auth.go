package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
	"github.com/tomo3110/gerbera/session"
	gu "github.com/tomo3110/gerbera/ui"
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
			gu.Theme(),
		),
		gd.Body(
			gu.ContainerNarrow(
				gu.Stack(
					gd.H1(gp.Value("Auth Demo")),
					gu.Card(
						gu.CardHeader(fmt.Sprintf("Welcome, %s!", v.Username)),
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
		renderLoginPage(w, sess, "Invalid username or password")
		return
	}

	renderLoginPage(w, sess, "")
}

func renderLoginPage(w http.ResponseWriter, sess *session.Session, errMsg string) {
	var csrfToken string
	if sess != nil {
		csrfToken = session.CSRFToken(sess)
		if csrfToken == "" {
			csrfToken = session.GenerateCSRFToken(sess)
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	g.ExecuteTemplate(w, "en",
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
	)
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
