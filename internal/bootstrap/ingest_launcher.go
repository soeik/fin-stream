package bootstrap

import (
	"context"
	"finstream/engine/internal/ingest"
	"finstream/engine/internal/market"
	"log"
	"sync"
)

func startSource(ctx context.Context, wg *sync.WaitGroup, source ingest.Ingester, output chan<- market.Tick) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := source.Run(ctx, output); err != nil {
			log.Printf("Ingest error: %v", err)
		}
	}()
}
