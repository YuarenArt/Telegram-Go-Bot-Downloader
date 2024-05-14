package tg_bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
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

	downloadingNotification = "⏳ Downloading... ⏳"
	sendingNotification     = " 🚀 Sending... 🚀"
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

	if strings.HasPrefix(message.Text, "https://www.youtube.com/") {
		err := b.handleYoutubeLink(message)
		if err != nil {
			log.Print("handleYoutubeLink return error!")
		}
		return
	}

	b.handleDefaultCommand(message)
	b.handleHelpCommand(message)

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
	return b.sendMessage(message, startMessage)
}

func (b *TgBot) handleHelpCommand(message *tgbotapi.Message) error {
	return b.sendMessage(message, helpMessage)
}

func (b *TgBot) handleDefaultCommand(message *tgbotapi.Message) error {
	return b.sendMessage(message, defaultMessage)
}

func (b *TgBot) handleYoutubeLink(message *tgbotapi.Message) error {

	// Send downloading notification
	resp, err := b.sendReplyMessage(message, downloadingNotification)
	if err != nil {
		return err
	}

	// Download the video and get the file path
	videoURL := message.Text
	videoPath, err := b.downloadVideo(videoURL)
	if err != nil {
		return err
	}

	// Send sending notification
	err = b.sendEditMessage(message.Chat.ID, resp.MessageID, sendingNotification)
	if err != nil {
		return err
	}

	// Send the video
	err = b.sendVideo(message.Chat.ID, resp.MessageID, videoPath)
	if err != nil {
		return err
	}

	return nil
}
