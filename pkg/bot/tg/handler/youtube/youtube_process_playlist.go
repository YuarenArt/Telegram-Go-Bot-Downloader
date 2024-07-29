package youtube

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kkdai/youtube/v2"
	"log"
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
		go sendAnswer(bot, callbackQuery, &resp, &path)
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
		go sendAnswer(bot, callbackQuery, &resp, &path)
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
	video, err := downloader.GetVideo(videoURL)
	if err != nil {
		log.Println("can't get video in processSingleVideo: " + err.Error())
	}
	formats := video.Formats
	keyboard, err := getKeyboardVideoFormats(formats, videoURL)
	if err != nil {
		log.Println("Error after getKeyboardVideoFormats in processSingleVideo: " + err.Error())
		send.SendReplyMessage(bot, callbackQuery.Message, "Something went wrong. Sorry!")
		return
	}

	send.SendKeyboardMessage(bot, callbackQuery.Message, keyboard)
}
