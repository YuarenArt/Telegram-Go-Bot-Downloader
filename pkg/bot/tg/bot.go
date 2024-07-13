package tg

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"youtube_downloader/pkg/bot/tg/handler"
	_ "youtube_downloader/pkg/database-client"
	database_client "youtube_downloader/pkg/database-client"
)

// TgBot uses telegram-Bot-api to maintain tg Bot
// It can download and send video with different formats (video/audio; quality) by handlers
type TgBot struct {
	Bot      *tgbotapi.BotAPI
	handlers []handler.Handler
	Client   *database_client.Client
}

// NewBot initializes a new TgBot instance with the given Telegram Bot API instance.
func NewBot(bot *tgbotapi.BotAPI) *TgBot {
	return &TgBot{
		Bot:    bot,
		Client: database_client.NewClient(),
	}
}

// StartBot starts the Bot by authorizing it and initiating the update handling process.
func (tb *TgBot) StartBot() error {
	log.Printf("Authorized on account %s", tb.Bot.Self.UserName)

	tb.initSupportedHandlers()

	updates := tb.initUpdatesChannel()
	tb.handleUpdates(updates)

	return nil
}

// initSupportedHandlers initializes all supported handlers for the Telegram bot
// according to SupportedHandlers
func (tb *TgBot) initSupportedHandlers() {
	for _, handlerType := range handler.SupportedHandlers {
		handler := handler.CreateHandler(handlerType)
		tb.registerHandler(&handler)
	}
}

// registerHandler registers a new handler to the TgBot
func (tb *TgBot) registerHandler(handler *handler.Handler) {
	tb.handlers = append(tb.handlers, *handler)
}

// initUpdatesChannel initializes the update channel for receiving updates from the Telegram server.
// It configures the update retrieval settings and returns the update channel.
func (tb *TgBot) initUpdatesChannel() tgbotapi.UpdatesChannel {
	update := tgbotapi.NewUpdate(0)
	update.Timeout = 60

	return tb.Bot.GetUpdatesChan(update)
}
