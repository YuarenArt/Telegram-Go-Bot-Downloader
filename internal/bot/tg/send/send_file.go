package send

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	youtube_downloader "youtube_downloader/internal/downloader/youtube/ytdl"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// SendFile send file according its type
func SendFile(bot *tgbotapi.BotAPI, message *tgbotapi.Message, filePath string) error {
	if bot == nil {
		return errors.New("bot cannot be nil")
	}
	if message == nil {
		return errors.New("message cannot be nil")
	}
	if filePath == "" {
		return errors.New("file path cannot be empty")
	}

	// Check if file exists
	if !fileExists(filePath) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	extension := strings.ToLower(filepath.Ext(filePath))

	switch extension {
	case ".mp4", ".avi", ".mov", ".mkv":
		return sendVideo(bot, message.Chat.ID, message.MessageID, filePath)
	case ".weba", ".mp3", ".m4a", ".wav", ".ogg":
		return sendAudio(bot, message.Chat.ID, message.MessageID, filePath)
	default:
		return fmt.Errorf("unsupported file extension: %s", extension)
	}
}

// sendVideo sends to user video by chatID and MessageID
func sendVideo(bot *tgbotapi.BotAPI, chatID int64, MessageID int, filePath string) error {
	if bot == nil {
		return errors.New("bot cannot be nil")
	}
	if filePath == "" {
		return errors.New("file path cannot be empty")
	}

	log.Printf("Start sending video: %s", filePath)

	video := tgbotapi.NewVideo(chatID, tgbotapi.FilePath(filePath))
	video.ReplyToMessageID = MessageID

	videoName := path.Base(filePath)
	video.Caption = videoName

	_, err := bot.Send(video)
	if err != nil {
		log.Printf("Can't send video file: %s", err.Error())
		return fmt.Errorf("failed to send video: %w", err)
	}

	log.Printf("Video sent successfully: %s", filePath)
	return nil
}

// sendAudio sends to user audio by chatID and MessageID
func sendAudio(bot *tgbotapi.BotAPI, chatID int64, MessageID int, filePath string) error {
	if bot == nil {
		return errors.New("bot cannot be nil")
	}
	if filePath == "" {
		return errors.New("file path cannot be empty")
	}

	log.Printf("Start sending audio: %s", filePath)

	// Handle file extension issues in Docker container
	actualFilePath, err := resolveAudioFilePath(filePath)
	if err != nil {
		return fmt.Errorf("failed to resolve audio file path: %w", err)
	}

	audio := tgbotapi.NewAudio(chatID, tgbotapi.FilePath(actualFilePath))
	audio.ReplyToMessageID = MessageID

	audioName := path.Base(actualFilePath)
	audio.Caption = audioName

	_, err = bot.Send(audio)
	if err != nil {
		log.Printf("Can't send audio file: %s", err.Error())
		return fmt.Errorf("failed to send audio: %w", err)
	}

	log.Printf("Audio sent successfully: %s", actualFilePath)
	return nil
}

// resolveAudioFilePath handles file extension issues in Docker containers
func resolveAudioFilePath(filePath string) (string, error) {
	// Check if the original file exists
	if fileExists(filePath) {
		return filePath, nil
	}

	// In Docker container, audio files might download with .mov extension
	// Try to find the file with different extensions
	basePath := strings.TrimSuffix(filePath, filepath.Ext(filePath))
	possibleExtensions := []string{".mov", ".m4a", ".weba", ".mp3"}

	for _, ext := range possibleExtensions {
		alternativePath := basePath + ext
		if fileExists(alternativePath) {
			log.Printf("Found audio file with alternative extension: %s", alternativePath)

			// Change extension to match expected format
			expectedExt := filepath.Ext(filePath)
			if err := youtube_downloader.ChangeFileExtension(alternativePath, expectedExt); err != nil {
				log.Printf("Warning: failed to change file extension from %s to %s: %v",
					alternativePath, expectedExt, err)
				// Continue with alternative path if extension change fails
				return alternativePath, nil
			}

			return basePath + expectedExt, nil
		}
	}

	return "", fmt.Errorf("audio file not found at %s or with alternative extensions", filePath)
}

// fileExists returns true if file with filePath exists
func fileExists(filePath string) bool {
	if filePath == "" {
		return false
	}

	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
