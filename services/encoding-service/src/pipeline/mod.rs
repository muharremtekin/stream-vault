pub mod orchestrator;
pub mod thumbnail;
pub mod transcoder;
pub mod uploader;
pub mod validator;

use std::path::PathBuf;

use crate::domain::profile::Quality;

/// Output of the validation step (ffprobe result + profile selection).
#[derive(Debug, Clone)]
pub struct VideoMetadata {
    pub codec: String,
    pub width: u32,
    pub height: u32,
    pub duration_secs: f64,
    pub file_size_bytes: u64,
    pub target_qualities: Vec<Quality>,
}

/// Output of the transcode+segment step for a single quality level.
#[derive(Debug, Clone)]
pub struct TranscodeResult {
    pub quality: Quality,
    pub output_dir: PathBuf,
    pub playlist_path: PathBuf,
    pub segment_count: u32,
}

/// Output of the thumbnail generation step.
#[derive(Debug, Clone)]
pub struct ThumbnailResult {
    pub poster_path: PathBuf,
    pub thumbnails: Vec<ThumbnailFile>,
}

#[derive(Debug, Clone)]
pub struct ThumbnailFile {
    pub name: String,
    pub path: PathBuf,
    pub width: u32,
    pub height: u32,
}
