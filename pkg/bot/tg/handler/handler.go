package handler

import (
	"youtube_downloader/pkg/bot/tg/handler/youtube"
	database_client "youtube_downloader/pkg/database-client"
	downloader_youtube "youtube_downloader/pkg/downloader/youtube"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HandlerType int

const (
	YoutubeHandler HandlerType = iota
)

var SupportedHandlers = []HandlerType{
	YoutubeHandler,
}

type Handler interface {
	HandleMessage(message *tgbotapi.Message) (*tgbotapi.InlineKeyboardMarkup, error)
	HandleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, translations *map[string]string)
}

// CreateHandler creates a new handler of the specified type
// cookiesPath is optional. If empty, cookies will not be used.
func CreateHandler(handlerType HandlerType, downloader downloader_youtube.Downloader, client *database_client.Client, cookiesPath string) Handler {
	switch handlerType {
	case YoutubeHandler:
		return youtube.NewYoutubeHandler(downloader, client, cookiesPath)
	default:
		return nil
	}
}
