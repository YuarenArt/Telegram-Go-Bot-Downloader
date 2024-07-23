package youtube

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kkdai/youtube/v2"
	"log"
	"strconv"
	"strings"
	"youtube_downloader/pkg/bot/tg/send"
	youtube_downloader "youtube_downloader/pkg/downloader/youtube"
)

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
			log.Printf("can't send reply message: %s", err.Error())
		}

		path, err := downloader.DownloadAudio(video)
		if err != nil {
			log.Printf("downloadAudio error: %v", err)
			continue
		}

		// start sending
		err = send.SendEditMessage(bot, resp.Chat.ID, resp.MessageID, send.SendingNotification)
		if err != nil {
			log.Printf("can't send edit message: %s", err.Error())
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
			log.Printf("can't send reply message: %s", err.Error())
		}

		path, err := downloader.DownloadVideo(video)
		if err != nil {
			log.Printf("downloadVideo error: %v", err)
			continue
		}

		// start sending
		err = send.SendEditMessage(bot, resp.Chat.ID, resp.MessageID, send.SendingNotification)
		if err != nil {
			log.Printf("can't send edit message: %s", err.Error())
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

	data := callbackQuery.Data
	dataParts := strings.Split(data, ",")
	videoID := dataParts[1] // video id
	var video *youtube.Video
	for _, playlistEntry := range playlist.Videos {
		if playlistEntry.ID == videoID {
			video, _ = downloader.GetVideoFromPlaylistEntry(playlistEntry)
			break
		}
	}

	videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", video.ID)
	formats := video.Formats.WithAudioChannels()

	keyboard, err := getKeyboardVideoFormats(formats, videoURL)
	if err != nil {
		log.Println("Error after getKeyboardVideoFormats in processSingleVideo: " + err.Error())
	}
	send.SendKeyboardMessage(bot, callbackQuery.Message, keyboard)
}
