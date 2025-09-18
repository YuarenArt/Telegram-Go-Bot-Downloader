package youtube

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"youtube_downloader/pkg/database-client"
	"youtube_downloader/pkg/downloader/youtube"
	"youtube_downloader/pkg/downloader/youtube/ytdl"

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
	Client     *database_client.Client
}

// groupKey group formats by (height, container, type)
type groupKey struct {
	height    int
	container string
	fType     string
}

// NewYoutubeHandler return new YoutubeHandler
// cookiesPath is optional. If empty, cookies will not be used.
func NewYoutubeHandler(downloader youtube.Downloader, client *database_client.Client, cookiesPath string) *YoutubeHandler {
	if downloader == nil {
		downloader = ytdl.NewYTDLBackend(cookiesPath)
	}
	return &YoutubeHandler{
		Downloader: downloader,
		Client:     client,
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
		return nil, fmt.Errorf("URL cannot be nil or empty")
	}
	if len(formats) == 0 {
		return nil, fmt.Errorf("no formats available")
	}

	// Step 1: filter out unwanted formats
	candidates := make([]youtube.Format, 0, len(formats))
	for _, f := range formats {
		if f.FormatID == "" && f.Itag == 0 {
			continue
		}
		lid := strings.ToLower(f.FormatID)
		if strings.HasPrefix(lid, "sb") || strings.Contains(strings.ToLower(f.MimeType), "mhtml") || strings.Contains(strings.ToLower(f.MimeType), "storyboard") {
			continue
		}
		mt := strings.ToLower(f.MimeType)
		if strings.Contains(mt, "text/") || (strings.Contains(mt, "application/") && !strings.Contains(mt, "video") && !strings.Contains(mt, "audio")) {
			continue
		}
		candidates = append(candidates, f)
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no user-friendly formats available")
	}

	// Step 2: choose best per group
	best := make(map[groupKey]youtube.Format)
	for _, f := range candidates {
		height := f.Height
		if height == 0 {
			height = parseHeightFromQuality(f.Quality)
		}
		container := f.Ext
		if container == "" {
			container = parseContainerFromMime(f.MimeType)
		}
		ftype := "video+audio"
		if f.AudioOnly {
			ftype = "audio"
		} else if f.VideoOnly {
			ftype = "video"
		}
		k := groupKey{height: height, container: strings.ToUpper(container), fType: ftype}
		cur, ok := best[k]
		if !ok || f.Bitrate > cur.Bitrate {
			best[k] = f
		}
	}

	// Step 3: build entries
	type displayEntry struct {
		Key      groupKey
		Format   youtube.Format
		Label    string
		SizeMB   float64
		SortRank int
	}
	entries := make([]displayEntry, 0, len(best))
	for k, f := range best {
		var size float64
		if f.FileSize > 0 {
			size = float64(f.FileSize) / 1024.0 / 1024.0
		} else if f.FileSizeApprox > 0 {
			size = float64(f.FileSizeApprox) / 1024.0 / 1024.0
		} else {
			size = estimateFormatSizeFromBitrate(f)
		}
		label := buildLabelForFormat(f, k.height, k.container, k.fType, size)
		rank := rankForGroup(k)
		entries = append(entries, displayEntry{
			Key:      k,
			Format:   f,
			Label:    label,
			SizeMB:   size,
			SortRank: rank,
		})
	}

	// Sort
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].SortRank != entries[j].SortRank {
			return entries[i].SortRank > entries[j].SortRank
		}
		return entries[i].SizeMB < entries[j].SizeMB
	})

	// Step 4: build keyboard
	keyboard := tgbotapi.NewInlineKeyboardMarkup()
	for _, e := range entries {
		formID := e.Format.FormatID
		if formID == "" {
			formID = strconv.Itoa(e.Format.Itag)
		}
		data := fmt.Sprintf("%s,%s", *url, formID)
		button := tgbotapi.NewInlineKeyboardButtonData(e.Label, data)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, []tgbotapi.InlineKeyboardButton{button})
	}
	return &keyboard, nil
}

func parseHeightFromQuality(q string) int {
	if q == "" {
		return 0
	}
	q = strings.TrimSpace(q)
	for i := 0; i < len(q); i++ {
		if q[i] >= '0' && q[i] <= '9' {
			j := i
			for j < len(q) && q[j] >= '0' && q[j] <= '9' {
				j++
			}
			if j < len(q) && (q[j] == 'p' || q[j] == 'P') {
				if h, err := strconv.Atoi(q[i:j]); err == nil {
					return h
				}
			}
			if h, err := strconv.Atoi(q[i:j]); err == nil {
				return h
			}
			break
		}
	}
	return 0
}

func parseContainerFromMime(mime string) string {
	if mime == "" {
		return "unknown"
	}
	parts := strings.Split(mime, "/")
	if len(parts) < 2 {
		return strings.TrimSpace(mime)
	}
	rest := parts[1]
	if idx := strings.Index(rest, ";"); idx >= 0 {
		rest = rest[:idx]
	}
	return strings.TrimSpace(rest)
}

func estimateFormatSizeFromBitrate(f youtube.Format) float64 {
	if f.Bitrate <= 0 {
		return 0.0
	}
	return float64(f.Bitrate) * DefaultDuration / BytesPerMB
}

func buildLabelForFormat(f youtube.Format, height int, container string, fType string, size float64) string {
	parts := make([]string, 0, 4)

	if height > 0 {
		parts = append(parts, fmt.Sprintf("%dp", height))
	} else if f.Quality != "" {
		parts = append(parts, f.Quality)
	}

	if container != "" && container != "unknown" {
		parts = append(parts, strings.ToUpper(container))
	}

	switch fType {
	case "audio":
		parts = append(parts, "Audio")
	case "video":
		parts = append(parts, "Video")
	default:
		parts = append(parts, "Video+Audio")
	}

	if size > 0 {
		parts = append(parts, fmt.Sprintf("%.1f Mb", size))
	}

	return strings.Join(parts, " ")
}

func rankForGroup(k groupKey) int {
	rank := k.height
	if k.fType == "audio" {
		rank -= 1
	} else if k.fType == "video" {
		rank -= 2
	}
	if strings.ToLower(k.container) == "mp4" {
		rank += 5
	}
	return rank
}
