package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"youtube_downloader/pkg/bot/tg/handler/youtube"
	database_client "youtube_downloader/pkg/database-client"
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
	HandleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, client *database_client.Client, translations *map[string]string)
}

func CreateHandler(handlerType HandlerType) Handler {
	switch handlerType {
	case YoutubeHandler:
		return youtube.NewYoutubeHandler()
	default:
		return nil
	}
}
