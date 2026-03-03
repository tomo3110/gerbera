package main

import (
	"sync"

	"github.com/tomo3110/gerbera/live"
)

// ChatMessage represents a single chat message exchanged via the Hub.
type ChatMessage struct {
	Author    string
	Content   string
	Timestamp string
	System    bool // system messages (join/leave notifications)
}

// Hub is a simple pub/sub that broadcasts ChatMessages to connected LiveView sessions.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]*live.Session
}

// NewHub creates an empty Hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*live.Session),
	}
}

// Join registers a LiveView session with the Hub.
func (h *Hub) Join(sess *live.Session) {
	h.mu.Lock()
	h.clients[sess.ID] = sess
	h.mu.Unlock()
}

// Leave removes a LiveView session from the Hub.
func (h *Hub) Leave(sessionID string) {
	h.mu.Lock()
	delete(h.clients, sessionID)
	h.mu.Unlock()
}

// OnlineCount returns the number of connected clients.
func (h *Hub) OnlineCount() int {
	h.mu.RLock()
	n := len(h.clients)
	h.mu.RUnlock()
	return n
}

// Broadcast sends a ChatMessage to all connected sessions except the sender.
func (h *Hub) Broadcast(msg ChatMessage, senderID string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for id, sess := range h.clients {
		if id == senderID {
			continue
		}
		sess.SendInfo(msg)
	}
}
