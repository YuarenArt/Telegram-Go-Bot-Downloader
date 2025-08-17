package youtube

import (
	"context"
	"fmt"
	"log"
	"youtube_downloader/internal/downloader/youtube"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	youtubeCheckPlaylist = "https://youtu.be/check" // for checking youtube link format
)

// handleYoutubePlaylist gets playlist,
// creates and return keyboard with all videos from it
func (yh *YoutubeHandler) handleYoutubePlaylist(message *tgbotapi.Message) (*tgbotapi.InlineKeyboardMarkup, error) {
	playlistURL := message.Text
	playlist, err := yh.Downloader.GetPlaylist(context.Background(), playlistURL)
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
		fmt.Sprintf("%s", "Download all: video"), youtubeCheckPlaylist+","+All_video)
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{button})

	button = tgbotapi.NewInlineKeyboardButtonData(
		fmt.Sprintf("%s", "Download all: audio"), youtubeCheckPlaylist+","+All_audio)
	keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{button})

	for _, video := range playlist.Videos {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s", video.Title), youtubeCheckPlaylist+","+video.ID)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{button})
	}

	return keyboard
}
