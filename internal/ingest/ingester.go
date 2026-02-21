package ingest

import (
	"context"
	"finstream/engine/internal/market"
)

type Ingester interface {
	Run(ctx context.Context, output chan<- market.Tick) error
}
