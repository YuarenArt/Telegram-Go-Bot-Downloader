package youtube

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kkdai/youtube/v2"
	"strconv"
	"strings"
	youtube_downloader "youtube_downloader/pkg/downloader/youtube"
)

const (
	All_video = "allVideo"
	All_audio = "allAudio"
)

// YoutubeHandler is a service for downloading video from youtube
type YoutubeHandler struct {
	Downloader youtube_downloader.YouTubeDownloader
}

// NewYoutubeHandler return new YoutubeHandler
func NewYoutubeHandler() *YoutubeHandler {
	downloader := youtube_downloader.NewYouTubeDownloader()
	return &YoutubeHandler{
		Downloader: *downloader,
	}
}

// HandleMessage handle YouTube link and return error
func (yh *YoutubeHandler) HandleMessage(message *tgbotapi.Message) (*tgbotapi.InlineKeyboardMarkup, error) {
	return yh.handleYoutubeLink(message)
}

// handleYoutubeLink checks the link type and calls the appropriate method
func (yh *YoutubeHandler) handleYoutubeLink(message *tgbotapi.Message) (*tgbotapi.InlineKeyboardMarkup, error) {

	videoURL := message.Text
	switch {
	case strings.HasPrefix(videoURL, "https://www.youtube.com/live/"):
		return yh.handleYoutubeStream(message)
	case strings.HasPrefix(videoURL, "https://youtube.com/playlist?"):
		return yh.handleYoutubePlaylist(message)
	default:
		return yh.handleYoutubeVideo(message)
	}
}

// getKeyboard return InlineKeyboardMarkup by all possible video formats. Button's data include video's url and ItagNo
func getKeyboardVideoFormats(formats youtube.FormatList, url string) (*tgbotapi.InlineKeyboardMarkup, error) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup()

	for _, format := range formats {

		mimeType := format.MimeType

		//ignore a .webm format
		if strings.HasPrefix(mimeType, "audio/webm") {
			continue
		}

		videoFormat := strings.Split(mimeType, ";")[0]
		qualityLabel := format.QualityLabel

		data := fmt.Sprintf("%s,%s", url, strconv.Itoa(format.ItagNo))

		size, err := getFileSize(format)
		size = size / (1024 * 1024)
		if err != nil {
			return &keyboard, err
		}

		sign := []string{videoFormat}
		if qualityLabel != "" {
			sign = append(sign, qualityLabel)
		}
		sign = append(sign, strconv.FormatFloat(size, 'f', 2, 64))

		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s Mb", strings.Join(sign, ", ")),
			data)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{button})
	}

	return &keyboard, nil
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
