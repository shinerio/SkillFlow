package notify

import "sync"

type Hub struct {
	mu          sync.RWMutex
	subscribers map[chan Event]struct{}
}

func NewHub() *Hub {
	return &Hub{subscribers: make(map[chan Event]struct{})}
}

func (h *Hub) Subscribe() <-chan Event {
	ch := make(chan Event, 32)
	h.mu.Lock()
	h.subscribers[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(ch <-chan Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for sub := range h.subscribers {
		if sub == ch {
			delete(h.subscribers, sub)
			close(sub)
			return
		}
	}
}

func (h *Hub) Publish(evt Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for sub := range h.subscribers {
		select {
		case sub <- evt:
		default: // drop if subscriber is slow
		}
	}
}
