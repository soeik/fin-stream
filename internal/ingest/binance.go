package ingest

import (
	"context"
	"encoding/json"
	"finstream/engine/internal/market"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type binanceCombinedFrame struct {
	Stream string         `json:"stream"`
	Data   binanceRawTick `json:"data"`
}

type binanceRawTick struct {
	Symbol    string `json:"s"`
	Price     string `json:"p"`
	Quantity  string `json:"q"`
	Timestamp int64  `json:"T"`
}

type BinanceSource struct {
	symbols []string
}

func (b *BinanceSource) streamAdapter(ctx context.Context, client *StreamClient, output chan<- market.Tick) {
	rawCh := make(chan []byte, 1000)

	go client.Connect(ctx, rawCh)

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-rawCh:
			if !ok {
				return
			}

			var frame binanceCombinedFrame
			if err := json.Unmarshal(msg, &frame); err != nil {
				log.Printf("JSON Unmarshal error: %v", err)
				continue
			}

			if frame.Data.Symbol == "" {
				continue
			}

			tick, err := b.convertToDomain(frame.Data)

			if err != nil {
				log.Printf("Conversion error: %v", err)
				continue
			}

			output <- tick
		}
	}
}

func (b *BinanceSource) convertToDomain(raw binanceRawTick) (market.Tick, error) {
	p, err := decimal.NewFromString(raw.Price)
	if err != nil {
		return market.Tick{}, err
	}

	q, err := decimal.NewFromString(raw.Quantity)
	if err != nil {
		return market.Tick{}, err
	}

	return market.Tick{
		Symbol:     strings.ToLower(raw.Symbol),
		Price:      p,
		Quantity:   q,
		Timestamp:  raw.Timestamp,
		ReceivedAt: time.Now().UnixNano(),
	}, nil
}

func NewBinanceSource(symbols []string) *BinanceSource {
	normalized := make([]string, len(symbols))
	for i, s := range symbols {
		normalized[i] = strings.ToLower(s)
	}
	return &BinanceSource{symbols: normalized}
}

func (b *BinanceSource) Run(ctx context.Context, output chan<- market.Tick) error {
	streams := make([]string, len(b.symbols))
	for i, s := range b.symbols {
		streams[i] = fmt.Sprintf("%s@aggTrade", s)
	}
	url := fmt.Sprintf("wss://stream.binance.com:9443/stream?streams=%s", strings.Join(streams, "/"))
	log.Printf("String %s", url)
	client := NewStreamClient(url)

	b.streamAdapter(ctx, client, output)
	return nil
}
