package youtube

// this code copied from github.com/kkdai/youtube/v2

import (
	"context"
	"errors"
	"github.com/kkdai/youtube/v2"
	"github.com/vbauerster/mpb/v5"
	"github.com/vbauerster/mpb/v5/decor"
	"io"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

var canonicals = map[string]string{
	"video/quicktime":  ".mov",
	"video/x-msvideo":  ".avi",
	"video/x-matroska": ".mkv",
	"video/mpeg":       ".mpeg",
	"video/webm":       ".webm",
	"video/3gpp2":      ".3g2",
	"video/x-flv":      ".flv",
	"video/3gpp":       ".3gp",
	"video/mp4":        ".mp4",
	"video/ogg":        ".ogv",
	"video/mp2t":       ".ts",
	"audio/mp4":        ".m4a",
}

const defaultExtension = ".mov"

type progress struct {
	contentLength     float64
	totalWrittenBytes float64
	downloadLevel     float64
}

func (dl *progress) Write(p []byte) (n int, err error) {
	n = len(p)
	dl.totalWrittenBytes = dl.totalWrittenBytes + float64(n)
	currentPercent := (dl.totalWrittenBytes / dl.contentLength) * 100
	if (dl.downloadLevel <= currentPercent) && (dl.downloadLevel < 100) {
		dl.downloadLevel++
	}
	return
}

// DownloadVideoWithFormatComposite downloads separately video and audio files then merges it.
// the quality, type of name, language can be empty string, then the download will be carried out with maximum quality.
func (ytd *YouTubeDownloader) DownloadVideoWithFormatComposite(ctx context.Context, outputFile string, v *youtube.Video, quality, mimetype, language string) (string, error) {
	videoFormat, audioFormat, err1 := getVideoAudioFormats(v, quality, mimetype, language)
	if err1 != nil {
		return "", err1
	}

	log := youtube.Logger.With("id", v.ID)

	log.Info(
		"Downloading composite video",
		"videoQuality", videoFormat.QualityLabel,
		"videoMimeType", videoFormat.MimeType,
		"audioMimeType", audioFormat.MimeType,
	)

	destFile, err := ytd.getOutputFile(v, videoFormat, outputFile)
	if err != nil {
		return "", err
	}
	outputDir := filepath.Dir(destFile)

	// Create temporary video file
	videoFile, err := os.CreateTemp(outputDir, "youtube_*.m4v")
	if err != nil {
		return "", err
	}
	defer func() {
		videoFile.Close()
		os.Remove(videoFile.Name())
	}()

	// Create temporary audio file
	audioFile, err := os.CreateTemp(outputDir, "youtube_*.m4a")
	if err != nil {
		return "", err
	}
	defer func() {
		audioFile.Close()
		os.Remove(audioFile.Name())
	}()

	log.Debug("Downloading video file...")
	err = ytd.videoDLWorker(ctx, videoFile, v, videoFormat)
	if err != nil {
		return "", err
	}

	log.Debug("Downloading audio file...")
	err = ytd.videoDLWorker(ctx, audioFile, v, audioFormat)
	if err != nil {
		return "", err
	}

	//nolint:gosec
	ffmpegVersionCmd := exec.Command("ffmpeg", "-y",
		"-i", videoFile.Name(),
		"-i", audioFile.Name(),
		"-c", "copy", // Just copy without re-encoding
		"-shortest", // Finish encoding when the shortest input stream ends
		destFile,
		"-loglevel", "warning",
	)
	ffmpegVersionCmd.Stderr = os.Stderr
	ffmpegVersionCmd.Stdout = os.Stdout
	log.Info("merging video and audio", "output", destFile)

	return destFile, ffmpegVersionCmd.Run()
}

func (ytd *YouTubeDownloader) videoDLWorker(ctx context.Context, out *os.File, video *youtube.Video, format *youtube.Format) error {
	stream, size, err := ytd.Downloader.GetStreamContext(ctx, video, format)
	if err != nil {
		return err
	}

	prog := &progress{
		contentLength: float64(size),
	}

	// create progress bar
	progress := mpb.New(mpb.WithWidth(64))
	bar := progress.AddBar(
		int64(prog.contentLength),

		mpb.PrependDecorators(
			decor.CountersKibiByte("% .2f / % .2f"),
			decor.Percentage(decor.WCSyncSpace),
		),
		mpb.AppendDecorators(
			decor.EwmaETA(decor.ET_STYLE_GO, 90),
			decor.Name(" ] "),
			decor.EwmaSpeed(decor.UnitKiB, "% .2f", 60),
		),
	)

	reader := bar.ProxyReader(stream)
	mw := io.MultiWriter(out, prog)
	_, err = io.Copy(mw, reader)
	if err != nil {
		return err
	}

	progress.Wait()
	return nil
}

func getVideoAudioFormats(v *youtube.Video, quality string, mimetype, language string) (*youtube.Format, *youtube.Format, error) {
	var videoFormats, audioFormats youtube.FormatList

	formats := v.Formats
	if mimetype != "" {
		formats = formats.Type(mimetype)
	}

	videoFormats = formats.Type("video/mp4").AudioChannels(0)
	audioFormats = formats.Type("audio/mp4")

	if quality != "" {
		videoFormats = videoFormats.Quality(quality)
	}

	if language != "" {
		audioFormats = audioFormats.Language(language)
	}

	if len(videoFormats) == 0 {
		return nil, nil, errors.New("no video format found after filtering")
	}

	if len(audioFormats) == 0 {
		return nil, nil, errors.New("no audio format found after filtering")
	}

	videoFormats.Sort()
	audioFormats.Sort()

	return &videoFormats[0], &audioFormats[0], nil
}

func (ytd *YouTubeDownloader) getOutputFile(v *youtube.Video, format *youtube.Format, outputFile string) (string, error) {
	if outputFile == "" {
		outputFile = SanitizeFilename(v.Title)
		outputFile += pickIdealFileExtension(format.MimeType)
	}

	if ytd.Downloader.OutputDir != "" {
		if err := os.MkdirAll(ytd.Downloader.OutputDir, 0o755); err != nil {
			return "", err
		}
		outputFile = filepath.Join(ytd.Downloader.OutputDir, outputFile)
	}

	return outputFile, nil
}

func pickIdealFileExtension(mediaType string) string {
	mediaType, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return defaultExtension
	}

	if extension, ok := canonicals[mediaType]; ok {
		return extension
	}

	// Our last resort is to ask the operating system, but these give multiple results and are rarely canonical.
	extensions, err := mime.ExtensionsByType(mediaType)
	if err != nil || extensions == nil {
		return defaultExtension
	}

	return extensions[0]
}

// SanitizeFilename clear all unsupported symbols for mac, linux, windows
func SanitizeFilename(fileName string) string {
	// Characters not allowed on mac
	//	:/
	// Characters not allowed on linux
	//	/
	// Characters not allowed on windows
	//	<>:"/\|?*

	// Ref https://docs.microsoft.com/en-us/windows/win32/fileio/naming-a-file#naming-conventions

	fileName = regexp.MustCompile(`[:/<>\:"\\|?*]`).ReplaceAllString(fileName, "")
	fileName = regexp.MustCompile(`\s+`).ReplaceAllString(fileName, " ")

	return fileName
}
