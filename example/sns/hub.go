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
type Hub struct {
	mu      sync.RWMutex
	clients map[string]*live.Session // key: live session ID
	userMap map[int64]string         // key: user DB ID → live session ID
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*live.Session),
		userMap: make(map[int64]string),
	}
}

// Join registers a LiveView session with the Hub.
func (h *Hub) Join(sess *live.Session, userID int64) {
	h.mu.Lock()
	h.clients[sess.ID] = sess
	h.userMap[userID] = sess.ID
	h.mu.Unlock()
}

// Leave removes a LiveView session from the Hub.
func (h *Hub) Leave(sessionID string, userID int64) {
	h.mu.Lock()
	delete(h.clients, sessionID)
	delete(h.userMap, userID)
	h.mu.Unlock()
}

// Notify sends a message to a specific user's LiveView session.
func (h *Hub) Notify(userID int64, msg any) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if sessID, ok := h.userMap[userID]; ok {
		if sess, ok := h.clients[sessID]; ok {
			sess.SendInfo(msg)
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
