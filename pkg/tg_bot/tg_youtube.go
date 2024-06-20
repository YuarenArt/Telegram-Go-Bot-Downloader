package tg_bot

import (
	"context"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kkdai/youtube/v2"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	youtube_downloader "youtube_downloader/pkg/downloader/youtube-downloader"
)

const (
	DOWNLOAD_PREFIX = "download/"

	VIDEO_PREFIX = "video/"
	AUDIO_PREFIX = "audio/"

	FORMAT_MP4 = ".mp4"
	FORMAT_MP3 = ".mp3"
)

func (tb *TgBot) downloadVideo(video *youtube.Video) (pathAndName string, err error) {

	dl := youtube_downloader.NewYouTubeDownloader()

	title := cleanVideoTitle(video.Title)
	pathAndName = DOWNLOAD_PREFIX + title + FORMAT_MP4

	formats := video.Formats.WithAudioChannels()
	formats, err = youtube_downloader.WithFormats(&formats, VIDEO_PREFIX)
	if err != nil {
		log.Printf("failed to get %s formats: %w", VIDEO_PREFIX, err)
	}
	formats.Sort()
	format := formats[len(formats)-1]

	ctx := context.Background()
	if err := dl.DownloadVideoWithFormat(ctx, video, &format, ""); err != nil {
		fmt.Println(err)
	}

	return pathAndName, nil
}

func (tb *TgBot) downloadAudio(video *youtube.Video) (pathAndName string, err error) {
	dl := youtube_downloader.NewYouTubeDownloader()

	title := cleanVideoTitle(video.Title)

	formats := video.Formats.WithAudioChannels()
	formats, err = youtube_downloader.WithFormats(&formats, AUDIO_PREFIX)
	if err != nil {
		log.Printf("failed to get %s formats: %w", VIDEO_PREFIX, err)
	}

	format := formats[0]
	ctx := context.Background()
	if err := dl.DownloadVideoWithFormat(ctx, video, &format, ""); err != nil {
		fmt.Println(err)
	}

	fileFormat, err := getFormatByMimeType(format.MimeType)
	pathAndName = DOWNLOAD_PREFIX + title + fileFormat

	// changes any extension except .mp4 to .mp3
	if fileFormat != ".mp4" {
		if err = changeFileExtensionToMp3(DOWNLOAD_PREFIX + title + fileFormat); err != nil {
			log.Println("can't rename file: " + err.Error())
		} else {
			fileFormat = ".mp3"
			pathAndName = DOWNLOAD_PREFIX + title + fileFormat
		}
	}

	return pathAndName, nil
}

// downloadWithFormat download a file by a link with a certain video format
func (tb *TgBot) downloadWithFormat(videoURL string, format youtube.Format) (pathAndName string, err error) {

	if !isAcceptableFileSize(format) {
		return "", fmt.Errorf("file's size to large. Acceptable size is %.2f Mb", maxFileSize/(1024*1024))
	}

	dl := youtube_downloader.NewYouTubeDownloader()

	video, err := dl.GetVideo(videoURL)
	if err != nil {
		log.Print(err)
		return "", err
	}

	title := cleanVideoTitle(video.Title)

	fileFormat, err := getFormatByMimeType(format.MimeType)
	if err != nil {
		return "", err
	}

	pathAndName = DOWNLOAD_PREFIX + title + fileFormat

	if tb.fileExists(pathAndName) {
		log.Print("File already exists, skipping download")
		return pathAndName, nil
	}

	ctx := context.Background()
	if err := dl.DownloadVideoWithFormat(ctx, video, &format, ""); err != nil {
		fmt.Println(err)
	}

	// changes any extension except .mp4 to .mp3
	if fileFormat != ".mp4" {
		if err = changeFileExtensionToMp3(DOWNLOAD_PREFIX + title + fileFormat); err != nil {
			log.Println("can't rename file: " + err.Error())
		} else {
			fileFormat = ".mp3"
			pathAndName = DOWNLOAD_PREFIX + title + fileFormat
		}
	}

	return pathAndName, nil
}

// sendFile send file according its type
func (tb *TgBot) sendFile(message *tgbotapi.Message, filePath string) error {

	switch filepath.Ext(filePath) {
	case ".mp4":
		return tb.sendVideo(message.Chat.ID, message.MessageID, filePath)
	case ".weba", ".mp3", ".m4a":
		return tb.sendAudio(message.Chat.ID, message.MessageID, filePath)
	default:
		return errors.New("unknown extension")
	}
}

// sendVideo sends to user video by chatID and MessageID
func (tb *TgBot) sendVideo(chatID int64, MessageID int, filePath string) error {

	log.Print("Start sending: " + filePath)

	video := tgbotapi.NewVideo(chatID, tgbotapi.FilePath(filePath))
	video.ReplyToMessageID = MessageID

	videoName := path.Base(filePath)
	video.Caption = videoName

	_, err := tb.bot.Send(video)
	if err != nil {
		log.Printf("Can't send file: %w", err.Error())
		return err
	}
	log.Print("Video has sent!")
	return err
}

func (tb *TgBot) sendAudio(chatID int64, MessageID int, filePath string) error {

	log.Print("Start sending: " + filePath)

	audio := tgbotapi.NewAudio(chatID, tgbotapi.FilePath(filePath))
	audio.ReplyToMessageID = MessageID

	audioName := path.Base(filePath)
	audio.Caption = audioName

	_, err := tb.bot.Send(audio)
	if err != nil {
		log.Printf("Can't send file: %w", err.Error())
		return err
	}
	log.Print("Audio has sent!")
	return err
}

// fileExists return true if file with filePath exist
func (tb *TgBot) fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// delete all unacceptable symbols for Mac, Windows, Ubuntu file system
func cleanVideoTitle(title string) string {
	title = regexp.MustCompile(`[/\\:*?"<>|]`).ReplaceAllString(title, "")
	title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")

	return title
}

// return a format of file (.mp4, .m4a, .weba) according to a mimeType
func getFormatByMimeType(mimeType string) (string, error) {
	switch {
	case strings.HasPrefix(mimeType, "video/mp4"):
		return ".mp4", nil
	case strings.HasPrefix(mimeType, "audio/mp4"):
		return ".m4a", nil
	case strings.HasPrefix(mimeType, "audio/webm"):
		return ".weba", nil
	default:
		return "", fmt.Errorf("unsupported mime type: %s", mimeType)
	}
}

func changeFileExtensionToMp3(filePath string) error {
	fileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	fileDir := filepath.Dir(filePath)
	newFilePath := fileDir + "/" + fileName + ".mp3"
	return os.Rename(filePath, newFilePath)
}
