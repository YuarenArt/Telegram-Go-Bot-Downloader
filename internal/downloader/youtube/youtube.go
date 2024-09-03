package youtube

import (
	"errors"
	"fmt"
	. "github.com/kkdai/youtube/v2"
	"github.com/kkdai/youtube/v2/downloader"
	"log"
	"os"
	"path/filepath"
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
			Client:    Client{},
			OutputDir: DOWNLOAD_DIR,
		},
	}
}

// GetVideo retrieves a YouTube video by its URL and returns a pointer to a
// YouTube.Video struct that contains the video's metadata
func (ytd *YouTubeDownloader) GetVideo(url string) (*Video, error) {
	log.Printf("Getting video from URL: %s", url)
	return ytd.Downloader.Client.GetVideo(url)
}

// GetPlaylist playlist return Playlist struct
func (ytd *YouTubeDownloader) GetPlaylist(url string) (*Playlist, error) {
	log.Printf("Getting playlist from URL: %s", url)
	return ytd.Downloader.Client.GetPlaylist(url)
}

// GetVideoFromPlaylistEntry return certain Video from playlist
func (ytd *YouTubeDownloader) GetVideoFromPlaylistEntry(entry *PlaylistEntry) (*Video, error) {
	log.Printf("Getting video from playlist: %s", entry.Title)
	return ytd.Downloader.Client.VideoFromPlaylistEntry(entry)
}

// WithFormats returns a new FormatList that contains only a formats
// from the given list that have the following prefix (i.e "video/", "audio/")
func WithFormats(list *FormatList, prefix string) (FormatList, error) {
	var result FormatList

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
func FormatWithAudioChannels(videoURL string) (FormatList, error) {
	client := Client{}
	video, err := client.GetVideo(videoURL)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	formats := video.Formats.WithAudioChannels()
	return formats, nil
}

// FormatWithAudioChannelsComposite return a new FormatList
// that contains only a formats with audio in various quality combinations
func FormatWithAudioChannelsComposite(videoURL string) (FormatList, error) {
	client := Client{}
	video, err := client.GetVideo(videoURL)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	var formats FormatList
	uniqueFormats := make(map[int]bool)
	for _, format := range video.Formats {
		if _, exists := uniqueFormats[format.ItagNo]; !exists {
			formats = append(formats, format)
			uniqueFormats[format.ItagNo] = true
		}
	}

	return formats, nil
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

// return a format of file according to it's mimeType
func getFormatByMimeType(mimeType string) (string, error) {
	mimeTypeParts := strings.Split(mimeType, ";")
	format := canonicals[mimeTypeParts[0]]
	if format == "" {
		return "", errors.New("unknown format")
	}
	return format, nil
}

// ChangeFileExtension changes to the specified extension
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
func isAcceptableFileSize(format Format) bool {
	fileSize, _ := getFileSize(format)
	return fileSize < MaxFileSize
}

// getFileSize return a file size in bite of certain format
func getFileSize(format Format) (float64, error) {

	// get durations in secs
	duration, err := strconv.ParseFloat(format.ApproxDurationMs, 64)
	if err != nil {
		return 0, err
	}
	duration /= 1000

	// size in bite
	contentLength := float64(format.Bitrate/8) * duration

	return contentLength, nil
}
