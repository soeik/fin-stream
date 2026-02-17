package process

import (
	"context"
	"finstream/engine/internal/models"
	"sync"
	"time"
)

type SymbolStats struct {
	trades    []models.Trade
	sumPrice  float64
	lastPrice float64
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

func (a *Aggregator) AddTrade(t models.Trade) {
	a.mu.Lock()
	defer a.mu.Unlock()

	s, ok := a.stats[t.Symbol]
	if !ok {
		s = &SymbolStats{trades: make([]models.Trade, 0)}
		a.stats[t.Symbol] = s
	}

	s.trades = append(s.trades, t)
	s.sumPrice += t.Price
	s.lastPrice = t.Price

	cutoff := time.Now().UnixMilli() - a.window.Milliseconds()

	i := 0
	for i < len(s.trades) && s.trades[i].EventTime < cutoff {
		oldPrice := s.trades[i].Price
		s.sumPrice -= oldPrice
		i++
	}
	s.trades = s.trades[i:]
}

func (a *Aggregator) GetSnapshot() []models.TradeStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	snapshot := make([]models.TradeStats, 0, len(a.stats))
	for symbol, stats := range a.stats {
		avg := 0.0
		if count := float64(len(stats.trades)); count > 0 {
			avg = stats.sumPrice / count
		}

		snapshot = append(snapshot, models.TradeStats{
			Symbol:   symbol,
			Price:    stats.lastPrice,
			AvgPrice: avg,
		})
	}
	return snapshot
}

func (a *Aggregator) RunSnapshotter(ctx context.Context, interval time.Duration, onTick func([]models.TradeStats)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			onTick(a.GetSnapshot())
		}
	}
}
