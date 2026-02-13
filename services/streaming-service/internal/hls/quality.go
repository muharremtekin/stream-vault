package hls

import "strings"

type QualityProfile struct {
	Label       string
	Width       int
	Height      int
	BitrateKbps int
	MinTier     string // "Basic", "Standard", "Premium"
}

var AllProfiles = []QualityProfile{
	{Label: "360p", Width: 640, Height: 360, BitrateKbps: 800, MinTier: "Basic"},
	{Label: "720p", Width: 1280, Height: 720, BitrateKbps: 2500, MinTier: "Basic"},
	{Label: "1080p", Width: 1920, Height: 1080, BitrateKbps: 5000, MinTier: "Standard"},
	{Label: "4k", Width: 3840, Height: 2160, BitrateKbps: 15000, MinTier: "Premium"},
}

func FilterByTier(tier string, available []QualityProfile) []QualityProfile {
	tierLevel := tierToLevel(tier)
	var filtered []QualityProfile
	for _, p := range available {
		if tierToLevel(p.MinTier) <= tierLevel {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func tierToLevel(tier string) int {
	switch strings.ToLower(tier) {
	case "premium", "subscription_tier_premium":
		return 4
	case "standard", "subscription_tier_standard":
		return 3
	case "basic", "subscription_tier_basic":
		return 2
	default:
		return 1
	}
}

func ProfileByLabel(label string) *QualityProfile {
	for _, p := range AllProfiles {
		if strings.EqualFold(p.Label, label) {
			return &p
		}
	}
	return nil
}
