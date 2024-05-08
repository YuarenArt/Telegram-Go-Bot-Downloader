package tg_bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

const (
	commandStart = "start"
	commandHelp  = "help"

	startMessage   = "I' am working"
	helpMessage    = "I can download video and audio from youtube, just send a link"
	defaultMessage = "I don't now this command"
)

func (b *TgBot) handleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message.IsCommand() {
			b.handleCommand(update.Message)
			continue
		}

		b.handleMessage(update.Message)
	}
}

func (b *TgBot) handleMessage(message *tgbotapi.Message) {
	log.Printf("[%s] %s", message.From.UserName, message.Text)

	if !strings.HasPrefix(message.Text, "https://www.youtube.com/") {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Please send a YouTube link")
		msg.ReplyToMessageID = message.MessageID

		b.bot.Send(msg)
		return
	}

	videoURL := message.Text
	downloadVideo(videoURL)
}

func (b *TgBot) handleCommand(message *tgbotapi.Message) error {
	switch message.Command() {
	case commandStart:
		return b.handleStartCommand(message)

	case commandHelp:
		return b.handleHelpCommand(message)

	default:
		return b.handleDefaultCommand(message)
	}
}

func (b *TgBot) handleStartCommand(message *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, startMessage)
	_, err := b.bot.Send(msg)
	return err
}

func (b *TgBot) handleHelpCommand(message *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, helpMessage)
	_, err := b.bot.Send(msg)
	return err
}

func (b *TgBot) handleDefaultCommand(message *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, defaultMessage)
	_, err := b.bot.Send(msg)
	return err
}
