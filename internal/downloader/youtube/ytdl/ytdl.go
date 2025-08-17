package ytdl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"youtube_downloader/internal/downloader/youtube"
)

// YTDLBackend orchestrates video and playlist metadata retrieval and downloading.
type YTDLBackend struct {
	executor *CommandExecutor
	parser   *MetadataParser
}

// NewYTDLBackend creates a new YTDLBackend with CommandExecutor and MetadataParser.
func NewYTDLBackend() *YTDLBackend {
	return &YTDLBackend{
		executor: NewCommandExecutor(NewCommandRunner()),
		parser:   NewMetadataParser(),
	}
}

// GetVideo retrieves metadata for a single video URL.
func (y *YTDLBackend) GetVideo(ctx context.Context, videoURL string) (*youtube.Video, error) {
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
	// Build and execute download command
	outputPath, err := y.executor.RunDownload(ctx, video.SourceURL, opts)
	if err != nil {
		return "", fmt.Errorf("failed to download video: %w", err)
	}

	return outputPath, nil
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

// fileExists return true if file with filePath exists
func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
