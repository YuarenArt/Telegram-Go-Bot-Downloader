package youtube

import (
	"context"
	"github.com/YuarenArt/tg-users-database/pkg/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kkdai/youtube/v2"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"youtube_downloader/pkg/bot/tg/send"
	database_client "youtube_downloader/pkg/database-client"
	youtube_downloader "youtube_downloader/pkg/downloader/youtube"
)

// TODO rework a way to get data for downloading

// TODO refactore indexes for data

// HandleCallbackQuery gets url from Bot's message with a replying link,
// then handle a link by its type: video (stream), playlist
func (yh *YoutubeHandler) HandleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, client *database_client.Client, translations *map[string]string) {

	// gets URL from a data
	text := callbackQuery.Data
	parts := strings.Split(text, ",")
	URL := parts[0]

	switch {
	// TODO fix that need to obtain link for handling playlist Button
	case strings.HasPrefix(URL, "https://youtube.com/playlist?") || URL == youtubeCheckPlaylist:
		yh.HandleCallbackQueryWithPlaylist(callbackQuery, bot, client, translations)
	default:
		yh.HandleCallbackQueryWithFormats(callbackQuery, bot, client, translations)
	}
}

// HandleCallbackQueryWithFormats gets a link on video by callbackQuery.Message.Text,
// gets ItagNo by callbackQuery.Data to find a correct format,
// gets possible formats by videoURL,
// and finally gets the format selected by the user.
// then download it with format
func (yh *YoutubeHandler) HandleCallbackQueryWithFormats(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI,
	client *database_client.Client, translations *map[string]string) {

	data := callbackQuery.Data
	dataParts := strings.Split(data, ",")
	videoURL := dataParts[0]

	formats, err := youtube_downloader.FormatWithAudioChannelsComposite(videoURL)
	if err != nil {
		log.Printf("FormatWithAudioChannels return %s in handleCallbackQuery", err)
	}

	// gets format by its TagNo
	tagNo, err := strconv.Atoi(dataParts[1])
	if err != nil {
		errorFormat := (*translations)["errorFormat"]
		send.SendReplyMessage(bot, callbackQuery.Message, &errorFormat)
		return
	}

	var formatFile youtube.Format
	for _, format := range formats {
		if format.ItagNo == tagNo {
			formatFile = format
			break
		}
	}

	if !checkTraffic(client, callbackQuery, &formatFile) {
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

	dl := youtube_downloader.NewYouTubeDownloader()
	video, _ := dl.GetVideo(videoURL)
	var pathAndName string
	if strings.HasPrefix(formatFile.MimeType, "audio") {
		pathAndName, err = dl.DownloadWithFormat(video, formatFile)
	} else {
		pathAndName, err = dl.DownloadVideoWithFormatComposite(context.Background(), "", video, formatFile.QualityLabel, "", "")
	}
	if err != nil {
		log.Printf(err.Error())
		errorFormat := (*translations)["errorFormat"]
		send.SendEditMessage(bot, resp.Chat.ID, resp.MessageID, &errorFormat)
		return
	}
	// start sending
	go sendAnswer(bot, callbackQuery, &resp, &pathAndName, client, nil, translations)
}

// HandleCallbackQueryWithPlaylist gets link on playlist by callbackQuery.Message.Text
// checks callbackQuery.Data
// if callbackQuery.Data include All_audio : download all videos from playlist in audio format
// if callbackQuery.Data include All_video : download all videos from playlist in video format
// else download a certain video by callbackQuery.Data
func (yh *YoutubeHandler) HandleCallbackQueryWithPlaylist(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, client *database_client.Client, translations *map[string]string) {
	lines := strings.Split(callbackQuery.Message.Text, "\n") // split the string into lines
	var playlistURL string
	for _, line := range lines {
		if strings.HasPrefix(line, "https://") {
			playlistURL = line
			break
		}
	}
	downloader := youtube_downloader.NewYouTubeDownloader()

	playlist, err := downloader.GetPlaylist(playlistURL)
	if err != nil {
		log.Printf("GetPlaylist in handleCallbackQueryWithPlaylist error: %v", err)
	}
	data := callbackQuery.Data
	dataParts := strings.Split(data, ",")

	switch {
	case dataParts[1] == All_audio:
		yh.processPlaylistAudio(bot, callbackQuery, playlist, client, translations)
	case dataParts[1] == All_video:
		yh.processPlaylistVideo(bot, callbackQuery, playlist, client, translations)
	default:
		yh.processSingleVideo(bot, callbackQuery, playlist, translations)
	}
}

func deleteFile(pathToFile string) error {
	return os.Remove(pathToFile)
}

func sendAnswer(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery, resp *tgbotapi.Message,
	path *string, client *database_client.Client, traffic *float64, translations *map[string]string) {

	sendingNotification := (*translations)["sendingNotification"]
	err := send.SendEditMessage(bot, resp.Chat.ID, resp.MessageID, &sendingNotification)
	if err != nil {
		log.Printf("can't send edit message: %s", err.Error())
	}

	defer func() {
		err = deleteFile(*path)
		if err != nil {
			log.Printf("deleteFile return %s in handleCallbackQuery", err)
		}
	}()

	err = send.SendFile(bot, callbackQuery.Message, *path)
	if err != nil {
		errorFormatSending := (*translations)["errorFormatSending"]
		send.SendEditMessage(bot, resp.Chat.ID, resp.MessageID, &errorFormatSending)
		log.Printf("sendFile return %s in handleCallbackQuery", err)
	} else {
		updateUserTraffic(callbackQuery, client, traffic)
	}
}

func updateUserTraffic(callbackQuery *tgbotapi.CallbackQuery, client *database_client.Client, traffic *float64) {
	log.Printf("Updating traffic for user: %s", callbackQuery.From.UserName)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	user, err := getOrCreateUser(ctx, client, callbackQuery)
	if err != nil || user == nil {
		log.Printf("Can't get or create user: %s error: %s", callbackQuery.From.UserName, err.Error())
		return
	}

	if traffic == nil {
		parsedTraffic, err := parseTrafficFromCallbackQuery(callbackQuery)
		if err != nil {
			log.Printf("Can't parse traffic: %s", err.Error())
			return
		}
		traffic = &parsedTraffic
	}

	err = client.UpdateTraffic(ctx, callbackQuery.From.UserName, user.Traffic+*traffic)
	if err != nil {
		log.Printf("Can't update user traffic user: %s; error: %s", user.Username, err.Error())
		return
	}

	log.Println("Successful updating")
}

func getOrCreateUser(ctx context.Context, client *database_client.Client, callbackQuery *tgbotapi.CallbackQuery) (*db.User, error) {
	user, err := client.GetUser(ctx, callbackQuery.From.UserName)
	if err != nil || user == nil {
		chatID := callbackQuery.Message.Chat.ID
		newUser := database_client.NewUser(callbackQuery.From.UserName, chatID)
		err = client.CreateUser(ctx, newUser)
		if err != nil {
			return nil, err
		}
		user, err = client.GetUser(ctx, callbackQuery.From.UserName)
		if err != nil {
			return nil, err
		}
	}
	return user, nil
}

func parseTrafficFromCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) (float64, error) {
	tokens := strings.Split(callbackQuery.Data, ",")
	itagNo := tokens[1]
	for _, row := range callbackQuery.Message.ReplyMarkup.InlineKeyboard {
		for _, keyboardButton := range row {
			tokens := strings.Split(*keyboardButton.CallbackData, ",")
			itag := tokens[len(tokens)-1]
			if itagNo == itag {
				tokens = strings.Split(keyboardButton.Text, ",")
				tokens = strings.Split(tokens[len(tokens)-1], " ")
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

func checkTraffic(client *database_client.Client, callbackQuery *tgbotapi.CallbackQuery, format *youtube.Format) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	user, err := client.GetUser(ctx, callbackQuery.From.UserName)
	if err != nil {
		log.Printf("can't get user by username: %s, error: %s", callbackQuery.Message.From.UserName, err.Error())
		return true
	} else if user == nil {
		log.Printf("Get nil user: %s", callbackQuery.Message.From.UserName)
		return true
	}
	fileSize, err := getFileSize(*format) // bite
	fileSize = fileSize / (1024 * 1024)   // Mb
	if err != nil {
		log.Printf("can't file size: %s", err.Error())
	}
	if user.Traffic+fileSize > TrafficLimit && user.Subscription.SubscriptionStatus != "active" {
		return false
	}
	return true
}
