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
)

// handleUpdates gets updates from telegramAPI and handles it
func (tb *TgBot) handleUpdates(updates tgbotapi.UpdatesChannel) {
	for {
		select {
		case <-tb.ctx.Done():
			return
		case update, ok := <-updates:
			if !ok {
				return
			}

			tb.wg.Add(1)
			go func(update tgbotapi.Update) {
				defer tb.wg.Done()

				ctx, cancel := context.WithTimeout(tb.ctx, 1*time.Minute)
				defer cancel()

				if update.Message != nil {
					if err := tb.ensureUserExists(ctx, update.Message); err != nil {
						log.Println(err)
					}
				}

				switch {
				case update.Message != nil && update.Message.SuccessfulPayment == nil:
					if update.Message.IsCommand() {
						tb.handleCommand(update.Message)
						return
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
					if update.Message != nil && update.Message.From != nil {
						tb.handleDefaultCommand(update.Message, update.Message.From.LanguageCode)
					}
				}
			}(update)
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

// ensureUserExists checks if a user exists in the database and creates it if not.
func (tb *TgBot) ensureUserExists(ctx context.Context, message *tgbotapi.Message) error {
	if message == nil || message.From == nil {
		log.Println("empty message or missing From field while ensureUserExists")
		return fmt.Errorf("invalid message")
	}

	username := message.From.UserName
	if username == "" {
		return fmt.Errorf("username is required")
	}

	exist, err := tb.Client.IsUserExist(ctx, username)
	if err != nil {
		return fmt.Errorf("error checking if user exists: %w", err)
	}

	if !exist {
		// Create user directly with username and chat ID
		_, err := tb.Client.CreateUser(ctx, username, message.Chat.ID)
		if err != nil {
			return fmt.Errorf("error creating new user: %w", err)
		}
		log.Printf("Created new user: %s", username)
	}
	return nil
}

func isYoutubeLink(link string) bool {
	link = strings.TrimSpace(link)
	return strings.HasPrefix(link, "https://www.youtube.com") ||
		strings.HasPrefix(link, "https://youtube.com") ||
		strings.HasPrefix(link, "https://youtu.be")
}
