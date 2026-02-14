use std::path::PathBuf;
use std::sync::Arc;
use std::time::Instant;

use async_trait::async_trait;
use tracing::{error, info};
use uuid::Uuid;

use crate::config::{EncodingConfig, MinIOConfig};
use crate::domain::job::{Job, JobOutput};
use crate::domain::status::JobStatus;
use crate::error::Result;
use crate::messaging::consumer::JobProcessor;
use crate::messaging::models::{EncodingJob, EncodingResult};
use crate::messaging::publisher::ResultPublisher;
use crate::storage::StorageClient;
use crate::store::{self, JobStore};

use super::thumbnail::ThumbnailGenerator;
use super::transcoder::Transcoder;
use super::uploader::Uploader;
use super::validator::Validator;

pub struct PipelineOrchestrator {
    validator: Validator,
    transcoder: Transcoder,
    thumbnail_gen: ThumbnailGenerator,
    uploader: Uploader,
    publisher: Arc<dyn ResultPublisher>,
    storage: Arc<dyn StorageClient>,
    raw_bucket: String,
    temp_dir: PathBuf,
    store: JobStore,
}

impl PipelineOrchestrator {
    pub fn new(
        encoding_config: &EncodingConfig,
        minio_config: &MinIOConfig,
        storage: Arc<dyn StorageClient>,
        publisher: Arc<dyn ResultPublisher>,
        store: JobStore,
    ) -> Self {
        let temp_dir = PathBuf::from(&encoding_config.temp_dir);
        Self {
            validator: Validator::new(encoding_config.ffprobe_path.clone()),
            transcoder: Transcoder::new(encoding_config.ffmpeg_path.clone(), temp_dir.clone()),
            thumbnail_gen: ThumbnailGenerator::new(encoding_config.ffmpeg_path.clone()),
            uploader: Uploader::new(
                Arc::clone(&storage),
                minio_config.encoded_bucket.clone(),
                minio_config.thumbnails_bucket.clone(),
            ),
            publisher,
            storage,
            raw_bucket: minio_config.raw_bucket.clone(),
            temp_dir,
            store,
        }
    }

    async fn run_pipeline(
        &self,
        job: &EncodingJob,
        job_dir: &std::path::Path,
    ) -> Result<Vec<crate::messaging::models::EncodingOutput>> {
        // Step 0: Create job temp directory
        tokio::fs::create_dir_all(job_dir).await?;

        // Step 1: Download source file from MinIO
        store::update_status(&self.store, &job.job_id, JobStatus::Processing, "downloading", 5.0).await;
        let source_path = job_dir.join("source.mp4");
        info!(job_id = %job.job_id, step = "download", "downloading source file");
        let bytes = self
            .storage
            .download_to_file(&self.raw_bucket, &job.source_path, &source_path)
            .await?;
        info!(job_id = %job.job_id, bytes, "source file downloaded");

        // Step 2: Validate
        store::update_status(&self.store, &job.job_id, JobStatus::Processing, "validating", 15.0).await;
        info!(job_id = %job.job_id, step = "validate", "validating source file");
        let metadata = self.validator.validate(&source_path).await?;

        // Step 3: Transcode + Segment (combined, parallel across qualities)
        store::update_status(&self.store, &job.job_id, JobStatus::Processing, "transcoding", 25.0).await;
        info!(job_id = %job.job_id, step = "transcode", qualities = metadata.target_qualities.len(), "starting transcode");
        let transcode_results = self
            .transcoder
            .transcode_all(&source_path, &job.job_id, &metadata)
            .await?;
        info!(job_id = %job.job_id, "transcode completed");

        // Step 4: Generate thumbnails
        store::update_status(&self.store, &job.job_id, JobStatus::Processing, "thumbnailing", 75.0).await;
        info!(job_id = %job.job_id, step = "thumbnail", "generating thumbnails");
        let thumb_dir = job_dir.join("thumbnails");
        let thumbnail_result = self
            .thumbnail_gen
            .generate(&source_path, metadata.duration_secs, &thumb_dir)
            .await?;

        // Step 5: Upload everything to MinIO
        store::update_status(&self.store, &job.job_id, JobStatus::Processing, "uploading", 85.0).await;
        info!(job_id = %job.job_id, step = "upload", "uploading to storage");
        let outputs = self
            .uploader
            .upload_all(&job.content_id, &transcode_results, &thumbnail_result)
            .await?;
        info!(job_id = %job.job_id, output_count = outputs.len(), "upload completed");

        Ok(outputs)
    }
}

#[async_trait]
impl JobProcessor for PipelineOrchestrator {
    async fn process(&self, job: EncodingJob) -> Result<()> {
        let start = Instant::now();
        let job_dir = self.temp_dir.join(&job.job_id);

        // Insert job into store
        let domain_job = Job::from_encoding_job(&job);
        store::insert_job(&self.store, domain_job).await;

        info!(
            job_id = %job.job_id,
            content_id = %job.content_id,
            source = %job.source_path,
            "starting encoding pipeline"
        );

        let result = self.run_pipeline(&job, &job_dir).await;

        match result {
            Ok(outputs) => {
                let elapsed = start.elapsed();
                info!(
                    job_id = %job.job_id,
                    content_id = %job.content_id,
                    qualities = outputs.len(),
                    elapsed_secs = elapsed.as_secs_f64(),
                    "encoding pipeline completed"
                );

                // Update store with completed status
                let domain_outputs: Vec<JobOutput> = outputs.iter().map(JobOutput::from).collect();
                store::complete_job(&self.store, &job.job_id, domain_outputs).await;

                let encoding_result = EncodingResult {
                    event_id: Uuid::new_v4().to_string(),
                    event_type: "encoding.job.completed".into(),
                    timestamp: chrono::Utc::now().to_rfc3339(),
                    source: "encoding-service".into(),
                    correlation_id: job.job_id.clone(),
                    job_id: job.job_id.clone(),
                    content_id: job.content_id.clone(),
                    status: "completed".into(),
                    outputs,
                    duration_seconds: elapsed.as_secs() as i64,
                    error_message: None,
                    completed_at: chrono::Utc::now().to_rfc3339(),
                };

                self.publisher.publish_completed(encoding_result).await?;
                Uploader::cleanup(&job_dir).await;
                Ok(())
            }
            Err(e) => {
                let elapsed = start.elapsed();
                error!(
                    job_id = %job.job_id,
                    content_id = %job.content_id,
                    error = %e,
                    elapsed_secs = elapsed.as_secs_f64(),
                    "encoding pipeline failed"
                );

                // Update store with failed status
                store::fail_job(&self.store, &job.job_id, &format!("{}", e)).await;

                let encoding_result = EncodingResult {
                    event_id: Uuid::new_v4().to_string(),
                    event_type: "encoding.job.failed".into(),
                    timestamp: chrono::Utc::now().to_rfc3339(),
                    source: "encoding-service".into(),
                    correlation_id: job.job_id.clone(),
                    job_id: job.job_id.clone(),
                    content_id: job.content_id.clone(),
                    status: "failed".into(),
                    outputs: vec![],
                    duration_seconds: elapsed.as_secs() as i64,
                    error_message: Some(format!("{}", e)),
                    completed_at: chrono::Utc::now().to_rfc3339(),
                };

                // Best-effort publish failure result
                if let Err(pub_err) = self.publisher.publish_failed(encoding_result).await {
                    error!(error = %pub_err, "failed to publish failure result");
                }

                Uploader::cleanup(&job_dir).await;

                // Return error so consumer can retry
                Err(e)
            }
        }
    }
}
