package ingest

import (
	"context"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type StreamClient struct {
	url string
}

func NewStreamClient(url string) *StreamClient {
	return &StreamClient{url: url}
}

func (c *StreamClient) Connect(ctx context.Context, rawEvents chan<- []byte) {
	backoff := time.Second

	for {
		select {
		case <-ctx.Done():
			log.Println("Ingest: stopping connection loop...")
			return
		default:
			log.Printf("Ingest: connecting to %s", c.url)

			conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
			if err != nil {
				log.Printf("Ingest: dial error: %v. Retrying in %v...", err, backoff)
				time.Sleep(backoff)

				if backoff < time.Minute {
					backoff *= 2
				}
				continue
			}

			backoff = time.Second

			go func() {
				<-ctx.Done()
				log.Println("Ingest: Closing WebSocket for shutdown...")
				conn.Close()
			}()

			log.Println("Ingest: successfully connected to Binance")

			c.readPump(ctx, conn, rawEvents)

			conn.Close()
		}
	}
}

func (c *StreamClient) readPump(ctx context.Context, conn *websocket.Conn, rawEvents chan<- []byte) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		case rawEvents <- message:
		}
	}
}
