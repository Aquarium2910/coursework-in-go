package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const dbURL = "postgres://user:password@localhost:5432/mydb"

func main() {
	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	// Очистимо таблицю перед заповненням (опціонально)
	_, _ = dbPool.Exec(ctx, "TRUNCATE TABLE orders")

	types := []string{"харчування", "транспорт", "розваги", "одяг", "електроніка"}
	currencies := []string{"UAH", "USD", "EUR"}

	// Початковий курс для USD
	currentUSDrate := 41.20

	fmt.Println("⏳ Заповнення бази даних...")

	// Генеруємо дані за останні 60 днів
	for i := 0; i < 150; i++ {
		// Випадкова дата в межах 60 днів
		randomDays := rand.Intn(60)
		date := time.Now().AddDate(0, 0, -randomDays)

		// Випадковий час (для тесту 8-годинних інтервалів)
		hour := rand.Intn(24)
		minute := rand.Intn(60)
		orderTime := fmt.Sprintf("%02d:%02d:00", hour, minute)

		// Випадковий тип та валюта
		orderType := types[rand.Intn(len(types))]
		currency := currencies[rand.Intn(len(currencies))]

		// Сума: робимо багато маленьких для "харчування" і кілька великих для топів
		var amount float64
		if orderType == "харчування" && rand.Float64() > 0.3 {
			amount = 10.0 + rand.Float64()*60.0 // Часті покупки біля 50 грн
		} else {
			amount = 100.0 + rand.Float64()*5000.0
		}

		// Курс валют: іноді змінюємо його для тесту умови №5
		rate := 1.0
		if currency == "USD" {
			if rand.Float64() > 0.8 { // 20% шанс, що курс сьогодні "змінився"
				currentUSDrate += (rand.Float64() - 0.5) // невелика флуктуація
			}
			rate = currentUSDrate
		} else if currency == "EUR" {
			rate = 44.50
		}

		_, err := dbPool.Exec(ctx, `
			INSERT INTO orders (orderdate, ordertime, ordertype, amount, currency, exchangerate)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			date, orderTime, orderType, amount, currency, rate)

		if err != nil {
			log.Printf("Помилка вставки: %v", err)
		}
	}

	fmt.Println("✅ База заповнена! Можна приступати до тестів.")
}
