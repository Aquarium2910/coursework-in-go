package main

import (
	"bytes"
	"coursework/internal/postgres"
	"testing"
)

//func TestOutput(t *testing.T) {
//	t.Run("menu output test", func(t *testing.T) {
//		buffer := &bytes.Buffer{}
//		spyReader := spyReader{}
//
//		menu(buffer)
//
//		got := buffer.String()
//		want := "1. List all orders\n2. Add new order\n"
//
//		if got != want {
//			t.Errorf("got %q, want %q", got, want)
//		}
//	})
//}

func TestPrintDb(t *testing.T) {
	t.Run("printing first 10 elements of db", func(t *testing.T) {
		controller := postgres.NewDbController(dbURL)

		buffer := &bytes.Buffer{}
		printDb(buffer, controller, 10)

		got := buffer.String()
		want := "Id\tDate and time of order\tOrder type\tPay amount\t\tCurrency \tExchange rate\t\n1 2025-12-08 21:04:00 +0000 UTC розваги 1639.970000 UAH 1.000000\n2 2025-12-19 18:09:00 +0000 UTC розваги 4390.890000 USD 41.200000\n3 2025-12-20 12:04:00 +0000 UTC харчування 3595.270000 USD 40.757892\n4 2026-01-26 22:28:00 +0000 UTC електроніка 3283.220000 UAH 1.000000\n5 2026-01-01 00:29:00 +0000 UTC транспорт 4275.460000 UAH 1.000000\n6 2025-12-15 13:15:00 +0000 UTC електроніка 2129.200000 USD 40.757892\n7 2025-12-20 09:50:00 +0000 UTC харчування 4991.050000 USD 40.757892\n8 2026-01-23 23:37:00 +0000 UTC харчування 52.290000 EUR 44.500000\n9 2026-01-14 21:13:00 +0000 UTC розваги 2725.680000 USD 40.757892\n10 2025-12-01 07:42:00 +0000 UTC транспорт 640.240000 USD 40.587195\n11 2025-12-07 17:40:00 +0000 UTC транспорт 1374.570000 EUR 44.500000\n12 2025-12-09 04:53:00 +0000 UTC розваги 394.880000 USD 40.587195\n13 2026-01-20 21:08:00 +0000 UTC розваги 4752.430000 EUR 44.500000\n14 2026-01-05 03:57:00 +0000 UTC розваги 3598.020000 UAH 1.000000\n15 2026-01-25 17:53:00 +0000 UTC розваги 2846.830000 UAH 1.000000\n16 2026-01-17 04:35:00 +0000 UTC електроніка 2305.070000 UAH 1.000000\n17 2026-01-28 20:15:00 +0000 UTC електроніка 2015.100000 EUR 44.500000\n18 2025-12-12 06:36:00 +0000 UTC одяг 1768.790000 USD 40.587195\n19 2026-01-08 12:06:00 +0000 UTC розваги 3622.600000 EUR 44.500000\n20 2026-01-07 01:06:00 +0000 UTC транспорт 1025.270000 USD 40.587195"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestInsertionToDb(t *testing.T) {
	t.Run("insertion to db", func(t *testing.T) {
		controller := postgres.NewDbController(dbURL)

	})
}
