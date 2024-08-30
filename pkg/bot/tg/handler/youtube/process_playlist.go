package youtube

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kkdai/youtube/v2"
	"log"
	"strings"
	"youtube_downloader/pkg/bot/tg/send"
	database_client "youtube_downloader/pkg/database-client"
	youtube_downloader "youtube_downloader/pkg/downloader/youtube"
)

func (yh *YoutubeHandler) processPlaylistAudio(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery,
	playlist *youtube.Playlist, client *database_client.Client, translations *map[string]string) {
	downloader := youtube_downloader.NewYouTubeDownloader()
	for _, playlistEntry := range playlist.Videos {
		video, err := downloader.GetVideoFromPlaylistEntry(playlistEntry)
		if err != nil {
			log.Printf("VideoFromPlaylistEntry error: %v", err)
			continue
		}

		formats := video.Formats.WithAudioChannels()
		formats.Sort()
		formats, err = youtube_downloader.WithFormats(&video.Formats, youtube_downloader.AUDIO_PREFIX)

		if !checkTraffic(client, callbackQuery, &formats[0]) {
			trafficLimit := (*translations)["trafficLimit"]
			_, err := send.SendReplyMessage(bot, callbackQuery.Message, &trafficLimit)
			if err != nil {
				log.Printf("can't send reply message: %s", err.Error())
			}
			return
		}

		// start downloading
		downloadingNotification := (*translations)["downloadingNotification"]
		resp, err := send.SendReplyMessage(bot, callbackQuery.Message, &downloadingNotification)
		if err != nil {
			log.Printf("can't send reply message: %s", err.Error())
		}

		path, err := downloader.DownloadAudio(video)
		if err != nil {
			log.Printf("downloadAudio error: %v", err)
			continue
		}
		fileSize, err := getFileSize(formats[0]) // bite
		fileSize = fileSize / (1024 * 1024)      // Mb
		if err != nil {
			log.Printf("can't file size: %s", err.Error())
		}

		// start sending
		go sendAnswer(bot, callbackQuery, &resp, &path, client, &fileSize, translations)
	}
}

func (yh *YoutubeHandler) processPlaylistVideo(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery,
	playlist *youtube.Playlist, client *database_client.Client, translations *map[string]string) {
	downloader := youtube_downloader.NewYouTubeDownloader()
	for _, playlistEntry := range playlist.Videos {
		video, err := downloader.GetVideoFromPlaylistEntry(playlistEntry)
		if err != nil {
			log.Printf("VideoFromPlaylistEntry error: %v", err)
			continue
		}

		formats := video.Formats.WithAudioChannels()
		formats.Sort()
		formats, err = youtube_downloader.WithFormats(&video.Formats, youtube_downloader.VIDEO_PREFIX)

		if !checkTraffic(client, callbackQuery, &video.Formats[len(formats)-1]) {
			trafficLimit := (*translations)["trafficLimit"]
			_, err := send.SendReplyMessage(bot, callbackQuery.Message, &trafficLimit)
			if err != nil {
				log.Printf("can't send reply message: %s", err.Error())
			}
			return
		}

		// start downloading
		downloadingNotification := (*translations)["downloadingNotification"]
		resp, err := send.SendReplyMessage(bot, callbackQuery.Message, &downloadingNotification)
		if err != nil {
			log.Printf("can't send reply message: %s", err.Error())
		}

		path, err := downloader.DownloadVideo(video)
		if err != nil {
			log.Printf("downloadVideo error: %v", err)
			continue
		}

		fileSize, err := getFileSize(formats[len(formats)-1]) // bite
		fileSize = fileSize / (1024 * 1024)                   // Mb
		if err != nil {
			log.Printf("can't file size: %s", err.Error())
		}

		// start sending
		go sendAnswer(bot, callbackQuery, &resp, &path, client, &fileSize, translations)
	}
}

func (yh *YoutubeHandler) processSingleVideo(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery,
	playlist *youtube.Playlist, translations *map[string]string) {
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
		return
	}
	formats := video.Formats
	keyboard, err := getKeyboardVideoFormats(&formats, &videoURL)
	if err != nil {
		log.Println("Error after getKeyboardVideoFormats in processSingleVideo: " + err.Error())
		somethingWentWrong := (*translations)["somethingWentWrong"]
		send.SendReplyMessage(bot, callbackQuery.Message, &somethingWentWrong)
		return
	}

	send.SendKeyboardMessage(bot, callbackQuery.Message, keyboard, translations)
}
