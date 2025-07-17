package router

import (
	"log"

	"github.com/go-telegram/bot"
	"github.com/merdernoty/stool-guru-bot/internal/bot/handlers/callbacks"
	"github.com/merdernoty/stool-guru-bot/internal/bot/handlers/commands"
)

type Router struct {
	startHandler *commands.StartHandler
	helpHandler  *commands.HelpHandler
	testHandler  *commands.TestHandler

	// Callback handlers
	callbackHandlers *callbacks.CallbackHandlers
}

func NewRouter(
	startHandler *commands.StartHandler,
	helpHandler *commands.HelpHandler,
	testHandler *commands.TestHandler,
	callbackHandlers *callbacks.CallbackHandlers,
) *Router {
	return &Router{
		startHandler:     startHandler,
		helpHandler:      helpHandler,
		testHandler:      testHandler,
		callbackHandlers: callbackHandlers,
	}
}

func (r *Router) RegisterHandlers(b *bot.Bot) {
	log.Println("📝 Registering handlers...")

	r.registerCommands(b)

	r.registerCallbacks(b)

	log.Println("✅ All handlers registered successfully")
}

func (r *Router) registerCommands(b *bot.Bot) {
	commands := []commands.CommandHandler{
		r.startHandler,
		r.helpHandler,
		r.testHandler,
	}

	for _, cmd := range commands {
		b.RegisterHandler(
			commands.HandlerTypeCommand,
			cmd.GetPattern(),
			cmd.GetMatchType(),
			cmd.Handle,
		)
		log.Printf("🔗 Registered command: %s", cmd.GetPattern())
	}
}

func (r *Router) registerCallbacks(b *bot.Bot) {
	callbackPatterns := r.callbackHandlers.GetCallbackPatterns()

	for pattern, handler := range callbackPatterns {
		b.RegisterHandler(
			bot.HandlerTypeCallbackQueryData,
			pattern,
			bot.MatchTypeExact,
			handler,
		)
		log.Printf("🔗 Registered callback: %s", pattern)
	}
}