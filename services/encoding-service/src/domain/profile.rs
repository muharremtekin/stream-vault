use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum Quality {
    #[serde(rename = "360p")]
    Q360p,
    #[serde(rename = "720p")]
    Q720p,
    #[serde(rename = "1080p")]
    Q1080p,
    #[serde(rename = "4k")]
    Q4K,
}

impl Quality {
    pub fn as_str(&self) -> &'static str {
        match self {
            Quality::Q360p => "360p",
            Quality::Q720p => "720p",
            Quality::Q1080p => "1080p",
            Quality::Q4K => "4k",
        }
    }

    pub fn all() -> &'static [Quality] {
        &[Quality::Q360p, Quality::Q720p, Quality::Q1080p, Quality::Q4K]
    }

    /// Return the profiles that make sense for a given source height.
    /// e.g. 720p source → only 360p + 720p
    pub fn for_source_height(height: u32) -> Vec<Quality> {
        Quality::all()
            .iter()
            .copied()
            .filter(|q| EncodingProfile::for_quality(*q).height as u32 <= height)
            .collect()
    }
}

#[derive(Debug, Clone)]
pub struct EncodingProfile {
    pub quality: Quality,
    pub width: i32,
    pub height: i32,
    pub video_bitrate_kbps: i32,
    pub audio_bitrate_kbps: i32,
    pub segment_duration_secs: u32,
}

impl EncodingProfile {
    pub fn for_quality(quality: Quality) -> Self {
        match quality {
            Quality::Q360p => Self {
                quality,
                width: 640,
                height: 360,
                video_bitrate_kbps: 800,
                audio_bitrate_kbps: 96,
                segment_duration_secs: 10,
            },
            Quality::Q720p => Self {
                quality,
                width: 1280,
                height: 720,
                video_bitrate_kbps: 2800,
                audio_bitrate_kbps: 128,
                segment_duration_secs: 10,
            },
            Quality::Q1080p => Self {
                quality,
                width: 1920,
                height: 1080,
                video_bitrate_kbps: 5000,
                audio_bitrate_kbps: 192,
                segment_duration_secs: 10,
            },
            Quality::Q4K => Self {
                quality,
                width: 3840,
                height: 2160,
                video_bitrate_kbps: 14000,
                audio_bitrate_kbps: 256,
                segment_duration_secs: 10,
            },
        }
    }

    pub fn all() -> Vec<Self> {
        Quality::all().iter().map(|q| Self::for_quality(*q)).collect()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_quality_for_source_height() {
        let profiles = Quality::for_source_height(720);
        assert_eq!(profiles.len(), 2);
        assert_eq!(profiles[0], Quality::Q360p);
        assert_eq!(profiles[1], Quality::Q720p);
    }

    #[test]
    fn test_quality_for_source_height_1080() {
        let profiles = Quality::for_source_height(1080);
        assert_eq!(profiles.len(), 3);
    }

    #[test]
    fn test_quality_for_source_height_4k() {
        let profiles = Quality::for_source_height(2160);
        assert_eq!(profiles.len(), 4);
    }

    #[test]
    fn test_encoding_profile_specs() {
        let p = EncodingProfile::for_quality(Quality::Q720p);
        assert_eq!(p.width, 1280);
        assert_eq!(p.height, 720);
        assert_eq!(p.video_bitrate_kbps, 2800);
        assert_eq!(p.audio_bitrate_kbps, 128);
        assert_eq!(p.segment_duration_secs, 10);
    }
}
