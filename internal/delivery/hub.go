package delivery

import (
	"context"
	"finstream/engine/internal/market"
	"sync"
)

type Hub struct {
	clients map[chan []market.TickStats]bool
	mu      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[chan []market.TickStats]bool),
	}
}

func (h *Hub) Register(ch chan []market.TickStats) {
	h.mu.Lock()
	h.clients[ch] = true
	h.mu.Unlock()
}

func (h *Hub) Unregister(ch chan []market.TickStats) {
	h.mu.Lock()
	if _, ok := h.clients[ch]; ok {
		delete(h.clients, ch)
		close(ch)
	}

	h.mu.Unlock()

}

func (h *Hub) Broadcast(snapshot []market.TickStats) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- snapshot:
		default:
			// Skip if full
		}
	}
}

func (h *Hub) Run(ctx context.Context) {
	<-ctx.Done()
	h.stop()
}

func (h *Hub) stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.clients {
		close(ch)
		delete(h.clients, ch)
	}
}
