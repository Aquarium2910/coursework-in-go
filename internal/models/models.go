package models

import "time"

type Order struct {
	Id           int
	TimeStamp    time.Time
	Type         string
	Amount       float64
	Currency     string
	ExchangeRate float64
}

type BiggestOrders struct {
	Date     time.Time
	TotalUah float64
}

type PeriodStats struct {
	TimePeriod string
	TotalSales int
	BigSales   int
	SmallSales int
}
