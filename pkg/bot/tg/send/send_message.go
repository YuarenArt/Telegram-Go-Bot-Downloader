package send

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	DownloadingNotification = "⏳ Downloading... ⏳"
	SendingNotification     = " 🚀 Sending... 🚀"
)

// SendMessage just sends text message using BotAPI Send
func SendMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, text string) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	_, err := bot.Send(msg)
	return err
}

// SendReplyMessage sends text message reply to message id
func SendReplyMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, text string) (resp tgbotapi.Message, err error) {
	replyMessage := tgbotapi.NewMessage(message.Chat.ID, text)
	replyMessage.ReplyToMessageID = message.MessageID
	resp, err = bot.Send(replyMessage)
	return resp, err
}

// SendEditMessage edits a message by its id
func SendEditMessage(bot *tgbotapi.BotAPI, chatID int64, messageID int, text string) error {
	editMessage := tgbotapi.NewEditMessageText(chatID, messageID, text)
	_, err := bot.Send(editMessage)
	return err
}

// SendKeyboardMessageReply sends user a keyboard in reply
func SendKeyboardMessageReply(bot *tgbotapi.BotAPI, message *tgbotapi.Message, keyboard *tgbotapi.InlineKeyboardMarkup) error {
	msg := tgbotapi.NewMessage(message.Chat.ID,
		fmt.Sprintf("Your link:\n"+
			"%s"+
			"\nChoose a format or video:", message.Text),
	)
	msg.ReplyMarkup = keyboard
	_, err := bot.Send(msg)
	return err
}

// SendKeyboardMessage sends user a keyboard
func SendKeyboardMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, keyboard *tgbotapi.InlineKeyboardMarkup) error {
	msg := tgbotapi.NewMessage(message.Chat.ID,
		fmt.Sprintf("Choose a format for video:"),
	)
	msg.ReplyMarkup = keyboard
	_, err := bot.Send(msg)
	return err
}

func SendKeyboardMessageReplyWithFormattedLink(bot *tgbotapi.BotAPI, message *tgbotapi.Message, keyboard *tgbotapi.InlineKeyboardMarkup, videoURL string) error {
	msg := tgbotapi.NewMessage(message.Chat.ID,
		fmt.Sprintf("Your link:\n"+
			"%s"+
			"\nChoose a format:", videoURL),
	)
	msg.ReplyMarkup = keyboard
	_, err := bot.Send(msg)
	return err
}
