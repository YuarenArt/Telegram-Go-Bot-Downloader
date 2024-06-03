package tg_bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

// TgBot uses telegram-bot-api to maintain tg bot
// It can download and send video with different formats (video/audio; quality)
type TgBot struct {
	bot *tgbotapi.BotAPI
}

// NewBot initializes a new TgBot instance with the given Telegram Bot API instance.
func NewBot(bot *tgbotapi.BotAPI) *TgBot {
	return &TgBot{bot: bot}
}

// StartBot starts the bot by authorizing it and initiating the update handling process.
func (tb *TgBot) StartBot() error {
	log.Printf("Authorized on account %s", tb.bot.Self.UserName)

	updates := tb.initUpdatesChannel()
	tb.handleUpdates(updates)

	return nil
}

// initUpdatesChannel initializes the updates channel for receiving updates from the Telegram server.
// It configures the update retrieval settings and returns the updates channel.
func (tb *TgBot) initUpdatesChannel() tgbotapi.UpdatesChannel {
	update := tgbotapi.NewUpdate(0)
	update.Timeout = 60

	return tb.bot.GetUpdatesChan(update)
}
