package hls

import (
	"fmt"
	"strings"
)

func GenerateMasterPlaylist(contentId string, profiles []QualityProfile) string {
	var b strings.Builder

	b.WriteString("#EXTM3U\n")
	b.WriteString("#EXT-X-VERSION:3\n")
	b.WriteString("\n")

	for _, p := range profiles {
		bandwidth := p.BitrateKbps * 1000
		b.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d,NAME=\"%s\"\n",
			bandwidth, p.Width, p.Height, p.Label))
		b.WriteString(fmt.Sprintf("/stream/%s/%s/playlist.m3u8\n", contentId, p.Label))
		b.WriteString("\n")
	}

	return b.String()
}
