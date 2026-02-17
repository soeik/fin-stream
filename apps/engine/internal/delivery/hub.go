package delivery

import (
	"context"
	"finstream/engine/internal/models"
	"sync"
)

type Hub struct {
	clients map[chan []models.TradeStats]bool
	mu      sync.RWMutex
	input   chan []models.TradeStats
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[chan []models.TradeStats]bool),
		input:   make(chan []models.TradeStats, 100),
	}
}

func (h *Hub) Register(ch chan []models.TradeStats) {
	h.mu.Lock()
	h.clients[ch] = true
	h.mu.Unlock()
}

func (h *Hub) Unregister(ch chan []models.TradeStats) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
	close(ch)
}

func (h *Hub) Broadcast(snapshot []models.TradeStats) {
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
