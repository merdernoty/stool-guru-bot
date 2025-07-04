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
	"github.com/merdernoty/stool-guru-bot/internal/config"
)

type StoolGuruBot struct {
	bot    *bot.Bot
	config *config.Config
	ctx    context.Context
	cancel context.CancelFunc
}

func NewBot(cfg *config.Config) (*StoolGuruBot, error) {
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

	stoolBot := &StoolGuruBot{
		bot:    b,
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}

	stoolBot.registerHandlers()

	log.Printf("✅ Bot initialized successfully")
	return stoolBot, nil
}

func (sb *StoolGuruBot) registerHandlers() {
	// Команды
	sb.bot.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, sb.handleStart)
	sb.bot.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, sb.handleHelp)
	sb.bot.RegisterHandler(bot.HandlerTypeMessageText, "/test", bot.MatchTypeExact, sb.handleTest)
	sb.bot.RegisterHandler(bot.HandlerTypeMessageText, "/analyze", bot.MatchTypeExact, sb.handleAnalyze)

	// Callback queries
	sb.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "test", bot.MatchTypeExact, sb.handleTestCallback)
	sb.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "help", bot.MatchTypeExact, sb.handleHelpCallback)
	sb.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "analyze", bot.MatchTypeExact, sb.handleAnalyzeCallback)
	sb.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "analyze_good", bot.MatchTypeExact, sb.handleAnalyzeGoodCallback)
	sb.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "analyze_normal", bot.MatchTypeExact, sb.handleAnalyzeNormalCallback)
	sb.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "analyze_bad", bot.MatchTypeExact, sb.handleAnalyzeBadCallback)

	log.Println("📝 Handlers registered successfully")
}

// Message handlers
func (sb *StoolGuruBot) handleStart(ctx context.Context, b *bot.Bot, update *models.Update) {
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "🧪 Тест", CallbackData: "test"},
				{Text: "❓ Помощь", CallbackData: "help"},
			},
			{
				{Text: "📊 Анализ", CallbackData: "analyze"},
			},
		},
	}

	text := "🤖 Stool Guru Bot запущен\n\nПривет, я готов помочь вам с анализом здоровья.\n\nВыберите действие в меню ниже:"

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        text,
		ReplyMarkup: keyboard,
	})
	if err != nil {
		log.Printf("Error sending start message: %v", err)
	}
}

func (sb *StoolGuruBot) handleHelp(ctx context.Context, b *bot.Bot, update *models.Update) {
	text := "🆘 Команды бота\n\n/start • Главное меню\n/help • Эта справка\n/test • Тест функций\n/analyze • Анализ здоровья\n\nБот работает отлично 🎉"

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   text,
	})
	if err != nil {
		log.Printf("Error sending help message: %v", err)
	}
}

func (sb *StoolGuruBot) handleTest(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("🧪 Test command received")

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "🧪 Тест прошел! HTTP клиент работает корректно.",
	})
	if err != nil {
		log.Printf("Error sending test message: %v", err)
	}
}

func (sb *StoolGuruBot) handleAnalyze(ctx context.Context, b *bot.Bot, update *models.Update) {
	keyboard := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "🟢 Хорошее", CallbackData: "analyze_good"},
				{Text: "🟡 Нормальное", CallbackData: "analyze_normal"},
			},
			{
				{Text: "🔴 Есть проблемы", CallbackData: "analyze_bad"},
			},
		},
	}

	text := "📊 Анализ состояния здоровья\n\nКак вы оцениваете ваше текущее состояние пищеварения?\n\nВыберите наиболее подходящий вариант:"

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        text,
		ReplyMarkup: keyboard,
	})
	if err != nil {
		log.Printf("Error sending analyze message: %v", err)
	}
}

func (sb *StoolGuruBot) handleTestCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "✅ Тест прошел отлично!",
		ShowAlert:       true,
	})
	if err != nil {
		log.Printf("Error answering callback: %v", err)
	}
}

func (sb *StoolGuruBot) handleHelpCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "📋 Команды: /start /help /test /analyze",
		ShowAlert:       false,
	})
	if err != nil {
		log.Printf("Error answering callback: %v", err)
	}
}

func (sb *StoolGuruBot) handleAnalyzeCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "📊 Выберите ваше состояние ниже",
		ShowAlert:       false,
	})
	if err != nil {
		log.Printf("Error answering callback: %v", err)
	}
}

func (sb *StoolGuruBot) handleAnalyzeGoodCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "🟢 Отлично! Продолжайте в том же духе! Пейте воду, ешьте клетчатку, двигайтесь!",
		ShowAlert:       true,
	})
	if err != nil {
		log.Printf("Error answering callback: %v", err)
	}
}

func (sb *StoolGuruBot) handleAnalyzeNormalCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "🟡 Нормально! Советы: больше пробиотиков, овощей, прогулки после еды",
		ShowAlert:       true,
	})
	if err != nil {
		log.Printf("Error answering callback: %v", err)
	}
}

func (sb *StoolGuruBot) handleAnalyzeBadCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "🔴 При серьезных симптомах обратитесь к врачу! Пейте воду, избегайте острого",
		ShowAlert:       true,
	})
	if err != nil {
		log.Printf("Error answering callback: %v", err)
	}
}

// Bot control methods
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

// Middleware
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

// Default handler для необработанных сообщений
func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message != nil && update.Message.Text != "" {
		log.Printf("📨 Unhandled message: %s", update.Message.Text)

		response := "🤔 Не понимаю эту команду.\n\nПопробуйте:\n• /start • главное меню\n• /help • справка\n• /test • тест\n• /analyze • анализ"

		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   response,
		})
		if err != nil {
			log.Printf("Error in default handler: %v", err)
		}
	}
}
