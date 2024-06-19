package tg_bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

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

// handleStartCommand sends message with startMessage text
func (tb *TgBot) handleStartCommand(message *tgbotapi.Message) error {
	return tb.sendMessage(message, startMessage)
}

// handleStartCommand sends message with helpMessage text
func (tb *TgBot) handleHelpCommand(message *tgbotapi.Message) error {
	return tb.sendMessage(message, helpMessage)
}

// handleStartCommand sends message with defaultMessage text
func (tb *TgBot) handleDefaultCommand(message *tgbotapi.Message) error {
	return tb.sendMessage(message, defaultMessage)
}
