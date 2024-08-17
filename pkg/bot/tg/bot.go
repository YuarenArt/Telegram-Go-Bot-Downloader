package tg

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"path/filepath"
	"youtube_downloader/pkg/bot/tg/handler"
	_ "youtube_downloader/pkg/database-client"
	database_client "youtube_downloader/pkg/database-client"
)

// TgBot uses telegram-Bot-api to maintain tg Bot
// It can download and send video with different formats (video/audio; quality) by handlers
type TgBot struct {
	Bot      *tgbotapi.BotAPI
	handlers []handler.Handler
	Client   *database_client.Client
}

// NewBot initializes a new TgBot instance with the given Telegram Bot API instance.
func NewBot(bot *tgbotapi.BotAPI) *TgBot {
	return &TgBot{
		Bot:    bot,
		Client: database_client.NewClient(bot.Token),
	}
}

// StartBot starts the Bot by authorizing it and initiating the update handling process.
func (tb *TgBot) StartBot() error {
	log.Printf("Authorized on account %s", tb.Bot.Self.UserName)
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if err = clearDownloadDirs(dir); err != nil {
		log.Println(err.Error())
	}

	tb.initSupportedHandlers()

	updates := tb.initUpdatesChannel()
	tb.handleUpdates(updates)

	return nil
}

// initSupportedHandlers initializes all supported handlers for the Telegram bot
// according to SupportedHandlers
func (tb *TgBot) initSupportedHandlers() {
	for _, handlerType := range handler.SupportedHandlers {
		handler := handler.CreateHandler(handlerType)
		tb.registerHandler(&handler)
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
