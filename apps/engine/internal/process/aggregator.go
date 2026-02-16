package process

import (
	"finstream/engine/internal/models"
	"sync"
	"time"
)

type SymbolStats struct {
	trades   []models.Trade
	sumPrice float64
}

type Aggregator struct {
	mu     sync.RWMutex
	window time.Duration
	stats  map[string]*SymbolStats
}

func NewAggregator(window time.Duration) *Aggregator {
	return &Aggregator{
		window: window,
		stats:  make(map[string]*SymbolStats),
	}
}

func (a *Aggregator) AddTrade(t models.Trade) float64 {
	a.mu.Lock()
	defer a.mu.Unlock()

	s, ok := a.stats[t.Symbol]
	if !ok {
		s = &SymbolStats{trades: make([]models.Trade, 0)}
		a.stats[t.Symbol] = s
	}

	s.trades = append(s.trades, t)
	s.sumPrice += t.Price

	cutoff := time.Now().UnixMilli() - a.window.Milliseconds()

	i := 0
	for i < len(s.trades) && s.trades[i].EventTime < cutoff {
		oldPrice := s.trades[i].Price
		s.sumPrice -= oldPrice
		i++
	}
	s.trades = s.trades[i:]

	if len(s.trades) > 0 {
		return s.sumPrice / float64(len(s.trades))
	}
	return 0
}
