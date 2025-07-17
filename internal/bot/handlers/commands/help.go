package commands

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type HelpHandler struct{}

func NewHelpHandler() *HelpHandler {
	return &HelpHandler{}
}

func (h *HelpHandler) Handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Printf("🆘 Help command received from @%s", update.Message.From.Username)

	text := `🆘 **Как пользоваться ботом:**

📸 **Отправьте фото** - бот автоматически проанализирует изображение

📋 **Команды:**
/start • Главное меню
/help • Эта справка  
/test • Тест функций
/analyze • Ручной анализ

🔬 Бот использует современный ИИ для анализа и дает рекомендации как опытный врач!

⚠️ **Важно:** Результаты анализа не заменяют консультацию врача. При серьезных симптомах обратитесь к специалисту.`

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: models.ParseModeMarkdown,
	})
	if err != nil {
		log.Printf("Error sending help message: %v", err)
	}
}

func (h *HelpHandler) GetPattern() string {
	return "/help"
}

func (h *HelpHandler) GetMatchType() bot.MatchType {
	return bot.MatchTypeExact
}