package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
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
	// start prof
	cpuFile, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer cpuFile.Close()

	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	memFile, err := os.Create("mem.prof")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer memFile.Close()

	defer func() {
		runtime.GC() // Принудительный запуск сборщика мусора для получения актуальных данных
		if err := pprof.WriteHeapProfile(memFile); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}()

	// end prof

	initConfig()

	botToken := viper.GetString("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN must be set")
	}

	//botAPIURL := "http://telegram-bot-api:8081/bot%s/%s"
	//botAPIURL := "http://localhost:8081/bot%s/%s"
	//bot, err := tgbotapi.NewBotAPIWithAPIEndpoint(botToken, botAPIURL)

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	tgBot := tg.NewBot(bot)
	if err := tgBot.StartBot(); err != nil {
		log.Fatal(err)
	}
}
