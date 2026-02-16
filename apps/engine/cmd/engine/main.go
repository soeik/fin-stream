package main

import (
	"context"
	"finstream/engine/internal/ingest"
	"finstream/engine/internal/models"
	"finstream/engine/internal/process"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	rawEvents := make(chan []byte, 100)

	log.Println("🚀 Starting FinStream Engine...")

	symbols := []string{"btcusdt", "ethusdt", "solusdt", "bnbusdt", "xrpusdt", "adausdt"}

	for _, s := range symbols {
		url := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@aggTrade", s)
		client := ingest.NewStreamClient(url)
		go client.Connect(ctx, rawEvents) // Все пишут в один канал rawEvents!
	}

	parsedTrades := make(chan *models.Trade, 1000)

	agg := process.NewAggregator(5 * time.Second)

	for range 5 {
		go process.Worker(rawEvents, parsedTrades, agg)
	}

	<-stop
	log.Println("⚠️ Shutting down...")

	cancel()

	log.Println("✅ Done.")
}
