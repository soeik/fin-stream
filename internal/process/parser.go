package process

import (
	"context"
	"finstream/engine/internal/market"
	"log"
	"sync/atomic"
)

var totalMessages uint64

func Worker(ctx context.Context, input <-chan market.Tick, agg *Aggregator) {
	for {
		select {
		case <-ctx.Done():
			return
		case tick, ok := <-input:
			if !ok {
				return
			}

			agg.AddTick(tick)

			count := atomic.AddUint64(&totalMessages, 1)
			if count%5000 == 0 {
				log.Printf("Engine Speed: %d ticks. Current: %s @ %s",
					count, tick.Symbol, tick.Price.String())
			}
		}
	}
}
