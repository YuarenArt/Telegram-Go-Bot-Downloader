package youtube

type Downloader interface {
	GetVideo(url string) (*Video, error)
	GetPlaylist(url string) (*Playlist, error)
	DownloadVideo(video *Video, format any) (string, error)
	DownloadAudio(video *Video, format any) (string, error)
}

type Video struct {
	ID       string
	Title    string
	Formats  []any
	Duration int64
}

type Playlist struct {
	ID     string
	Title  string
	Videos []Video
}
