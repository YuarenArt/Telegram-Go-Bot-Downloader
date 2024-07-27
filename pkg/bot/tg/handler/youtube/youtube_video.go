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
	formats, err := youtube_downloader.FormatWithAudioChannelsComposite(videoURL)
	if err != nil {
		log.Printf("FormatWithAudioChannels return %s", err)
		return nil, err
	}

	keyboard, err := getKeyboardVideoFormats(formats, videoURL)
	if err != nil {
		log.Printf("GetKeyboard return %s", err)
		return nil, err
	}
	return keyboard, nil
}
