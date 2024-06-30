package youtube

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/url"
	"strings"
	youtube_downloader "youtube_downloader/pkg/downloader/youtube"
)

// handleYoutubeStream transforms live/ link into common video link
// download it and return
func (yh *YoutubeHandler) handleYoutubeStream(message *tgbotapi.Message) (*tgbotapi.InlineKeyboardMarkup, error) {

	videoURLWithLivePrefix := message.Text
	videoURL := FormatYouTubeURLOnStream(videoURLWithLivePrefix)
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

// FormatYouTubeURLonStream instead of live/ links return link on video
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
