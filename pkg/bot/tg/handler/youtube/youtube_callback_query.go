package youtube

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kkdai/youtube/v2"
	"log"
	"os"
	"strconv"
	"strings"
	"youtube_downloader/pkg/bot/tg/send"
	youtube_downloader "youtube_downloader/pkg/downloader/youtube"
)

// TODO rework a way to get data for downloading

// HandleCallbackQuery gets url from Bot's message with a replying link,
// then handle a link by its type: video (stream), playlist
func (yh *YoutubeHandler) HandleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) {

	// gets URL from a data
	text := callbackQuery.Data
	parts := strings.Split(text, ",")
	URL := parts[0]

	switch {
	// TODO fix that need to obtainn link for handling playlist Button
	case strings.HasPrefix(URL, "https://youtube.com/playlist?") || URL == "https://youtu.be":
		yh.HandleCallbackQueryWithPlaylist(callbackQuery, bot)
	default:
		yh.HandleCallbackQueryWithFormats(callbackQuery, bot)
	}
}

// HandleCallbackQueryWithFormats gets a link on video by callbackQuery.Message.Text,
// gets ItagNo by callbackQuery.Data to find a correct format,
// gets possible formats by videoURL,
// and finally gets the format selected by the user.
// then download it with format
func (yh *YoutubeHandler) HandleCallbackQueryWithFormats(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) {

	data := callbackQuery.Data
	dataParts := strings.Split(data, ",")
	videoURL := dataParts[0]

	formats, err := youtube_downloader.FormatWithAudioChannels(videoURL)
	if err != nil {
		log.Printf("FormatWithAudioChannels return %s in handleCallbackQuery", err)
	}

	// gets format by its TagNo
	tagNo, err := strconv.Atoi(dataParts[1])
	if err != nil {
		send.SendReplyMessage(bot, callbackQuery.Message, "Error! Try others formats, sorry (")
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
	resp, err := send.SendReplyMessage(bot, callbackQuery.Message, send.DownloadingNotification)
	if err != nil {
		log.Printf("can't send reply message: %s", err.Error())
	}

	dl := youtube_downloader.NewYouTubeDownloader()
	pathAndName, err := dl.DownloadWithFormat(videoURL, formatFile)

	if err != nil {
		log.Printf(err.Error())
		send.SendEditMessage(bot, resp.Chat.ID, resp.MessageID, "Error! Try others formats, sorry (")
		return
	}
	// start sending
	err = send.SendEditMessage(bot, resp.Chat.ID, resp.MessageID, send.SendingNotification)
	if err != nil {
		log.Printf("can't send edit message: %s", err.Error())
	}

	// deletes file after sending
	defer func() {
		err := deleteFile(pathAndName)
		if err != nil {
			log.Printf("deleteFile return %s in handleCallbackQuery", err)
		}
	}()

	err = send.SendFile(bot, callbackQuery.Message, pathAndName)
	if err != nil {
		send.SendEditMessage(bot, resp.Chat.ID, resp.MessageID, "I can't send the file. Sorry, something went wrong. Please, try others format")
		log.Printf("sendFile return %s in handleCallbackQuery", err)
	}

}

// HandleCallbackQueryWithPlaylist gets link on playlist by callbackQuery.Message.Text
// checks callbackQuery.Data
// if callbackQuery.Data include All_audio : download all videos from playlist in audio format
// if callbackQuery.Data include All_video : download all videos from playlist in video format
// else download a certain video by callbackQuery.Data
func (yh *YoutubeHandler) HandleCallbackQueryWithPlaylist(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI) {
	lines := strings.Split(callbackQuery.Message.Text, "\n") // split the string into lines
	var playlistURL string
	for _, line := range lines {
		if strings.HasPrefix(line, "https://") {
			playlistURL = line
			break
		}
	}
	downloader := youtube_downloader.NewYouTubeDownloader()

	playlist, err := downloader.GetPlaylist(playlistURL)
	if err != nil {
		log.Printf("GetPlaylist in handleCallbackQueryWithPlaylist error: %v", err)
		return
	}
	data := callbackQuery.Data
	dataParts := strings.Split(data, ",")

	switch {
	case dataParts[1] == All_audio:
		yh.processPlaylistAudio(bot, callbackQuery, playlist)
	case dataParts[1] == All_video:
		yh.processPlaylistVideo(bot, callbackQuery, playlist)
	default:
		yh.processSingleVideo(bot, callbackQuery, playlist)
	}
}

func deleteFile(pathToFile string) error {
	return os.Remove(pathToFile)
}

func extractPlaylistURL(text string) string {
	parts := strings.Split(text, "\n")
	for _, part := range parts {
		if strings.HasPrefix(part, "https://") {
			return part
		}
	}
	return ""
}
