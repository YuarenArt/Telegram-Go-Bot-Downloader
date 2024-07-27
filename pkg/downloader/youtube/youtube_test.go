package youtube

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/kkdai/youtube/v2"
	"github.com/kkdai/youtube/v2/downloader"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestCase struct {
	name          string
	videoURL      string
	expectedErr   bool
	expectedTitle string
}

type PlaylistTestCase struct {
	name            string
	playlistURL     string
	expectedErr     bool
	expectedTitle   string
	expectedEntries int
}

var playlistTestCases = []PlaylistTestCase{
	{
		name:            "Test playlist case 1",
		playlistURL:     "https://youtube.com/playlist?list=PLGWn6fd74osw8DeWrcvgVopsaRiZHt84P&si=EenX2YfEXYPJ8FCM",
		expectedErr:     false,
		expectedTitle:   "Warhammer 40k Gym Music",
		expectedEntries: 22,
	},
	{
		name:            "Test playlist case 2",
		playlistURL:     "https://www.youtube.com/playlist?list=INVALID_ID",
		expectedErr:     true,
		expectedTitle:   "",
		expectedEntries: 0,
	},
}

var testCases = []TestCase{
	{
		name:          "Test case 1",
		videoURL:      "https://youtu.be/wPdX66-Ag2s?si=ldaZVeoXdaPMzFyf",
		expectedErr:   false,
		expectedTitle: "JoJo's Bizarre Adventure - Opening 1 [4K 60FPS | Creditless | CC]",
	},
	{
		name:          "Test case 2",
		videoURL:      "https://youtu.be/HkJezWe8naI?si=pAYzWiGHFXn8-9MY",
		expectedErr:   false,
		expectedTitle: "The Kitchen Massacre | Sausage Party | CineClips",
	},
	{
		name:          "Test case 3",
		videoURL:      "https://youtu.be/HkJezGHFXn8-9MY",
		expectedErr:   true,
		expectedTitle: "",
	},
}

var ytd = &YouTubeDownloader{
	Downloader: downloader.Downloader{
		Client: youtube.Client{},
	},
}

func TestGetPlaylist(t *testing.T) {
	for _, tc := range playlistTestCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			playlist, err := ytd.GetPlaylist(tc.playlistURL)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedTitle, playlist.Title)
				assert.NotNil(t, playlist)
				assert.NotEmpty(t, playlist.Title)
				assert.Equal(t, tc.expectedEntries, len(playlist.Videos))
			}
		})
	}
}

func TestGetVideo(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			video, err := ytd.GetVideo(tc.videoURL)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedTitle, video.Title)
				assert.NotNil(t, video)
				assert.NotEmpty(t, video.Title)
			}
		})
	}
}

func TestDownloadVideoWithFormatComposite(t *testing.T) {
	downloader := NewYouTubeDownloader()
	video, _ := downloader.Downloader.GetVideo("https://youtu.be/LXb3EKWsInQ?si=b5U_ILrEtBbOa_zZ")
	downloader.DownloadVideoWithFormatComposite(context.Background(), "", video,
		"720", "", "")
	// video/mp4; codecs="avc1.640028"
}
