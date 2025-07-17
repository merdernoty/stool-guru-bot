package bot

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/merdernoty/stool-guru-bot/internal/bot/handlers/callbacks"
	"github.com/merdernoty/stool-guru-bot/internal/bot/handlers/commands"
	"github.com/merdernoty/stool-guru-bot/internal/bot/router"
	"github.com/merdernoty/stool-guru-bot/internal/bot/services/gemini"
	"github.com/merdernoty/stool-guru-bot/internal/config"
)

type StoolGuruBot struct {
	bot           *bot.Bot
	config        *config.Config
	router        *router.Router
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewBot(cfg *config.Config, geminiService *gemini.GeminiService) (*StoolGuruBot, error) {
	ctx, cancel := context.WithCancel(context.Background())
	httpClient := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: cfg.Timeout,
		},
	}

	log.Printf("🔄 Creating bot with custom HTTP client (timeout: %v)", cfg.Timeout)

	opts := []bot.Option{
		bot.WithDefaultHandler(defaultHandler),
		bot.WithCheckInitTimeout(cfg.Timeout),
		bot.WithHTTPClient(30*time.Second, httpClient),
	}

	if cfg.Debug {
		opts = append(opts, bot.WithMiddlewares(debugMiddleware))
		log.Println("🔍 Debug mode enabled")
	}

	b, err := bot.New(cfg.TelegramToken, opts...)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	startHandler := commands.NewStartHandler()
	helpHandler := commands.NewHelpHandler()
	callbackHandlers := callbacks.NewCallbackHandlers()


	botRouter := router.NewRouter(
		startHandler,
		helpHandler,

		callbackHandlers,
	)

	stoolBot := &StoolGuruBot{
		bot:    b,
		config: cfg,
		router: botRouter,
		ctx:    ctx,
		cancel: cancel,
	}

	stoolBot.router.RegisterHandlers(stoolBot.bot)

	log.Printf("✅ Bot initialized successfully")
	return stoolBot, nil
}

func (sb *StoolGuruBot) StartPolling() error {
	log.Println("🔄 Starting bot in polling mode...")

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("🛑 Received shutdown signal...")
		sb.cancel()
	}()

	log.Println("✅ Bot started! Попробуйте отправить /start в Telegram")
	sb.bot.Start(sb.ctx)
	log.Println("✅ Bot stopped gracefully")
	return nil
}

func (sb *StoolGuruBot) SetWebhook() error {
	if sb.config.WebhookURL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	webhookURL := sb.config.WebhookURL + "/bot"

	ctxWithTimeout, cancel := context.WithTimeout(sb.ctx, sb.config.Timeout)
	defer cancel()

	_, err := sb.bot.SetWebhook(ctxWithTimeout, &bot.SetWebhookParams{
		URL: webhookURL,
	})
	if err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	log.Printf("📡 Webhook set to: %s", webhookURL)
	return nil
}

func (sb *StoolGuruBot) ProcessWebhookUpdate(update *models.Update) error {
	sb.bot.ProcessUpdate(sb.ctx, update)
	return nil
}

func (sb *StoolGuruBot) SetGlobalBot() {
	// TODO: Убрать когда перенесем обработку фото в отдельный хэндлер
	globalStoolBot = sb
}


func debugMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message != nil {
			log.Printf("🔍 Message from @%s: %s",
				update.Message.From.Username,
				update.Message.Text)
		}
		if update.CallbackQuery != nil {
			log.Printf("🔍 Callback from @%s: %s",
				update.CallbackQuery.From.Username,
				update.CallbackQuery.Data)
		}
		next(ctx, b, update)
	}
}

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message != nil && len(update.Message.Photo) > 0 {
		log.Printf("📸 Photo received from @%s", update.Message.From.Username)
		
		// TODO: Перенести в отдельный хэндлер фотографий
		if globalStoolBot != nil {
			globalStoolBot.handlePhoto(ctx, b, update)
		}
		return
	}

	if update.Message != nil && update.Message.Text != "" {
		log.Printf("📨 Unhandled message: %s", update.Message.Text)

		response := `🤔 Не понимаю эту команду.

**Доступные команды:**
• /start - главное меню
• /help - справка
• /test - тест функций

📸 Или просто отправьте фото для анализа!`

		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      response,
			ParseMode: models.ParseModeMarkdown,
		})
		if err != nil {
			log.Printf("Error in default handler: %v", err)
		}
	}
}

// TODO: Временная глобальная переменная, убрать после рефакторинга фото хэндлера
var globalStoolBot *StoolGuruBot

func (sb *StoolGuruBot) handlePhoto(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO: Вынести в handlers/media/photo.go
	log.Println("📸 Photo received for analysis - using legacy handler")
	
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "📸 Фото получено! Обработка фотографий будет реализована в следующей итерации рефакторинга.",
	})
	if err != nil {
		log.Printf("Error sending photo response: %v", err)
	}
}