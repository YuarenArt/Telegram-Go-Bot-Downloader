package ytdl

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"unicode"
	"youtube_downloader/pkg/downloader/youtube"
)

// CommandRunner executes external yt-dlp commands.
type CommandRunner struct{}

// NewCommandRunner creates a new CommandRunner.
func NewCommandRunner() *CommandRunner {
	return &CommandRunner{}
}

// Run executes the specified command with arguments, returning stdout bytes or an error.
// stdout and stderr are captured separately. We return stdout (possibly sanitized) and
// log stderr so that warnings do not break JSON parsing.
func (c *CommandRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	if name == "" {
		return nil, fmt.Errorf("command name cannot be empty")
	}

	log.Printf("[yt-dlp] Running command: %s %v", name, args)

	cmd := exec.CommandContext(ctx, name, args...)
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	// Always log stderr content (warnings, retry messages etc.) so operator can inspect it.
	if stderrBuf.Len() > 0 {
		// Trim leading/trailing whitespace for cleaner logs
		stderrStr := bytes.TrimSpace(stderrBuf.Bytes())
		log.Printf("[yt-dlp][stderr] %s", stderrStr)
	}

	// If command failed, return combined error but still try to return stdout if available.
	if err != nil {
		if ctx.Err() != nil {
			log.Printf("[yt-dlp] Context error: %v", ctx.Err())
			return nil, ctx.Err()
		}
		// If we have stdout content, return it along with command error so upper layer can attempt to parse.
		if stdoutBuf.Len() > 0 {
			log.Printf("[yt-dlp] Command finished with error but stdout available: %v", err)
			out := sanitizeJSONOutput(stdoutBuf.Bytes())
			return out, fmt.Errorf("command execution failed: %w", err)
		}
		log.Printf("[yt-dlp] Command error: %v", err)
		return nil, fmt.Errorf("command execution failed: %w", err)
	}

	// Command succeeded: return (sanitized) stdout
	out := sanitizeJSONOutput(stdoutBuf.Bytes())
	log.Printf("[yt-dlp] Command output size: %d bytes", len(out))
	return out, nil
}

// sanitizeJSONOutput trims any non-json leading characters and whitespace from output.
// It looks for the first '{' or '[' and returns from there; otherwise returns original bytes.
func sanitizeJSONOutput(b []byte) []byte {
	if len(b) == 0 {
		return b
	}
	// find first '{' or '['
	for i, ch := range b {
		if ch == '{' || ch == '[' {
			// Trim left-side whitespace/newlines before returning for neatness
			return bytes.TrimLeftFunc(b[i:], func(r rune) bool { return unicode.IsSpace(r) })
		}
		// also skip common logging prefixes like "WARNING:" which are ASCII letters,
		// so loop continues until JSON start.
	}
	// fallback — return trimmed original
	return bytes.TrimSpace(b)
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

	// Add --no-progress and --no-warnings if you want to reduce noise from yt-dlp itself;
	// we still capture stderr for debugging.
	args := append(e.buildBaseArgs(), "--no-progress", "--dump-json", videoURL)
	log.Printf("[yt-dlp] Getting video metadata for: %s", videoURL)

	output, err := e.client.Run(ctx, "yt-dlp", args...)
	if err != nil {
		// Return output even on error if available (caller may attempt to parse it).
		return output, fmt.Errorf("failed to execute yt-dlp --dump-json: %w", err)
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
		"--no-progress",
		"--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.101 Safari/537.36",
		"--dump-single-json", playlistURL,
	)
	log.Printf("[yt-dlp] Getting playlist metadata for: %s", playlistURL)

	output, err := e.client.Run(ctx, "yt-dlp", args...)
	if err != nil {
		return output, fmt.Errorf("failed to execute yt-dlp --dump-single-json: %w", err)
	}
	return output, nil
}

// RunDownload executes yt-dlp to download a video or audio file.
func (e *CommandExecutor) RunDownload(ctx context.Context, videoURL string, opts youtube.DownloadOptions) (string, error) {
	if videoURL == "" {
		return "", fmt.Errorf("video URL cannot be empty")
	}
	// If caller forgot to set defaults, reject — caller should call validateDownloadOptions beforehand.
	if opts.OutputDir == "" {
		return "", fmt.Errorf("output directory cannot be empty")
	}
	if opts.Filename == "" {
		return "", fmt.Errorf("filename cannot be empty")
	}

	e.validateCookiesFile()

	// build arguments
	args := e.buildBaseArgs()
	args = append(args,
		"--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.101 Safari/537.36",
		"--no-progress",
	)

	// If a format string is provided, pass it to yt-dlp using -f
	if opts.Format != "" {
		args = append(args, "-f", opts.Format)
	}

	// Use output template that allows yt-dlp to add the correct extension.
	// The template will result in files like /path/filename.mp4 or /path/filename.mp3
	outTemplate := filepath.Join(opts.OutputDir, opts.Filename+".%(ext)s")
	args = append(args, "-o", outTemplate)

	// Audio extraction if requested
	if opts.AudioOnly {
		args = append(args, "--extract-audio", "--audio-format", "mp3")
	}

	// Merge audio+video into mp4 if requested
	if opts.Merge {
		args = append(args, "--merge-output-format", "mp4")
	}

	if len(opts.ExtraArgs) > 0 {
		args = append(args, opts.ExtraArgs...)
	}

	// finally the URL
	args = append(args, videoURL)

	log.Printf("[yt-dlp] Downloading: %s with options: %+v", videoURL, opts)

	stdout, err := e.client.Run(ctx, "yt-dlp", args...)
	// Log stdout for debug (may contain progress or final info). Do not treat it as error if parsing failed earlier.
	if len(stdout) > 0 {
		log.Printf("[yt-dlp] Download stdout size: %d bytes", len(stdout))
	}
	if err != nil {
		// If there was an error but stdout/created file exists, we can still try to locate file.
		log.Printf("[yt-dlp] Download command returned error: %v", err)
	}

	// Try to discover resulting file.
	// Because we passed "%(ext)s", search for common extensions.
	basePath := filepath.Join(opts.OutputDir, opts.Filename)
	candidates := []string{
		basePath + ".mp4",
		basePath + ".mkv",
		basePath + ".webm",
		basePath + ".mp3",
		basePath + ".m4a",
		basePath + ".aac",
		basePath + ".opus",
		basePath + ".flac",
	}

	for _, c := range candidates {
		if fileExists(c) {
			return c, nil
		}
	}

	// As a fallback, try to glob any file that starts with basePath
	matches, _ := filepath.Glob(basePath + ".*")
	if len(matches) > 0 {
		// return first match
		return matches[0], nil
	}

	// final fallback: check exact template without extension (some setups may write exactly this)
	if fileExists(basePath) {
		return basePath, nil
	}

	// nothing found -> return error (include run error if present)
	if err != nil {
		return "", fmt.Errorf("failed to execute yt-dlp download: %w", err)
	}
	return "", fmt.Errorf("download completed but file not found at: %s (checked common extensions)", basePath)
}
