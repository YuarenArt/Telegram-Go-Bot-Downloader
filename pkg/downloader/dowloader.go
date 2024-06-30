package downloader

type VideoInfo interface {
}

type Downloader interface {
	GetVideo(videoURL string) (VideoInfo, error)
}
