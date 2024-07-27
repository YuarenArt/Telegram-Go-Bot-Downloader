package youtube

import (
	"errors"
	"fmt"
	"github.com/kkdai/youtube/v2"
	"github.com/kkdai/youtube/v2/downloader"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	DOWNLOAD_DIR = "download/"

	VIDEO_PREFIX = "video/"
	AUDIO_PREFIX = "audio/"

	FORMAT_MP4 = ".mp4"
	FORMAT_MP3 = ".mp3"

	MaxFileSize = 2147483648.0 // in bites (2 Gb)

)

var SupportedPrefixesFormat = []string{
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
			OutputDir: DOWNLOAD_DIR,
		},
	}
}

// GetVideo retrieves a YouTube video by its URL and returns a pointer to a
// YouTube.Video struct that contains the video's metadata
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

func (ytd *YouTubeDownloader) GetPlaylist(url string) (*youtube.Playlist, error) {
	return ytd.Downloader.Client.GetPlaylist(url)
}

func (ytd *YouTubeDownloader) GetVideoFromPlaylistEntry(entry *youtube.PlaylistEntry) (*youtube.Video, error) {
	return ytd.Downloader.Client.VideoFromPlaylistEntry(entry)
}

// WithFormats returns a new FormatList that contains only a formats
// from the given list that have the following prefix (i.e "video/", "audio/")
func WithFormats(list *youtube.FormatList, prefix string) (youtube.FormatList, error) {
	var result youtube.FormatList

	if !contains(SupportedPrefixesFormat, prefix) {
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

// FormatWithAudioChannels return a new FormatList that contains only a formats with audio
func FormatWithAudioChannels(videoURL string) (youtube.FormatList, error) {
	client := youtube.Client{}
	video, err := client.GetVideo(videoURL)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	formats := video.Formats.WithAudioChannels()
	return formats, nil
}

func FormatWithAudioChannelsComposite(videoURL string) (youtube.FormatList, error) {
	client := youtube.Client{}
	video, err := client.GetVideo(videoURL)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	var formats youtube.FormatList
	uniqueFormats := make(map[int]bool)
	for _, format := range video.Formats {
		if _, exists := uniqueFormats[format.ItagNo]; !exists {
			formats = append(formats, format)
			uniqueFormats[format.ItagNo] = true
		}
	}

	return formats, nil
}

func FormatWithAudioChannelsByVideo(video *youtube.Video) youtube.FormatList {
	formats := video.Formats.WithAudioChannels()
	return formats
}

// contains return true if item in slice
func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

// fileExists return true if file with filePath exists
func fileExists(filePath string) bool {
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
	format := canonicals[mimeType]
	if format == "" {
		return "", errors.New("unknown format")
	}
	return format, nil
}

// change any file extension on .mp3
func ChangeFileExtension(filePath, extension string) error {
	// Check if the file exists
	if _, err := os.Stat(filePath); err != nil {
		if !fileExists(filePath) {
			return fmt.Errorf("file %s does not exist", filePath)
		}
		return err
	}

	fileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	fileDir := filepath.Dir(filePath)
	newFilePath := filepath.Join(fileDir, fileName+extension)

	if err := os.Rename(filePath, newFilePath); err != nil {
		return fmt.Errorf("error renaming file: %s", err)
	}

	return nil
}

// isAcceptableFileSize return true if file size less than possible size to send to tg API
func isAcceptableFileSize(format youtube.Format) bool {
	fileSize, _ := getFileSize(format)
	return fileSize < MaxFileSize
}

// getFileSize return a file size in bite of certain format
func getFileSize(format youtube.Format) (float64, error) {

	// get durations in secs
	duration, err := strconv.ParseFloat(format.ApproxDurationMs, 64)
	if err != nil {
		return 0, err
	}
	duration /= 1000

	// get bitrate in bite\sec
	bitrate := format.Bitrate

	// size in bite
	contentLength := float64(bitrate/8) * duration

	return contentLength, nil
}
