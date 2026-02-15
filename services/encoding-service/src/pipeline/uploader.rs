use std::path::Path;
use std::sync::Arc;

use tracing::{info, warn};

use crate::domain::profile::EncodingProfile;
use crate::error::Result;
use crate::messaging::models::EncodingOutput;
use crate::storage::StorageClient;

use super::{ThumbnailResult, TranscodeResult};

pub struct Uploader {
    storage: Arc<dyn StorageClient>,
    encoded_bucket: String,
    thumbnails_bucket: String,
}

impl Uploader {
    pub fn new(
        storage: Arc<dyn StorageClient>,
        encoded_bucket: String,
        thumbnails_bucket: String,
    ) -> Self {
        Self {
            storage,
            encoded_bucket,
            thumbnails_bucket,
        }
    }

    /// Upload all transcoded segments, playlists, and thumbnails to MinIO.
    /// Returns the list of EncodingOutput for the result message.
    #[tracing::instrument(skip_all, fields(content_id = %content_id))]
    pub async fn upload_all(
        &self,
        content_id: &str,
        transcode_results: &[TranscodeResult],
        thumbnail_result: &ThumbnailResult,
    ) -> Result<Vec<EncodingOutput>> {
        let mut outputs = Vec::new();

        // Upload each quality's segments + playlist
        for result in transcode_results {
            let profile = EncodingProfile::for_quality(result.quality);
            let quality_str = result.quality.as_str();
            let base_key = format!("{}/{}", content_id, quality_str);

            // Upload playlist
            let playlist_key = format!("{}/playlist.m3u8", base_key);
            self.storage
                .upload_from_file(
                    &self.encoded_bucket,
                    &playlist_key,
                    &result.playlist_path,
                    "application/vnd.apple.mpegurl",
                )
                .await?;

            // Upload segments
            let mut entries = tokio::fs::read_dir(&result.output_dir).await?;
            while let Some(entry) = entries.next_entry().await? {
                let file_name = entry.file_name();
                let file_name_str = file_name.to_string_lossy();

                if file_name_str.ends_with(".ts") {
                    let segment_key = format!("{}/{}", base_key, file_name_str);
                    self.storage
                        .upload_from_file(
                            &self.encoded_bucket,
                            &segment_key,
                            &entry.path(),
                            "video/mp2t",
                        )
                        .await?;
                }
            }

            info!(
                quality = quality_str,
                segments = result.segment_count,
                "uploaded quality"
            );

            outputs.push(EncodingOutput {
                quality: quality_str.to_string(),
                width: profile.width,
                height: profile.height,
                bitrate_kbps: profile.video_bitrate_kbps,
                segment_count: result.segment_count as i32,
                playlist_path: playlist_key,
            });
        }

        // Upload thumbnails
        self.upload_thumbnails(content_id, thumbnail_result).await?;

        Ok(outputs)
    }

    async fn upload_thumbnails(
        &self,
        content_id: &str,
        thumbnail_result: &ThumbnailResult,
    ) -> Result<()> {
        let base_key = content_id;

        // Upload poster
        let poster_key = format!("{}/poster.jpg", base_key);
        self.storage
            .upload_from_file(
                &self.thumbnails_bucket,
                &poster_key,
                &thumbnail_result.poster_path,
                "image/jpeg",
            )
            .await?;

        // Upload thumbnails
        for thumb in &thumbnail_result.thumbnails {
            let thumb_key = format!("{}/{}", base_key, thumb.name);
            self.storage
                .upload_from_file(
                    &self.thumbnails_bucket,
                    &thumb_key,
                    &thumb.path,
                    "image/jpeg",
                )
                .await?;
        }

        info!(
            count = thumbnail_result.thumbnails.len() + 1,
            "uploaded thumbnails"
        );

        Ok(())
    }

    /// Clean up the job's temp directory.
    pub async fn cleanup(job_dir: &Path) {
        if let Err(e) = tokio::fs::remove_dir_all(job_dir).await {
            warn!(path = ?job_dir, error = %e, "failed to clean up temp directory");
        } else {
            info!(path = ?job_dir, "cleaned up temp directory");
        }
    }
}
