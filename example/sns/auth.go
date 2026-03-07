package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"regexp"
	"strings"

	g "github.com/tomo3110/gerbera"
	gd "github.com/tomo3110/gerbera/dom"
	"github.com/tomo3110/gerbera/expr"
	gp "github.com/tomo3110/gerbera/property"
	"github.com/tomo3110/gerbera/session"
	gs "github.com/tomo3110/gerbera/styles"
	gu "github.com/tomo3110/gerbera/ui"
)

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_]{1,30}$`)

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

func renderAuthPage(title string, csrfToken, errMsg string, isRegister bool) g.Components {
	actionURL := "/login"
	heading := "Sign In"
	switchText := "Don't have an account?"
	switchLink := "/register"
	switchLabel := "Register"
	if isRegister {
		actionURL = "/register"
		heading = "Create Account"
		switchText = "Already have an account?"
		switchLink = "/login"
		switchLabel = "Sign In"
	}

	var formFields g.Components
	formFields = append(formFields,
		expr.If(csrfToken != "",
			gd.Input(gp.Attr("type", "hidden"), gp.Name("csrf_token"), gp.Attr("value", csrfToken)),
		),
	)

	if isRegister {
		formFields = append(formFields,
			gu.FormGroup(
				gu.FormLabel("Username", "username"),
				gu.FormInput("username", gp.ID("username"), gp.Attr("type", "text"),
					gp.Attr("required", "required"), gp.Attr("maxlength", "30"),
					gp.Attr("pattern", "[a-zA-Z0-9_]{1,30}"),
					gp.Attr("title", "Letters, numbers, and underscores only"),
					gp.Placeholder("Choose a username")),
			),
			gu.FormGroup(
				gu.FormLabel("Display Name", "display_name"),
				gu.FormInput("display_name", gp.ID("display_name"), gp.Attr("type", "text"),
					gp.Attr("required", "required"), gp.Attr("maxlength", "50"),
					gp.Placeholder("Your display name")),
			),
			gu.FormGroup(
				gu.FormLabel("Email", "email"),
				gu.FormInput("email", gp.ID("email"), gp.Attr("type", "email"),
					gp.Attr("required", "required"),
					gp.Placeholder("you@example.com")),
			),
		)
	} else {
		formFields = append(formFields,
			gu.FormGroup(
				gu.FormLabel("Username", "username"),
				gu.FormInput("username", gp.ID("username"), gp.Attr("type", "text"),
					gp.Attr("required", "required"),
					gp.Placeholder("Username")),
			),
		)
	}

	formFields = append(formFields,
		gu.FormGroup(
			gu.FormLabel("Password", "password"),
			gu.FormInput("password", gp.ID("password"), gp.Attr("type", "password"),
				gp.Attr("required", "required"),
				gp.Placeholder("Password")),
		),
		gu.Button(heading, gu.ButtonPrimary,
			gp.Attr("type", "submit"),
			gp.Attr("style", "width: 100%"),
		),
	)

	return g.Components{
		gd.Head(
			gd.Title(title),
			gd.Meta(gp.Attr("name", "viewport"), gp.Attr("content", "width=device-width, initial-scale=1")),
			gu.Theme(),
			gs.CSS(snsCSS),
		),
		gd.Body(
			gu.Center(
				gp.Attr("style", "min-height: 100vh"),
				gu.ContainerNarrow(
					gu.Stack(
						gd.H1(gp.Attr("style", "text-align:center"), gp.Value("SNS")),
						expr.If(errMsg != "",
							gu.Alert(errMsg, "danger"),
						),
						gu.Card(
							gu.CardHeader(heading),
							gu.CardBody(
								gd.Form(
									gp.Attr("method", "POST"),
									gp.Attr("action", actionURL),
									gu.Stack(formFields...),
								),
							),
							gu.CardFooter(
								gd.P(
									gp.Attr("style", "color: var(--g-text-secondary); font-size: 0.9em; margin: 0; text-align: center"),
									gp.Value(switchText+" "),
									gd.A(gp.Attr("href", switchLink), gp.Value(switchLabel)),
								),
							),
						),
					),
				),
			),
		),
	}
}

func loginPage(r *http.Request) g.Components {
	sess := session.FromContext(r.Context())
	var csrfToken string
	if sess != nil {
		csrfToken = session.CSRFToken(sess)
		if csrfToken == "" {
			csrfToken = session.GenerateCSRFToken(sess)
		}
	}
	return renderAuthPage("Login — SNS", csrfToken, "", false)
}

func registerPage(r *http.Request) g.Components {
	sess := session.FromContext(r.Context())
	var csrfToken string
	if sess != nil {
		csrfToken = session.CSRFToken(sess)
		if csrfToken == "" {
			csrfToken = session.GenerateCSRFToken(sess)
		}
	}
	return renderAuthPage("Register — SNS", csrfToken, "", true)
}

func loginPostHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := session.FromContext(r.Context())
		r.ParseForm()
		username := strings.TrimSpace(r.FormValue("username"))
		password := r.FormValue("password")
		csrfToken := r.FormValue("csrf_token")

		if sess != nil && !session.ValidCSRFToken(sess, csrfToken) {
			http.Error(w, "invalid CSRF token", http.StatusForbidden)
			return
		}

		user, err := dbGetUserByUsername(db, username)
		if err != nil || !verifyPassword(user.PasswordHash, password) {
			var newCSRF string
			if sess != nil {
				newCSRF = session.GenerateCSRFToken(sess)
			}
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			g.ExecuteTemplate(w, "en", renderAuthPage("Login — SNS", newCSRF, "Invalid username or password", false)...)
			return
		}

		sess.Set("user_id", user.ID)
		sess.Set("username", user.Username)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func registerPostHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := session.FromContext(r.Context())
		r.ParseForm()
		username := strings.TrimSpace(r.FormValue("username"))
		displayName := strings.TrimSpace(r.FormValue("display_name"))
		email := strings.TrimSpace(r.FormValue("email"))
		password := r.FormValue("password")
		csrfToken := r.FormValue("csrf_token")

		if sess != nil && !session.ValidCSRFToken(sess, csrfToken) {
			http.Error(w, "invalid CSRF token", http.StatusForbidden)
			return
		}

		if username == "" || displayName == "" || email == "" || password == "" {
			var newCSRF string
			if sess != nil {
				newCSRF = session.GenerateCSRFToken(sess)
			}
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			g.ExecuteTemplate(w, "en", renderAuthPage("Register — SNS", newCSRF, "All fields are required", true)...)
			return
		}

		if !usernameRe.MatchString(username) {
			var newCSRF string
			if sess != nil {
				newCSRF = session.GenerateCSRFToken(sess)
			}
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			g.ExecuteTemplate(w, "en", renderAuthPage("Register — SNS", newCSRF, "Username must contain only letters, numbers, and underscores (a-z, 0-9, _)", true)...)
			return
		}

		hash := hashPassword(password)
		userID, err := dbCreateUser(db, username, displayName, email, hash)
		if err != nil {
			var newCSRF string
			if sess != nil {
				newCSRF = session.GenerateCSRFToken(sess)
			}
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			g.ExecuteTemplate(w, "en", renderAuthPage("Register — SNS", newCSRF, "Username or email already taken", true)...)
			return
		}

		sess.Set("user_id", userID)
		sess.Set("username", username)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func logoutHandler(store session.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := session.FromContext(r.Context())
		if sess != nil {
			store.Destroy(w, r, sess)
		}
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}
