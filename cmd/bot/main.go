package main

import (
	"log"
	"youtube_downloader/pkg/tg_bot"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NewBotAPI("6979900763:AAFH_B1QpdIJXA87LXTRqwvhxgji8LAm9g4")
	if err != nil {
		log.Panic(err)
	}

	tgBot := tg_bot.NewBot(bot)
	if err := tgBot.StartBot(); err != nil {
		log.Fatal(err)
	}
}
