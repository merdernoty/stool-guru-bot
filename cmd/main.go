package main

import (
	"log"

	"github.com/merdernoty/stool-guru-bot/internal/app"
)

func main() {
	log.Println("🤖 Starting Stool Guru Bot...")

	application, err := app.New()
	if err != nil {
		log.Fatal("Failed to create application:", err)
	}

	if err := application.Start(); err != nil {
		log.Fatal("Application failed to start:", err)
	}
}
