package models

type TradeStats struct {
	Symbol   string  `json:"s"`
	Price    float64 `json:"p"`
	AvgPrice float64 `json:"a"`
}
