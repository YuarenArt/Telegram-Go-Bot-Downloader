package tg

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"youtube_downloader/pkg/bot/tg/handler"
	_ "youtube_downloader/pkg/database-client"
	database_client "youtube_downloader/pkg/database-client"
	kkdaiDownloader "youtube_downloader/pkg/downloader/youtube/ytdl"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TgBot uses telegram-Bot-api to maintain tg Bot
// It can download and send video with different formats (video/audio; quality) by handlers
type TgBot struct {
	Bot          *tgbotapi.BotAPI
	handlers     []handler.Handler
	Client       *database_client.Client
	translations map[string]map[string]string
	cookiesPath  string
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

var (
	instance *TgBot
	once     sync.Once
)

// LoadTranslations loads language files into memory
func (tb *TgBot) LoadTranslations() error {
	languages := []string{"en", "ru"}
	tb.translations = make(map[string]map[string]string)

	for _, lang := range languages {
		filePath := fmt.Sprintf("cmd/locales/%s.json", lang)
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("could not open translation file: %v", err)
		}
		defer file.Close()

		var translation map[string]string
		if err := json.NewDecoder(file).Decode(&translation); err != nil {
			return fmt.Errorf("could not decode translation file: %v", err)
		}

		tb.translations[lang] = translation
	}
	return nil
}

// newBot initializes a new TgBot instance with the given Telegram Bot API instance.
func newBot(bot *tgbotapi.BotAPI, cookiesPath, dbToken string) *TgBot {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize database client
	dbClient := database_client.NewClient(dbToken)

	return &TgBot{
		Bot:          bot,
		handlers:     make([]handler.Handler, 0),
		Client:       dbClient,
		translations: make(map[string]map[string]string),
		cookiesPath:  cookiesPath,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// BotInstance returns the singleton instance of TgBot.
// If the instance does not exist, it initializes it.
// cookiesPath is optional. If empty, cookies will not be used.
// dbToken is required for database operations.
func BotInstance(bot *tgbotapi.BotAPI, cookiesPath, dbToken string) *TgBot {
	once.Do(func() {
		instance = newBot(bot, cookiesPath, dbToken)
	})
	return instance
}

// Start starts the Bot by authorizing it and initiating the update handling process.
func (tb *TgBot) Start() error {
	log.Printf("Authorized on account %s", tb.Bot.Self.UserName)

	if err := tb.LoadTranslations(); err != nil {
		return fmt.Errorf("error loading translations: %w", err)
	}

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting working directory: %w", err)
	}

	if err = clearDownloadDirs(dir); err != nil {
		log.Printf("Warning: error clearing download dirs: %v", err)
	}

	tb.initSupportedHandlers()

	updates := tb.initUpdatesChannel()
	tb.wg.Add(1)
	go func() {
		defer tb.wg.Done()
		tb.handleUpdates(updates)
	}()

	return nil
}

// initSupportedHandlers initializes all supported handlers for the Telegram bot
// according to SupportedHandlers
func (tb *TgBot) initSupportedHandlers() {
	for _, handlerType := range handler.SupportedHandlers {
		var h handler.Handler
		switch handlerType {
		case handler.YoutubeHandler:
			ytDownloader := kkdaiDownloader.NewYTDLBackend(tb.cookiesPath)
			h = handler.CreateHandler(handlerType, ytDownloader, tb.Client, tb.cookiesPath)
		}
		tb.registerHandler(&h)
	}
}

// registerHandler registers a new handler to the TgBot
func (tb *TgBot) registerHandler(handler *handler.Handler) {
	tb.handlers = append(tb.handlers, *handler)
}

// initUpdatesChannel initializes the update channel for receiving updates from the Telegram server.
// It configures the update retrieval settings and returns the update channel.
func (tb *TgBot) initUpdatesChannel() tgbotapi.UpdatesChannel {
	update := tgbotapi.NewUpdate(0)
	update.Timeout = 60

	return tb.Bot.GetUpdatesChan(update)
}

// Stop gracefully stops the bot and waits for all operations to complete
func (tb *TgBot) Stop() {
	if tb.cancel != nil {
		tb.cancel()
	}
	tb.wg.Wait()
}

func clearDownloadDir() error {
	dir := "download"
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(0)
	if err != nil {
		return err
	}

	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}

	log.Println("Download dir is cleaned")
	return nil
}

func clearDownloadDirs(rootDir string) error {
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == "download" {
			files, err := os.ReadDir(path)
			if err != nil {
				return err
			}
			for _, file := range files {
				err = os.RemoveAll(filepath.Join(path, file.Name()))
				if err != nil {
					return err
				}
			}
			log.Printf("Download dir %s is cleaned\n", path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// SetCommands sets the commands for the bot
func (tb *TgBot) SetCommands() {
	commands := []tgbotapi.BotCommand{
		{Command: commandStart, Description: "Start the bot"},
		{Command: commandHelp, Description: "Get help"},
		{Command: commandPay, Description: "Subscribe to premium features"},
		{Command: commandStatus, Description: "Send user premium subscription status"},
	}

	config := tgbotapi.NewSetMyCommands(commands...)
	if _, err := tb.Bot.Request(config); err != nil {
		log.Println("Error setting bot commands:", err)
	}
}
