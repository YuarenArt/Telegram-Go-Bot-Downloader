package tg_bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kkdai/youtube/v2"
	"log"
	"strconv"
	"strings"
	youtube_downloader "youtube_downloader/pkg/downloader/youtube-downloader"
)

func (tb *TgBot) handleYoutubePlaylist(message *tgbotapi.Message) error {

	playlistURL := message.Text

	downloader := youtube_downloader.NewYouTubeDownloader()
	playlist, err := downloader.GetPlaylist(playlistURL)
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

func (tb *TgBot) processPlaylistAudio(callbackQuery *tgbotapi.CallbackQuery, playlist *youtube.Playlist) {
	downloader := youtube_downloader.NewYouTubeDownloader()
	for _, playlistEntry := range playlist.Videos {
		video, err := downloader.GetVideoFromPlaylistEntry(playlistEntry)
		if err != nil {
			log.Printf("VideoFromPlaylistEntry error: %v", err)
			continue
		}

		// start downloading
		resp, err := tb.sendReplyMessage(callbackQuery.Message, downloadingNotification)
		if err != nil {
			log.Printf("can't send reply message: %w", err.Error())
		}

		path, err := tb.downloadAudio(video)
		if err != nil {
			log.Printf("downloadAudio error: %v", err)
			continue
		}

		// start sending
		err = tb.sendEditMessage(resp.Chat.ID, resp.MessageID, sendingNotification)
		if err != nil {
			log.Printf("can't send edit message: %w", err.Error())
		}

		if err := tb.sendFile(callbackQuery.Message, path); err != nil {
			log.Printf("sendFile error: %v", err)
			tb.sendReplyMessage(callbackQuery.Message, "File Too Large: max files size is "+strconv.Itoa(maxFileSize/(1024*1024))+" Mb")
		}
	}
}

func (tb *TgBot) processPlaylistVideo(callbackQuery *tgbotapi.CallbackQuery, playlist *youtube.Playlist) {
	downloader := youtube_downloader.NewYouTubeDownloader()
	for _, playlistEntry := range playlist.Videos {
		video, err := downloader.GetVideoFromPlaylistEntry(playlistEntry)
		if err != nil {
			log.Printf("VideoFromPlaylistEntry error: %v", err)
			continue
		}

		// start downloading
		resp, err := tb.sendReplyMessage(callbackQuery.Message, downloadingNotification)
		if err != nil {
			log.Printf("can't send reply message: %w", err.Error())
		}

		path, err := tb.downloadVideo(video)
		if err != nil {
			log.Printf("downloadVideo error: %v", err)
			continue
		}

		// start sending
		err = tb.sendEditMessage(resp.Chat.ID, resp.MessageID, sendingNotification)
		if err != nil {
			log.Printf("can't send edit message: %w", err.Error())
		}

		if err := tb.sendFile(callbackQuery.Message, path); err != nil {
			log.Printf("sendFile error: %v", err)
			tb.sendReplyMessage(callbackQuery.Message, "File Too Large: max files size is "+strconv.Itoa(maxFileSize))
		}

		if err := deleteFile(path); err != nil {
			log.Printf("deleteFile error: %v", err)
		}
	}
}

func (tb *TgBot) processSingleVideo(callbackQuery *tgbotapi.CallbackQuery, playlist *youtube.Playlist) {
	downloader := youtube_downloader.NewYouTubeDownloader()
	var video *youtube.Video
	for _, playlistEntry := range playlist.Videos {
		if playlistEntry.ID == callbackQuery.Data {
			video, _ = downloader.GetVideoFromPlaylistEntry(playlistEntry)
			break
		}
	}

	// start downloading
	resp, err := tb.sendReplyMessage(callbackQuery.Message, downloadingNotification)
	if err != nil {
		log.Printf("can't send reply message: %w", err.Error())
	}

	path, err := tb.downloadVideo(video)
	if err != nil {
		log.Printf("downloadVideo error: %v", err)
		return
	}

	// start sending
	err = tb.sendEditMessage(resp.Chat.ID, resp.MessageID, sendingNotification)
	if err != nil {
		log.Printf("can't send edit message: %w", err.Error())
	}

	if err := tb.sendFile(callbackQuery.Message, path); err != nil {
		log.Printf("sendFile error: %v", err)
		tb.sendReplyMessage(callbackQuery.Message, "Request Entity Too Large")
	}

	defer func() {
		if err := deleteFile(path); err != nil {
			log.Printf("deleteFile error: %v", err)
		}
	}()
}

func extractPlaylistURL(text string) string {
	parts := strings.Split(text, "\n")
	for _, part := range parts {
		if strings.HasPrefix(part, "https://") {
			return part
		}
	}
	return ""
}
