package youtube

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"youtube_downloader/internal/downloader/youtube"
	"youtube_downloader/internal/downloader/youtube/ytdl"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type YouTubeType int

const (
	Unknown YouTubeType = iota
	Video
	Playlist
	Stream
)

// YoutubeHandler is a service for downloading video from youtube
type YoutubeHandler struct {
	Downloader youtube.Downloader
}

// NewYoutubeHandler return new YoutubeHandler
func NewYoutubeHandler(downloader youtube.Downloader) *YoutubeHandler {
	if downloader == nil {
		downloader = ytdl.NewYTDLBackend()
	}
	return &YoutubeHandler{
		Downloader: downloader,
	}
}

// HandleMessage handle YouTube link and return error
func (yh *YoutubeHandler) HandleMessage(message *tgbotapi.Message) (*tgbotapi.InlineKeyboardMarkup, error) {
	if message == nil || message.Text == "" {
		return nil, errors.New("invalid message")
	}
	return yh.handleYoutubeLink(message)
}

// handleYoutubeLink normalizes the URL and routes to the correct handler
func (yh *YoutubeHandler) handleYoutubeLink(message *tgbotapi.Message) (*tgbotapi.InlineKeyboardMarkup, error) {
	normalizedURL, typ, err := normalizeYouTubeURL(message.Text)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize URL: %w", err)
	}

	message.Text = normalizedURL

	switch typ {
	case Stream:
		return yh.handleYoutubeStream(message)
	case Playlist:
		return yh.handleYoutubePlaylist(message)
	case Video:
		return yh.handleYoutubeVideo(message)
	default:
		return nil, errors.New("unknown or unsupported YouTube link type")
	}
}

// normalizeYouTubeURL parses and transforms a URL into a yt-dlp compatible format
func normalizeYouTubeURL(raw string) (string, YouTubeType, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", Unknown, errors.New("URL cannot be empty")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", Unknown, fmt.Errorf("invalid URL format: %w", err)
	}

	host := strings.ToLower(u.Host)
	switch {
	case strings.HasPrefix(host, YouTubeShort):
		id := strings.TrimPrefix(u.Path, "/")
		if id == "" {
			return "", Unknown, errors.New("empty video ID in youtu.be")
		}
		return "https://youtu.be/" + id, Video, nil

	case strings.HasPrefix(host, YouTubeWWW) || strings.HasPrefix(host, YouTubeDomain):
		switch {
		case strings.HasPrefix(u.Path, YouTubeWatchPath):
			id := u.Query().Get("v")
			if id == "" {
				return "", Unknown, errors.New("missing v parameter")
			}
			return "https://youtu.be/" + id, Video, nil

		case strings.HasPrefix(u.Path, YouTubePlaylistPath):
			list := u.Query().Get("list")
			if list == "" {
				return "", Unknown, errors.New("missing list parameter")
			}
			return "https://www.youtube.com/playlist?list=" + list, Playlist, nil

		case strings.HasPrefix(u.Path, YouTubeLivePath):
			id := strings.TrimPrefix(u.Path, YouTubeLivePath+"/")
			if id == "" {
				return "", Unknown, errors.New("missing live stream ID")
			}
			return "https://www.youtube.com/live/" + id, Stream, nil

		default:
			return "", Unknown, errors.New("unsupported youtube.com path")
		}

	default:
		return "", Unknown, errors.New("unsupported host")
	}
}

// getKeyboardVideoFormats builds keyboard by available formats
func getKeyboardVideoFormats(formats []youtube.Format, url *string) (*tgbotapi.InlineKeyboardMarkup, error) {
	if url == nil || *url == "" {
		return nil, errors.New("URL cannot be nil or empty")
	}
	if len(formats) == 0 {
		return nil, errors.New("no formats available")
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup()

	// Find best audio format for size estimation
	audioSize := estimateAudioSize(formats)

	for _, format := range formats {
		// Skip WebM formats as they can cause issues
		if strings.HasPrefix(format.MimeType, "audio/webm") || strings.HasPrefix(format.MimeType, "video/webm") {
			continue
		}

		formatType := determineFormatType(format)
		itagNo := format.Itag
		data := fmt.Sprintf("%s,%d", *url, itagNo)

		size := estimateFormatSize(format, audioSize)

		sign := buildFormatDescription(format, formatType, size)

		button := tgbotapi.NewInlineKeyboardButtonData(strings.Join(sign, ", "), data)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{button})
	}

	return &keyboard, nil
}

// estimateAudioSize estimates the size of audio track for video formats
func estimateAudioSize(formats []youtube.Format) float64 {
	for _, format := range formats {
		if format.AudioOnly && format.Bitrate > 0 {
			return float64(format.Bitrate) * DefaultDuration / BytesPerMB
		}
	}
	return 0.0
}

// determineFormatType determines the type of format (Audio, Video, or Video+Audio)
func determineFormatType(format youtube.Format) string {
	if format.AudioOnly {
		return "Audio"
	} else if format.VideoOnly {
		return "Video"
	}
	return "Video+Audio"
}

// estimateFormatSize estimates the file size for a given format
func estimateFormatSize(format youtube.Format, audioSize float64) float64 {
	size := 0.0
	if format.Bitrate > 0 {
		size = float64(format.Bitrate) * DefaultDuration / BytesPerMB
	}

	// Add audio size for video-only formats
	if !format.AudioOnly && format.VideoOnly {
		size += audioSize
	}

	return size
}

// buildFormatDescription builds the description text for a format button
func buildFormatDescription(format youtube.Format, formatType string, size float64) []string {
	sign := []string{formatType}

	if format.Quality != "" {
		sign = append(sign, format.Quality)
	}

	if format.VideoCodec != "" && !format.AudioOnly {
		sign = append(sign, format.VideoCodec)
	}

	if format.AudioCodec != "" && format.AudioOnly {
		sign = append(sign, format.AudioCodec)
	}

	sign = append(sign, fmt.Sprintf("%.1f Mb", size))

	return sign
}

// getFileSizeGeneric estimates file size from metadata
func getFileSizeGeneric(format map[string]any) (float64, error) {
	if cl, ok := format["ContentLength"]; ok && cl.(int) > 0 {
		return float64(cl.(int)), nil
	}

	duration, err := strconv.ParseFloat(fmt.Sprintf("%v", format["ApproxDurationMs"]), 64)
	if err != nil {
		return 0, err
	}
	duration /= 1000

	bitrate := format["Bitrate"].(int)
	if ab, ok := format["AverageBitrate"]; ok && ab.(int) > 0 {
		bitrate = ab.(int)
	}

	contentLength := float64(bitrate/8) * duration
	return contentLength, nil
}
