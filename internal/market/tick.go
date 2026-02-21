package market

import "github.com/shopspring/decimal"

type Tick struct {
	Symbol     string          `json:"s"`
	Price      decimal.Decimal `json:"p"`
	Quantity   decimal.Decimal `json:"q"`
	Timestamp  int64           `json:"T"` // Event occured
	ReceivedAt int64           `json:"r"` // Event Received
}

type TickStats struct {
	Symbol        string          `json:"s"`
	Price         decimal.Decimal `json:"p"`  // Last Price
	AvgPrice      decimal.Decimal `json:"a"`  // Average in window
	MinPrice      decimal.Decimal `json:"mi"` // Min in window
	MaxPrice      decimal.Decimal `json:"ma"` // Max price in window
	IsVolumeSpike bool            `json:"vs"`
}
