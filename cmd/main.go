package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/labstack/echo/v4"
	"github.com/merdernoty/stool-guru-bot/internal/app"
)

func main() {
	fmt.Println("Starting Stool Guru Bot...")
	app.Start()

	botToken := os.Getenv("TELEGRAM_TOKEN")
	webhookURL := os.Getenv("WEBHOOK_URL") // например, https://yourdomain.com/bot

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	_, err = bot.Request(tgbotapi.NewWebhook(webhookURL))
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()

	e.POST("/bot", func(c echo.Context) error {
		update := tgbotapi.Update{}
		if err := c.Bind(&update); err != nil {
			return err
		}

		if update.Message != nil && update.Message.IsCommand() && update.Message.Command() == "start" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет!")
			// Добавим инлайн-кнопку
			button := tgbotapi.NewInlineKeyboardButtonData("Нажми меня", "button_clicked")
			keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(button))
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
		}

		if update.CallbackQuery != nil && update.CallbackQuery.Data == "button_clicked" {
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "Кнопка нажата!")
			bot.Request(callback)
		}

		return c.NoContent(http.StatusOK)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
