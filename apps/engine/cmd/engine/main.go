package main

import (
	"context"
	"finstream/engine/internal/delivery"
	"finstream/engine/internal/ingest"
	"finstream/engine/internal/process"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Initialize
	rawEvents := make(chan []byte, 100)
	agg := process.NewAggregator(60 * time.Second)
	hub := delivery.NewHub()

	var wg sync.WaitGroup

	// Start everything
	go agg.RunSnapshotter(ctx, 100*time.Millisecond, hub.Broadcast)
	startWorkers(ctx, &wg, rawEvents, agg)
	startIngestion(ctx, &wg, rawEvents)

	server := startHTTPServer(hub)

	// Wait and shutdown
	<-ctx.Done()
	return shutdown(server, rawEvents, &wg)
}

func startWorkers(ctx context.Context, wg *sync.WaitGroup, input <-chan []byte, agg *process.Aggregator) {
	for range runtime.NumCPU() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			process.Worker(ctx, input, agg)
		}()
	}
}

func startIngestion(ctx context.Context, wg *sync.WaitGroup, output chan<- []byte) {
	symbols := []string{
		"btcusdt",
		"ethusdt",
		"solusdt",
		"bnbusdt",
		"xrpusdt",
		"adausdt",
		"dogeusdt",
		"dotusdt",
		"maticusdt",
		"avaxusdt",
	}
	for _, s := range symbols {
		wg.Add(1)
		url := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@aggTrade", s)
		client := ingest.NewStreamClient(url)
		go func() {
			defer wg.Done()
			client.Connect(ctx, output)
		}()
	}
}

func startHTTPServer(hub *delivery.Hub) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", hub.HandleWS)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Println("🌍 WebSocket server started on :8080/ws")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	return server
}

func shutdown(server *http.Server, rawEvents chan []byte, wg *sync.WaitGroup) error {
	log.Println("Shutting down gracefully...")

	// Create a timeout context so we don't wait forever
	shutdownCtx, forceCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer forceCancel()

	// Stop the HTTP/WS server (stops accepting new connections)
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Wait for all goroutines (Hub, Workers, Ingest) to call wg.Done()
	wg.Wait()

	log.Println("All systems stopped. Clean exit.")
	return nil
}
