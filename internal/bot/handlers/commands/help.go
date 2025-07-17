package commands

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type HelpHandler struct {
	BaseHandler
}

func NewHelpHandler() *HelpHandler {
	return &HelpHandler{
		BaseHandler: NewBaseHandler("/help", bot.MatchTypeExact),
	}
}

func (h *HelpHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Printf("❓ Help command received from %s", update.Message.From.Username)

	text := `🆘 <b>Как пользоваться ботом:</b>

📸 <b>Отправьте фото</b> - бот автоматически проанализирует изображение

📋 <b>Команды:</b>
/start • Главное меню
/help • Эта справка  
/test • Тест функций
/analyze • Ручной анализ

🔬 Бот использует современный ИИ для анализа и дает рекомендации как опытный врач!

💡 <b>Совет:</b> Для лучшего анализа убедитесь, что фото четкое и хорошо освещенное.`

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		log.Printf("Error sending help message: %v", err)
		sendErrorMessage(ctx, b, update.Message.Chat.ID, "Ошибка отправки справки")
	}
}