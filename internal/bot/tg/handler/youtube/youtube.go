package youtube

import (
	"fmt"
	"strconv"
	"strings"
	"youtube_downloader/internal/downloader/youtube"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	All_video    = "allVideo"
	All_audio    = "allAudio"
	TrafficLimit = 5000.0 // Mb
)

// YoutubeHandler is a service for downloading video from youtube
// Теперь использует интерфейс Downloader
type YoutubeHandler struct {
	Downloader youtube.Downloader
}

// NewYoutubeHandler return new YoutubeHandler
func NewYoutubeHandler(downloader youtube.Downloader) *YoutubeHandler {
	return &YoutubeHandler{
		Downloader: downloader,
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
func getKeyboardVideoFormats(formats []any, url *string) (*tgbotapi.InlineKeyboardMarkup, error) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup()

	// getting the size of audio
	audioSize := 0.0
	for _, f := range formats {
		format := f.(map[string]any)
		if format["QualityLabel"].(string) == "" {
			size, err := getFileSizeGeneric(format)
			if err == nil {
				audioSize = size / (1024 * 1024)
				break
			}
		}
	}

	for _, f := range formats {
		format := f.(map[string]any)

		mimeType := format["MimeType"].(string)
		if strings.HasPrefix(mimeType, "audio/webm") || strings.HasPrefix(mimeType, "video/webm") {
			continue
		}

		videoFormat := strings.Split(mimeType, ";")[0]
		itagNo := format["ItagNo"].(int)
		data := fmt.Sprintf("%s,%d", *url, itagNo)

		size, err := getFileSizeGeneric(format)
		size = size / (1024 * 1024)

		if strings.HasPrefix(mimeType, "video") {
			size = size + audioSize
		}

		if err != nil {
			return &keyboard, err
		}

		sign := []string{videoFormat}
		if ql := format["QualityLabel"].(string); ql != "" {
			sign = append(sign, ql)
		}
		sign = append(sign, strconv.FormatFloat(size, 'f', 2, 64))

		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s Mb", strings.Join(sign, ", ")),
			data)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{button})
	}

	return &keyboard, nil
}

// getFileSizeGeneric return a file size in bite of certain format (map[string]any)
func getFileSizeGeneric(format map[string]any) (float64, error) {
	if cl, ok := format["ContentLength"]; ok && cl.(int) > 0 {
		return float64(cl.(int)), nil
	}

	duration, err := strconv.ParseFloat(fmt.Sprintf("%v", format["ApproxDurationMs"]), 64)
	if err != nil {
		return 0, err
	}
	duration /= 1000

	bitrate := format["Bitrate"].(int)
	if ab, ok := format["AverageBitrate"]; ok && ab.(int) > 0 {
		bitrate = ab.(int)
	}

	contentLength := float64(bitrate/8) * duration

	return contentLength, nil
}
