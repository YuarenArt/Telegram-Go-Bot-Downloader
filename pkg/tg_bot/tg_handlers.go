package tg_bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kkdai/youtube/v2"
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

	maxFileSize = 52428800.0 // in bites (50 Mb)
)

// TODO сделать многопоточную раюоту так чтобы после того видео загрузилось
// и начло отправляться можно было скачивать и отправлять нвоое

// handleUpdates gets updates from telegramAPI and handle it
func (tb *TgBot) handleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				tb.handleCommand(update.Message)
				continue
			}
			tb.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			tb.handleCallbackQuery(update.CallbackQuery)
		}
	}
}

// handleMessage processes an incoming message.
// If the message contains a command, it handles the command.
// If the message contains a link, it handles the link.
// Otherwise, it handles default and help commands.
func (tb *TgBot) handleMessage(message *tgbotapi.Message) {
	log.Printf("[%s] %s", message.From.UserName, message.Text)

	// handle youtube link
	if isYoutubeLink(message.Text) {
		err := tb.handleYoutubeLink(message)
		if err != nil {
			log.Print(err)
			errMsg := err.Error()
			if errMsg == "Request Entity Too Large" {
				tb.sendReplyMessage(message, "Your file too large")
			} else if errMsg == "extractVideoID failed: invalid characters in video id" {
				tb.sendReplyMessage(message, "Your link incorrect. Just send a link")
			} else {
				tb.sendReplyMessage(message, "Something went wrong")
			}
		}
		return
	}

	tb.handleDefaultCommand(message)
	tb.handleHelpCommand(message)
}

// checks the link type and calls the appropriate method
func (tb *TgBot) handleYoutubeLink(message *tgbotapi.Message) error {

	videoURL := message.Text
	switch {
	case strings.HasPrefix(videoURL, "https://www.youtube.com/live/"):
		return tb.handleYoutubeStream(message)
	case strings.HasPrefix(videoURL, "https://youtube.com/playlist?"):
		return tb.handleYoutubePlaylist(message)
	default:
		return tb.handleYoutubeVideo(message)
	}
}

// isAcceptableFileSize return true if file size less than possible size to send to tg API
func isAcceptableFileSize(format youtube.Format) bool {
	fileSize, _ := getFileSize(format)
	return fileSize < maxFileSize
}

func isYoutubeLink(link string) bool {
	return strings.HasPrefix(link, "https://www.youtube.com/") ||
		strings.HasPrefix(link, "https://youtube.com/playlist") ||
		strings.HasPrefix(link, "https://youtu.be")
}
