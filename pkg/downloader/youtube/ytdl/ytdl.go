package ytdl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"youtube_downloader/pkg/downloader/youtube"
)

// YTDLBackend orchestrates video and playlist metadata retrieval and downloading.
type YTDLBackend struct {
	executor *CommandExecutor
	parser   *MetadataParser
}

// NewYTDLBackend creates a new YTDLBackend with CommandExecutor and MetadataParser.
// cookiesPath is optional. If empty, cookies will not be used.
func NewYTDLBackend(cookiesPath string) *YTDLBackend {
	return &YTDLBackend{
		executor: NewCommandExecutor(NewCommandRunner(), cookiesPath),
		parser:   NewMetadataParser(),
	}
}

// GetVideo retrieves metadata for a single video URL.
func (y *YTDLBackend) GetVideo(ctx context.Context, videoURL string) (*youtube.Video, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}
	if videoURL == "" {
		return nil, fmt.Errorf("video URL cannot be empty")
	}

	// Execute youtube-dl command to get metadata
	output, err := y.executor.RunVideoMetadata(ctx, videoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get video metadata: %w", err)
	}

	// Parse JSON into Video struct
	video, err := y.parser.ParseVideo(output, videoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse video metadata: %w", err)
	}

	return video, nil
}

// GetPlaylist retrieves metadata for a playlist URL.
func (y *YTDLBackend) GetPlaylist(ctx context.Context, playlistURL string) (*youtube.Playlist, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}
	if playlistURL == "" {
		return nil, fmt.Errorf("playlist URL cannot be empty")
	}

	// Execute youtube-dl command to get playlist metadata
	output, err := y.executor.RunPlaylistMetadata(ctx, playlistURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist metadata: %w", err)
	}

	// Parse JSON into Playlist struct
	playlist, err := y.parser.ParsePlaylist(output, playlistURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse playlist metadata: %w", err)
	}

	return playlist, nil
}

// Download fetches the video or audio file according to the provided options.
func (y *YTDLBackend) Download(ctx context.Context, video *youtube.Video, opts youtube.DownloadOptions) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("context cannot be nil")
	}
	if video == nil {
		return "", fmt.Errorf("video cannot be nil")
	}
	if video.SourceURL == "" {
		return "", fmt.Errorf("video source URL cannot be empty")
	}

	// Validate and set default options
	if err := y.validateDownloadOptions(&opts); err != nil {
		return "", fmt.Errorf("invalid download options: %w", err)
	}

	// Build and execute download command
	outputPath, err := y.executor.RunDownload(ctx, video.SourceURL, opts)
	if err != nil {
		return "", fmt.Errorf("failed to download video: %w", err)
	}

	return outputPath, nil
}

// validateDownloadOptions validates and sets default values for download options
func (y *YTDLBackend) validateDownloadOptions(opts *youtube.DownloadOptions) error {
	if opts.OutputDir == "" {
		opts.OutputDir = "download"
	}

	if opts.Filename == "" {
		opts.Filename = "video"
	}

	// Ensure output directory exists
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return nil
}

// ChangeFileExtension changes the file extension to the specified one
func ChangeFileExtension(filePath, newExtension string) error {
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}
	if newExtension == "" {
		return fmt.Errorf("new extension cannot be empty")
	}

	// Ensure new extension starts with a dot
	if !strings.HasPrefix(newExtension, ".") {
		newExtension = "." + newExtension
	}

	// Check if the original file exists
	if !fileExists(filePath) {
		return fmt.Errorf("file %s does not exist", filePath)
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Check if it's a regular file
	if fileInfo.IsDir() {
		return fmt.Errorf("path %s is a directory, not a file", filePath)
	}

	// Build new file path
	fileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	fileDir := filepath.Dir(filePath)
	newFilePath := filepath.Join(fileDir, fileName+newExtension)

	// Check if target file already exists
	if fileExists(newFilePath) {
		// Remove existing file first
		if err := os.Remove(newFilePath); err != nil {
			return fmt.Errorf("failed to remove existing file %s: %w", newFilePath, err)
		}
	}

	// Rename the file
	if err := os.Rename(filePath, newFilePath); err != nil {
		return fmt.Errorf("failed to rename file from %s to %s: %w", filePath, newFilePath, err)
	}

	return nil
}

// fileExists returns true if a file with the given path exists
func fileExists(filePath string) bool {
	if filePath == "" {
		return false
	}

	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
