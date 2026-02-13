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

func TestGenerateMasterPlaylist_EmptyProfiles(t *testing.T) {
	result := GenerateMasterPlaylist("movie-1", nil)

	if !strings.HasPrefix(result, "#EXTM3U\n") {
		t.Error("empty playlist should still have #EXTM3U header")
	}
	if strings.Contains(result, "EXT-X-STREAM-INF") {
		t.Error("empty playlist should not have stream entries")
	}
}

func TestFilterByTier(t *testing.T) {
	tests := []struct {
		name     string
		tier     string
		expected []string
	}{
		{"Premium gets all", "Premium", []string{"360p", "720p", "1080p", "4k"}},
		{"Standard gets up to 1080p", "Standard", []string{"360p", "720p", "1080p"}},
		{"Basic gets up to 720p", "Basic", []string{"360p", "720p"}},
		{"Free gets nothing", "Free", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := FilterByTier(tt.tier, AllProfiles)
			if len(filtered) != len(tt.expected) {
				t.Fatalf("FilterByTier(%q) returned %d profiles, want %d", tt.tier, len(filtered), len(tt.expected))
			}
			for i, p := range filtered {
				if p.Label != tt.expected[i] {
					t.Errorf("profile[%d].Label = %s, want %s", i, p.Label, tt.expected[i])
				}
			}
		})
	}
}

func TestProfileByLabel(t *testing.T) {
	tests := []struct {
		label string
		found bool
		width int
	}{
		{"360p", true, 640},
		{"720p", true, 1280},
		{"1080p", true, 1920},
		{"4k", true, 3840},
		{"4K", true, 3840}, // case-insensitive
		{"240p", false, 0},
		{"", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			p := ProfileByLabel(tt.label)
			if tt.found && p == nil {
				t.Fatalf("ProfileByLabel(%q) returned nil, expected profile", tt.label)
			}
			if !tt.found && p != nil {
				t.Fatalf("ProfileByLabel(%q) returned profile, expected nil", tt.label)
			}
			if tt.found && p.Width != tt.width {
				t.Errorf("Width = %d, want %d", p.Width, tt.width)
			}
		})
	}
}
