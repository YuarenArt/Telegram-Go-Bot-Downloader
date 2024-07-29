package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
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

// startProfiling initializes CPU and memory profiling and sets up signal handling for graceful shutdown.
func startProfiling(cpuProfile, memProfile string) (cleanup func(), err error) {
	// Start CPU profiling
	cpuFile, err := os.Create(cpuProfile)
	if err != nil {
		return nil, err
	}
	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		cpuFile.Close()
		return nil, err
	}

	// Start memory profiling
	memFile, err := os.Create(memProfile)
	if err != nil {
		cpuFile.Close()
		pprof.StopCPUProfile()
		return nil, err
	}

	// Handle system signals for graceful shutdown
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Println(sig)
		pprof.StopCPUProfile()
		cpuFile.Close()
		runtime.GC() // Forcing garbage collection to get accurate memory profile
		if err := pprof.WriteHeapProfile(memFile); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		memFile.Close()
		done <- true
	}()

	// Cleanup function to stop profiling and close files
	cleanup = func() {
		pprof.StopCPUProfile()
		cpuFile.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(memFile); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
		memFile.Close()
		done <- true
	}

	return cleanup, nil
}

func main() {

	cleanup, err := startProfiling("cpu.prof", "mem.prof")
	if err != nil {
		log.Fatal("Error starting profiling: ", err)
	}
	defer cleanup()

	initConfig()
	botToken := viper.GetString("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN must be set")
	}

	//botAPIURL := "http://telegram-bot-api:8081/bot%s/%s"
	//botAPIURL := "http://localhost:8081/bot%s/%s"
	//bot, err := tgbotapi.NewBotAPIWithAPIEndpoint(botToken, botAPIURL)

	botAPI, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	tgBot := tg.NewBot(botAPI)
	if err := tgBot.StartBot(); err != nil {
		log.Fatal(err)
	}
}
