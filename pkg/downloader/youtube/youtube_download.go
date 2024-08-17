package youtube

import (
	"context"
	"fmt"
	"github.com/kkdai/youtube/v2"
	"log"
	"strings"
)

// DownloadVideoWithFormat download a video according to a format
func (ytd *YouTubeDownloader) DownloadVideoWithFormat(
	ctx context.Context,
	video *youtube.Video,
	format *youtube.Format,
	outputFile string) error {

	if err := ytd.Downloader.Download(ctx, video, format, outputFile); err != nil {
		log.Printf("Error after Download : %s", err)
		return err
	}
	return nil
}

// DownloadVideo downloads video with the lowest quality
func (ytd *YouTubeDownloader) DownloadVideo(video *youtube.Video) (pathAndName string, err error) {
	title := SanitizeFilename(video.Title)
	pathAndName = DOWNLOAD_DIR + title + FORMAT_MP4

	formats := video.Formats.WithAudioChannels()
	formats, err = WithFormats(&formats, VIDEO_PREFIX)
	if err != nil {
		log.Printf("failed to get %s formats: %s", VIDEO_PREFIX, err)
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

	formats := video.Formats.WithAudioChannels()
	formats, err = WithFormats(&formats, AUDIO_PREFIX)
	if err != nil {
		log.Printf("failed to get %s formats: %s", VIDEO_PREFIX, err)
	}
	formats.Sort()
	format := formats[0]
	if err := ytd.DownloadVideoWithFormat(context.Background(), video, &format, ""); err != nil {
		fmt.Println(err)
	}

	title := SanitizeFilename(video.Title)
	fileFormat, err := getFormatByMimeType(format.MimeType)
	pathAndName = DOWNLOAD_DIR + title + fileFormat

	return pathAndName, nil
}

// DownloadWithFormat downloads a file by a link with a certain video format
func (ytd *YouTubeDownloader) DownloadWithFormat(video *youtube.Video, format youtube.Format) (pathAndName string, err error) {
	if !isAcceptableFileSize(format) {
		return "", fmt.Errorf("file's size too large. Acceptable size is %.2f Mb", MaxFileSize/(1024*1024))
	}

	title := SanitizeFilename(video.Title)
	mimeType := format.MimeType
	mimeTypeParts := strings.Split(mimeType, ";")
	mimeType = mimeTypeParts[0]
	pathAndName = DOWNLOAD_DIR + title + canonicals[mimeType]

	err = ytd.DownloadVideoWithFormat(context.Background(), video, &format, "")
	if err != nil {
		log.Println(err)
		return pathAndName, err
	}

	log.Printf("DownloadWithFormat return path: %s", pathAndName)

	return pathAndName, nil
}

// DownloadWithFormatComposite downloads a file by a link with a certain video format and returns a path to file
func (ytd *YouTubeDownloader) DownloadWithFormatComposite(videoURL string, format youtube.Format) (pathAndName string, err error) {
	if !isAcceptableFileSize(format) {
		return "", fmt.Errorf("file's size too large. Acceptable size is %.2f Mb", MaxFileSize/(1024*1024))
	}

	video, err := ytd.GetVideo(videoURL)
	if err != nil {
		log.Print(err)
		return "", err
	}

	ctx := context.Background()
	pathAndName, err = ytd.DownloadVideoWithFormatComposite(ctx, "", video, format.QualityLabel, format.MimeType, "")
	if err != nil {
		log.Println(err)
		return pathAndName, err
	}

	log.Printf("DownloadWithFormat return path: %s", pathAndName)

	return pathAndName, nil
}
