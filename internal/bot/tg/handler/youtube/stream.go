package youtube

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/url"
	"strings"
	youtube_downloader "youtube_downloader/internal/downloader/youtube"
)

// handleYoutubeStream transforms live/ link into common video link
// creates a keyboard and return it
func (yh *YoutubeHandler) handleYoutubeStream(message *tgbotapi.Message) (*tgbotapi.InlineKeyboardMarkup, error) {

	videoURLWithLivePrefix := message.Text
	videoURL := FormatYouTubeURLOnStream(videoURLWithLivePrefix)
	formats, err := youtube_downloader.FormatWithAudioChannels(videoURL)
	if err != nil {
		log.Printf("FormatWithAudioChannels return %s", err)
		return nil, err
	}

	keyboard, err := getKeyboardVideoFormats(&formats, &videoURL)
	if err != nil {
		log.Printf("GetKeyboard return %s", err)
		return nil, err
	}

	return keyboard, nil
}

// FormatYouTubeURLOnStream instead of live/ links return link on video
func FormatYouTubeURLOnStream(inputURL string) string {
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
