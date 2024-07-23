package tg

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
	"youtube_downloader/pkg/bot/tg/handler"
	"youtube_downloader/pkg/bot/tg/handler/youtube"
	"youtube_downloader/pkg/bot/tg/send"
	database_client "youtube_downloader/pkg/database-client"
)

// handleUpdates gets updates from telegramAPI and handles it
func (tb *TgBot) handleUpdates(updates tgbotapi.UpdatesChannel) {

	for update := range updates {
		ctx := context.Background()
		if err := tb.ensureUserExists(ctx, update.Message); err != nil {
			log.Println(err)
		}

		switch {
		case update.Message != nil:
			if update.Message.IsCommand() {
				tb.handleCommand(update.Message)
				continue
			}
			tb.handleMessage(update.Message)
		case update.CallbackQuery != nil:
			tb.handleCallbackQuery(update.CallbackQuery)
		default:
			log.Println("unknown case")
			tb.handleDefaultCommand(update.Message)
		}

	}
}

// handleMessage processes an incoming message.
// If the message contains a command, it handles the command.
// If the message contains a link, it handles the link.
// Otherwise, it handles default and help commands.
func (tb *TgBot) handleMessage(message *tgbotapi.Message) {
	log.Printf("[%s] %s", message.From.UserName, message.Text)

	switch {
	case isYoutubeLink(message.Text):
		keyboard, err := tb.handlers[handler.YoutubeHandler].HandleMessage(message)
		if err != nil {
			log.Print(err)
			errMsg := err.Error()
			if errMsg == "Request Entity Too Large" {
				send.SendReplyMessage(tb.Bot, message, "Your file too large")
			} else if errMsg == "extractVideoID failed: invalid characters in video id" {
				send.SendReplyMessage(tb.Bot, message, "Your link incorrect. Just send a link")
			} else {
				send.SendReplyMessage(tb.Bot, message, "Something went wrong")
			}
		}

		if strings.HasPrefix(message.Text, "https://www.youtube.com/live/") {
			videoURL := youtube.FormatYouTubeURLOnStream(message.Text)
			send.SendKeyboardMessageReplyWithFormattedLink(tb.Bot, message, keyboard, videoURL)
		} else {
			send.SendKeyboardMessageReply(tb.Bot, message, keyboard)
		}

	default:
		tb.handleDefaultCommand(message)
		tb.handleHelpCommand(message)
	}
}

// createUserIfNotExists checks if a user exists in the database and creates it if not.
func (tb *TgBot) ensureUserExists(ctx context.Context, message *tgbotapi.Message) error {

	if message == nil {
		log.Println("empty message while ensureUserExists")
		return nil
	}

	username := message.From.UserName

	exist, err := tb.Client.IsUserExist(ctx, username)
	if err != nil {
		return fmt.Errorf("error checking if user exists: %w", err)
	}
	if !exist {
		newUser := database_client.NewUser(username)
		if err := tb.Client.CreateUser(ctx, newUser); err != nil {
			return fmt.Errorf("error creating new user: %w", err)
		}
	}
	return nil
}

func isYoutubeLink(link string) bool {
	link = strings.TrimSpace(link)
	return strings.HasPrefix(link, "https://www.youtube.com") ||
		strings.HasPrefix(link, "https://youtube.com") ||
		strings.HasPrefix(link, "https://youtu.be")
}
