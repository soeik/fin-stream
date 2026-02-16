package process

import (
	"encoding/json"
	"finstream/engine/internal/models"
	"log"
	"sync"
	"sync/atomic"
)

var tradePool = sync.Pool{
	New: func() any {
		// Мы создаем пустую структуру, которую будем переиспользовать
		return new(models.Trade)
	},
}

var totalMessages uint64

func Worker(input <-chan []byte, output chan<- *models.Trade, agg *Aggregator) {
	for data := range input {
		t := tradePool.Get().(*models.Trade)

		if err := json.Unmarshal(data, t); err != nil {
			tradePool.Put(t)
			log.Printf("Parser: error unmarshaling: %v", err)
			continue
		}

		agg.AddTrade(*t)

		count := atomic.AddUint64(&totalMessages, 1)

		if count%1000 == 0 {
			log.Printf("🚀 Processed %d messages. Current symbol: %s", count, t.Symbol)
		}

		*t = models.Trade{}
		tradePool.Put(t)

	}
}
