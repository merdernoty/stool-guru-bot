package commands

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type CommandHandler interface {
	Handle(ctx context.Context, b *bot.Bot, update *models.Update)
	GetPattern() string
	GetMatchType() bot.MatchType
}

const (
	HandlerTypeCommand = bot.HandlerTypeMessageText
)