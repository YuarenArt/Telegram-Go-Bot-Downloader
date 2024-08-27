package tg

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
	. "youtube_downloader/pkg/bot/tg/handler"
	"youtube_downloader/pkg/bot/tg/send"
)

// handleCallbackQuery gets url from Bot's message with a replying link,
// then handle a link by its type: video (stream), playlist
func (tb *TgBot) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {

	data := callbackQuery.Data
	parts := strings.Split(data, ",")
	data = parts[0]

	switch {
	case strings.HasPrefix(data, "pay_"):
		subscriptionType := strings.TrimPrefix(data, "pay_")
		tb.processPayment(callbackQuery.Message, subscriptionType)
	case isYoutubeLink(data):
		tb.handlers[YoutubeHandler].HandleCallbackQuery(callbackQuery, tb.Bot, tb.Client)
	default:
		log.Printf("handleCallbackQuery get default case with %s link", data)
		send.SendReplyMessage(tb.Bot, callbackQuery.Message, "Something went wrong")
	}
}
