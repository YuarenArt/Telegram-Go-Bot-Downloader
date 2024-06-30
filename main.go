package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
	"log"
	"youtube_downloader/pkg/bot/tg"
)

func initConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
}

func main() {
	initConfig()

	botToken := viper.GetString("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN must be set")
	}

	botAPIURL := "http://localhost:8081/bot%s/%s"
	bot, err := tgbotapi.NewBotAPIWithAPIEndpoint(botToken, botAPIURL)

	//bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	tgBot := tg.NewBot(bot)
	if err := tgBot.StartBot(); err != nil {
		log.Fatal(err)
	}
}
