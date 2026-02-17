package process

import (
	"context"
	"encoding/json"
	"finstream/engine/internal/models"
	"log"
	"sync"
	"sync/atomic"
)

var tradePool = sync.Pool{
	New: func() any {
		return new(models.Trade)
	},
}

var totalMessages uint64

func Worker(ctx context.Context, input <-chan []byte, agg *Aggregator) {
	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-input:
			if !ok {
				return
			}

			t := tradePool.Get().(*models.Trade)

			if err := json.Unmarshal(data, t); err != nil {
				tradePool.Put(t)
				continue
			}

			agg.AddTrade(*t)

			count := atomic.AddUint64(&totalMessages, 1)
			if count%1000 == 0 {
				log.Printf("🚀 Processed %d messages. Symbol: %s", count, t.Symbol)
			}

			*t = models.Trade{}
			tradePool.Put(t)
		}
	}
}
