package main

import (
	"context"
	"finstream/engine/internal/bootstrap"
	"finstream/engine/internal/delivery"
	"finstream/engine/internal/market"
	"finstream/engine/internal/process"
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
	tickStream := make(chan market.Tick, 1000)
	agg := process.NewAggregator(60 * time.Second)
	hub := delivery.NewHub()

	var wg sync.WaitGroup

	bootstrap.InitIngesters(ctx, &wg, tickStream)
	startWorkers(ctx, &wg, tickStream, agg)

	go bootstrap.InitStrategy(ctx, agg)
	go agg.RunSnapshotter(ctx, 200*time.Millisecond, hub.Broadcast)

	server := startHTTPServer(hub)

	// Wait and shutdown
	<-ctx.Done()
	return shutdown(server, &wg)
}

func startWorkers(ctx context.Context, wg *sync.WaitGroup, input <-chan market.Tick, agg *process.Aggregator) {
	for range runtime.NumCPU() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			process.Worker(ctx, input, agg)
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
		log.Println("WebSocket server started on :8080/ws")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	return server
}

func shutdown(server *http.Server, wg *sync.WaitGroup) error {
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
