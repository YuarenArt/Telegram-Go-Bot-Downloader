package youtube

import (
	"context"
	"fmt"
	"github.com/kkdai/youtube/v2"
	"log"
)

// DownloadVideoWithFormat download a video according to a format
func (ytd *YouTubeDownloader) DownloadVideoWithFormat(
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

// DownloadVideo downloads video with the lowest quality
func (ytd *YouTubeDownloader) DownloadVideo(video *youtube.Video) (pathAndName string, err error) {
	title := cleanVideoTitle(video.Title)
	pathAndName = DOWNLOAD_PREFIX + title + FORMAT_MP4

	formats := video.Formats.WithAudioChannels()
	formats, err = WithFormats(&formats, VIDEO_PREFIX)
	if err != nil {
		log.Printf("failed to get %s formats: %w", VIDEO_PREFIX, err)
	}
	formats.Sort()
	format := formats[len(formats)-1]

	ctx := context.Background()
	if err := ytd.DownloadVideoWithFormat(ctx, video, &format, ""); err != nil {
		fmt.Println(err)
	}

	return pathAndName, nil
}

// DownloadAudio downloads audio with the highest quality
func (ytd *YouTubeDownloader) DownloadAudio(video *youtube.Video) (pathAndName string, err error) {
	title := cleanVideoTitle(video.Title)

	formats := video.Formats.WithAudioChannels()
	formats, err = WithFormats(&formats, AUDIO_PREFIX)
	if err != nil {
		log.Printf("failed to get %s formats: %w", VIDEO_PREFIX, err)
	}
	formats.Sort()
	format := formats[0]
	ctx := context.Background()
	if err := ytd.DownloadVideoWithFormat(ctx, video, &format, ""); err != nil {
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

// DownloadWithFormat downloads a file by a link with a certain video format
func (ytd *YouTubeDownloader) DownloadWithFormat(videoURL string, format youtube.Format) (pathAndName string, err error) {
	if !isAcceptableFileSize(format) {
		return "", fmt.Errorf("file's size too large. Acceptable size is %.2f Mb", MaxFileSize/(1024*1024))
	}

	video, err := ytd.GetVideo(videoURL)
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

	if fileExists(pathAndName) {
		log.Print("File already exists, skipping download")
		return pathAndName, nil
	}

	ctx := context.Background()
	if err := ytd.DownloadVideoWithFormat(ctx, video, &format, ""); err != nil {
		fmt.Println(err)
	}

	// changes any extension except .mp4 to .mp3
	if fileFormat != ".mp4" {
		if err := changeFileExtensionToMp3(pathAndName); err != nil {
			return "", fmt.Errorf("can't rename file: %v", err)
		}
		pathAndName = DOWNLOAD_PREFIX + title + ".mp3"
	}

	return pathAndName, nil
}
