package bot

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/merdernoty/stool-guru-bot/internal/config"
	"github.com/merdernoty/stool-guru-bot/pkg/gemini"
)

type StoolGuruBot struct {
	bot           *bot.Bot
	config        *config.Config
	geminiService *gemini.GeminiService
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

	stoolBot := &StoolGuruBot{
		bot:           b,
		config:        cfg,
		geminiService: geminiService,
		ctx:           ctx,
		cancel:        cancel,
	}

	stoolBot.registerHandlers()

	log.Printf("✅ Bot initialized successfully")
	return stoolBot, nil
}

func (sb *StoolGuruBot) registerHandlers() {
	sb.bot.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, sb.handleStart)
	sb.bot.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, sb.handleHelp)
	sb.bot.RegisterHandler(bot.HandlerTypeMessageText, "/test", bot.MatchTypeExact, sb.handleTest)
	sb.bot.RegisterHandler(bot.HandlerTypeMessageText, "/analyze", bot.MatchTypeExact, sb.handleAnalyze)

	sb.bot.RegisterHandler(bot.HandlerTypeMessagePhoto, "", bot.MatchTypeExact, sb.handlePhoto)

	// Callback queries
	sb.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "test", bot.MatchTypeExact, sb.handleTestCallback)
	sb.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "help", bot.MatchTypeExact, sb.handleHelpCallback)
	sb.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "analyze", bot.MatchTypeExact, sb.handleAnalyzeCallback)
	sb.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "analyze_good", bot.MatchTypeExact, sb.handleAnalyzeGoodCallback)
	sb.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "analyze_normal", bot.MatchTypeExact, sb.handleAnalyzeNormalCallback)
	sb.bot.RegisterHandler(bot.HandlerTypeCallbackQueryData, "analyze_bad", bot.MatchTypeExact, sb.handleAnalyzeBadCallback)

	log.Println("📝 Handlers registered successfully")
}

func (sb *StoolGuruBot) handlePhoto(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("📸 Photo received for analysis")

	loadingMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "🔬 Анализирую ваше фото... Это может занять несколько секунд.",
	})
	if err != nil {
		log.Printf("Error sending loading message: %v", err)
		return
	}

	var photo *models.PhotoSize
	if len(update.Message.Photo) > 0 {
		photo = &update.Message.Photo[len(update.Message.Photo)-1] 
	} else {
		sb.sendErrorMessage(ctx, b, update.Message.Chat.ID, "Фото не найдено")
		return
	}

	imageBytes, mimeType, err := sb.downloadFile(ctx, b, photo.FileID)
	if err != nil {
		log.Printf("Error downloading file: %v", err)
		sb.sendErrorMessage(ctx, b, update.Message.Chat.ID, "Ошибка загрузки фото")
		return
	}

	result, err := sb.geminiService.AnalyzeImage(ctx, imageBytes, mimeType)
	if err != nil {
		log.Printf("Error analyzing image: %v", err)
		sb.sendErrorMessage(ctx, b, update.Message.Chat.ID, "Ошибка анализа фото")
		return
	}

	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    update.Message.Chat.ID,
		MessageID: loadingMsg.MessageID,
	})

	responseText := fmt.Sprintf("🔬 **Результат анализа:**\n\n%s", result.Text)
	
	if len(responseText) > 4000 {
		responseText = responseText[:4000] + "...\n\n✂️ *Результат сокращен*"
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      responseText,
		ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		log.Printf("Error sending analysis result: %v", err)
		sb.sendErrorMessage(ctx, b, update.Message.Chat.ID, "Ошибка отправки результата")
	}

	log.Println("✅ Photo analysis completed and sent")
}

func (sb *StoolGuruBot) downloadFile(ctx context.Context, b *bot.Bot, fileID string) ([]byte, string, error) {
	file, err := b.GetFile(ctx, &bot.GetFileParams{FileID: fileID})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get file info: %w", err)
	}

	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", sb.config.TelegramToken, file.FilePath)

	resp, err := http.Get(fileURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to download file, status: %d", resp.StatusCode)
	}

	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file content: %w", err)
	}

	mimeType := "image/jpeg"
	if len(imageBytes) > 0 {
		if len(imageBytes) > 3 && imageBytes[0] == 0x89 && imageBytes[1] == 0x50 && imageBytes[2] == 0x4E {
			mimeType = "image/png"
		}
	}

	log.Printf("📁 File downloaded: %d bytes, type: %s", len(imageBytes), mimeType)
	return imageBytes, mimeType, nil
}

func (sb *StoolGuruBot) sendErrorMessage(ctx context.Context, b *bot.Bot, chatID int64, errorText string) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   fmt.Sprintf("❌ %s\n\nПопробуйте еще раз или обратитесь в поддержку.", errorText),
	})
	if err != nil {
		log.Printf("Error sending error message: %v", err)
	}
}

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

	text := "🤖 Stool Guru Bot запущен\n\nПривет! Я готов помочь вам с анализом здоровья.\n\n📸 **Просто отправьте мне фото для анализа!**\n\nИли выберите действие в меню ниже:"

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        text,
		ReplyMarkup: keyboard,
		ParseMode:   models.ParseModeMarkdown,
	})
	if err != nil {
		log.Printf("Error sending start message: %v", err)
	}
}

func (sb *StoolGuruBot) handleHelp(ctx context.Context, b *bot.Bot, update *models.Update) {
	text := `🆘 **Как пользоваться ботом:**

📸 **Отправьте фото** - бот автоматически проанализирует изображение

📋 **Команды:**
/start • Главное меню
/help • Эта справка  
/test • Тест функций
/analyze • Ручной анализ

🔬 Бот использует современный ИИ для анализа и дает рекомендации как опытный врач!`

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		log.Printf("Error sending help message: %v", err)
	}
}

func (sb *StoolGuruBot) handleTest(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println("🧪 Test command received")
	testResult, err := sb.geminiService.SendTextMessage(ctx, "Привет! Это тест подключения к Gemini.")
	if err != nil {
		log.Printf("Gemini test failed: %v", err)
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "❌ Тест не прошел! Проблема с Gemini API.",
		})
	} else {
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("✅ Тест прошел!\n\n🤖 Gemini ответил: %s", testResult.Text),
		})
	}

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