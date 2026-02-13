package hls

import (
	"strings"
	"testing"
)

func TestGenerateMasterPlaylist_AbsoluteURIs(t *testing.T) {
	profiles := []QualityProfile{
		{Label: "720p", Width: 1280, Height: 720, BitrateKbps: 2500},
		{Label: "1080p", Width: 1920, Height: 1080, BitrateKbps: 5000},
	}

	result := GenerateMasterPlaylist("movie-123", profiles)

	if !strings.Contains(result, "/stream/movie-123/720p/playlist.m3u8") {
		t.Errorf("expected absolute URI with contentId for 720p, got:\n%s", result)
	}
	if !strings.Contains(result, "/stream/movie-123/1080p/playlist.m3u8") {
		t.Errorf("expected absolute URI with contentId for 1080p, got:\n%s", result)
	}
}

func TestGenerateMasterPlaylist_ContainsStreamInf(t *testing.T) {
	profiles := []QualityProfile{
		{Label: "360p", Width: 640, Height: 360, BitrateKbps: 800},
	}

	result := GenerateMasterPlaylist("content-abc", profiles)

	if !strings.Contains(result, "#EXTM3U") {
		t.Error("missing #EXTM3U header")
	}
	if !strings.Contains(result, "#EXT-X-STREAM-INF:BANDWIDTH=800000,RESOLUTION=640x360") {
		t.Errorf("missing or incorrect STREAM-INF tag, got:\n%s", result)
	}
	if !strings.Contains(result, "/stream/content-abc/360p/playlist.m3u8") {
		t.Errorf("expected absolute URI for 360p, got:\n%s", result)
	}
}

func TestGenerateMasterPlaylist_NoRelativeURIs(t *testing.T) {
	profiles := []QualityProfile{
		{Label: "720p", Width: 1280, Height: 720, BitrateKbps: 2500},
	}

	result := GenerateMasterPlaylist("test-id", profiles)

	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.HasSuffix(line, "/playlist.m3u8") && !strings.HasPrefix(line, "/stream/") {
			t.Errorf("found relative URI: %s — expected absolute path starting with /stream/", line)
		}
	}
}
