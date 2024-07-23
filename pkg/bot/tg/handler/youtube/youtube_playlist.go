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
		log.Printf("GetPlaylist in handleYoutubePlaylist: %s", err)
		return nil, err
	}

	keyboard := getKeyboardPlaylist(playlist)
	return &keyboard, nil
}

// getKeyboardPlaylist return a keyboard with all videos from playlist. Button's data include youtube url (for checking link while handling)
// and playlistEntry.ID for certain videos, and All_video and All_audio for downloading all playlist
func getKeyboardPlaylist(playlist *youtube.Playlist) tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup()

	button := tgbotapi.NewInlineKeyboardButtonData(
		fmt.Sprintf("%s", "Download all: video"), "https://youtu.be"+","+All_video)
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{button})

	button = tgbotapi.NewInlineKeyboardButtonData(
		fmt.Sprintf("%s", "Download all: audio"), "https://youtu.be"+","+All_audio)
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{button})

	for _, playlistEntry := range playlist.Videos {

		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s", playlistEntry.Title), "https://youtu.be"+","+playlistEntry.ID)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{button})
	}

	return keyboard
}
