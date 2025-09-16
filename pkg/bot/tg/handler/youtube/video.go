package youtube

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleYoutubeVideo gets all possible formats of the video by a link,
// creates a keyboard and return it
func (yh *YoutubeHandler) handleYoutubeVideo(message *tgbotapi.Message) (*tgbotapi.InlineKeyboardMarkup, error) {
	videoURL := message.Text
	video, err := yh.Downloader.GetVideo(context.Background(), videoURL)
	if err != nil {
		log.Printf("GetVideo return %s", err)
		return nil, err
	}
	keyboard, err := getKeyboardVideoFormats(video.Formats, &videoURL)
	if err != nil {
		log.Printf("GetKeyboard return %s", err)
		return nil, err
	}
	return keyboard, nil
}
