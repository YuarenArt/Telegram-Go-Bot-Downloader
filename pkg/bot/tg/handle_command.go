package tg

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"youtube_downloader/pkg/bot/tg/send"
)

const (
	commandStart = "start"
	commandHelp  = "help"

	startMessage = "🤖 I'm working! 🤖"
	helpMessage  = "I can do the following things:\n\n" +
		" Download videos from YouTube\n" +
		" Download audio from YouTube\n" +
		" Convert videos to audio\n\n" +
		"Just send me a link to the video or audio you want to download."
	defaultMessage = "🤔 I don't know this command. 🤔"
)

// handleCommand handles supported commands
func (tb *TgBot) handleCommand(message *tgbotapi.Message) error {
	switch message.Command() {
	case commandStart:
		return tb.handleStartCommand(message)
	case commandHelp:
		return tb.handleHelpCommand(message)
	default:
		return tb.handleDefaultCommand(message)
	}
}

// handleStartCommand sends a message with startMessage text
func (tb *TgBot) handleStartCommand(message *tgbotapi.Message) error {
	return send.SendMessage(tb.Bot, message, startMessage)
}

// handleStartCommand sends a message with helpMessage text
func (tb *TgBot) handleHelpCommand(message *tgbotapi.Message) error {
	return send.SendMessage(tb.Bot, message, helpMessage)
}

// handleStartCommand sends a message with defaultMessage text
func (tb *TgBot) handleDefaultCommand(message *tgbotapi.Message) error {
	return send.SendMessage(tb.Bot, message, defaultMessage)
}
