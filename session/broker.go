package session

import "sync"

// Broker distributes session invalidation events to subscribers.
// When a session is destroyed or expires, all subscribers for that
// session ID are notified via their channels.
type Broker struct {
	mu          sync.RWMutex
	subscribers map[string]map[uint64]chan struct{}
	nextID      uint64
}

// NewBroker creates a new Broker.
func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[string]map[uint64]chan struct{}),
	}
}

// Subscribe registers for invalidation notifications on the given session ID.
// Returns a channel that will be closed when the session is invalidated,
// and an unsubscribe function that must be called when done (typically via defer).
func (b *Broker) Subscribe(sessionID string) (ch <-chan struct{}, unsubscribe func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := b.nextID
	b.nextID++

	c := make(chan struct{}, 1)
	if b.subscribers[sessionID] == nil {
		b.subscribers[sessionID] = make(map[uint64]chan struct{})
	}
	b.subscribers[sessionID][id] = c

	return c, func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		delete(b.subscribers[sessionID], id)
		if len(b.subscribers[sessionID]) == 0 {
			delete(b.subscribers, sessionID)
		}
	}
}

// Invalidate notifies all subscribers for the given session ID.
// Each subscriber's channel receives a signal (non-blocking send).
func (b *Broker) Invalidate(sessionID string) {
	b.mu.RLock()
	subs := b.subscribers[sessionID]
	if len(subs) == 0 {
		b.mu.RUnlock()
		return
	}
	// Copy channels under read lock to avoid holding lock during sends
	chs := make([]chan struct{}, 0, len(subs))
	for _, ch := range subs {
		chs = append(chs, ch)
	}
	b.mu.RUnlock()

	for _, ch := range chs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
