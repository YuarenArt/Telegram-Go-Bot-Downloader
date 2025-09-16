package youtube

import (
	"context"
	"net/http"
	"time"
)

// Downloader defines core methods to fetch metadata and download.
type Downloader interface {
	// GetVideo retrieves metadata for a single video URL.
	GetVideo(ctx context.Context, videoURL string) (*Video, error)
	// GetPlaylist retrieves metadata for a playlist URL.
	GetPlaylist(ctx context.Context, playlistURL string) (*Playlist, error)
	// Download fetches raw bytes of a video or audio according to options.
	Download(ctx context.Context, video *Video, opts DownloadOptions) (string, error)
}

// Video holds generic metadata about a single video.
type Video struct {
	ID        string
	Title     string
	Duration  time.Duration
	Formats   []Format
	SourceURL string
}

// Playlist holds metadata about a playlist.
type Playlist struct {
	ID     string
	Title  string
	Videos []*Video
}

// Format describes a specific stream format.
type Format struct {
	Itag           int
	FormatID       string
	Ext            string
	Height         int
	Width          int
	AudioOnly      bool
	VideoOnly      bool
	Bitrate        int
	MimeType       string
	Quality        string
	AudioCodec     string
	VideoCodec     string
	FileSize       int64
	FileSizeApprox int64
}

// DownloadOptions groups parameters for Download.
type DownloadOptions struct {
	Format    string // Itag or custom identifier
	AudioOnly bool
	OutputDir string
	Filename  string
	Merge     bool // merge audio+video if separate
	ExtraArgs []string
}

// CookieManager handles loading, saving and applying cookies.
type CookieManager interface {
	// LoadCookies loads cookies from a Netscape‐format file.
	LoadCookies(path string) error
	// SaveCookies writes current Jar to a file.
	SaveCookies(path string) error
	// HTTPClient returns an http.Client with cookiejar applied.
	HTTPClient() *http.Client
}
