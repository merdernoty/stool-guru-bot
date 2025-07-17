// internal/bot/handlers/commands/interface.go
package commands

import (
	"context"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type CommandHandler interface {
	GetPattern() string
	GetMatchType() bot.MatchType
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
}

type BaseHandler struct {
	pattern   string
	matchType bot.MatchType
}

func NewBaseHandler(pattern string, matchType bot.MatchType) BaseHandler {
	return BaseHandler{
		pattern:   pattern,
		matchType: matchType,
	}
}

func (h *BaseHandler) GetPattern() string {
	return h.pattern
}

func (h *BaseHandler) GetMatchType() bot.MatchType {
	return h.matchType
}


func escapeMarkdownV2(text string) string {
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	
	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}

func sendErrorMessage(ctx context.Context, b *bot.Bot, chatID int64, errorText string) {
	text := "❌ " + errorText + "\n\nПопробуйте еще раз или обратитесь в поддержку."
	
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	if err != nil {
		// log.Printf("Error sending error message: %v", err)
	}
}