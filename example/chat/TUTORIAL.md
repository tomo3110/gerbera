# Chat Example Tutorial

This example demonstrates a multi-user real-time chat room using `InfoReceiver`, `Unmounter`, and `Session.SendInfo()` to broadcast messages between connected LiveView sessions.

## Overview

The chat example implements:

- **Join form** — users enter a username to join the chat room
- **Real-time messaging** — messages are broadcast to all connected users instantly
- **Online count** — displays the number of connected users
- **System notifications** — join/leave events are shown as system messages
- **Auto-scroll** — chat automatically scrolls to the latest message

## Key Concepts

### Hub (Pub/Sub)

The `Hub` struct (`hub.go`) is a simple pub/sub that tracks connected LiveView sessions and broadcasts messages:

```go
type Hub struct {
    mu      sync.RWMutex
    clients map[string]*live.Session
}
```

- `Join(sess)` — registers a LiveView session
- `Leave(sessionID)` — removes a session
- `Broadcast(msg, senderID)` — sends a message to all sessions except the sender via `sess.SendInfo(msg)`
- `OnlineCount()` — returns the number of connected clients

A single shared `Hub` instance is used across all connections:

```go
var hub = NewHub()
```

### InfoReceiver

`InfoReceiver` is a LiveView interface that allows a View to receive server-side messages. When another user sends a chat message, it is delivered via `HandleInfo()`:

```go
type InfoReceiver interface {
    View
    HandleInfo(msg any) error
}
```

```go
func (v *ChatView) HandleInfo(msg any) error {
    if cm, ok := msg.(ChatMessage); ok {
        v.Messages = append(v.Messages, cm)
        v.ScrollIntoPct("#chat-messages", "1.0")
    }
    return nil
}
```

### Session.SendInfo()

`SendInfo()` pushes a message to a LiveView session's info channel. The message is delivered to the View's `HandleInfo()` method, triggering a re-render:

```go
func (h *Hub) Broadcast(msg ChatMessage, senderID string) {
    for id, sess := range h.clients {
        if id == senderID {
            continue
        }
        sess.SendInfo(msg)
    }
}
```

### Unmounter

`Unmounter` is called when the WebSocket connection closes (e.g., user closes the tab). The chat uses it to leave the Hub and broadcast a departure notification:

```go
type Unmounter interface {
    Unmount()
}
```

```go
func (v *ChatView) Unmount() {
    if v.Username != "" && v.session != nil {
        v.hub.Leave(v.session.ID)
        v.hub.Broadcast(ChatMessage{
            Author:  v.Username,
            Content: fmt.Sprintf("%s left the room", v.Username),
            System:  true,
        }, v.session.ID)
    }
}
```

### LiveSession Access

The `Mount()` method stores the `LiveSession` reference from `params.Conn.LiveSession`. This is the `*live.Session` that provides `SendInfo()`:

```go
func (v *ChatView) Mount(params gl.Params) error {
    v.hub = hub
    v.session = params.Conn.LiveSession
    return nil
}
```

## Walkthrough

### chat.go — ChatView

1. **Mount** — saves references to the shared Hub and the LiveSession
2. **HandleEvent** — processes three events:
   - `"join"` — sets the username, joins the Hub, broadcasts a join notification
   - `"input"` — updates the draft message text
   - `"send"` / `"keydown"` (Enter) — appends the message locally, clears the draft, broadcasts to others
3. **HandleInfo** — receives `ChatMessage` from other users, appends to the message list
4. **Unmount** — leaves the Hub and broadcasts a leave notification
5. **Render** — shows either the join form or the chat interface depending on whether `Username` is set

### hub.go — Hub

A thread-safe map of session ID to `*live.Session`. `Broadcast()` iterates over all clients except the sender and calls `sess.SendInfo(msg)`.

### Message Flow

```
User A types message
  → HandleEvent("send") on User A's View
  → hub.Broadcast(msg, A.sessionID)
  → For each other session: sess.SendInfo(msg)
  → User B's HandleInfo(msg) is called
  → User B's View re-renders with the new message
```

## Running

```bash
go run example/chat/chat.go          # http://localhost:8920
go run example/chat/chat.go -debug   # with debug panel
```

Open multiple browser tabs to simulate multiple users. Each tab can join with a different username.
