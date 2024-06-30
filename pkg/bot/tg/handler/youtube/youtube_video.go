package youtube

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	youtube_downloader "youtube_downloader/pkg/downloader/youtube"
)

// handleYoutubeVideo gets all possible formats of the video by a link
// creates a keyboard and return it
func (yh *YoutubeHandler) handleYoutubeVideo(message *tgbotapi.Message) (*tgbotapi.InlineKeyboardMarkup, error) {
	videoURL := message.Text
	formats, err := youtube_downloader.FormatWithAudioChannels(videoURL)
	if err != nil {
		log.Printf("FormatWithAudioChannels return %w", err)
		return nil, err
	}

	keyboard, err := getKeyboardVideoFormats(formats)
	if err != nil {
		log.Printf("GetKeyboard return %w", err)
		return nil, err
	}
	return &keyboard, nil
}
