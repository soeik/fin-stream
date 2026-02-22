package bootstrap

import (
	"context"
	"finstream/engine/internal/ingest"
	"finstream/engine/internal/market"
	"sync"
)

var symbols = []string{
	"btcusdt", "ethusdt", "solusdt", "bnbusdt", "xrpusdt",
	"adausdt", "dogeusdt", "dotusdt", "maticusdt", "avaxusdt",
	"trxusdt", "shibusdt", "ltcusdt", "linkusdt", "nearusdt",
	"atomusdt", "aptusdt", "arbusdt", "opusdt", "ldousdt",
}

func InitIngesters(ctx context.Context, wg *sync.WaitGroup, output chan<- market.Tick) {
	binance := ingest.NewBinanceSource(symbols)

	startSource(ctx, wg, binance, output)
}
