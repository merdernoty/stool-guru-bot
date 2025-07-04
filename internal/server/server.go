package server

import (
	"fmt"
	"net/http"

	"github.com/go-telegram/bot/models"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/merdernoty/stool-guru-bot/internal/bot"
	"github.com/merdernoty/stool-guru-bot/internal/config"
)

type Server struct {
	echo   *echo.Echo
	bot    *bot.StoolGuruBot
	config *config.Config
}

func NewServer(cfg *config.Config, bot *bot.StoolGuruBot) *Server {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())

	e.HideBanner = true

	return &Server{
		echo:   e,
		bot:    bot,
		config: cfg,
	}
}

func (s *Server) SetupRoutes() {
	// Health check endpoint
	s.echo.GET("/health", s.healthCheck)

	// Webhook endpoint
	s.echo.POST("/bot", s.handleWebhook)

	// Bot info endpoint
	s.echo.GET("/", s.botInfo)

	// Metrics endpoint
	s.echo.GET("/metrics", s.metrics)
}

func (s *Server) handleWebhook(c echo.Context) error {
	var update models.Update
	if err := c.Bind(&update); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	if err := s.bot.ProcessWebhookUpdate(&update); err != nil {
		c.Logger().Error("Error processing webhook update:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	}

	return c.NoContent(http.StatusOK)
}

func (s *Server) healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"bot":     "stool-guru-bot",
		"version": "2.0.0",
		"library": "github.com/go-telegram/bot",
	})
}

func (s *Server) botInfo(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"name":        "Stool Guru Bot",
		"version":     "2.0.0",
		"description": "Modern Telegram bot for stool health insights",
		"status":      "running",
		"library":     "github.com/go-telegram/bot",
		"features": []string{
			"Health analysis",
			"Personal recommendations",
			"Interactive callbacks",
			"Smart polling/webhook",
		},
	})
}

func (s *Server) metrics(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"uptime": "running",
		"mode":   s.config.Debug,
	})
}

func (s *Server) Start() error {
	s.SetupRoutes()

	fmt.Printf("🚀 Starting modern server on port %s\n", s.config.Port)
	return s.echo.Start(":" + s.config.Port)
}

func (s *Server) Shutdown() error {
	return s.echo.Close()
}
