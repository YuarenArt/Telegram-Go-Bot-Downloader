package send

import (
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	youtube_downloader "youtube_downloader/pkg/downloader/youtube"
)

// sendFile send file according its type
func SendFile(bot *tgbotapi.BotAPI, message *tgbotapi.Message, filePath string) error {

	switch filepath.Ext(filePath) {
	case ".mp4":
		return sendVideo(bot, message.Chat.ID, message.MessageID, filePath)
	case ".weba", ".mp3", ".m4a":
		log.Printf("Start sending audio with extension: %s", filepath.Ext(filePath))
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
		log.Printf("Can't send file: %s", err.Error())
		return err
	}
	log.Print("Video has sent!")
	return err
}

// sendAudio sends to user audio by chatID and MessageID
func sendAudio(bot *tgbotapi.BotAPI, chatID int64, MessageID int, filePath string) error {

	log.Print("Start sending: " + filePath)

	// in docker container audio files downloading with .mov extension(I don't know why),
	// so if it is true, we changed the extension on original
	if !fileExists(filePath) {
		log.Printf("sendAudio: filo by file pat not exist: %s", filePath)
		fileExtension := filepath.Ext(filePath)
		tmpFilePath := strings.TrimSuffix(filePath, fileExtension) + ".mov"

		if err := youtube_downloader.ChangeFileExtension(tmpFilePath, fileExtension); err != nil {
			log.Printf("Can't change extension for: %s", tmpFilePath)
			return err
		}

	}

	audio := tgbotapi.NewAudio(chatID, tgbotapi.FilePath(filePath))
	audio.ReplyToMessageID = MessageID

	audioName := path.Base(filePath)
	audio.Caption = audioName

	_, err := bot.Send(audio)
	if err != nil {
		log.Printf("Can't send file: %s", err.Error())
		return err
	}
	log.Print("Audio has sent!")
	return err
}

func fileExists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}
