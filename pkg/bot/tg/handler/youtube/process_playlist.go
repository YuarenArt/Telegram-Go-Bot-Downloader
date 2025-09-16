package youtube

import (
	"context"
	"fmt"
	"log"
	"strings"
	"youtube_downloader/pkg/bot/tg/send"
	database_client "youtube_downloader/pkg/database-client"
	"youtube_downloader/pkg/downloader/youtube"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// processPlaylistAudio downloads all videos from playlist in audio format
func (yh *YoutubeHandler) processPlaylistAudio(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery,
	playlist *youtube.Playlist, client *database_client.Client, translations *map[string]string) {

	if playlist == nil || len(playlist.Videos) == 0 {
		log.Println("Empty playlist for audio processing")
		return
	}

	for _, video := range playlist.Videos {
		if err := yh.processSingleVideoAudio(bot, callbackQuery, video, client, translations); err != nil {
			log.Printf("Failed to process video %s: %v", video.ID, err)
			continue
		}
	}
}

// processPlaylistVideo downloads all videos from playlist in video format
func (yh *YoutubeHandler) processPlaylistVideo(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery,
	playlist *youtube.Playlist, client *database_client.Client, translations *map[string]string) {

	if playlist == nil || len(playlist.Videos) == 0 {
		log.Println("Empty playlist for video processing")
		return
	}

	for _, video := range playlist.Videos {
		if err := yh.processSingleVideoVideo(bot, callbackQuery, video, client, translations); err != nil {
			log.Printf("Failed to process video %s: %v", video.ID, err)
			continue
		}
	}
}

// processSingleVideoAudio processes a single video for audio download
func (yh *YoutubeHandler) processSingleVideoAudio(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery,
	video *youtube.Video, client *database_client.Client, translations *map[string]string) error {

	if video == nil {
		return fmt.Errorf("video is nil")
	}

	audioFormat := findBestAudioFormat(video.Formats)
	if audioFormat == nil {
		return fmt.Errorf("no audio format found for video %s", video.ID)
	}

	// start downloading
	downloadingNotification := (*translations)["downloadingNotification"]
	resp, err := send.SendReplyMessage(bot, callbackQuery.Message, &downloadingNotification)
	if err != nil {
		return fmt.Errorf("failed to send downloading notification: %w", err)
	}

	opts := youtube.DownloadOptions{
		FormatID:  audioFormat.Itag,
		AudioOnly: true,
	}

	path, err := yh.Downloader.Download(context.Background(), video, opts)
	if err != nil {
		return fmt.Errorf("failed to download audio: %w", err)
	}

	// start sending
	go sendAnswer(bot, callbackQuery, &resp, &path, client, nil, translations)
	return nil
}

// processSingleVideoVideo processes a single video for video download
func (yh *YoutubeHandler) processSingleVideoVideo(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery,
	video *youtube.Video, client *database_client.Client, translations *map[string]string) error {

	if video == nil {
		return fmt.Errorf("video is nil")
	}

	videoFormat := findBestVideoFormat(video.Formats)
	if videoFormat == nil {
		return fmt.Errorf("no video format found for video %s", video.ID)
	}

	// start downloading
	downloadingNotification := (*translations)["downloadingNotification"]
	resp, err := send.SendReplyMessage(bot, callbackQuery.Message, &downloadingNotification)
	if err != nil {
		return fmt.Errorf("failed to send downloading notification: %w", err)
	}

	opts := youtube.DownloadOptions{
		FormatID:  videoFormat.Itag,
		AudioOnly: false,
	}

	path, err := yh.Downloader.Download(context.Background(), video, opts)
	if err != nil {
		return fmt.Errorf("failed to download video: %w", err)
	}

	// start sending
	go sendAnswer(bot, callbackQuery, &resp, &path, client, nil, translations)
	return nil
}

// processSingleVideo processes a single video from playlist and shows format options
func (yh *YoutubeHandler) processSingleVideo(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery,
	playlist *youtube.Playlist, translations *map[string]string) {

	if playlist == nil || len(playlist.Videos) == 0 {
		log.Println("Empty playlist for single video processing")
		return
	}

	data := callbackQuery.Data
	dataParts := strings.Split(data, ",")
	if len(dataParts) < 2 {
		log.Printf("Invalid callback data format: %s", data)
		return
	}

	videoID := dataParts[1]
	video := findVideoByID(playlist.Videos, videoID)
	if video == nil {
		log.Printf("Video not found in playlist: %s", videoID)
		return
	}

	videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", video.ID)
	keyboard, err := getKeyboardVideoFormats(video.Formats, &videoURL)
	if err != nil {
		log.Printf("Error getting keyboard formats: %v", err)
		somethingWentWrong := (*translations)["somethingWentWrong"]
		send.SendReplyMessage(bot, callbackQuery.Message, &somethingWentWrong)
		return
	}

	send.SendKeyboardMessage(bot, callbackQuery.Message, keyboard, translations)
}

// findBestAudioFormat finds the best audio format from available formats
func findBestAudioFormat(formats []youtube.Format) *youtube.Format {
	var bestFormat *youtube.Format
	var bestBitrate int

	for i, f := range formats {
		if f.AudioOnly && f.Bitrate > bestBitrate {
			bestFormat = &formats[i]
			bestBitrate = f.Bitrate
		}
	}

	return bestFormat
}

// findBestVideoFormat finds the best video format from available formats
func findBestVideoFormat(formats []youtube.Format) *youtube.Format {
	var bestFormat *youtube.Format
	var bestBitrate int

	for i, f := range formats {
		if !f.AudioOnly && f.Bitrate > bestBitrate {
			bestFormat = &formats[i]
			bestBitrate = f.Bitrate
		}
	}

	return bestFormat
}

// findVideoByID finds a video in playlist by its ID
func findVideoByID(videos []*youtube.Video, videoID string) *youtube.Video {
	for _, v := range videos {
		if v.ID == videoID {
			return v
		}
	}
	return nil
}
