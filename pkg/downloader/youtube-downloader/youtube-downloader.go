package youtube_downloader

import (
	"context"
	"fmt"
	"github.com/kkdai/youtube/v2"
	"github.com/kkdai/youtube/v2/downloader"
	"log"
	"strings"
)

var SuppotedPrefixesFormat = []string{
	"video/",
	"audio/",
}

type YouTubeDownloader struct {
	Downloader downloader.Downloader
}

// SetDownloadDir sets dir to download
func (ytd *YouTubeDownloader) SetDownloadDir(dir string) {
	ytd.Downloader.OutputDir = dir
}

func NewYouTubeDownloader() *YouTubeDownloader {
	return &YouTubeDownloader{
		Downloader: downloader.Downloader{
			Client:    youtube.Client{},
			OutputDir: "",
		},
	}
}

// GetVideo retrieves a YouTube video by its URL and returns a pointer to a
// youtube.Video struct that contains the video's metadata
func (ytd *YouTubeDownloader) GetVideo(videoURL string) (*youtube.Video, error) {
	log.Printf("Getting video from URL: %s", videoURL)

	video, err := ytd.Downloader.Client.GetVideo(videoURL)
	if err != nil {
		log.Printf("Failed to get video from URL: %s, error: %s", videoURL, err)
		return nil, err
	}
	log.Printf("Got video: %s", video.Title)
	return video, err
}

func (ytd *YouTubeDownloader) DownloadVideo(
	ctx context.Context,
	video *youtube.Video,
	format *youtube.Format,
	outputFile string) error {

	if err := ytd.Downloader.Download(ctx, video, format, outputFile); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// WithFormats returns a new FormatList that contains only the formats
// from the given list that have the following prefix (i.e "video/", "audio/")
func WithFormats(list *youtube.FormatList, prefix string) (youtube.FormatList, error) {
	var result youtube.FormatList

	if !contains(SuppotedPrefixesFormat, prefix) {
		return nil, fmt.Errorf("unsupported prefix: %s", prefix)
	}

	for _, format := range *list {
		// If the format has video, add it to the result list.
		if strings.HasPrefix(format.MimeType, prefix) {
			result = append(result, format)
		}
	}
	return result, nil
}

func GetFormatWithAudioChannels(videoURL string) (youtube.FormatList, error) {
	client := youtube.Client{}
	video, err := client.GetVideo(videoURL)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	formats := video.Formats.WithAudioChannels()
	return formats, nil
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}
