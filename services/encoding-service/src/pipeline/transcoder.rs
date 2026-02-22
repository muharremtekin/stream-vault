use std::path::{Path, PathBuf};
use std::sync::atomic::{AtomicU32, Ordering};
use std::sync::Arc;

use tokio::io::{AsyncBufReadExt, BufReader};
use tokio::sync::Semaphore;
use tokio_util::sync::CancellationToken;
use tracing::{info, warn};

use crate::domain::profile::EncodingProfile;
use crate::domain::status::JobStatus;
use crate::error::{EncodingError, Result};
use crate::store::{self, JobStore};

use super::{TranscodeResult, VideoMetadata};

/// Maximum number of ffmpeg processes running concurrently.
/// Keeps peak memory usage within container limits (2 GB).
const MAX_CONCURRENT_FFMPEG: usize = 2;

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
    /// Runs profiles in parallel via tokio::spawn, updating progress in the
    /// job store as each quality reports FFmpeg progress.
    ///
    /// Progress is mapped to the 15%–75% range (60% budget split across qualities).
    #[tracing::instrument(skip_all, fields(job_id = %job_id, qualities = metadata.target_qualities.len()))]
    pub async fn transcode_all(
        &self,
        source_path: &Path,
        job_id: &str,
        metadata: &VideoMetadata,
        store: &JobStore,
    ) -> Result<Vec<TranscodeResult>> {
        let total_qualities = metadata.target_qualities.len();

        // Shared progress slots: one AtomicU32 per quality (0–100 each).
        // Each parallel FFmpeg task writes its own slot; aggregate is computed
        // from the average of all slots mapped to the 15%–75% range.
        let progress_slots: Arc<Vec<AtomicU32>> = Arc::new(
            (0..total_qualities).map(|_| AtomicU32::new(0)).collect(),
        );

        // Limit concurrent ffmpeg processes to avoid OOM kills in
        // memory-constrained containers (e.g. 4K + 1080p + 720p + 360p).
        let semaphore = Arc::new(Semaphore::new(MAX_CONCURRENT_FFMPEG));

        // Cancellation token: when one quality fails, cancel the others
        // so we don't waste CPU/memory on work that will be discarded.
        let cancel = CancellationToken::new();

        let mut handles = Vec::new();

        for (qi, quality) in metadata.target_qualities.iter().enumerate() {
            let profile = EncodingProfile::for_quality(*quality);
            let output_dir = self.temp_dir.join(job_id).join(quality.as_str());
            tokio::fs::create_dir_all(&output_dir).await?;

            let ffmpeg_path = self.ffmpeg_path.clone();
            let source = source_path.to_path_buf();
            let duration_secs = metadata.duration_secs;
            let quality_name = quality.as_str().to_string();
            let slots = Arc::clone(&progress_slots);
            let store_clone = store.clone();
            let jid = job_id.to_string();
            let sem = Arc::clone(&semaphore);
            let task_cancel = cancel.clone();

            let handle = tokio::spawn(async move {
                let _permit = sem.acquire().await.map_err(|e| {
                    EncodingError::Transcode(format!("semaphore closed: {}", e))
                })?;

                // Check if another quality already failed before starting ffmpeg.
                if task_cancel.is_cancelled() {
                    return Err(EncodingError::Transcode(format!(
                        "{} cancelled: another quality failed", quality_name
                    )));
                }

                let result = transcode_quality(
                    ffmpeg_path, source, output_dir, profile, duration_secs, quality_name.clone(),
                    store_clone, jid, qi, total_qualities, slots, task_cancel.clone(),
                ).await;

                // On failure, signal all sibling tasks to stop.
                if result.is_err() {
                    warn!(quality = %quality_name, "transcode failed, cancelling remaining qualities");
                    task_cancel.cancel();
                }

                result
            });
            handles.push((*quality, handle));
        }

        let mut results = Vec::new();
        let mut first_error: Option<EncodingError> = None;
        for (quality, handle) in handles {
            match handle.await {
                Ok(Ok(result)) => results.push(result),
                Ok(Err(e)) => {
                    if first_error.is_none() {
                        first_error = Some(e);
                    }
                    // Cancel remaining tasks on first real failure.
                    cancel.cancel();
                }
                Err(e) => {
                    if first_error.is_none() {
                        first_error = Some(EncodingError::Transcode(
                            format!("{} transcode task panicked: {}", quality.as_str(), e),
                        ));
                    }
                    cancel.cancel();
                }
            }
        }

        if let Some(err) = first_error {
            return Err(err);
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
    store: JobStore,
    job_id: String,
    quality_index: usize,
    total_qualities: usize,
    progress_slots: Arc<Vec<AtomicU32>>,
    cancel: CancellationToken,
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
        // Kill the ffmpeg process if the future is dropped (e.g. on cancellation).
        .kill_on_drop(true)
        .spawn()
        .map_err(|e| EncodingError::Transcode(format!("failed to start ffmpeg: {}", e)))?;

    // Parse progress from stdout and update job store in real-time.
    // Each quality writes its own slot (0–100); aggregate progress across
    // all parallel qualities is mapped to the 15%–75% overall range.
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

                            // Write this quality's progress to its slot
                            if let Some(slot) = progress_slots.get(quality_index) {
                                slot.store(pct as u32, Ordering::Relaxed);
                            }

                            // Update store every ~10% per quality
                            if decile > last_logged_decile {
                                // Aggregate: average of all qualities → map to 15%–75%
                                let avg: f64 = progress_slots
                                    .iter()
                                    .map(|s| s.load(Ordering::Relaxed) as f64)
                                    .sum::<f64>()
                                    / total_qualities as f64;
                                let overall_pct = 15.0 + (avg * 60.0 / 100.0);
                                let step = format!("transcoding {} {:.0}%", qname, pct);
                                store::update_status(
                                    &store, &job_id, JobStatus::Processing, &step, overall_pct,
                                )
                                .await;

                                info!(
                                    quality = %qname,
                                    progress = format!("{:.0}%", pct),
                                    overall = format!("{:.0}%", overall_pct),
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

    let output = tokio::select! {
        result = child.wait_with_output() => {
            result.map_err(|e| EncodingError::Transcode(format!("ffmpeg process error: {}", e)))?
        }
        _ = cancel.cancelled() => {
            // Another quality failed — kill_on_drop will terminate ffmpeg
            // when `child` is dropped at the end of this scope.
            warn!(quality = %quality_name, "cancelling ffmpeg due to sibling failure");
            // child is already moved into the first branch's future;
            // tokio::select! will drop it (triggering kill_on_drop) when this branch wins.
            return Err(EncodingError::Transcode(format!(
                "{} cancelled: another quality failed", quality_name
            )));
        }
    };

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
