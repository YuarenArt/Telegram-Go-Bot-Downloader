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

const (
	All_video = "allVideo"
	All_audio = "allAudio"
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
	var videoURL string
	for _, part := range parts {
		if strings.HasPrefix(part, "https://") {
			videoURL = part
			break
		}
	}

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

// handleCallbackQueryWithPlaylist gets link on playlist by callbackQuery.Message.Text
// checks callbackQuery.Data
// if callbackQuery.Data == All_audio : download all videos from playlist in audio format
// if callbackQuery.Data == All_video : download all videos from playlist in video format
// else download a certain video by callbackQuery.Data(playlistEntry.ID)
func (tb *TgBot) handleCallbackQueryWithPlaylist(callbackQuery *tgbotapi.CallbackQuery) {

	text := callbackQuery.Message.Text
	parts := strings.Split(text, "\n")
	var playlistURL string
	for _, part := range parts {
		if strings.HasPrefix(part, "https://") {
			playlistURL = part
			break
		}
	}

	downloader := youtube_downloader.NewYouTubeDownloader()
	playlist, err := downloader.Downloader.Client.GetPlaylist(playlistURL)
	if err != nil {
		log.Printf("GetPlaylist in handleCallbackQueryWithPlaylist error: %w", err)
	}

	switch {
	case callbackQuery.Data == All_audio:
		var video *youtube.Video
		for _, playlistEntry := range playlist.Videos {
			video, err = downloader.Downloader.Client.VideoFromPlaylistEntry(playlistEntry)

			path, err := tb.downloadAudio(video)
			if err != nil {
				log.Printf("downloadVideo in handleCallbackQueryWithPlaylist error: %w", err)
			}

			err = tb.sendFile(callbackQuery.Message, path)
			if err != nil {
				log.Printf("sendFile in handleCallbackQueryWithPlaylist error: %s", err)
				tb.sendReplyMessage(callbackQuery.Message, "File Too Large: max files size is "+strconv.Itoa(maxFileSize/(1024*1024))+" Mb")
			}
		}

	case callbackQuery.Data == All_video:

		var video *youtube.Video
		for _, playlistEntry := range playlist.Videos {
			video, err = downloader.Downloader.Client.VideoFromPlaylistEntry(playlistEntry)

			path, err := tb.downloadVideo(video)
			if err != nil {
				log.Printf("downloadVideo in handleCallbackQueryWithPlaylist error: %w", err)
			}

			err = tb.sendFile(callbackQuery.Message, path)
			if err != nil {
				log.Printf("sendFile in handleCallbackQueryWithPlaylist error: %w", err)
				tb.sendReplyMessage(callbackQuery.Message, "File Too Large: max files size is "+strconv.Itoa(maxFileSize))
			}

			err = deleteFile(path)
			if err != nil {
				log.Printf("deleteFile return %w in handleCallbackQuery", err)
			}
		}

	default:

		var video *youtube.Video
		for _, playlistEntry := range playlist.Videos {
			if playlistEntry.ID == callbackQuery.Data {
				video, err = downloader.Downloader.Client.VideoFromPlaylistEntry(playlistEntry)
				break
			}
		}

		path, err := tb.downloadVideo(video)
		if err != nil {
			log.Printf("downloadVideo in handleCallbackQueryWithPlaylist error: %w", err)
		}

		err = tb.sendFile(callbackQuery.Message, path)
		if err != nil {
			log.Printf("sendFile in handleCallbackQueryWithPlaylist error: %w", err)
			tb.sendReplyMessage(callbackQuery.Message, "Request Entity Too Large")
		}

		// deletes file after sending
		defer func() {
			err := deleteFile(path)
			if err != nil {
				log.Printf("deleteFile return %w in handleCallbackQuery", err)
			}
		}()

	}
}

//func (tb *TgBot)

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

func getKeyboardPlaylist(playlist *youtube.Playlist) tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup()

	button := tgbotapi.NewInlineKeyboardButtonData(
		fmt.Sprintf("%s", "Download all: video"), All_video)
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{button})

	button = tgbotapi.NewInlineKeyboardButtonData(
		fmt.Sprintf("%s", "Download all: audio"), All_audio)
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{button})

	for _, playlistEntry := range playlist.Videos {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s", playlistEntry.Title), playlistEntry.ID)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{button})
	}

	return keyboard
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
