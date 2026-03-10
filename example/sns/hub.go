package main

import (
	"sync"

	"github.com/tomo3110/gerbera/live"
)

// Notification message types sent via Hub.
type NewLikeNotif struct {
	PostID    int64
	ActorName string
}
type NewRetweetNotif struct {
	PostID    int64
	ActorName string
}
type NewFollowNotif struct {
	ActorName string
}
type NewMessageNotif struct {
	SenderID int64
	Content  string
}
type NewCommentNotif struct {
	PostID    int64
	ActorName string
}
type NewPostNotif struct {
	AuthorName string
}
type NewCommentOnViewedPostNotif struct {
	PostID    int64
	ActorName string
}

// Hub is a pub/sub that routes real-time notifications to connected LiveView sessions.
// A single user may have multiple sessions (one per View / WebSocket).
type Hub struct {
	mu       sync.RWMutex
	clients  map[string]*live.Session         // sessionID → Session
	userSess map[int64]map[string]struct{} // userID → set of sessionIDs
}

func NewHub() *Hub {
	return &Hub{
		clients:  make(map[string]*live.Session),
		userSess: make(map[int64]map[string]struct{}),
	}
}

// Join registers a LiveView session with the Hub.
func (h *Hub) Join(sess *live.Session, userID int64) {
	h.mu.Lock()
	h.clients[sess.ID] = sess
	if h.userSess[userID] == nil {
		h.userSess[userID] = make(map[string]struct{})
	}
	h.userSess[userID][sess.ID] = struct{}{}
	h.mu.Unlock()
}

// Leave removes a LiveView session from the Hub.
func (h *Hub) Leave(sessionID string, userID int64) {
	h.mu.Lock()
	delete(h.clients, sessionID)
	if sids, ok := h.userSess[userID]; ok {
		delete(sids, sessionID)
		if len(sids) == 0 {
			delete(h.userSess, userID)
		}
	}
	h.mu.Unlock()
}

// Notify sends a message to all of a specific user's LiveView sessions.
func (h *Hub) Notify(userID int64, msg any) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if sids, ok := h.userSess[userID]; ok {
		for sid := range sids {
			if sess, ok := h.clients[sid]; ok {
				sess.SendInfo(msg)
			}
		}
	}
}

// Broadcast sends a message to all connected sessions except the sender.
func (h *Hub) Broadcast(msg any, senderSessionID string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for id, sess := range h.clients {
		if id == senderSessionID {
			continue
		}
		sess.SendInfo(msg)
	}
}
