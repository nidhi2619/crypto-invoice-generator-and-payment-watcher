package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/user/crypto-invoice-generator/backend/internal/application"
	"github.com/user/crypto-invoice-generator/backend/internal/config"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load Config
	cfg := config.NewConfig()

	// Start Application
	application.StartApp(cfg)
}
