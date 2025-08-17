package ytdl

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"youtube_downloader/internal/downloader/youtube"
)

// CommandRunner executes external yt-dlp commands.
// It logs command, arguments, and output for debugging.
type CommandRunner struct{}

// NewCommandRunner creates a new CommandRunner.
func NewCommandRunner() *CommandRunner {
	return &CommandRunner{}
}

// Run executes the specified command with arguments, returning combined stdout/stderr or an error.
func (c *CommandRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	log.Printf("[yt-dlp] Running command: %s %v", name, args)
	cmd := exec.CommandContext(ctx, name, args...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	log.Printf("[yt-dlp] Command output: %s", output.String())
	if err != nil {
		if ctx.Err() != nil {
			log.Printf("[yt-dlp] Context error: %v", ctx.Err())
			return nil, ctx.Err()
		}
		log.Printf("[yt-dlp] Command error: %v", err)
		return output.Bytes(), err
	}
	return output.Bytes(), nil
}

// CommandExecutor handles execution of yt-dlp commands.
type CommandExecutor struct {
	client *CommandRunner
}

// NewCommandExecutor creates a new CommandExecutor.
func NewCommandExecutor(runner *CommandRunner) *CommandExecutor {
	return &CommandExecutor{client: runner}
}

// RunVideoMetadata executes yt-dlp --dump-json to get video metadata.
func (e *CommandExecutor) RunVideoMetadata(ctx context.Context, videoURL string) ([]byte, error) {
	args := []string{
		"--cookies", "cookies.txt",
		"--no-check-certificates",
		"--dump-json", videoURL,
	}
	log.Printf("[yt-dlp] Getting video metadata for: %s", videoURL)
	log.Printf("[yt-dlp] Using cookies file: cookies.txt")

	// Check if cookies file exists
	if _, err := os.Stat("cookies.txt"); os.IsNotExist(err) {
		log.Printf("[yt-dlp] WARNING: cookies.txt file not found!")
	} else {
		log.Printf("[yt-dlp] cookies.txt file found and will be used")
	}

	output, err := e.client.Run(ctx, "yt-dlp", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute yt-dlp --dump-json: %w", err)
	}
	return output, nil
}

// RunPlaylistMetadata executes yt-dlp --dump-single-json to get playlist metadata.
func (e *CommandExecutor) RunPlaylistMetadata(ctx context.Context, playlistURL string) ([]byte, error) {
	args := []string{
		"--cookies", "cookies.txt",
		"--no-check-certificates",
		"--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.101 Safari/537.36",
		"--dump-single-json", playlistURL,
	}
	log.Printf("[yt-dlp] Getting playlist metadata for: %s", playlistURL)
	log.Printf("[yt-dlp] Using cookies file: cookies.txt")

	// Check if cookies file exists
	if _, err := os.Stat("cookies.txt"); os.IsNotExist(err) {
		log.Printf("[yt-dlp] WARNING: cookies.txt file not found!")
	} else {
		log.Printf("[yt-dlp] cookies.txt file found and will be used")
	}

	output, err := e.client.Run(ctx, "yt-dlp", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute yt-dlp --dump-single-json: %w", err)
	}
	return output, nil
}

// RunDownload executes yt-dlp to download a video or audio file.
func (e *CommandExecutor) RunDownload(ctx context.Context, videoURL string, opts youtube.DownloadOptions) (string, error) {
	args := []string{
		"--cookies", "cookies.txt",
		"--no-check-certificates",
		"--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.101 Safari/537.36",
		"-f", fmt.Sprintf("%d", opts.FormatID),
		"-o", filepath.Join(opts.OutputDir, opts.Filename),
	}

	// Add audio extraction for audio-only formats
	if opts.AudioOnly {
		args = append(args, "--extract-audio")
		args = append(args, "--audio-format", "mp3")
	}

	if opts.Merge {
		args = append(args, "--merge-output-format", "mp4")
	}
	args = append(args, opts.ExtraArgs...)
	args = append(args, videoURL)
	log.Printf("[yt-dlp] Downloading: %s with options: %+v", videoURL, opts)
	log.Printf("[yt-dlp] Using cookies file: cookies.txt")

	// Check if cookies file exists
	if _, err := os.Stat("cookies.txt"); os.IsNotExist(err) {
		log.Printf("[yt-dlp] WARNING: cookies.txt file not found!")
	} else {
		log.Printf("[yt-dlp] cookies.txt file found and will be used")
	}

	_, err := e.client.Run(ctx, "yt-dlp", args...)
	if err != nil {
		return "", fmt.Errorf("failed to execute yt-dlp download: %w", err)
	}
	return filepath.Join(opts.OutputDir, opts.Filename), nil
}
