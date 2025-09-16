package youtube

// Common constants used across the YouTube handler
const (
	// Download types
	All_video = "allVideo"
	All_audio = "allAudio"

	// Traffic limits
	TrafficLimit = 5000.0 // Mb

	// Default values
	DefaultDuration = 180.0 // seconds for size estimation

	// File extensions
	SupportedVideoExtensions = ".mp4,.avi,.mov,.mkv"
	SupportedAudioExtensions = ".weba,.mp3,.m4a,.wav,.ogg"

	// YouTube URL patterns
	YouTubeDomain       = "youtube.com"
	YouTubeWWW          = "www.youtube.com"
	YouTubeShort        = "youtu.be"
	YouTubeWatchPath    = "/watch"
	YouTubePlaylistPath = "/playlist"
	YouTubeLivePath     = "/live"

	// Command timeouts
	DefaultContextTimeout = 60  // seconds
	DownloadTimeout       = 300 // seconds

	// File size calculation
	BytesPerMB = 8 * 1024 * 1024 // 8 MB in bits
)
