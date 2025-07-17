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
	"github.com/merdernoty/stool-guru-bot/internal/bot/services/gemini"
	"github.com/merdernoty/stool-guru-bot/internal/config"
	"github.com/merdernoty/stool-guru-bot/internal/server"
)

type App struct {
	config        *config.Config
	server        *server.Server
	bot           *bot.StoolGuruBot
	geminiService *gemini.GeminiService
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	log.Printf("📋 Loaded config: %s", cfg.String())

	
	geminiService, err := gemini.NewGeminiService(cfg.GeminiAPIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini service: %w", err)
	}

	botInstance, err := bot.NewBot(cfg, geminiService)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	botInstance.SetGlobalBot()

	serverInstance := server.NewServer(cfg, botInstance)

	return &App{
		config:        cfg,
		server:        serverInstance,
		bot:           botInstance,
		geminiService: geminiService,
	}, nil
}

func (a *App) Start() error {
	defer func() {
		if err := a.geminiService.Close(); err != nil {
			log.Printf("Error closing Gemini service: %v", err)
		}
	}()

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