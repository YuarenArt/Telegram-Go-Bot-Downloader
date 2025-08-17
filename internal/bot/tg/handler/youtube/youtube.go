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

const (
	All_video    = "allVideo"
	All_audio    = "allAudio"
	TrafficLimit = 5000.0 // Mb
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
	return &YoutubeHandler{
		Downloader: ytdl.NewYTDLBackend(),
	}
}

// HandleMessage handle YouTube link and return error
func (yh *YoutubeHandler) HandleMessage(message *tgbotapi.Message) (*tgbotapi.InlineKeyboardMarkup, error) {
	return yh.handleYoutubeLink(message)
}

// handleYoutubeLink normalizes the URL and routes to the correct handler
func (yh *YoutubeHandler) handleYoutubeLink(message *tgbotapi.Message) (*tgbotapi.InlineKeyboardMarkup, error) {
	normalizedURL, typ, err := normalizeYouTubeURL(message.Text)
	if err != nil {
		return nil, err
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
	u, err := url.Parse(raw)
	if err != nil {
		return "", Unknown, err
	}

	host := strings.ToLower(u.Host)
	switch {
	case strings.HasPrefix(host, "youtu.be"):
		id := strings.TrimPrefix(u.Path, "/")
		if id == "" {
			return "", Unknown, errors.New("empty video ID in youtu.be")
		}
		return "https://youtu.be/" + id, Video, nil

	case strings.HasPrefix(host, "www.youtube.com") || strings.HasPrefix(host, "youtube.com"):
		switch {
		case strings.HasPrefix(u.Path, "/watch"):
			id := u.Query().Get("v")
			if id == "" {
				return "", Unknown, errors.New("missing v parameter")
			}
			return "https://youtu.be/" + id, Video, nil

		case strings.HasPrefix(u.Path, "/playlist"):
			list := u.Query().Get("list")
			if list == "" {
				return "", Unknown, errors.New("missing list parameter")
			}
			return "https://www.youtube.com/playlist?list=" + list, Playlist, nil

		case strings.HasPrefix(u.Path, "/live/"):
			id := strings.TrimPrefix(u.Path, "/live/")
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
	keyboard := tgbotapi.NewInlineKeyboardMarkup()
	audioSize := 0.0

	for _, format := range formats {
		if format.AudioOnly && format.Bitrate > 0 {
			duration := 180.0
			audioSize = float64(format.Bitrate) * duration / (8 * 1024 * 1024)
			break
		}
	}

	for _, format := range formats {
		mimeType := format.MimeType
		if strings.HasPrefix(mimeType, "audio/webm") || strings.HasPrefix(mimeType, "video/webm") {
			continue
		}

		formatType := "Unknown"
		if format.AudioOnly {
			formatType = "Audio"
		} else if format.VideoOnly {
			formatType = "Video"
		} else {
			formatType = "Video+Audio"
		}

		itagNo := format.Itag
		data := fmt.Sprintf("%s,%d", *url, itagNo)

		size := 0.0
		if format.Bitrate > 0 {
			duration := 180.0
			size = float64(format.Bitrate) * duration / (8 * 1024 * 1024)
		}
		if !format.AudioOnly {
			size += audioSize
		}

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

		button := tgbotapi.NewInlineKeyboardButtonData(strings.Join(sign, ", "), data)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{button})
	}

	return &keyboard, nil
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
