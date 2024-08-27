package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"youtube_downloader/pkg/bot/tg"
)

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

	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN must be set")
	}

	host := os.Getenv("HOST") // telegram-bot-api:8081 (docker network) || localhost:8081 || empty "" for default without botAPIEndpoint
	var botAPIEndpoint string
	var botAPI *tgbotapi.BotAPI
	if host != "" {
		botAPIEndpoint = fmt.Sprintf("http://%s/bot%%s/%%s", host)
		botAPI, err = tgbotapi.NewBotAPIWithAPIEndpoint(botToken, botAPIEndpoint)
	} else {
		botAPI, err = tgbotapi.NewBotAPI(botToken)
	}

	if err != nil {
		log.Fatal(err)
	}

	tgBot := tg.GetBotInstance(botAPI)
	tgBot.SetCommands()
	if err := tgBot.StartBot(); err != nil {
		log.Fatal(err)
	}
}
