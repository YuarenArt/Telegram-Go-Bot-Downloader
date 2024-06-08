package tg_bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// sendMessage just sends text message using BotAPI Send
func (tb *TgBot) sendMessage(message *tgbotapi.Message, text string) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	_, err := tb.bot.Send(msg)
	return err
}

// sendReplyMessage sends text message reply to message id
func (tb *TgBot) sendReplyMessage(message *tgbotapi.Message, text string) (resp tgbotapi.Message, err error) {
	replyMessage := tgbotapi.NewMessage(message.Chat.ID, text)
	replyMessage.ReplyToMessageID = message.MessageID
	resp, err = tb.bot.Send(replyMessage)
	return resp, err
}

// sendEditMessage edits a message by its id
func (tb *TgBot) sendEditMessage(chatID int64, messageID int, text string) error {
	editMessage := tgbotapi.NewEditMessageText(chatID, messageID, text)
	_, err := tb.bot.Send(editMessage)
	return err
}

// sendKeyboardMessage sends user a keyboard
func (tb *TgBot) sendKeyboardMessage(message *tgbotapi.Message, keyboard tgbotapi.InlineKeyboardMarkup) error {
	msg := tgbotapi.NewMessage(message.Chat.ID,
		fmt.Sprintf("Your link:\n"+
			"%s"+
			"\nChoose a format:", message.Text),
	)
	msg.ReplyMarkup = keyboard
	_, err := tb.bot.Send(msg)
	return err
}

func (tb *TgBot) sendKeyboardMessageWithFormattedLink(message *tgbotapi.Message, keyboard tgbotapi.InlineKeyboardMarkup, videoURL string) error {
	msg := tgbotapi.NewMessage(message.Chat.ID,
		fmt.Sprintf("Your link:\n"+
			"%s"+
			"\nChoose a format:", videoURL),
	)
	msg.ReplyMarkup = keyboard
	_, err := tb.bot.Send(msg)
	return err
}
