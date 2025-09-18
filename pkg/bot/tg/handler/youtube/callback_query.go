package youtube

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"youtube_downloader/pkg/bot/tg/send"
	"youtube_downloader/pkg/bot/tg/util"
	"youtube_downloader/pkg/database/models"
	"youtube_downloader/pkg/downloader/youtube"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleCallbackQuery gets url from Bot's message with a replying link,
// then handle a link by its type: video (stream), playlist
func (yh *YoutubeHandler) HandleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, translations *map[string]string) {
	if callbackQuery == nil || callbackQuery.Data == "" {
		log.Println("Invalid callback query")
		return
	}

	text := callbackQuery.Data
	parts := strings.Split(text, ",")
	if len(parts) < 2 {
		log.Printf("Invalid callback data format: %s", text)
		return
	}

	URL := parts[0]

	switch {
	case strings.HasPrefix(URL, "https://youtube.com/playlist?") || URL == youtubeCheckPlaylist:
		yh.HandleCallbackQueryWithPlaylist(callbackQuery, bot, translations)
	default:
		yh.HandleCallbackQueryWithFormats(callbackQuery, bot, translations)
	}
}

// HandleCallbackQueryWithFormats gets a link on video by callbackQuery.Message.Text,
// gets ItagNo by callbackQuery.Data to find a correct format,
// gets possible formats by videoURL,
// and finally gets the format selected by the user.
// then download it with format
func (yh *YoutubeHandler) HandleCallbackQueryWithFormats(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, translations *map[string]string) {
	// Create a context with timeout for the entire operation
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	data := callbackQuery.Data
	dataParts := strings.Split(data, ",")
	if len(dataParts) < 2 {
		log.Printf("Invalid callback data format: %s", data)
		errorMsg := (*translations)["errorFormat"]
		send.SendReplyMessage(bot, callbackQuery.Message, &errorMsg)
		return
	}

	videoURL := dataParts[0]
	tagNoStr := dataParts[1]

	// Notify user that we're processing their request
	processingMsg := (*translations)["processingRequest"]
	if processingMsg == "" {
		processingMsg = "🔄 Processing your request..."
	}
	processing, err := send.SendReplyMessage(bot, callbackQuery.Message, &processingMsg)
	if err != nil {
		log.Printf("Failed to send processing message: %v", err)
	}

	// Get video metadata
	video, err := yh.Downloader.GetVideo(ctx, videoURL)
	if err != nil {
		log.Printf("Failed to get video: %v", err)
		errorMsg := (*translations)["errorGettingVideo"]
		if errorMsg == "" {
			errorMsg = "❌ Failed to get video information. Please try again later."
		}
		send.SendEditMessage(bot, processing.Chat.ID, processing.MessageID, &errorMsg)
		return
	}

	tagNo, err := strconv.Atoi(tagNoStr)
	if err != nil {
		log.Printf("Invalid tag number: %s", tagNoStr)
		errorMsg := (*translations)["errorFormat"]
		if errorMsg == "" {
			errorMsg = "❌ Invalid format selected. Please try again."
		}
		send.SendEditMessage(bot, processing.Chat.ID, processing.MessageID, &errorMsg)
		return
	}

	formatFile := findFormatByItag(video.Formats, tagNo)
	if formatFile == nil {
		log.Printf("Format not found for tag: %d", tagNo)
		errorMsg := (*translations)["errorFormat"]
		if errorMsg == "" {
			errorMsg = "❌ Selected format is not available. Please try another format."
		}
		send.SendEditMessage(bot, processing.Chat.ID, processing.MessageID, &errorMsg)
		return
	}

	// Check traffic limits
	if !yh.checkTraffic(callbackQuery, formatFile) {
		trafficLimit := (*translations)["trafficLimit"]
		if trafficLimit == "" {
			trafficLimit = "⚠️ You've reached your download limit. Please try again later."
		}
		send.SendEditMessage(bot, processing.Chat.ID, processing.MessageID, &trafficLimit)
		return
	}

	// Prepare download options with proper filename
	filename := util.SanitizeFilename(video.Title)
	if filename == "" {
		filename = "video"
	}

	// Add quality/format info to filename if available
	if formatFile.Quality != "" {
		filename += "_" + formatFile.Quality
	}
	if formatFile.AudioOnly {
		filename += "_audio"
	}

	// Create a temporary directory for downloads if it doesn't exist
	tempDir := filepath.Join(os.TempDir(), "youtube-dl-bot")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Printf("Failed to create temp directory: %v", err)
		errorMsg := "❌ Internal server error. Please try again later."
		send.SendEditMessage(bot, processing.Chat.ID, processing.MessageID, &errorMsg)
		return
	}

	// Start downloading with progress updates
	downloadingMsg := (*translations)["downloadingNotification"]
	if downloadingMsg == "" {
		downloadingMsg = "⏬ Downloading..."
	}
	send.SendEditMessage(bot, processing.Chat.ID, processing.MessageID, &downloadingMsg)

	opts := youtube.DownloadOptions{
		Format:    formatFile.FormatID,
		AudioOnly: formatFile.AudioOnly,
		Filename:  filename,
		OutputDir: tempDir,
	}

	// Download the file
	pathAndName, err := yh.Downloader.Download(ctx, video, opts)
	if err != nil {
		log.Printf("Download failed: %v", err)
		errorMsg := (*translations)["downloadFailed"]
		if errorMsg == "" {
			errorMsg = "❌ Failed to download the video. Please try again later."
		}
		send.SendEditMessage(bot, processing.Chat.ID, processing.MessageID, &errorMsg)
		return
	}

	// Send the file to the user
	go yh.sendAnswer(bot, callbackQuery, processing, &pathAndName, nil, translations)
}

// findFormatByItag finds a format by its itag number
func findFormatByItag(formats []youtube.Format, itag int) *youtube.Format {
	for i, f := range formats {
		if f.Itag == itag {
			return &formats[i]
		}
	}
	return nil
}

// HandleCallbackQueryWithPlaylist gets link on playlist by callbackQuery.Message.Text
// checks callbackQuery.Data
// if callbackQuery.Data include All_audio : download all videos from playlist in audio format
// if callbackQuery.Data include All_video : download all videos from playlist in video format
// else download a certain video by callbackQuery.Data
func (yh *YoutubeHandler) HandleCallbackQueryWithPlaylist(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, translations *map[string]string) {
	lines := strings.Split(callbackQuery.Message.Text, "\n")
	var playlistURL string
	for _, line := range lines {
		if strings.HasPrefix(line, "https://") {
			playlistURL = line
			break
		}
	}

	if playlistURL == "" {
		log.Println("No playlist URL found in message")
		return
	}

	playlist, err := yh.Downloader.GetPlaylist(context.Background(), playlistURL)
	if err != nil {
		log.Printf("GetPlaylist in handleCallbackQueryWithPlaylist error: %v", err)
		return
	}

	data := callbackQuery.Data
	dataParts := strings.Split(data, ",")
	if len(dataParts) < 2 {
		log.Printf("Invalid callback data format: %s", data)
		return
	}

	switch {
	case dataParts[1] == All_audio:
		yh.processPlaylistAudio(bot, callbackQuery, playlist, translations)
	case dataParts[1] == All_video:
		yh.processPlaylistVideo(bot, callbackQuery, playlist, translations)
	default:
		yh.processSingleVideo(bot, callbackQuery, playlist, translations)
	}
}

func deleteFile(pathToFile string) error {
	if pathToFile == "" {
		return nil
	}
	return os.Remove(pathToFile)
}

// sendAnswer sends the downloaded file to the user and handles cleanup
func (yh *YoutubeHandler) sendAnswer(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, resp *tgbotapi.Message,
	path *string, traffic *float64, translations *map[string]string) {

	// Ensure we have a valid file path
	if path == nil || *path == "" {
		errorMsg := "❌ Error: No file to send"
		send.SendEditMessage(bot, resp.Chat.ID, resp.MessageID, &errorMsg)
		log.Println("Invalid file path for sending")
		return
	}

	// Ensure the file exists before attempting to send
	if _, err := os.Stat(*path); os.IsNotExist(err) {
		errorMsg := "❌ Error: The downloaded file was not found"
		send.SendEditMessage(bot, resp.Chat.ID, resp.MessageID, &errorMsg)
		log.Printf("File not found: %s", *path)
		return
	}

	// Update user that we're sending the file
	sendingMsg := (*translations)["sendingNotification"]
	if sendingMsg == "" {
		sendingMsg = "📤 Sending file..."
	}
	err := send.SendEditMessage(bot, resp.Chat.ID, resp.MessageID, &sendingMsg)
	if err != nil {
		log.Printf("Failed to update sending status: %v", err)
	}

	// Defer file cleanup
	defer func() {
		if err := deleteFile(*path); err != nil {
			log.Printf("Failed to delete file %s: %v", *path, err)
		}
	}()

	// Get file info for logging
	fileInfo, err := os.Stat(*path)
	if err == nil {
		log.Printf("Sending file: %s (%.2f MB)", *path, float64(fileInfo.Size())/1024/1024)
	}

	// Send the file
	err = send.SendFile(bot, callbackQuery.Message, *path)
	if err != nil {
		log.Printf("Failed to send file: %v", err)
		errorMsg := (*translations)["errorSendingFile"]
		if errorMsg == "" {
			errorMsg = "❌ Failed to send the file. Please try again."
		}
		send.SendEditMessage(bot, resp.Chat.ID, resp.MessageID, &errorMsg)
		log.Printf("Failed to send file: %v", err)
	} else {
		// File sent successfully, update user traffic
		if traffic != nil {
			yh.updateUserTraffic(callbackQuery, traffic)
		}
		// Send completion message
		completionMsg := (*translations)["downloadComplete"]
		if completionMsg == "" {
			completionMsg = "✅ Download complete! Enjoy! 🎉"
		}
		send.SendEditMessage(bot, resp.Chat.ID, resp.MessageID, &completionMsg)
	}
}

func (yh *YoutubeHandler) updateUserTraffic(callbackQuery *tgbotapi.CallbackQuery, traffic *float64) {
	log.Printf("Updating traffic for user: %s", callbackQuery.From.UserName)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	user, err := yh.getOrCreateUser(ctx, callbackQuery)
	if err != nil || user == nil {
		log.Printf("Can't get or create user: %s error: %s", callbackQuery.From.UserName, err.Error())
		return
	}

	var trafficToAdd float64
	if traffic != nil {
		trafficToAdd = *traffic
	} else {
		parsedTraffic, err := parseTrafficFromCallbackQuery(callbackQuery)
		if err != nil {
			log.Printf("Can't parse traffic: %s", err.Error())
			return
		}
		trafficToAdd = parsedTraffic
	}

	err = yh.Client.UpdateTraffic(ctx, callbackQuery.From.UserName, int64(user.Traffic+trafficToAdd))
	if err != nil {
		log.Printf("Can't update user traffic user: %s; error: %s", user.Username, err.Error())
		return
	}

	log.Println("Successful updating")
}

func (yh *YoutubeHandler) getOrCreateUser(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery) (*models.User, error) {
	userResp, err := yh.Client.GetUser(ctx, callbackQuery.From.UserName)
	if err != nil || userResp == nil {
		chatID := callbackQuery.Message.Chat.ID
		// Create a new user directly with the required fields
		userResp, err = yh.Client.CreateUser(ctx, callbackQuery.From.UserName, chatID)
		if err != nil {
			return nil, err
		}
	}

	// Convert UserResponse to models.User
	user := &models.User{
		Username: userResp.Username,
		Traffic:  float64(userResp.Traffic),
		ChatID:   userResp.ChatID,
		Subscription: models.Subscription{
			Duration:          userResp.Subscription.Duration,
			StartSubscription: userResp.Subscription.StartSubscription,
			EndSubscription:   userResp.Subscription.EndSubscription,
		},
	}
	return user, nil
}

func parseTrafficFromCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) (float64, error) {
	if callbackQuery.Message == nil || callbackQuery.Message.ReplyMarkup == nil {
		return 0, fmt.Errorf("invalid message or reply markup")
	}

	tokens := strings.Split(callbackQuery.Data, ",")
	if len(tokens) < 2 {
		return 0, fmt.Errorf("invalid callback data format")
	}

	itagNo := tokens[1]

	for _, row := range callbackQuery.Message.ReplyMarkup.InlineKeyboard {
		for _, keyboardButton := range row {
			if keyboardButton.CallbackData == nil {
				continue
			}

			tokens := strings.Split(*keyboardButton.CallbackData, ",")
			if len(tokens) < 2 {
				continue
			}

			itag := tokens[len(tokens)-1]
			if itagNo == itag {
				tokens = strings.Split(keyboardButton.Text, ",")
				if len(tokens) == 0 {
					continue
				}

				tokens = strings.Split(tokens[len(tokens)-1], " ")
				if len(tokens) < 2 {
					continue
				}

				traffic, err := strconv.ParseFloat(tokens[1], 64)
				if err != nil {
					return 0, err
				}
				return traffic, nil
			}
		}
	}
	return 0, nil
}

func (yh *YoutubeHandler) checkTraffic(callbackQuery *tgbotapi.CallbackQuery, format *youtube.Format) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	userResp, err := yh.Client.GetUser(ctx, callbackQuery.From.UserName)
	if err != nil {
		log.Printf("can't get user by username: %s, error: %s", callbackQuery.Message.From.UserName, err.Error())
		return true
	} else if userResp == nil {
		log.Printf("Get nil user: %s", callbackQuery.Message.From.UserName)
		return true
	}

	fileSize := estimateFileSize(format)
	if float64(userResp.Traffic)+fileSize > TrafficLimit && userResp.Subscription.Duration == "" {
		return false
	}
	return true
}

// estimateFileSize estimates file size from format metadata
func estimateFileSize(format *youtube.Format) float64 {
	if format == nil {
		return 0
	}

	// Use actual duration if available, otherwise use default
	duration := DefaultDuration
	if format.Bitrate > 0 {
		fileSize := float64(format.Bitrate) * duration / BytesPerMB
		return fileSize
	}

	return 0
}
