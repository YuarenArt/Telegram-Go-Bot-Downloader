package youtube

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kkdai/youtube/v2"
	"log"
	youtube_downloader "youtube_downloader/pkg/downloader/youtube"
)

// handleYoutubePlaylist gets playlist,
// creates and return keyboard with all videos from it
func (yh *YoutubeHandler) handleYoutubePlaylist(message *tgbotapi.Message) (*tgbotapi.InlineKeyboardMarkup, error) {

	playlistURL := message.Text

	downloader := youtube_downloader.NewYouTubeDownloader()
	playlist, err := downloader.GetPlaylist(playlistURL)
	if err != nil {
		log.Printf("GetPlaylist in handleYoutubePlaylist: %w", err)
		return nil, err
	}

	keyboard := getKeyboardPlaylist(playlist)
	return &keyboard, nil
}

// getKeyboardPlaylist return a keyboard with all videos from playlist
func getKeyboardPlaylist(playlist *youtube.Playlist) tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup()

	button := tgbotapi.NewInlineKeyboardButtonData(
		fmt.Sprintf("%s", "Download all: video"), All_video)
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{button})

	button = tgbotapi.NewInlineKeyboardButtonData(
		fmt.Sprintf("%s", "Download all: audio"), All_audio)
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{button})

	for _, playlistEntry := range playlist.Videos {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s", playlistEntry.Title), playlistEntry.ID)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{button})
	}

	return keyboard
}
