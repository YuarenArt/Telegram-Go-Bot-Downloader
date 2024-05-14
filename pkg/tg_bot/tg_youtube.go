package tg_bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kkdai/youtube/v2"
	"io"
	"log"
	"os"
	"path"
	"regexp"
)

const (
	DOWNLOAD_VIDEO_PREFIX = "download/video/"

	FORMAT_MP4 = ".mp4"
)

func (b *TgBot) downloadVideo(videoURL string) (pathAndName string, err error) {
	client := youtube.Client{}

	video, err := client.GetVideo(videoURL)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	// delete all unacceptable symbols
	re := regexp.MustCompile(`[/\\:*?"<>|]`)
	title := re.ReplaceAllString(video.Title, "")

	pathAndName = DOWNLOAD_VIDEO_PREFIX + title + FORMAT_MP4
	if b.fileExists(pathAndName) {
		log.Print("File already exists, skipping download")
		return pathAndName, nil
	}

	formats := video.Formats.WithAudioChannels() // only get videos with audio
	stream, _, err := client.GetStream(video, &formats[0])
	if err != nil {
		panic(err)
		return "", err
	}
	defer stream.Close()

	file, err := os.Create(pathAndName)
	if err != nil {
		panic(err)
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, stream)
	if err != nil {
		panic(err)
		return "", err
	}

	return pathAndName, nil
}

func (b *TgBot) sendVideo(chatID int64, MessageID int, videoPath string) error {

	log.Print("Start sending: " + videoPath)

	video := tgbotapi.NewVideo(chatID, tgbotapi.FilePath(videoPath))
	video.ReplyToMessageID = MessageID

	videoName := path.Base(videoPath)
	video.Caption = videoName

	_, err := b.bot.Send(video)
	log.Print("Video has sent!")
	return err
}

func (b *TgBot) fileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	return !os.IsNotExist(err)
}
