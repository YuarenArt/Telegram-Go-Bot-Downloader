package ytdl

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
	"youtube_downloader/internal/downloader/youtube"
)

// MetadataParser handles parsing of youtube-dl JSON output into structured data.
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

	durationMs, ok := videoData["duration"].(float64)
	if !ok || durationMs <= 0 {
		// Duration might be missing for some videos, use default
		durationMs = 0
	}

	formats, err := p.parseFormats(videoData["formats"])
	if err != nil {
		return nil, fmt.Errorf("failed to parse formats: %w", err)
	}

	return &youtube.Video{
		ID:        id,
		Title:     title,
		Duration:  time.Duration(durationMs) * time.Second,
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
	itag, ok := formatData["format_id"].(string)
	if !ok || itag == "" {
		return youtube.Format{}, errors.New("invalid format_id")
	}

	itagInt, err := strconv.Atoi(itag)
	if err != nil {
		return youtube.Format{}, fmt.Errorf("invalid format_id format: %w", err)
	}

	mimeType, _ := formatData["mime_type"].(string)
	quality, _ := formatData["format_note"].(string)

	var bitrate int
	if bitrateRaw, ok := formatData["tbr"].(float64); ok {
		bitrate = int(bitrateRaw)
	}

	audioCodec, _ := formatData["acodec"].(string)
	videoCodec, _ := formatData["vcodec"].(string)

	// Determine format type
	audioOnly := audioCodec != "" && videoCodec == "none"
	videoOnly := videoCodec != "" && audioCodec == "none"

	return youtube.Format{
		Itag:       itagInt,
		MimeType:   mimeType,
		Quality:    quality,
		Bitrate:    bitrate,
		AudioOnly:  audioOnly,
		VideoOnly:  videoOnly,
		AudioCodec: audioCodec,
		VideoCodec: videoCodec,
	}, nil
}
