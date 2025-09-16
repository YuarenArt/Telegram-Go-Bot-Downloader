package ytdl

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"youtube_downloader/pkg/downloader/youtube"
)

// CommandRunner executes external yt-dlp commands.
type CommandRunner struct{}

// NewCommandRunner creates a new CommandRunner.
func NewCommandRunner() *CommandRunner {
	return &CommandRunner{}
}

// Run executes the specified command with arguments, returning combined stdout/stderr or an error.
func (c *CommandRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	if name == "" {
		return nil, fmt.Errorf("command name cannot be empty")
	}

	log.Printf("[yt-dlp] Running command: %s %v", name, args)

	cmd := exec.CommandContext(ctx, name, args...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()

	if err != nil {
		if ctx.Err() != nil {
			log.Printf("[yt-dlp] Context error: %v", ctx.Err())
			return nil, ctx.Err()
		}
		log.Printf("[yt-dlp] Command error: %v", err)
		return output.Bytes(), fmt.Errorf("command execution failed: %w", err)
	}

	log.Printf("[yt-dlp] Command output: %s", output.String())
	return output.Bytes(), nil
}

// CommandExecutor handles execution of yt-dlp commands.
type CommandExecutor struct {
	client *CommandRunner
}

// NewCommandExecutor creates a new CommandExecutor.
func NewCommandExecutor(runner *CommandRunner) *CommandExecutor {
	if runner == nil {
		runner = NewCommandRunner()
	}
	return &CommandExecutor{client: runner}
}

// validateCookiesFile checks if cookies file exists and logs appropriate messages
func (e *CommandExecutor) validateCookiesFile() {
	if _, err := os.Stat("cookies.txt"); os.IsNotExist(err) {
		log.Printf("[yt-dlp] WARNING: cookies.txt file not found!")
	} else {
		log.Printf("[yt-dlp] cookies.txt file found and will be used")
	}
}

// buildBaseArgs creates common arguments for yt-dlp commands
func (e *CommandExecutor) buildBaseArgs() []string {
	return []string{
		"--cookies", "cookies.txt",
		"--no-check-certificates",
	}
}

// RunVideoMetadata executes yt-dlp --dump-json to get video metadata.
func (e *CommandExecutor) RunVideoMetadata(ctx context.Context, videoURL string) ([]byte, error) {
	if videoURL == "" {
		return nil, fmt.Errorf("video URL cannot be empty")
	}

	e.validateCookiesFile()

	args := append(e.buildBaseArgs(), "--dump-json", videoURL)
	log.Printf("[yt-dlp] Getting video metadata for: %s", videoURL)

	output, err := e.client.Run(ctx, "yt-dlp", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute yt-dlp --dump-json: %w", err)
	}
	return output, nil
}

// RunPlaylistMetadata executes yt-dlp --dump-single-json to get playlist metadata.
func (e *CommandExecutor) RunPlaylistMetadata(ctx context.Context, playlistURL string) ([]byte, error) {
	if playlistURL == "" {
		return nil, fmt.Errorf("playlist URL cannot be empty")
	}

	e.validateCookiesFile()

	args := append(e.buildBaseArgs(),
		"--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.101 Safari/537.36",
		"--dump-single-json", playlistURL,
	)
	log.Printf("[yt-dlp] Getting playlist metadata for: %s", playlistURL)

	output, err := e.client.Run(ctx, "yt-dlp", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute yt-dlp --dump-single-json: %w", err)
	}
	return output, nil
}

// RunDownload executes yt-dlp to download a video or audio file.
func (e *CommandExecutor) RunDownload(ctx context.Context, videoURL string, opts youtube.DownloadOptions) (string, error) {
	if videoURL == "" {
		return "", fmt.Errorf("video URL cannot be empty")
	}
	if opts.OutputDir == "" {
		return "", fmt.Errorf("output directory cannot be empty")
	}
	if opts.Filename == "" {
		return "", fmt.Errorf("filename cannot be empty")
	}

	e.validateCookiesFile()

	args := append(e.buildBaseArgs(),
		"--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.101 Safari/537.36",
		"-f", fmt.Sprintf("%d", opts.FormatID),
		"-o", filepath.Join(opts.OutputDir, opts.Filename),
	)

	// Add audio extraction for audio-only formats
	if opts.AudioOnly {
		args = append(args, "--extract-audio", "--audio-format", "mp3")
	}

	if opts.Merge {
		args = append(args, "--merge-output-format", "mp4")
	}

	if len(opts.ExtraArgs) > 0 {
		args = append(args, opts.ExtraArgs...)
	}

	args = append(args, videoURL)

	log.Printf("[yt-dlp] Downloading: %s with options: %+v", videoURL, opts)

	_, err := e.client.Run(ctx, "yt-dlp", args...)
	if err != nil {
		return "", fmt.Errorf("failed to execute yt-dlp download: %w", err)
	}

	outputPath := filepath.Join(opts.OutputDir, opts.Filename)

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("download completed but file not found at: %s", outputPath)
	}

	return outputPath, nil
}
