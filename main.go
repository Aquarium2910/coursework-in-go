package main

import (
	"coursework/internal/app"
	"coursework/internal/postgres"
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	cleanup, err := app.StartupLogger()
	if err != nil {
		panic(err)
	}
	defer cleanup()

	slog.Info("Logger initialized")

	_ = godotenv.Load()
	dbURL := os.Getenv("DB_URL")

	slog.Info("Loaded environment variables")

	controller := postgres.NewDbController(dbURL)
	defer controller.Close()

	slog.Info("✅ Connected to DB")

	err = app.Menu(os.Stdout, os.Stdin, controller)

	if err != nil {
		log.Fatal(err)
	}
}
