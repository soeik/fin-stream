package process

import (
	"context"
	"finstream/engine/internal/market"
	"sync"
	"time"
	"log"

	"github.com/shopspring/decimal"
)

type Aggregator struct {
	mu     sync.RWMutex
	window time.Duration
	stats  map[string]*SymbolStats
}

type SymbolStats struct {
	ticks         []market.Tick
	sumPrice      decimal.Decimal
	sumVolume     decimal.Decimal
	lastPrice     decimal.Decimal
	minPrice      decimal.Decimal
	maxPrice      decimal.Decimal
	IsVolumeSpike bool
}

func NewAggregator(window time.Duration) *Aggregator {
	return &Aggregator{
		window: window,
		stats:  make(map[string]*SymbolStats),
	}
}

func (a *Aggregator) AddTick(t market.Tick) {
	log.Printf("AddTick called for %s", t.Symbol)
	a.mu.Lock()
	defer a.mu.Unlock()

	s, ok := a.stats[t.Symbol]
	if !ok {
		s = &SymbolStats{
			ticks:     make([]market.Tick, 0, 1000),
			minPrice:  t.Price,
			maxPrice:  t.Price,
			lastPrice: t.Price,
		}
		a.stats[t.Symbol] = s
	}

	avgVolume := decimal.Zero
	if len(s.ticks) > 0 {
		avgVolume = s.sumVolume.Div(decimal.NewFromInt(int64(len(s.ticks))))
	}

	threshold := decimal.NewFromInt(5)
	if avgVolume.GreaterThan(decimal.Zero) && t.Quantity.GreaterThan(avgVolume.Mul(threshold)) {
		s.IsVolumeSpike = true
	} else {
		s.IsVolumeSpike = false
	}

	s.ticks = append(s.ticks, t)
	s.sumPrice = s.sumPrice.Add(t.Price)
	s.sumVolume = s.sumVolume.Add(t.Quantity)
	s.lastPrice = t.Price

	if t.Price.LessThan(s.minPrice) {
		s.minPrice = t.Price
	}
	if t.Price.GreaterThan(s.maxPrice) {
		s.maxPrice = t.Price
	}

	cutoff := time.Now().UnixMilli() - a.window.Milliseconds()

	i := 0
	for i < len(s.ticks) && s.ticks[i].Timestamp < cutoff {
		oldTick := s.ticks[i]
		s.sumPrice = s.sumPrice.Sub(oldTick.Price)
		s.sumVolume = s.sumVolume.Sub(oldTick.Quantity)
		i++
	}
	s.ticks = s.ticks[i:]
}

func (a *Aggregator) GetSnapshot() []market.TickStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	snapshot := make([]market.TickStats, 0, len(a.stats))
	for symbol, s := range a.stats {
		count := int64(len(s.ticks))
		avg := s.lastPrice

		if count > 0 {
			avg = s.sumPrice.Div(decimal.NewFromInt(count))
		}

		snapshot = append(snapshot, market.TickStats{
			Symbol:   symbol,
			Price:    s.lastPrice,
			AvgPrice: avg,
			MinPrice: s.minPrice,
			MaxPrice: s.maxPrice,
		})
	}
	return snapshot
}

func (a *Aggregator) RunSnapshotter(ctx context.Context, interval time.Duration, onTick func([]market.TickStats)) {
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
