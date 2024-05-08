package tg_bot

import (
	"github.com/kkdai/youtube/v2"
	"io"
	"log"
	"os"
)

const (
	DOWNLOAD_VIDEO_PREFIX = "download/video/"

	FORMAT_MP4 = "mp4"
)

func downloadVideo(videoURL string) {
	client := youtube.Client{}

	video, err := client.GetVideo(videoURL)
	if err != nil {
		log.Fatal(err)
	}

	formats := video.Formats.WithAudioChannels()           // only get videos with audio
	stream, _, err := client.GetStream(video, &formats[0]) // TODO: добавить информацию о размере файла
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	file, err := os.Create(DOWNLOAD_VIDEO_PREFIX + video.Title + FORMAT_MP4)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = io.Copy(file, stream)
	if err != nil {
		panic(err)
	}
}
