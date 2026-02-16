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
				log.Println("Ingest: successfully connected to Binance")

				c.readPump(ctx, conn, rawEvents)
				
				conn.Close()
		}
	}
}

func (c *StreamClient) readPump(ctx context.Context, conn *websocket.Conn, rawEvents chan<- []byte) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Устанавливаем таймаут на чтение (опционально, для контроля "мертвых" сессий)
			// conn.SetReadDeadline(time.Now().Add(60 * time.Second))

			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Ingest: read error: %v", err)
				return 
			}

			rawEvents <- message
		}
	}
}
