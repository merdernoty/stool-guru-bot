package callbacks

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type CallbackHandlers struct{}

func NewCallbackHandlers() *CallbackHandlers {
	return &CallbackHandlers{}
}

func (h *CallbackHandlers) HandleTestCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Printf("🧪 Test callback received from @%s", update.CallbackQuery.From.Username)

	_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "✅ Используйте команду /test для полного тестирования!",
		ShowAlert:       true,
	})
	if err != nil {
		log.Printf("Error answering test callback: %v", err)
	}
}

func (h *CallbackHandlers) HandleHelpCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Printf("❓ Help callback received from @%s", update.CallbackQuery.From.Username)

	_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "📋 Команды: /start /help /test /analyze",
		ShowAlert:       false,
	})
	if err != nil {
		log.Printf("Error answering help callback: %v", err)
	}
}

func (h *CallbackHandlers) HandleAnalyzeCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Printf("📊 Analyze callback received from @%s", update.CallbackQuery.From.Username)

	_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "📊 Выберите ваше состояние ниже",
		ShowAlert:       false,
	})
	if err != nil {
		log.Printf("Error answering analyze callback: %v", err)
	}
}

func (h *CallbackHandlers) HandleAnalyzeGoodCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Printf("🟢 Analyze good callback from @%s", update.CallbackQuery.From.Username)

	_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "🟢 Отлично! Продолжайте в том же духе! Пейте воду, ешьте клетчатку, двигайтесь!",
		ShowAlert:       true,
	})
	if err != nil {
		log.Printf("Error answering good callback: %v", err)
	}
}

func (h *CallbackHandlers) HandleAnalyzeNormalCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Printf("🟡 Analyze normal callback from @%s", update.CallbackQuery.From.Username)

	_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "🟡 Нормально! Советы: больше пробиотиков, овощей, прогулки после еды",
		ShowAlert:       true,
	})
	if err != nil {
		log.Printf("Error answering normal callback: %v", err)
	}
}

func (h *CallbackHandlers) HandleAnalyzeBadCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Printf("🔴 Analyze bad callback from @%s", update.CallbackQuery.From.Username)

	_, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            "🔴 При серьезных симптомах обратитесь к врачу! Пейте воду, избегайте острого",
		ShowAlert:       true,
	})
	if err != nil {
		log.Printf("Error answering bad callback: %v", err)
	}
}

func (h *CallbackHandlers) GetCallbackPatterns() map[string]func(context.Context, *bot.Bot, *models.Update) {
	return map[string]func(context.Context, *bot.Bot, *models.Update){
		"test":           h.HandleTestCallback,
		"help":           h.HandleHelpCallback,
		"analyze":        h.HandleAnalyzeCallback,
		"analyze_good":   h.HandleAnalyzeGoodCallback,
		"analyze_normal": h.HandleAnalyzeNormalCallback,
		"analyze_bad":    h.HandleAnalyzeBadCallback,
	}
}