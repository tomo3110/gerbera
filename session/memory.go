package session

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

// MemoryStore is an in-memory session store.
// It implements both Store and BrokerStore interfaces.
type MemoryStore struct {
	mu         sync.RWMutex
	sessions   map[string]*Session
	key        []byte
	maxAge     time.Duration
	cookie     CookieConfig
	gcInterval time.Duration
	stopGC     chan struct{}
	broker     *Broker
}

// MemoryOption configures a MemoryStore.
type MemoryOption func(*MemoryStore)

// WithMaxAge sets the session max age.
func WithMaxAge(d time.Duration) MemoryOption {
	return func(s *MemoryStore) {
		s.maxAge = d
		s.cookie.MaxAge = int(d.Seconds())
	}
}

// WithCookie sets the cookie configuration.
func WithCookie(cfg CookieConfig) MemoryOption {
	return func(s *MemoryStore) {
		s.cookie = cfg
	}
}

// WithGCInterval sets the garbage collection interval.
func WithGCInterval(d time.Duration) MemoryOption {
	return func(s *MemoryStore) {
		s.gcInterval = d
	}
}

// NewMemoryStore creates a new in-memory session store.
// The key is used for HMAC-SHA256 cookie signing.
func NewMemoryStore(key []byte, opts ...MemoryOption) *MemoryStore {
	s := &MemoryStore{
		sessions:   make(map[string]*Session),
		key:        key,
		maxAge:     24 * time.Hour,
		cookie:     defaultCookieConfig(),
		gcInterval: 10 * time.Minute,
		stopGC:     make(chan struct{}),
		broker:     NewBroker(),
	}
	for _, opt := range opts {
		opt(s)
	}
	go s.gc()
	return s
}

// Get retrieves the session associated with the request.
func (s *MemoryStore) Get(r *http.Request) (*Session, error) {
	cookie, err := r.Cookie(s.cookie.Name)
	if err != nil {
		return s.newSession(), nil
	}
	id, err := verify(cookie.Value, s.key)
	if err != nil {
		return s.newSession(), nil
	}
	s.mu.RLock()
	sess, ok := s.sessions[id]
	s.mu.RUnlock()
	if !ok {
		return s.newSession(), nil
	}
	sess.mu.Lock()
	if time.Now().After(sess.expiresAt) {
		sess.mu.Unlock()
		s.mu.Lock()
		delete(s.sessions, id)
		s.mu.Unlock()
		return s.newSession(), nil
	}
	sess.mu.Unlock()
	return sess, nil
}

// Save persists the session and writes the session cookie.
func (s *MemoryStore) Save(w http.ResponseWriter, r *http.Request, sess *Session) error {
	sess.mu.Lock()
	sess.expiresAt = time.Now().Add(s.maxAge)
	sess.IsNew = false
	sess.modified = false
	sess.mu.Unlock()

	s.mu.Lock()
	s.sessions[sess.ID] = sess
	s.mu.Unlock()

	writeCookie(w, s.cookie, sign(sess.ID, s.key))
	return nil
}

// Destroy removes the session from the store and clears the cookie.
// It also notifies all Broker subscribers for this session.
func (s *MemoryStore) Destroy(w http.ResponseWriter, r *http.Request, sess *Session) error {
	s.mu.Lock()
	delete(s.sessions, sess.ID)
	s.mu.Unlock()
	s.broker.Invalidate(sess.ID)
	clearCookie(w, s.cookie)
	return nil
}

// Broker returns the session invalidation broker.
// This satisfies the BrokerStore interface.
func (s *MemoryStore) Broker() *Broker {
	return s.broker
}

// Close stops the background GC goroutine.
func (s *MemoryStore) Close() {
	close(s.stopGC)
}

func (s *MemoryStore) gc() {
	ticker := time.NewTicker(s.gcInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopGC:
			return
		case <-ticker.C:
			now := time.Now()
			var expired []string
			s.mu.Lock()
			for id, sess := range s.sessions {
				sess.mu.RLock()
				isExpired := now.After(sess.expiresAt)
				sess.mu.RUnlock()
				if isExpired {
					delete(s.sessions, id)
					expired = append(expired, id)
				}
			}
			s.mu.Unlock()
			for _, id := range expired {
				s.broker.Invalidate(id)
			}
		}
	}
}

func (s *MemoryStore) newSession() *Session {
	return &Session{
		ID:     generateID(),
		Values: make(map[string]any),
		IsNew:  true,
	}
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
