package process

import (
	"encoding/json"
	"finstream/engine/internal/models"
	"log"
)

func Worker(input <-chan []byte, output chan<- *models.Trade, agg *Aggregator) {
	for data := range input {
		var trade models.Trade

		if err := json.Unmarshal(data, &trade); err != nil {
			log.Printf("Parser: error unmarshaling: %v", err)
			continue
		}

		avgPrice := agg.AddTrade(trade)

		log.Printf("📊 [%s] 5s Avg Price: %.2f", trade.Symbol, avgPrice)

		output <- &trade
	}
}
