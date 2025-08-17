package youtube

import (
	"context"
	"fmt"
	"log"
	"strings"
	"youtube_downloader/internal/bot/tg/send"
	database_client "youtube_downloader/internal/database-client"
	"youtube_downloader/internal/downloader/youtube"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (yh *YoutubeHandler) processPlaylistAudio(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery,
	playlist *youtube.Playlist, client *database_client.Client, translations *map[string]string) {
	for _, video := range playlist.Videos {
		formats := video.Formats

		var audioFormat *youtube.Format
		for _, f := range formats {
			if f.AudioOnly {
				audioFormat = &f
				break
			}
		}
		if audioFormat == nil {
			continue
		}

		// start downloading
		downloadingNotification := (*translations)["downloadingNotification"]
		resp, err := send.SendReplyMessage(bot, callbackQuery.Message, &downloadingNotification)
		if err != nil {
			log.Printf("can't send reply message: %s", err.Error())
		}
		opts := youtube.DownloadOptions{
			FormatID:  audioFormat.Itag,
			AudioOnly: true,
		}
		path, err := yh.Downloader.Download(context.Background(), video, opts)
		if err != nil {
			log.Printf("downloadAudio error: %v", err)
			continue
		}
		// start sending
		go sendAnswer(bot, callbackQuery, &resp, &path, client, nil, translations)
	}
}

func (yh *YoutubeHandler) processPlaylistVideo(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery,
	playlist *youtube.Playlist, client *database_client.Client, translations *map[string]string) {
	for _, video := range playlist.Videos {
		formats := video.Formats

		var videoFormat *youtube.Format
		for _, f := range formats {
			if !f.AudioOnly {
				videoFormat = &f
			}
		}
		if videoFormat == nil {
			continue
		}

		// start downloading
		downloadingNotification := (*translations)["downloadingNotification"]
		resp, err := send.SendReplyMessage(bot, callbackQuery.Message, &downloadingNotification)
		if err != nil {
			log.Printf("can't send reply message: %s", err.Error())
		}
		opts := youtube.DownloadOptions{
			FormatID:  videoFormat.Itag,
			AudioOnly: false,
		}
		path, err := yh.Downloader.Download(context.Background(), video, opts)
		if err != nil {
			log.Printf("downloadVideo error: %v", err)
			continue
		}
		// start sending
		go sendAnswer(bot, callbackQuery, &resp, &path, client, nil, translations)
	}
}

func (yh *YoutubeHandler) processSingleVideo(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery,
	playlist *youtube.Playlist, translations *map[string]string) {
	data := callbackQuery.Data
	dataParts := strings.Split(data, ",")
	videoID := dataParts[1] // video id
	var video *youtube.Video
	for _, v := range playlist.Videos {
		if v.ID == videoID {
			video = v
			break
		}
	}
	if video == nil {
		return
	}
	videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", video.ID)
	keyboard, err := getKeyboardVideoFormats(video.Formats, &videoURL)
	if err != nil {
		log.Println("Error after getKeyboardVideoFormats in processSingleVideo: " + err.Error())
		somethingWentWrong := (*translations)["somethingWentWrong"]
		send.SendReplyMessage(bot, callbackQuery.Message, &somethingWentWrong)
		return
	}

	send.SendKeyboardMessage(bot, callbackQuery.Message, keyboard, translations)
}
