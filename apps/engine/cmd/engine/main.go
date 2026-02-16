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

	// go func() {
	// 	for trade := range parsedTrades {
	// 		log.Printf("💰 Trade: %s | Price: %s | Qty: %s",
	// 			trade.Symbol, trade.Price, trade.Quantity)
	// 	}
	// }()

	// go func() {
	// 	var count int
	// 	ticker := time.NewTicker(time.Second) // Срабатывает раз в секунду
	// 	defer ticker.Stop()

	// 	for {
	// 		select {
	// 			case <-ctx.Done():
	// 				return
	// 			case <-parsedTrades: // Просто забираем сделку и инкрементируем счетчик
	// 				count++
	// 			case <-ticker.C: // Каждую секунду выводим результат
	// 				log.Printf("📊 Throughput: %d trades/sec", count)
	// 				count = 0 // Сбрасываем для следующей секунды
	// 			}
	// 	}
	// }()111111111

	<-stop
	log.Println("⚠️ Shutting down...")

	cancel()

	log.Println("✅ Done.")
}
