package tg_bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kkdai/youtube/v2"
	"log"
	"os"
	"strconv"
	"strings"
	youtube_downloader "youtube_downloader/pkg/downloader/youtube-downloader"
)

func (tb *TgBot) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {

	text := callbackQuery.Message.Text
	parts := strings.Split(text, "\n")
	URL := parts[1]

	switch {
	case strings.HasPrefix(URL, "https://youtube.com/playlist?"):
		tb.handleCallbackQueryWithPlaylist(callbackQuery)
	default:
		tb.handleCallbackQueryWithFormats(callbackQuery)
	}
}

// handleCallbackQuery gets link on video by callbackQuery.Message.Text,
// gets ItagNo by callbackQuery.Data to find correct format,
// gets possible formats by videoURL,
// and finally gets the format selected by the user.
// then download it with format
func (tb *TgBot) handleCallbackQueryWithFormats(callbackQuery *tgbotapi.CallbackQuery) {

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

func (tb *TgBot) handleCallbackQueryWithPlaylist(callbackQuery *tgbotapi.CallbackQuery) {

}

// getKeyboard return InlineKeyboardMarkup by all possible video formats
func getKeyboardVideoFormats(formats youtube.FormatList) (tgbotapi.InlineKeyboardMarkup, error) {
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

func deleteFile(pathToFile string) error {
	return os.Remove(pathToFile)
}
