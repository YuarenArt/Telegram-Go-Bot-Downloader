package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	config "youtube_downloader/internal/config"
	"youtube_downloader/pkg/bot/tg"
)

// TgBotApp holds the application state and resources.
type TgBotApp struct {
	bot     *tg.TgBot
	cleanup func()
	errChan chan error
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewApp initializes the application with configuration, bot, and context.
func NewApp(cfg *config.Config) (*TgBotApp, error) {

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	botAPI, err := createBotAPI(cfg)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	tgBot := tg.BotInstance(botAPI)
	tgBot.SetCommands()

	return &TgBotApp{
		bot:     tgBot,
		cleanup: func() {},
		errChan: make(chan error, 1),
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

// createBotAPI creates a new bot API instance based on configuration
func createBotAPI(cfg *config.Config) (*tgbotapi.BotAPI, error) {
	if cfg.TelegramBotToken == "" {
		return nil, fmt.Errorf("telegram bot token must be provided")
	}

	var botAPI *tgbotapi.BotAPI
	var err error

	if cfg.APIEndpoint != "" {
		botAPI, err = tgbotapi.NewBotAPIWithAPIEndpoint(cfg.TelegramBotToken,
			fmt.Sprintf("http://%s/bot%%s/%%s", cfg.APIEndpoint))
	} else {
		botAPI, err = tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	}

	return botAPI, err
}

// Run starts the bot and handles shutdown.
func (a *TgBotApp) Run() error {
	defer a.cleanup()
	defer a.gracefulShutdown()

	go func() {
		if err := a.bot.Start(); err != nil {
			a.errChan <- fmt.Errorf("bot error: %w", err)
		}
	}()

	select {
	case <-a.ctx.Done():
		log.Println("Shutdown signal received")
		return nil
	case err := <-a.errChan:
		return err
	}
}

// gracefulShutdown stops the bot and cleans up resources.
func (a *TgBotApp) gracefulShutdown() {
	a.cancel()

	if a.bot != nil {
		a.bot.Stop()
	}
	close(a.errChan)
}
