package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/merdernoty/stool-guru-bot/internal/bot"
	"github.com/merdernoty/stool-guru-bot/internal/config"
	"github.com/merdernoty/stool-guru-bot/internal/server"
)

type App struct {
	config *config.Config
	server *server.Server
	bot    *bot.StoolGuruBot
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	log.Printf("📋 Loaded config: %s", cfg.String())

	botInstance, err := bot.NewBot(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	serverInstance := server.NewServer(cfg, botInstance)

	return &App{
		config: cfg,
		server: serverInstance,
		bot:    botInstance,
	}, nil
}

func (a *App) Start() error {
	if a.config.Debug {
		log.Println("🔄 Debug mode: using polling instead of webhook")
		return a.bot.StartPolling()
	}

	if err := a.bot.SetWebhook(); err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	go func() {
		if err := a.server.Start(); err != nil {
			log.Fatal("Server failed to start:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")


	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	log.Println("✅ Server stopped gracefully")
	return nil
}
