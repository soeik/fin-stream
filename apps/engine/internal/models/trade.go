package models

type Trade struct {
	Symbol    string  `json:"s"`
	Price     float64 `json:"p,string"`
	EventTime int64   `json:"T"`
}
