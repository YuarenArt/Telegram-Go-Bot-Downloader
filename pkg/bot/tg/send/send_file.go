package send

import (
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"path"
	"path/filepath"
)

// sendFile send file according its type
func SendFile(bot *tgbotapi.BotAPI, message *tgbotapi.Message, filePath string) error {

	switch filepath.Ext(filePath) {
	case ".mp4":
		return sendVideo(bot, message.Chat.ID, message.MessageID, filePath)
	case ".weba", ".mp3", ".m4a":
		return sendAudio(bot, message.Chat.ID, message.MessageID, filePath)
	default:
		return errors.New("unknown extension")
	}
}

// sendVideo sends to user video by chatID and MessageID
func sendVideo(bot *tgbotapi.BotAPI, chatID int64, MessageID int, filePath string) error {

	log.Print("Start sending: " + filePath)

	video := tgbotapi.NewVideo(chatID, tgbotapi.FilePath(filePath))
	video.ReplyToMessageID = MessageID

	videoName := path.Base(filePath)
	video.Caption = videoName

	_, err := bot.Send(video)
	if err != nil {
		log.Printf("Can't send file: %w", err.Error())
		return err
	}
	log.Print("Video has sent!")
	return err
}

// sendAudio sends to user audio by chatID and MessageID
func sendAudio(bot *tgbotapi.BotAPI, chatID int64, MessageID int, filePath string) error {

	log.Print("Start sending: " + filePath)

	audio := tgbotapi.NewAudio(chatID, tgbotapi.FilePath(filePath))
	audio.ReplyToMessageID = MessageID

	audioName := path.Base(filePath)
	audio.Caption = audioName

	_, err := bot.Send(audio)
	if err != nil {
		log.Printf("Can't send file: %w", err.Error())
		return err
	}
	log.Print("Audio has sent!")
	return err
}
