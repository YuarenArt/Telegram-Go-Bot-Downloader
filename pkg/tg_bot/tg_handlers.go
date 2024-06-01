package tg_bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kkdai/youtube/v2"
	"log"
	"strconv"
	"strings"
	youtube_downloader "youtube_downloader/pkg/downloader/youtube-downloader"
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

// TODO если пользовательно нажемет кнопку остановить и блокировать то происходит panic
// handleUpdates gets updates from telegramAPI and handle it
func (b *TgBot) handleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				b.handleCommand(update.Message)
				continue
			}
			b.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			b.handleCallbackQuery(update.CallbackQuery)
		}
	}
}

// handleMessage checks the message
// If it is a command, handle command
// If it is a link, handle the link
// Unless, just handles default and help command
func (b *TgBot) handleMessage(message *tgbotapi.Message) {
	log.Printf("[%s] %s", message.From.UserName, message.Text)

	if strings.HasPrefix(message.Text, "https://www.youtube.com/") {
		err := b.handleYoutubeLink(message)
		if err != nil {
			log.Print(err)
			errMsg := err.Error()
			if errMsg == "Request Entity Too Large" {
				b.sendReplyMessage(message, "Your file too large")
			} else {
				b.sendReplyMessage(message, "Something went wrong")
			}
		}
		return
	}

	b.handleDefaultCommand(message)
	b.handleHelpCommand(message)

}

// handleCommand handles supported commands
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

// handleStartCommand sends message with startMessage text
func (b *TgBot) handleStartCommand(message *tgbotapi.Message) error {
	return b.sendMessage(message, startMessage)
}

// handleStartCommand sends message with helpMessage text
func (b *TgBot) handleHelpCommand(message *tgbotapi.Message) error {
	return b.sendMessage(message, helpMessage)
}

// handleStartCommand sends message with defaultMessage text
func (b *TgBot) handleDefaultCommand(message *tgbotapi.Message) error {
	return b.sendMessage(message, defaultMessage)
}

// handleYoutubeLink gets all possible formats of the video by link
// creates a keyboard and sends it to user's chat
func (b *TgBot) handleYoutubeLink(message *tgbotapi.Message) error {

	videoURL := message.Text
	formats, err := youtube_downloader.GetFormatWithAudioChannels(videoURL)
	if err != nil {
		log.Printf("GetFormatWithAudioChannels return %w", err)
		return err
	}

	keyboard, err := getKeyboard(formats)
	if err != nil {
		log.Printf("GetKeyboard return %w", err)
		return err
	}

	err = b.sendKeyboardMessage(message, keyboard)
	if err != nil {
		log.Printf("sendKeyboardMessage %w", err)
		return err
	}

	return nil
}

// handleCallbackQuery gets link on video by callbackQuery.Message.Text,
// gets mimeType by callbackQuery.Data,
// gets possible formats by videoURL,
// and finally gets the format selected by the user.
// then download it with format
func (b *TgBot) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {

	text := callbackQuery.Message.Text
	parts := strings.Split(text, "\n")
	videoURL := parts[1]

	formats, err := youtube_downloader.GetFormatWithAudioChannels(videoURL)
	if err != nil {
		log.Printf("GetFormatWithAudioChannels return %w in handleCallbackQuery", err)
	}

	mimeType := callbackQuery.Data
	var formatFile youtube.Format
	for _, format := range formats {
		if format.MimeType == mimeType {
			formatFile = format
			break
		}
	}
	pathAndName, err := b.downloadWithFormat(videoURL, formatFile)
	err = b.sendFile(callbackQuery.Message, pathAndName)
	if err != nil {
		log.Printf("sendFile return %w in handleCallbackQuery", err)
	}

	//TODO удалить файл после отправки

}

// getKeyboard return InlineKeyboardMarkup by all possible video formats
func getKeyboard(formats youtube.FormatList) (tgbotapi.InlineKeyboardMarkup, error) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup()

	for _, format := range formats {

		mimeType := format.MimeType
		videoFormat := strings.Split(mimeType, ";")[0]
		qualityLabel := format.QualityLabel

		size, err := getFileSize(format)
		size = size / (1024 * 1024)
		if err != nil {
			return keyboard, err
		}

		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s, %s, %s Mb", videoFormat, qualityLabel, strconv.FormatFloat(size, 'f', 2, 64)),
			mimeType)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{button})
	}

	return keyboard, nil
}

// getFileSize return a file size in bite
func getFileSize(format youtube.Format) (float64, error) {

	// get durations in secs
	duration, err := strconv.ParseFloat(format.ApproxDurationMs, 64)
	if err != nil {
		return 0, err
	}
	duration /= 1000

	// get bitrate in bite\sec
	bitrate := format.Bitrate

	// size in bite
	contentLength := float64(bitrate/8) * duration

	return contentLength, nil
}
