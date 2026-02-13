use std::path::{Path, PathBuf};

use tokio::io::{AsyncBufReadExt, BufReader};
use tracing::info;

use crate::domain::profile::EncodingProfile;
use crate::error::{EncodingError, Result};

use super::{TranscodeResult, VideoMetadata};

pub struct Transcoder {
    ffmpeg_path: String,
    temp_dir: PathBuf,
}

impl Transcoder {
    pub fn new(ffmpeg_path: String, temp_dir: PathBuf) -> Self {
        Self {
            ffmpeg_path,
            temp_dir,
        }
    }

    /// Transcode source video into HLS segments for all target qualities.
    /// Runs profiles in parallel via tokio::spawn.
    pub async fn transcode_all(
        &self,
        source_path: &Path,
        job_id: &str,
        metadata: &VideoMetadata,
    ) -> Result<Vec<TranscodeResult>> {
        let mut handles = Vec::new();

        for quality in &metadata.target_qualities {
            let profile = EncodingProfile::for_quality(*quality);
            let output_dir = self.temp_dir.join(job_id).join(quality.as_str());
            tokio::fs::create_dir_all(&output_dir).await?;

            let ffmpeg_path = self.ffmpeg_path.clone();
            let source = source_path.to_path_buf();
            let duration_secs = metadata.duration_secs;
            let quality_name = quality.as_str().to_string();

            let handle = tokio::spawn(async move {
                transcode_quality(ffmpeg_path, source, output_dir, profile, duration_secs, quality_name).await
            });
            handles.push((*quality, handle));
        }

        let mut results = Vec::new();
        for (quality, handle) in handles {
            let result = handle.await.map_err(|e| {
                EncodingError::Transcode(format!("{} transcode task panicked: {}", quality.as_str(), e))
            })??;
            results.push(result);
        }

        Ok(results)
    }
}

async fn transcode_quality(
    ffmpeg_path: String,
    source_path: PathBuf,
    output_dir: PathBuf,
    profile: EncodingProfile,
    duration_secs: f64,
    quality_name: String,
) -> Result<TranscodeResult> {
    let playlist_path = output_dir.join("playlist.m3u8");
    let segment_pattern = output_dir.join("segment_%03d.ts");

    let video_bitrate = format!("{}k", profile.video_bitrate_kbps);
    let max_rate = format!("{}k", (profile.video_bitrate_kbps as f64 * 1.1) as i32);
    let buf_size = format!("{}k", profile.video_bitrate_kbps * 2);
    let audio_bitrate = format!("{}k", profile.audio_bitrate_kbps);
    let scale_filter = format!(
        "scale={}:{}:force_original_aspect_ratio=decrease,pad={}:{}:(ow-iw)/2:(oh-ih)/2",
        profile.width, profile.height, profile.width, profile.height
    );
    let hls_time = profile.segment_duration_secs.to_string();

    info!(quality = %quality_name, "starting transcode");

    let mut child = tokio::process::Command::new(&ffmpeg_path)
        .args([
            "-i",
            source_path.to_str().unwrap_or_default(),
            "-c:v",
            "libx264",
            "-b:v",
            &video_bitrate,
            "-maxrate",
            &max_rate,
            "-bufsize",
            &buf_size,
            "-vf",
            &scale_filter,
            "-c:a",
            "aac",
            "-b:a",
            &audio_bitrate,
            "-ar",
            "48000",
            "-preset",
            "medium",
            "-profile:v",
            "main",
            "-level",
            "4.0",
            "-f",
            "hls",
            "-hls_time",
            &hls_time,
            "-hls_list_size",
            "0",
            "-hls_segment_type",
            "mpegts",
            "-hls_segment_filename",
            segment_pattern.to_str().unwrap_or_default(),
            "-progress",
            "pipe:1",
            "-y",
            playlist_path.to_str().unwrap_or_default(),
        ])
        .stdout(std::process::Stdio::piped())
        .stderr(std::process::Stdio::piped())
        .spawn()
        .map_err(|e| EncodingError::Transcode(format!("failed to start ffmpeg: {}", e)))?;

    // Parse progress from stdout
    let stdout = child.stdout.take();
    let total_duration_us = (duration_secs * 1_000_000.0) as u64;
    let qname = quality_name.clone();

    let progress_handle = tokio::spawn(async move {
        if let Some(stdout) = stdout {
            let reader = BufReader::new(stdout);
            let mut lines = reader.lines();
            let mut last_logged_decile: u32 = 0;

            while let Ok(Some(line)) = lines.next_line().await {
                if let Some(time_str) = line.strip_prefix("out_time_us=") {
                    if let Ok(out_time_us) = time_str.trim().parse::<u64>() {
                        if total_duration_us > 0 {
                            let pct =
                                (out_time_us as f64 / total_duration_us as f64 * 100.0).min(100.0);
                            let decile = pct as u32 / 10;
                            if decile > last_logged_decile {
                                info!(
                                    quality = %qname,
                                    progress = format!("{:.0}%", pct),
                                    "transcoding progress"
                                );
                                last_logged_decile = decile;
                            }
                        }
                    }
                }
            }
        }
    });

    let output = child
        .wait_with_output()
        .await
        .map_err(|e| EncodingError::Transcode(format!("ffmpeg process error: {}", e)))?;

    // Wait for progress reader to finish (ignore errors)
    let _ = progress_handle.await;

    if !output.status.success() {
        let stderr = String::from_utf8_lossy(&output.stderr);
        return Err(EncodingError::Transcode(format!(
            "{} transcode failed (exit {}): {}",
            quality_name,
            output.status,
            stderr.chars().take(1000).collect::<String>()
        )));
    }

    // Count .ts segment files
    let segment_count = count_segments(&output_dir).await?;

    if segment_count == 0 {
        return Err(EncodingError::Transcode(format!(
            "{} transcode produced no segments",
            quality_name
        )));
    }

    info!(
        quality = %quality_name,
        segment_count,
        "transcode completed"
    );

    Ok(TranscodeResult {
        quality: profile.quality,
        output_dir,
        playlist_path,
        segment_count,
    })
}

async fn count_segments(dir: &Path) -> Result<u32> {
    let mut count = 0u32;
    let mut entries = tokio::fs::read_dir(dir).await?;
    while let Some(entry) = entries.next_entry().await? {
        if entry
            .file_name()
            .to_string_lossy()
            .ends_with(".ts")
        {
            count += 1;
        }
    }
    Ok(count)
}
