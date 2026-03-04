package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	g "github.com/tomo3110/gerbera"
	gl "github.com/tomo3110/gerbera/live"
	"github.com/tomo3110/gerbera/session"
)

var hub = NewHub()

func main() {
	addr := flag.String("addr", ":8930", "listen address")
	dsn := flag.String("dsn", "sns:snspass@tcp(127.0.0.1:3306)/sns?parseTime=true", "MySQL DSN")
	debug := flag.Bool("debug", false, "enable debug panel")
	flag.Parse()

	// Environment variable overrides flag (useful for Docker Compose)
	if envDSN := os.Getenv("DATABASE_DSN"); envDSN != "" {
		*dsn = envDSN
	}
	if envAddr := os.Getenv("LISTEN_ADDR"); envAddr != "" {
		*addr = envAddr
	}

	db, err := openDB(*dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := autoMigrate(db); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	fmt.Println("database migration complete")

	key := []byte("sns-example-secret-key-change-me!")
	store := session.NewMemoryStore(key)
	defer store.Close()

	sessionMW := session.Middleware(store)
	authGuard := session.RequireKey("user_id", "/login")

	// Ensure upload directory exists
	if err := os.MkdirAll("uploads/avatars", 0755); err != nil {
		log.Fatalf("failed to create upload directory: %v", err)
	}

	mux := http.NewServeMux()

	// Static file serving for avatars
	mux.Handle("GET /avatars/", http.StripPrefix("/avatars/", http.FileServer(http.Dir("uploads/avatars"))))

	// Static auth pages
	mux.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		if sess := session.FromContext(r.Context()); sess != nil {
			if _, ok := sess.Get("user_id").(int64); ok {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		components := loginPage(r)
		g.ExecuteTemplate(w, "en", components...)
	})

	mux.HandleFunc("POST /login", loginPostHandler(db))

	mux.HandleFunc("GET /register", func(w http.ResponseWriter, r *http.Request) {
		if sess := session.FromContext(r.Context()); sess != nil {
			if _, ok := sess.Get("user_id").(int64); ok {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		components := registerPage(r)
		g.ExecuteTemplate(w, "en", components...)
	})

	mux.HandleFunc("POST /register", registerPostHandler(db))
	mux.HandleFunc("/logout", logoutHandler(store))

	// LiveView route (protected)
	liveOpts := []gl.Option{
		gl.WithSessionStore(store),
	}
	if *debug {
		liveOpts = append(liveOpts, gl.WithDebug())
	}

	mux.Handle("/", authGuard(gl.Handler(func(_ context.Context) gl.View {
		return &SNSView{
			db:  db,
			hub: hub,
		}
	}, liveOpts...)))

	handler := sessionMW(mux)

	log.Printf("SNS running on http://localhost%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, handler))
}
