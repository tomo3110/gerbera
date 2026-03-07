package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	g "github.com/tomo3110/gerbera"
	_ "github.com/tomo3110/gerbera/assets"
	gd "github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	gl "github.com/tomo3110/gerbera/live"
	gp "github.com/tomo3110/gerbera/property"
	"github.com/tomo3110/gerbera/session"
	gu "github.com/tomo3110/gerbera/ui"
)

// ---------------------------------------------------------------------------
// DashboardView — LiveView embedded via LiveMount in the SSR shell
// ---------------------------------------------------------------------------

type DashboardView struct {
	gl.CommandQueue
	Username       string
	SessionExpired bool
}

func (v *DashboardView) Mount(params gl.Params) error {
	if params.Conn.Session != nil {
		v.Username = params.Conn.Session.GetString("username")
	}
	return nil
}

func (v *DashboardView) Render() g.Components {
	return g.Components{
		gd.Body(
			gu.Stack(
				expr.If(v.SessionExpired,
					gu.Card(
						gu.CardBody(
							gp.Attr("style", "border-color: var(--g-warning-border); background: var(--g-warning-bg)"),
							gd.P(
								gp.Attr("style", "color: var(--g-warning); margin: 0"),
								gp.Value("Session expired. Redirecting to login..."),
							),
						),
					),
				),
				gu.Card(
					gu.CardHeader(fmt.Sprintf("Welcome, %s!", v.Username)),
					gu.CardBody(
						gd.P(gp.Value("You are logged in. This page is a LiveView protected by session middleware.")),
						gd.P(
							gp.Attr("style", "color: var(--g-text-secondary); font-size: 0.9em"),
							gp.Value("Try logging out from another tab to see push-based session invalidation."),
						),
					),
					gu.CardFooter(
						gd.A(gp.Attr("href", "/logout"),
							gu.Button("Logout", gu.ButtonDanger),
						),
					),
				),
			),
		),
	}
}

func (v *DashboardView) HandleEvent(_ string, _ gl.Payload) error {
	return nil
}

// OnSessionExpired is called by the Broker when the session is invalidated.
func (v *DashboardView) OnSessionExpired() error {
	v.SessionExpired = true
	v.Navigate("/login")
	return nil
}

// ---------------------------------------------------------------------------
// SSR pages
// ---------------------------------------------------------------------------

func dashboardPage() g.Components {
	return g.Components{
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

func loginPage(csrfToken string, errMsg string) g.Components {
	return g.Components{
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
								gu.CardBody(
									gp.Attr("style", "border-color: var(--g-danger-border); background: var(--g-danger-bg)"),
									gd.P(
										gp.Attr("style", "color: var(--g-danger); margin: 0"),
										gp.Value(errMsg),
									),
								),
							),
						),
						gu.Card(
							gu.CardHeader("Sign In"),
							gu.CardBody(
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

func getCSRFToken(r *http.Request) string {
	sess := session.FromContext(r.Context())
	if sess == nil {
		return ""
	}
	token := session.CSRFToken(sess)
	if token == "" {
		token = session.GenerateCSRFToken(sess)
	}
	return token
}

// ---------------------------------------------------------------------------
// main
// ---------------------------------------------------------------------------

func main() {
	addr := flag.String("addr", ":8895", "listen address")
	debug := flag.Bool("debug", false, "enable debug panel")
	flag.Parse()

	key := []byte("example-secret-key-change-in-prod")
	store := session.NewMemoryStore(key)
	defer store.Close()

	sessionMW := session.Middleware(store)
	authGuard := session.RequireKey("username", "/login")

	mux := http.NewServeMux()

	// GET /login — SSR login form (redirect if already logged in)
	mux.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		if sess := session.FromContext(r.Context()); sess != nil && sess.GetString("username") != "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		g.Handler(loginPage(getCSRFToken(r), "")...).ServeHTTP(w, r)
	})

	// POST /login — handle login form submission
	mux.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
		sess := session.FromContext(r.Context())

		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")
		csrfToken := r.FormValue("csrf_token")

		if sess != nil && !session.ValidCSRFToken(sess, csrfToken) {
			http.Error(w, "invalid CSRF token", http.StatusForbidden)
			return
		}

		if username != "" && password == "password" {
			sess.Set("username", username)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		newCSRF := getCSRFToken(r)
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		g.ExecuteTemplate(w, "en", loginPage(newCSRF, "Invalid username or password")...)
	})

	// GET /logout — destroy session and redirect
	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		sess := session.FromContext(r.Context())
		if sess != nil {
			store.Destroy(w, r, sess)
		}
		http.Redirect(w, r, "/login", http.StatusFound)
	})

	// GET / — SSR shell with LiveMount for dashboard
	liveOpts := []gl.Option{
		gl.WithSessionStore(store),
	}
	if *debug {
		liveOpts = append(liveOpts, gl.WithDebug())
	}
	mux.Handle("GET /{$}", authGuard(g.Handler(dashboardPage()...)))
	mux.Handle("/live/dashboard", authGuard(gl.Handler(func(_ context.Context) gl.View {
		return &DashboardView{}
	}, liveOpts...)))

	handler := sessionMW(mux)

	log.Printf("auth demo running on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, g.Serve(handler)))
}
