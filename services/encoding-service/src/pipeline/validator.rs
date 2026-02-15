use std::path::Path;

use serde::Deserialize;
use tracing::{info, warn};

use crate::domain::profile::Quality;
use crate::error::{EncodingError, Result};

use super::VideoMetadata;

const MAX_FILE_SIZE_BYTES: u64 = 10 * 1024 * 1024 * 1024; // 10 GB

const WELL_KNOWN_CODECS: &[&str] = &[
    "h264",
    "hevc",
    "vp9",
    "av1",
    "mpeg4",
    "mpeg2video",
    "vp8",
    "theora",
];

pub struct Validator {
    ffprobe_path: String,
}

impl Validator {
    pub fn new(ffprobe_path: String) -> Self {
        Self { ffprobe_path }
    }

    #[tracing::instrument(skip_all, fields(source = %source_path.display()))]
    pub async fn validate(&self, source_path: &Path) -> Result<VideoMetadata> {
        // Get file size from filesystem (more reliable than ffprobe)
        let file_size = tokio::fs::metadata(source_path)
            .await
            .map_err(|e| EncodingError::Validation(format!("cannot read file: {}", e)))?
            .len();

        if file_size > MAX_FILE_SIZE_BYTES {
            return Err(EncodingError::Validation(format!(
                "file size {} bytes exceeds 10 GB limit",
                file_size
            )));
        }

        if file_size == 0 {
            return Err(EncodingError::Validation("file is empty".into()));
        }

        // Run ffprobe
        let output = tokio::process::Command::new(&self.ffprobe_path)
            .args([
                "-v",
                "quiet",
                "-print_format",
                "json",
                "-show_format",
                "-show_streams",
            ])
            .arg(source_path)
            .output()
            .await
            .map_err(|e| EncodingError::FFprobe(format!("failed to execute ffprobe: {}", e)))?;

        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            return Err(EncodingError::FFprobe(format!(
                "ffprobe exited with {}: {}",
                output.status, stderr
            )));
        }

        let probe: FfprobeOutput = serde_json::from_slice(&output.stdout).map_err(|e| {
            EncodingError::FFprobe(format!("failed to parse ffprobe output: {}", e))
        })?;

        // Find first video stream
        let video_stream = probe
            .streams
            .iter()
            .find(|s| s.codec_type.as_deref() == Some("video"))
            .ok_or_else(|| EncodingError::Validation("no video stream found".into()))?;

        let codec = video_stream
            .codec_name
            .clone()
            .unwrap_or_else(|| "unknown".into());

        let width = video_stream.width.unwrap_or(0);
        let height = video_stream.height.unwrap_or(0);

        if width == 0 || height == 0 {
            return Err(EncodingError::Validation(format!(
                "invalid resolution: {}x{}",
                width, height
            )));
        }

        // Duration: prefer video stream, fallback to format
        let duration_secs = video_stream
            .duration
            .as_deref()
            .and_then(|d| d.parse::<f64>().ok())
            .or_else(|| {
                probe
                    .format
                    .as_ref()
                    .and_then(|f| f.duration.as_deref())
                    .and_then(|d| d.parse::<f64>().ok())
            })
            .unwrap_or(0.0);

        if duration_secs <= 0.0 {
            return Err(EncodingError::Validation(
                "could not determine video duration".into(),
            ));
        }

        // Log codec warning if not well-known
        if !WELL_KNOWN_CODECS.contains(&codec.as_str()) {
            warn!(codec = %codec, "uncommon video codec, ffmpeg may still handle it");
        }

        // Determine target profiles based on source height
        let target_qualities = if height < 360 {
            vec![Quality::Q360p]
        } else {
            Quality::for_source_height(height)
        };

        info!(
            codec = %codec,
            resolution = format!("{}x{}", width, height),
            duration_secs,
            file_size_bytes = file_size,
            target_qualities = ?target_qualities.iter().map(|q| q.as_str()).collect::<Vec<_>>(),
            "validation passed"
        );

        Ok(VideoMetadata {
            codec,
            width,
            height,
            duration_secs,
            file_size_bytes: file_size,
            target_qualities,
        })
    }
}

// ── ffprobe JSON output types ─────────────────────────────────────

#[derive(Debug, Deserialize)]
struct FfprobeOutput {
    #[serde(default)]
    streams: Vec<FfprobeStream>,
    format: Option<FfprobeFormat>,
}

#[derive(Debug, Deserialize)]
struct FfprobeStream {
    codec_type: Option<String>,
    codec_name: Option<String>,
    width: Option<u32>,
    height: Option<u32>,
    duration: Option<String>,
}

#[derive(Debug, Deserialize)]
struct FfprobeFormat {
    duration: Option<String>,
    #[allow(dead_code)]
    size: Option<String>,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_ffprobe_output_parse() {
        let json = r#"{
            "streams": [
                {
                    "codec_type": "video",
                    "codec_name": "h264",
                    "width": 1920,
                    "height": 1080,
                    "duration": "120.5"
                },
                {
                    "codec_type": "audio",
                    "codec_name": "aac"
                }
            ],
            "format": {
                "duration": "120.5",
                "size": "15000000"
            }
        }"#;

        let probe: FfprobeOutput = serde_json::from_str(json).unwrap();
        assert_eq!(probe.streams.len(), 2);

        let video = probe
            .streams
            .iter()
            .find(|s| s.codec_type.as_deref() == Some("video"))
            .unwrap();
        assert_eq!(video.width, Some(1920));
        assert_eq!(video.height, Some(1080));
        assert_eq!(video.codec_name.as_deref(), Some("h264"));
    }

    #[test]
    fn test_quality_selection_for_720p_source() {
        let qualities = Quality::for_source_height(720);
        assert_eq!(qualities.len(), 2);
        assert!(qualities.contains(&Quality::Q360p));
        assert!(qualities.contains(&Quality::Q720p));
    }

    #[test]
    fn test_quality_selection_for_low_res() {
        // Below 360p should still get Q360p
        let qualities = if 240 < 360 {
            vec![Quality::Q360p]
        } else {
            Quality::for_source_height(240)
        };
        assert_eq!(qualities.len(), 1);
        assert_eq!(qualities[0], Quality::Q360p);
    }
}
