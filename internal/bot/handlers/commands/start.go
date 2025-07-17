package commands

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type StartHandler struct {
	BaseHandler
}

func NewStartHandler() *StartHandler {
	return &StartHandler{
		BaseHandler: NewBaseHandler("/start", bot.MatchTypeExact),
	}
}

func (h *StartHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	username := "Anonymous"
	if update.Message.From.Username != "" {
		username = "@" + update.Message.From.Username
	}
	
	log.Printf("📋 Start command received from %s", username)
	
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

	text := `🤖 <b>Stool Guru Bot запущен</b>

Привет! Я готов помочь вам с анализом здоровья.

📸 <b>Просто отправьте мне фото для анализа!</b>

Или выберите действие в меню ниже:`

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        text,
		ReplyMarkup: keyboard,
		ParseMode:   models.ParseModeHTML,
	})
	if err != nil {
		log.Printf("Error sending start message: %v", err)
		sendErrorMessage(ctx, b, update.Message.Chat.ID, "Ошибка отправки приветственного сообщения")
	}
}