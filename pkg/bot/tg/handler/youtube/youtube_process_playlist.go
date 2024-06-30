package youtube

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kkdai/youtube/v2"
	"log"
	"strconv"
	"youtube_downloader/pkg/bot/tg/send"
	youtube_downloader "youtube_downloader/pkg/downloader/youtube"
)

// TODO добавить возможность выбирать качесвто видео

func (yh *YoutubeHandler) processPlaylistAudio(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, playlist *youtube.Playlist) {
	downloader := youtube_downloader.NewYouTubeDownloader()
	for _, playlistEntry := range playlist.Videos {
		video, err := downloader.GetVideoFromPlaylistEntry(playlistEntry)
		if err != nil {
			log.Printf("VideoFromPlaylistEntry error: %v", err)
			continue
		}

		// start downloading
		resp, err := send.SendReplyMessage(bot, callbackQuery.Message, send.DownloadingNotification)
		if err != nil {
			log.Printf("can't send reply message: %w", err.Error())
		}

		path, err := downloader.DownloadAudio(video)
		if err != nil {
			log.Printf("downloadAudio error: %v", err)
			continue
		}

		// start sending
		err = send.SendEditMessage(bot, resp.Chat.ID, resp.MessageID, send.SendingNotification)
		if err != nil {
			log.Printf("can't send edit message: %w", err.Error())
		}

		if err := send.SendFile(bot, callbackQuery.Message, path); err != nil {
			log.Printf("sendFile error: %v", err)
			send.SendReplyMessage(bot, callbackQuery.Message, "File Too Large: max files size is "+strconv.Itoa(youtube_downloader.MaxFileSize/(1024*1024))+" Mb")
		}

		if err := deleteFile(path); err != nil {
			log.Printf("deleteFile error: %v", err)
		}
	}
}

func (yh *YoutubeHandler) processPlaylistVideo(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, playlist *youtube.Playlist) {
	downloader := youtube_downloader.NewYouTubeDownloader()
	for _, playlistEntry := range playlist.Videos {
		video, err := downloader.GetVideoFromPlaylistEntry(playlistEntry)
		if err != nil {
			log.Printf("VideoFromPlaylistEntry error: %v", err)
			continue
		}

		// start downloading
		resp, err := send.SendReplyMessage(bot, callbackQuery.Message, send.DownloadingNotification)
		if err != nil {
			log.Printf("can't send reply message: %w", err.Error())
		}

		path, err := downloader.DownloadVideo(video)
		if err != nil {
			log.Printf("downloadVideo error: %v", err)
			continue
		}

		// start sending
		err = send.SendEditMessage(bot, resp.Chat.ID, resp.MessageID, send.SendingNotification)
		if err != nil {
			log.Printf("can't send edit message: %w", err.Error())
		}

		if err := send.SendFile(bot, callbackQuery.Message, path); err != nil {
			log.Printf("sendFile error: %v", err)
			send.SendReplyMessage(bot, callbackQuery.Message, "File Too Large: max files size is "+strconv.Itoa(youtube_downloader.MaxFileSize))
		}

		if err := deleteFile(path); err != nil {
			log.Printf("deleteFile error: %v", err)
		}
	}
}

func (yh *YoutubeHandler) processSingleVideo(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, playlist *youtube.Playlist) {
	downloader := youtube_downloader.NewYouTubeDownloader()
	var video *youtube.Video
	for _, playlistEntry := range playlist.Videos {
		if playlistEntry.ID == callbackQuery.Data {
			video, _ = downloader.GetVideoFromPlaylistEntry(playlistEntry)
			break
		}
	}

	// start downloading
	resp, err := send.SendReplyMessage(bot, callbackQuery.Message, send.DownloadingNotification)
	if err != nil {
		log.Printf("can't send reply message: %w", err.Error())
	}

	path, err := downloader.DownloadVideo(video)
	if err != nil {
		log.Printf("downloadVideo error: %v", err)
		return
	}

	defer func() {
		if err := deleteFile(path); err != nil {
			log.Printf("deleteFile error: %v", err)
		}
	}()

	// start sending
	err = send.SendEditMessage(bot, resp.Chat.ID, resp.MessageID, send.SendingNotification)
	if err != nil {
		log.Printf("can't send edit message: %w", err.Error())
	}

	if err := send.SendFile(bot, callbackQuery.Message, path); err != nil {
		log.Printf("sendFile error: %v", err)
		send.SendReplyMessage(bot, callbackQuery.Message, "Request Entity Too Large")
	}

}
