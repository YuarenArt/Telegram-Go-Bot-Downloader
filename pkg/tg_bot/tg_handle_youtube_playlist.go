package tg_bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	youtube_downloader "youtube_downloader/pkg/downloader/youtube-downloader"
)

func (tb *TgBot) handleYoutubePlaylist(message *tgbotapi.Message) error {

	playlistURL := message.Text

	downloader := youtube_downloader.NewYouTubeDownloader()
	playlist, err := downloader.Downloader.Client.GetPlaylist(playlistURL)
	if err != nil {
		log.Printf("GetPlaylist in handleYoutubePlaylist: %w", err)
		return err
	}

	keyboard := getKeyboardPlaylist(playlist)
	err = tb.sendKeyboardMessage(message, keyboard)
	if err != nil {
		log.Printf("sendKeyboardMessage in handleYoutubePlaylist: %w", err)
		return err
	}

	return nil
}
