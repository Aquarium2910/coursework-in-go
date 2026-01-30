package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

const dbURL = "postgres://user:password@localhost:5432/mydb"

func main() {
	// 1. Підключення
	ctx := context.Background()
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	fmt.Println("✅ Підключено до БД для міграції")

	// 2. Створення таблиці (SQL DDL)
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS orders (
       id SERIAL PRIMARY KEY,
       orderDate DATE NOT NULL,
       orderTime TIME NOT NULL,
	   orderType VARCHAR(50) NOT NULL,
	   amount NUMERIC (15, 2),
	   currency CHAR(3) NOT NULL,
	   exchangeRate NUMERIC (10, 6)
    );`

	_, err = dbPool.Exec(ctx, createTableSQL)
	if err != nil {
		log.Fatal("Помилка створення таблиці: ", err)
	}
	fmt.Println("🔨 Таблиця 'orders' перевірена/створена успішно")
}
