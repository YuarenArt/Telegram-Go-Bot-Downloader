package ytdl

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
	"youtube_downloader/pkg/downloader/youtube"
)

// MetadataParser handles parsing of yt-dlp JSON output into structured data.
type MetadataParser struct{}

// NewMetadataParser creates a new MetadataParser.
func NewMetadataParser() *MetadataParser {
	return &MetadataParser{}
}

// ParseVideo converts JSON data into a Video struct.
func (p *MetadataParser) ParseVideo(data []byte, sourceURL string) (*youtube.Video, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data provided")
	}
	if sourceURL == "" {
		return nil, errors.New("source URL cannot be empty")
	}

	var videoData map[string]interface{}
	if err := json.Unmarshal(data, &videoData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	id, ok := videoData["id"].(string)
	if !ok || id == "" {
		return nil, errors.New("invalid or missing video ID")
	}

	title, ok := videoData["title"].(string)
	if !ok || title == "" {
		return nil, errors.New("invalid or missing video title")
	}

	// duration in seconds (may be float)
	var duration time.Duration
	if durRaw, ok := videoData["duration"].(float64); ok && durRaw > 0 {
		// keep fractional seconds accurate
		duration = time.Duration(math.Round(durRaw*1000)) * time.Millisecond
	} else {
		duration = 0
	}

	formats, err := p.parseFormats(videoData["formats"])
	if err != nil {
		// don't fail completely if formats parsing failed; return video with zero formats for caller to decide
		return &youtube.Video{
			ID:        id,
			Title:     title,
			Duration:  duration,
			Formats:   nil,
			SourceURL: sourceURL,
		}, fmt.Errorf("failed to parse formats: %w", err)
	}

	return &youtube.Video{
		ID:        id,
		Title:     title,
		Duration:  duration,
		Formats:   formats,
		SourceURL: sourceURL,
	}, nil
}

// ParsePlaylist converts JSON data into a Playlist struct.
func (p *MetadataParser) ParsePlaylist(data []byte, playlistURL string) (*youtube.Playlist, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data provided")
	}
	if playlistURL == "" {
		return nil, errors.New("playlist URL cannot be empty")
	}

	var playlistData map[string]interface{}
	if err := json.Unmarshal(data, &playlistData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	id, ok := playlistData["id"].(string)
	if !ok || id == "" {
		return nil, errors.New("invalid or missing playlist ID")
	}

	title, ok := playlistData["title"].(string)
	if !ok || title == "" {
		return nil, errors.New("invalid or missing playlist title")
	}

	entries, ok := playlistData["entries"].([]interface{})
	if !ok {
		return nil, errors.New("invalid or missing playlist entries")
	}

	videos := make([]*youtube.Video, 0, len(entries))
	for i, entry := range entries {
		videoData, ok := entry.(map[string]interface{})
		if !ok {
			log.Printf("Skipping invalid video entry at index %d", i)
			continue
		}

		videoID, ok := videoData["id"].(string)
		if !ok || videoID == "" {
			log.Printf("Skipping video entry with invalid ID at index %d", i)
			continue
		}

		videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)

		videoJSON, err := json.Marshal(videoData)
		if err != nil {
			log.Printf("Failed to marshal video data for entry %d: %v", i, err)
			continue
		}

		video, err := p.ParseVideo(videoJSON, videoURL)
		if err != nil {
			log.Printf("Failed to parse video entry %d: %v", i, err)
			continue
		}
		videos = append(videos, video)
	}

	if len(videos) == 0 {
		return nil, errors.New("no valid videos found in playlist")
	}

	return &youtube.Playlist{
		ID:     id,
		Title:  title,
		Videos: videos,
	}, nil
}

// parseFormats converts the formats array from JSON into a slice of Format structs.
func (p *MetadataParser) parseFormats(formatsData interface{}) ([]youtube.Format, error) {
	formatsRaw, ok := formatsData.([]interface{})
	if !ok {
		return nil, errors.New("invalid formats data")
	}

	if len(formatsRaw) == 0 {
		return nil, errors.New("no formats available")
	}

	formats := make([]youtube.Format, 0, len(formatsRaw))
	for i, f := range formatsRaw {
		formatData, ok := f.(map[string]interface{})
		if !ok {
			log.Printf("Skipping invalid format at index %d", i)
			continue
		}

		format, err := p.parseSingleFormat(formatData)
		if err != nil {
			log.Printf("Skipping format at index %d: %v", i, err)
			continue
		}

		formats = append(formats, format)
	}

	if len(formats) == 0 {
		return nil, errors.New("no valid formats found")
	}

	return formats, nil
}

// parseSingleFormat parses a single format from the formats array
func (p *MetadataParser) parseSingleFormat(formatData map[string]interface{}) (youtube.Format, error) {
	var f youtube.Format

	// format_id → строка
	if raw, ok := formatData["format_id"]; ok {
		f.FormatID = fmt.Sprintf("%v", raw)
	}

	// itag из format_id
	if num := strings.TrimLeft(f.FormatID, "0123456789"); len(f.FormatID) > len(num) {
		if v, err := strconv.Atoi(f.FormatID[:len(f.FormatID)-len(num)]); err == nil {
			f.Itag = v
		}
	}

	if v, ok := formatData["ext"].(string); ok {
		f.Ext = v
	}
	if v, ok := formatData["height"].(float64); ok {
		f.Height = int(v)
	}
	if v, ok := formatData["width"].(float64); ok {
		f.Width = int(v)
	}
	if v, ok := formatData["filesize"].(float64); ok {
		f.FileSize = int64(v)
	}
	if v, ok := formatData["filesize_approx"].(float64); ok {
		f.FileSizeApprox = int64(v)
	}
	if v, ok := formatData["mime_type"].(string); ok {
		f.MimeType = v
	}
	if v, ok := formatData["format_note"].(string); ok {
		f.Quality = v
	}
	if v, ok := formatData["tbr"].(float64); ok {
		f.Bitrate = int(v)
	}
	if v, ok := formatData["acodec"].(string); ok {
		f.AudioCodec = v
	}
	if v, ok := formatData["vcodec"].(string); ok {
		f.VideoCodec = v
	}

	f.AudioOnly = (f.AudioCodec != "" && f.AudioCodec != "none") && (f.VideoCodec == "" || f.VideoCodec == "none")
	f.VideoOnly = (f.VideoCodec != "" && f.VideoCodec != "none") && (f.AudioCodec == "" || f.AudioCodec == "none")

	return f, nil
}
