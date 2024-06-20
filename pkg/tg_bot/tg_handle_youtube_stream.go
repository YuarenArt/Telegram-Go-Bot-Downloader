package tg_bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/url"
	"strings"
	youtube_downloader "youtube_downloader/pkg/downloader/youtube-downloader"
)

// handleYoutubeStream transforms live/ link into common video link
// download it and send
func (tb *TgBot) handleYoutubeStream(message *tgbotapi.Message) error {

	videoURLWithLivePrefix := message.Text
	videoURL := formatYouTubeURLOnStream(videoURLWithLivePrefix)
	formats, err := youtube_downloader.FormatWithAudioChannels(videoURL)
	if err != nil {
		log.Printf("FormatWithAudioChannels return %w", err)
		return err
	}

	keyboard, err := getKeyboardVideoFormats(formats)
	if err != nil {
		log.Printf("GetKeyboard return %w", err)
		return err
	}

	err = tb.sendKeyboardMessageWithFormattedLink(message, keyboard, videoURL)
	if err != nil {
		log.Printf("sendKeyboardMessage %w", err)
		return err
	}

	return nil
}

// formatYouTubeURLonStream instead of live/ links return link on video
func formatYouTubeURLOnStream(inputURL string) string {
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
