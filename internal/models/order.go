package models

import "time"

type Order struct {
	Id                  int
	OrderTimeStamp      time.Time
	OrderType           string
	Amount              float64
	Currency            string
	ExchangeRate        float64
	OrderTimeStampInput string
	TotalUah            float64
}
