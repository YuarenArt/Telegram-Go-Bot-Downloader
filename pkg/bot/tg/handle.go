package tg

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
	"time"
	"youtube_downloader/pkg/bot/tg/handler"
	"youtube_downloader/pkg/bot/tg/handler/youtube"
	"youtube_downloader/pkg/bot/tg/send"
	database_client "youtube_downloader/pkg/database-client"
)

// handleUpdates gets updates from telegramAPI and handles it
func (tb *TgBot) handleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		if err := tb.ensureUserExists(ctx, update.Message); err != nil {
			log.Println(err)
		}

		switch {
		case update.Message != nil && update.Message.SuccessfulPayment == nil:
			if update.Message.IsCommand() {
				tb.handleCommand(update.Message)
				continue
			}
			tb.handleMessage(update.Message)
		case update.CallbackQuery != nil:
			tb.handleCallbackQuery(update.CallbackQuery)
		case update.PreCheckoutQuery != nil:
			tb.handlePreCheckoutQuery(update.PreCheckoutQuery)
		case update.Message != nil && update.Message.SuccessfulPayment != nil:
			tb.handleSuccessfulPayment(update.Message)
		default:
			log.Println("unknown user's message")
			tb.handleDefaultCommand(update.Message, update.Message.From.LanguageCode)
		}
	}
}

func (tb *TgBot) handlePreCheckoutQuery(preCheckoutQuery *tgbotapi.PreCheckoutQuery) {
	preCheckoutConfig := tgbotapi.PreCheckoutConfig{
		PreCheckoutQueryID: preCheckoutQuery.ID,
		OK:                 true,
	}
	if _, err := tb.Bot.Request(preCheckoutConfig); err != nil {
		log.Println("Error handling pre-checkout query:", err)
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
		lang := message.From.LanguageCode
		if err != nil {
			log.Print(err)
			errMsg := err.Error()
			if errMsg == "Request Entity Too Large" {
				fileTooLarge := tb.translations[lang]["fileTooLarge"]
				send.SendReplyMessage(tb.Bot, message, &fileTooLarge)
			} else if errMsg == "extractVideoID failed: invalid characters in video id" {
				invalidLink := tb.translations[lang]["invalidLink"]
				send.SendReplyMessage(tb.Bot, message, &invalidLink)
			} else {
				somethingWentWrong := tb.translations[lang]["somethingWentWrong"]
				send.SendReplyMessage(tb.Bot, message, &somethingWentWrong)
			}
			return
		}

		if strings.HasPrefix(message.Text, "https://www.youtube.com/live/") {
			videoURL := youtube.FormatYouTubeURLOnStream(message.Text)
			translations := tb.translations[lang]
			send.SendKeyboardMessageReplyWithFormattedLink(tb.Bot, message, keyboard, videoURL, translations)
		} else {
			translations := tb.translations[lang]
			send.SendKeyboardMessageReply(tb.Bot, message, keyboard, &translations)
		}

	default:
		tb.handleDefaultCommand(message, message.From.LanguageCode)
		tb.handleHelpCommand(message, message.From.LanguageCode)
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
		newUser := database_client.NewUser(username, message.Chat.ID)
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
