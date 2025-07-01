package youtube

import (
	"youtube_downloader/internal/downloader/youtube"

	yt "github.com/kkdai/youtube/v2"
)

type KKDAIDownloader struct {
	client *yt.Client
}

func NewKKDAIDownloader() *KKDAIDownloader {
	return &KKDAIDownloader{
		client: &yt.Client{},
	}
}

func (d *KKDAIDownloader) GetVideo(url string) (*youtube.Video, error) {
	v, err := d.client.GetVideo(url)
	if err != nil {
		return nil, err
	}
	return convertVideoToGeneric(v), nil
}

func (d *KKDAIDownloader) GetPlaylist(url string) (*youtube.Playlist, error) {
	p, err := d.client.GetPlaylist(url)
	if err != nil {
		return nil, err
	}
	videos := make([]youtube.Video, 0, len(p.Videos))
	for _, entry := range p.Videos {
		v, err := d.client.VideoFromPlaylistEntry(entry)
		if err != nil {
			continue
		}
		videos = append(videos, *convertVideoToGeneric(v))
	}
	return &youtube.Playlist{
		ID:     p.ID,
		Title:  p.Title,
		Videos: videos,
	}, nil
}

func (d *KKDAIDownloader) DownloadVideo(video *youtube.Video, format any) (string, error) {
	kkdaiVideo := convertGenericToKKDAIVideo(video)
	kkdaiFormat := convertAnyToKKDAIFormat(format)
	ytd := NewYouTubeDownloader()
	return ytd.DownloadWithFormat(kkdaiVideo, kkdaiFormat)
}

func (d *KKDAIDownloader) DownloadAudio(video *youtube.Video, format any) (string, error) {
	kkdaiVideo := convertGenericToKKDAIVideo(video)
	kkdaiFormat := convertAnyToKKDAIFormat(format)
	ytd := NewYouTubeDownloader()
	return ytd.DownloadWithFormat(kkdaiVideo, kkdaiFormat)
}

func convertVideoToGeneric(v *yt.Video) *youtube.Video {
	formats := make([]any, len(v.Formats))
	for i, f := range v.Formats {
		formats[i] = map[string]any{
			"ItagNo":        f.ItagNo,
			"MimeType":      f.MimeType,
			"QualityLabel":  f.QualityLabel,
			"Bitrate":       f.Bitrate,
			"AudioChannels": f.AudioChannels,
		}
	}
	return &youtube.Video{
		ID:       v.ID,
		Title:    v.Title,
		Formats:  formats,
		Duration: int64(v.Duration.Seconds()),
	}
}

func convertGenericToKKDAIVideo(v *youtube.Video) *yt.Video {
	return &yt.Video{
		ID:    v.ID,
		Title: v.Title,
	}
}

func convertAnyToKKDAIFormat(a any) yt.Format {
	m, ok := a.(map[string]any)
	if !ok {
		return yt.Format{}
	}
	return yt.Format{
		ItagNo:        m["ItagNo"].(int),
		MimeType:      m["MimeType"].(string),
		QualityLabel:  m["QualityLabel"].(string),
		Bitrate:       m["Bitrate"].(int),
		AudioChannels: m["AudioChannels"].(int),
	}
}
