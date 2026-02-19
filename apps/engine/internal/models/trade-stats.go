package models

type TradeStats struct {
	Symbol   string  `json:"s"`
	Price    float64 `json:"p"`
	AvgPrice float64 `json:"a"`
	MinPrice float64 `json:"m"`
	MaxPrice float64 `json:"x"`
}
