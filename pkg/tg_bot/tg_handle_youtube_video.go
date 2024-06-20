package tg_bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	youtube_downloader "youtube_downloader/pkg/downloader/youtube-downloader"
)

// handleYoutubeVideo gets all possible formats of the video by a link
// creates a keyboard and sends it to user's chat
func (tb *TgBot) handleYoutubeVideo(message *tgbotapi.Message) error {
	videoURL := message.Text
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

	err = tb.sendKeyboardMessage(message, keyboard)
	if err != nil {
		log.Printf("sendKeyboardMessage %w", err)
		return err
	}

	return nil
}
