package commands

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type StartHandler struct{}

func NewStartHandler() *StartHandler {
	return &StartHandler{}
}

func (h *StartHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Printf("📋 Start command received from @%s", update.Message.From.Username)

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

	text := `🤖 Stool Guru Bot запущен

Привет! Я готов помочь вам с анализом здоровья.

📸 **Просто отправьте мне фото для анализа!**

Или выберите действие в меню ниже:`

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

func (h *StartHandler) GetPattern() string {
	return "/start"
}

func (h *StartHandler) GetMatchType() bot.MatchType {
	return bot.MatchTypeExact
}