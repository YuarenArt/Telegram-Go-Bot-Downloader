package tg

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
	"youtube_downloader/pkg/bot/tg/handler"
	"youtube_downloader/pkg/bot/tg/send"
)

// handleCallbackQuery gets url from Bot's message with a replying link,
// then handle a link by its type: video (stream), playlist
func (tb *TgBot) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {

	text := callbackQuery.Message.Text
	parts := strings.Split(text, "\n")
	URL := parts[1]

	switch {
	case isYoutubeLink(URL):
		tb.handlers[handler.YoutubeHandler].HandleCallbackQuery(callbackQuery, tb.Bot)
	default:
		log.Printf("handleCallbackQuery get default case with %s link", URL)
		send.SendReplyMessage(tb.Bot, callbackQuery.Message, "Something went wrong")
	}
}
