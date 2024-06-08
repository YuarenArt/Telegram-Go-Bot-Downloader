package tg_bot

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kkdai/youtube/v2"
	"log"
	"net/url"
	"os"
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

	if strings.HasPrefix(message.Text, "https://www.youtube.com/") {

		err := tb.handleYoutubeLink(message)
		if err != nil {
			log.Print(err)
			errMsg := err.Error()
			if errMsg == "Request Entity Too Large" {
				tb.sendReplyMessage(message, "Your file too large")
			} else {
				tb.sendReplyMessage(message, "Something went wrong")
			}
		}
		return
	}

	tb.handleDefaultCommand(message)
	tb.handleHelpCommand(message)
}

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

// checks the link type and calls the appropriate method
func (tb *TgBot) handleYoutubeLink(message *tgbotapi.Message) error {

	videoURL := message.Text
	switch {
	case strings.HasPrefix(videoURL, "https://www.youtube.com/live/"):
		return tb.handleYoutubeStream(message)
	case strings.HasPrefix(videoURL, "https://youtube.com/playlist?"):
	default:
		return tb.handleYoutubeVideo(message)
	}

	return errors.New("out of switch body in handleYoutubeLink method")
}

// handleYoutubeVideo gets all possible formats of the video by a link
// creates a keyboard and sends it to user's chat
func (tb *TgBot) handleYoutubeVideo(message *tgbotapi.Message) error {
	videoURL := message.Text
	formats, err := youtube_downloader.FormatWithAudioChannels(videoURL)
	if err != nil {
		log.Printf("FormatWithAudioChannels return %w", err)
		return err
	}

	keyboard, err := getKeyboard(formats)
	if err != nil {
		log.Printf("GetKeyboard return %w", err)
		return err
	}

	err = tb.sendKeyboardMessage(message, keyboard)
	if err != nil {
		log.Printf("sendKeyboardMessage %w", err)
		return err
	}

	return nil
}

func (tb *TgBot) handleYoutubeStream(message *tgbotapi.Message) error {

	videoURLWithLivePrefix := message.Text
	videoURL := formatYouTubeURLOnStream(videoURLWithLivePrefix)
	formats, err := youtube_downloader.FormatWithAudioChannels(videoURL)
	if err != nil {
		log.Printf("FormatWithAudioChannels return %w", err)
		return err
	}

	keyboard, err := getKeyboard(formats)
	if err != nil {
		log.Printf("GetKeyboard return %w", err)
		return err
	}

	err = tb.sendKeyboardMessageWithFormattedLink(message, keyboard, videoURL)
	if err != nil {
		log.Printf("sendKeyboardMessage %w", err)
		return err
	}

	return nil

}

// handleCallbackQuery gets link on video by callbackQuery.Message.Text,
// gets ItagNo by callbackQuery.Data to find correct format,
// gets possible formats by videoURL,
// and finally gets the format selected by the user.
// then download it with format
func (tb *TgBot) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {

	text := callbackQuery.Message.Text
	parts := strings.Split(text, "\n")
	videoURL := parts[1]

	formats, err := youtube_downloader.FormatWithAudioChannels(videoURL)
	if err != nil {
		log.Printf("FormatWithAudioChannels return %w in handleCallbackQuery", err)
	}

	// gets format by its TagNo
	tagNo, err := strconv.Atoi(callbackQuery.Data)
	if err != nil {
		tb.sendReplyMessage(callbackQuery.Message, "Error! Try others formats, sorry (")
		return
	}

	var formatFile youtube.Format
	for _, format := range formats {
		if format.ItagNo == tagNo {
			formatFile = format
			break
		}
	}

	// start downloading
	resp, err := tb.sendReplyMessage(callbackQuery.Message, downloadingNotification)
	if err != nil {
		log.Printf("can't send reply message: %w", err.Error())
	}
	pathAndName, err := tb.downloadWithFormat(videoURL, formatFile)
	if err != nil {
		log.Printf(err.Error())
		tb.sendEditMessage(resp.Chat.ID, resp.MessageID, err.Error())
		return
	}
	// start sending
	err = tb.sendEditMessage(resp.Chat.ID, resp.MessageID, sendingNotification)
	if err != nil {
		log.Printf("can't send edit message: %w", err.Error())
	}

	err = tb.sendFile(callbackQuery.Message, pathAndName)
	if err != nil {
		log.Printf("sendFile return %w in handleCallbackQuery", err)
	}

	// deletes file after sending
	defer func() {
		err := deleteFile(pathAndName)
		if err != nil {
			log.Printf("deleteFile return %w in handleCallbackQuery", err)
		}
	}()
}

func deleteFile(pathToFile string) error {
	return os.Remove(pathToFile)
}

// getKeyboard return InlineKeyboardMarkup by all possible video formats
func getKeyboard(formats youtube.FormatList) (tgbotapi.InlineKeyboardMarkup, error) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup()

	for _, format := range formats {

		mimeType := format.MimeType
		videoFormat := strings.Split(mimeType, ";")[0]
		qualityLabel := format.QualityLabel
		data := strconv.Itoa(format.ItagNo) // to download video in correct format

		size, err := getFileSize(format)
		size = size / (1024 * 1024)
		if err != nil {
			return keyboard, err
		}

		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s, %s, %s Mb", videoFormat, qualityLabel, strconv.FormatFloat(size, 'f', 2, 64)),
			data)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{button})
	}

	return keyboard, nil
}

// getFileSize return a file size in bite of certain format
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

// isAcceptableFileSize return true if file size less than possible size to send to tg API
func isAcceptableFileSize(format youtube.Format) bool {
	fileSize, _ := getFileSize(format)
	return fileSize < maxFileSize
}

// formatYouTubeURLonStream instead of live/ links return link on video
func formatYouTubeURLOnStream(inputURL string) string {
	u, err := url.Parse(inputURL)
	if err != nil {
		return inputURL
	}

	parts := strings.Split(u.Path, "/")
	if len(parts) < 2 || parts[1] != "live" {
		return inputURL
	}

	videoID := parts[2]
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
}
