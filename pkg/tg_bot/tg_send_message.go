package tg_bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *TgBot) sendMessage(message *tgbotapi.Message, text string) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	_, err := b.bot.Send(msg)
	return err
}

func (b *TgBot) sendReplyMessage(message *tgbotapi.Message, text string) (resp tgbotapi.Message, err error) {
	replyMessage := tgbotapi.NewMessage(message.Chat.ID, text)
	replyMessage.ReplyToMessageID = message.MessageID
	resp, err = b.bot.Send(replyMessage)
	return resp, err
}

func (b *TgBot) sendEditMessage(chatID int64, messageID int, text string) error {
	editMessage := tgbotapi.NewEditMessageText(chatID, messageID, text)
	_, err := b.bot.Send(editMessage)
	return err
}
