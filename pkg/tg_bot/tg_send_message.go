package tg_bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// sendMessage just sends text message using BotAPI Send
func (b *TgBot) sendMessage(message *tgbotapi.Message, text string) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	_, err := b.bot.Send(msg)
	return err
}

// sendReplyMessage sends text message reply to message id
func (b *TgBot) sendReplyMessage(message *tgbotapi.Message, text string) (resp tgbotapi.Message, err error) {
	replyMessage := tgbotapi.NewMessage(message.Chat.ID, text)
	replyMessage.ReplyToMessageID = message.MessageID
	resp, err = b.bot.Send(replyMessage)
	return resp, err
}

// sendEditMessage edits a message by its id
func (b *TgBot) sendEditMessage(chatID int64, messageID int, text string) error {
	editMessage := tgbotapi.NewEditMessageText(chatID, messageID, text)
	_, err := b.bot.Send(editMessage)
	return err
}

// sendKeyboardMessage sends user a keyboard
func (b *TgBot) sendKeyboardMessage(message *tgbotapi.Message, keyboard tgbotapi.InlineKeyboardMarkup) error {
	msg := tgbotapi.NewMessage(message.Chat.ID,
		fmt.Sprintf("Your link:\n"+
			"%s"+
			"\nChoose a format:", message.Text),
	)
	msg.ReplyMarkup = keyboard
	_, err := b.bot.Send(msg)
	return err
}
