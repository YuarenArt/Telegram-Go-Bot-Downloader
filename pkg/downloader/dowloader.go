package downloader

type VideoInfo interface {
}

type Downloader interface {
	VideoInfo(videoURL string) (VideoInfo, error)
}
