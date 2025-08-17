package ytdl

import (
	"encoding/json"
	"errors"
	"fmt"
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
	var videoData map[string]interface{}
	if err := json.Unmarshal(data, &videoData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	id, ok := videoData["id"].(string)
	if !ok {
		return nil, errors.New("invalid or missing video ID")
	}

	title, ok := videoData["title"].(string)
	if !ok {
		return nil, errors.New("invalid or missing video title")
	}

	durationMs, ok := videoData["duration"].(float64)
	if !ok {
		return nil, errors.New("invalid or missing video duration")
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
	var playlistData map[string]interface{}
	if err := json.Unmarshal(data, &playlistData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	id, ok := playlistData["id"].(string)
	if !ok {
		return nil, errors.New("invalid or missing playlist ID")
	}

	title, ok := playlistData["title"].(string)
	if !ok {
		return nil, errors.New("invalid or missing playlist title")
	}

	entries, ok := playlistData["entries"].([]interface{})
	if !ok {
		return nil, errors.New("invalid or missing playlist entries")
	}

	videos := make([]*youtube.Video, 0, len(entries))
	for _, entry := range entries {
		videoData, ok := entry.(map[string]interface{})
		if !ok {
			return nil, errors.New("invalid video entry in playlist")
		}

		videoID, ok := videoData["id"].(string)
		if !ok {
			return nil, errors.New("invalid video ID in playlist entry")
		}
		videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)

		videoJSON, err := json.Marshal(videoData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal video data: %w", err)
		}

		video, err := p.ParseVideo(videoJSON, videoURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse video entry: %w", err)
		}
		videos = append(videos, video)
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

	formats := make([]youtube.Format, 0, len(formatsRaw))
	for _, f := range formatsRaw {
		formatData, ok := f.(map[string]interface{})
		if !ok {
			continue
		}

		itag, ok := formatData["format_id"].(string)
		if !ok {
			continue
		}
		itagInt, err := strconv.Atoi(itag)
		if err != nil {
			continue
		}

		mimeType, ok := formatData["mime_type"].(string)
		if !ok {
			mimeType = ""
		}

		quality, ok := formatData["format_note"].(string)
		if !ok {
			quality = ""
		}

		var bitrate int
		if bitrateRaw, ok := formatData["tbr"].(float64); ok {
			bitrate = int(bitrateRaw)
		}

		var audioCodec, videoCodec string
		if acodec, ok := formatData["acodec"].(string); ok {
			audioCodec = acodec
		}
		if vcodec, ok := formatData["vcodec"].(string); ok {
			videoCodec = vcodec
		}

		audioOnly := audioCodec != "" && videoCodec == "none"
		videoOnly := videoCodec != "" && audioCodec == "none"

		formats = append(formats, youtube.Format{
			Itag:       itagInt,
			MimeType:   mimeType,
			Quality:    quality,
			Bitrate:    bitrate,
			AudioOnly:  audioOnly,
			VideoOnly:  videoOnly,
			AudioCodec: audioCodec,
			VideoCodec: videoCodec,
		})
	}

	if len(formats) == 0 {
		return nil, errors.New("no valid formats found")
	}

	return formats, nil
}
