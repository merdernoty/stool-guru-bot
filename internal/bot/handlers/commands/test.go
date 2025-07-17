package commands

import (
	"context"
	"fmt"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/merdernoty/stool-guru-bot/internal/bot/services/gemini"
)

type TestHandler struct {
	geminiService *gemini.GeminiService
}

func NewTestHandler(geminiService *gemini.GeminiService) *TestHandler {
	return &TestHandler{
		geminiService: geminiService,
	}
}

func (h *TestHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Printf("🧪 Test command received from @%s", update.Message.From.Username)

	loadingMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "🧪 Тестирую подключение к AI сервисам...",
	})
	if err != nil {
		log.Printf("Error sending loading message: %v", err)
		return
	}

	testResult, err := h.geminiService.SendTextMessage(ctx, "Привет! Это тест подключения к Gemini.")
	
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    update.Message.Chat.ID,
		MessageID: loadingMsg.ID,
	})

	var responseText string
	if err != nil {
		log.Printf("Gemini test failed: %v", err)
		responseText = `❌ **Тест не прошел!**

🔴 Проблема с Gemini API:
` + "`" + err.Error() + "`" + `

🔧 **Возможные причины:**
• Неверный API ключ
• Превышен лимит запросов
• Проблемы с сетью

💡 **Решение:** Проверьте настройки в .env файле`
	} else {
		responseText = fmt.Sprintf(`✅ **Тест прошел успешно!**

🤖 **Gemini AI ответил:**
%s

🔧 **Состояние системы:**
• ✅ Telegram Bot API - работает
• ✅ Gemini AI API - работает
• ✅ HTTP клиент - работает

🚀 Бот готов к анализу фотографий!`, testResult.Text)
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      responseText,
		ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		log.Printf("Error sending test result: %v", err)
	}
}

func (h *TestHandler) GetPattern() string {
	return "/test"
}
	
func (h *TestHandler) GetMatchType() bot.MatchType {
	return bot.MatchTypeExact
}