package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	g "github.com/tomo3110/gerbera"
	_ "github.com/tomo3110/gerbera/assets"
	gl "github.com/tomo3110/gerbera/live"
	"github.com/tomo3110/gerbera/session"
)

var hub = NewHub()

func main() {
	addr := flag.String("addr", ":8930", "listen address")
	dsn := flag.String("dsn", "sns:snspass@tcp(127.0.0.1:3306)/sns?parseTime=true", "MySQL DSN")
	debug := flag.Bool("debug", false, "enable debug panel")
	flag.Parse()

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

	if err := os.MkdirAll("uploads/avatars", 0755); err != nil {
		log.Fatalf("failed to create upload directory: %v", err)
	}

	mux := http.NewServeMux()

	// Static file serving for avatars
	mux.Handle("GET /avatars/", http.StripPrefix("/avatars/", http.FileServer(http.Dir("uploads/avatars"))))

	// Auth pages
	mux.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		if sess := session.FromContext(r.Context()); sess != nil {
			if _, ok := sess.Get("user_id").(int64); ok {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		}
		g.Handler(loginPage(r)...).ServeHTTP(w, r)
	})
	mux.HandleFunc("POST /login", loginPostHandler(db))
	mux.HandleFunc("GET /register", func(w http.ResponseWriter, r *http.Request) {
		if sess := session.FromContext(r.Context()); sess != nil {
			if _, ok := sess.Get("user_id").(int64); ok {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		}
		g.Handler(registerPage(r)...).ServeHTTP(w, r)
	})
	mux.HandleFunc("POST /register", registerPostHandler(db))
	mux.HandleFunc("/logout", logoutHandler(store))

	// LiveView options
	liveOpts := []gl.Option{
		gl.WithSessionStore(store),
	}
	if *debug {
		liveOpts = append(liveOpts, gl.WithDebug())
	}

	// --- WebSocket Multiplexing ---
	registry := gl.NewViewRegistry()
	registry.Register("/live/timeline", func(_ context.Context) gl.View { return NewTimelineView(db, hub) })
	registry.Register("/live/profile", func(_ context.Context) gl.View { return NewProfileView(db, hub) })
	registry.Register("/live/post", func(_ context.Context) gl.View { return NewPostDetailView(db, hub) })
	registry.Register("/live/messages", func(_ context.Context) gl.View { return NewMessagesView(db, hub) })
	registry.Register("/live/search", func(_ context.Context) gl.View { return NewSearchView(db, hub) })
	registry.Register("/live/settings", func(_ context.Context) gl.View { return NewSettingsView(db, hub) })
	muxHandler := gl.NewMultiplexHandler(registry, liveOpts...)

	// --- SSR pages (render layout shell with LiveMount) ---

	mux.Handle("GET /{$}", authGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		badges := fetchBadgeCounts(db, r)
		g.Handler(snsPage("Timeline", "home", "/live/timeline", badges)...).ServeHTTP(w, r)
	})))

	mux.Handle("GET /profile", authGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		badges := fetchBadgeCounts(db, r)
		g.Handler(snsPage("Profile", "profile", "/live/profile", badges)...).ServeHTTP(w, r)
	})))

	mux.Handle("GET /profile/{id}", authGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		badges := fetchBadgeCounts(db, r)
		id := r.PathValue("id")
		g.Handler(snsPage("Profile", "profile", "/live/profile?id="+id, badges)...).ServeHTTP(w, r)
	})))

	mux.Handle("GET /post/{id}", authGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		badges := fetchBadgeCounts(db, r)
		id := r.PathValue("id")
		g.Handler(snsPage("Post", "post", "/live/post?id="+id, badges)...).ServeHTTP(w, r)
	})))

	mux.Handle("GET /messages", authGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		badges := fetchBadgeCounts(db, r)
		g.Handler(snsPage("Messages", "messages", "/live/messages", badges)...).ServeHTTP(w, r)
	})))

	mux.Handle("GET /messages/{id}", authGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		badges := fetchBadgeCounts(db, r)
		id := r.PathValue("id")
		g.Handler(snsPage("Messages", "messages", "/live/messages?id="+id, badges)...).ServeHTTP(w, r)
	})))

	mux.Handle("GET /search", authGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		badges := fetchBadgeCounts(db, r)
		endpoint := "/live/search"
		if kw := r.URL.Query().Get("keyword"); kw != "" {
			endpoint += "?keyword=" + kw
		}
		g.Handler(snsPage("Search", "search", endpoint, badges)...).ServeHTTP(w, r)
	})))

	mux.Handle("GET /settings", authGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		badges := fetchBadgeCounts(db, r)
		g.Handler(snsPage("Settings", "settings", "/live/settings", badges)...).ServeHTTP(w, r)
	})))

	// --- LiveView endpoints ---

	mux.Handle("/live/timeline", authGuard(gl.Handler(func(_ context.Context) gl.View {
		return NewTimelineView(db, hub)
	}, liveOpts...)))

	mux.Handle("/live/profile", authGuard(gl.Handler(func(_ context.Context) gl.View {
		return NewProfileView(db, hub)
	}, liveOpts...)))

	mux.Handle("/live/post", authGuard(gl.Handler(func(_ context.Context) gl.View {
		return NewPostDetailView(db, hub)
	}, liveOpts...)))

	mux.Handle("/live/messages", authGuard(gl.Handler(func(_ context.Context) gl.View {
		return NewMessagesView(db, hub)
	}, liveOpts...)))

	mux.Handle("/live/search", authGuard(gl.Handler(func(_ context.Context) gl.View {
		return NewSearchView(db, hub)
	}, liveOpts...)))

	mux.Handle("/live/settings", authGuard(gl.Handler(func(_ context.Context) gl.View {
		return NewSettingsView(db, hub)
	}, liveOpts...)))

	// sessionMW must wrap g.Serve so that /_gerbera/ws (multiplex endpoint)
	// also has the session loaded into context for authGuard to work.
	handler := sessionMW(g.Serve(mux, g.WithMultiplex(authGuard(muxHandler))))

	log.Printf("SNS running on http://localhost%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, handler))
}

// fetchBadgeCounts retrieves notification and message badge counts for the current user.
func fetchBadgeCounts(db *sql.DB, r *http.Request) badgeCounts {
	sess := session.FromContext(r.Context())
	if sess == nil {
		return badgeCounts{}
	}
	uid, ok := sess.Get("user_id").(int64)
	if !ok {
		return badgeCounts{}
	}
	notif, _ := dbUnreadNotificationCount(db, uid)
	msgs, _ := dbUnreadMessageCount(db, uid)
	return badgeCounts{Notifications: notif, Messages: msgs}
}

// parseID parses a string to int64, returning 0 on failure.
func parseID(s string) int64 {
	id, _ := strconv.ParseInt(s, 10, 64)
	return id
}
