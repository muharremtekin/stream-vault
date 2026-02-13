use std::path::Path;

use tracing::info;

use crate::error::{EncodingError, Result};

use super::{ThumbnailFile, ThumbnailResult};

pub struct ThumbnailGenerator {
    ffmpeg_path: String,
}

impl ThumbnailGenerator {
    pub fn new(ffmpeg_path: String) -> Self {
        Self { ffmpeg_path }
    }

    /// Generate poster and thumbnails from the source video.
    /// Extracts a frame at 50% of duration, then resizes to thumbnail dimensions.
    pub async fn generate(
        &self,
        source_path: &Path,
        duration_secs: f64,
        output_dir: &Path,
    ) -> Result<ThumbnailResult> {
        tokio::fs::create_dir_all(output_dir).await?;

        // Extract poster at 50% of duration
        let seek_time = duration_secs / 2.0;
        let poster_path = output_dir.join("poster.jpg");

        self.extract_frame(source_path, seek_time, &poster_path)
            .await?;

        info!("poster extracted");

        // Generate thumbnails at two sizes
        let thumb_specs = [(300u32, 170u32), (600u32, 340u32)];
        let mut thumbnails = Vec::new();

        for (w, h) in thumb_specs {
            let name = format!("thumb_{}x{}.jpg", w, h);
            let thumb_path = output_dir.join(&name);

            self.resize_image(&poster_path, &thumb_path, w, h).await?;

            thumbnails.push(ThumbnailFile {
                name,
                path: thumb_path,
                width: w,
                height: h,
            });
        }

        info!(count = thumbnails.len(), "thumbnails generated");

        Ok(ThumbnailResult {
            poster_path,
            thumbnails,
        })
    }

    async fn extract_frame(
        &self,
        source_path: &Path,
        seek_secs: f64,
        output_path: &Path,
    ) -> Result<()> {
        let seek = format!("{:.3}", seek_secs);

        let output = tokio::process::Command::new(&self.ffmpeg_path)
            .args([
                "-ss",
                &seek,
                "-i",
                source_path.to_str().unwrap_or_default(),
                "-frames:v",
                "1",
                "-q:v",
                "2",
                "-y",
                output_path.to_str().unwrap_or_default(),
            ])
            .output()
            .await
            .map_err(|e| {
                EncodingError::Transcode(format!("failed to start ffmpeg for thumbnail: {}", e))
            })?;

        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            return Err(EncodingError::Transcode(format!(
                "frame extraction failed: {}",
                stderr.chars().take(500).collect::<String>()
            )));
        }

        Ok(())
    }

    async fn resize_image(
        &self,
        input_path: &Path,
        output_path: &Path,
        width: u32,
        height: u32,
    ) -> Result<()> {
        let scale_filter = format!(
            "scale={}:{}:force_original_aspect_ratio=decrease,pad={}:{}:(ow-iw)/2:(oh-ih)/2",
            width, height, width, height
        );

        let output = tokio::process::Command::new(&self.ffmpeg_path)
            .args([
                "-i",
                input_path.to_str().unwrap_or_default(),
                "-vf",
                &scale_filter,
                "-q:v",
                "3",
                "-y",
                output_path.to_str().unwrap_or_default(),
            ])
            .output()
            .await
            .map_err(|e| {
                EncodingError::Transcode(format!("failed to start ffmpeg for resize: {}", e))
            })?;

        if !output.status.success() {
            let stderr = String::from_utf8_lossy(&output.stderr);
            return Err(EncodingError::Transcode(format!(
                "thumbnail resize failed: {}",
                stderr.chars().take(500).collect::<String>()
            )));
        }

        Ok(())
    }
}
