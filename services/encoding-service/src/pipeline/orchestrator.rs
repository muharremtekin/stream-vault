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
use crate::metrics;
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

    #[tracing::instrument(skip_all, fields(job_id = %job.job_id))]
    async fn run_pipeline(
        &self,
        job: &EncodingJob,
        job_dir: &std::path::Path,
    ) -> Result<Vec<crate::messaging::models::EncodingOutput>> {
        // Step 0: Create job temp directory
        tokio::fs::create_dir_all(job_dir).await?;

        // Step 1: Download source file from MinIO (0% → 5%)
        store::update_status(&self.store, &job.job_id, JobStatus::Processing, "downloading", 5.0).await;
        let source_path = job_dir.join("source.mp4");
        info!(job_id = %job.job_id, step = "download", "downloading source file");
        let bytes = self
            .storage
            .download_to_file(&self.raw_bucket, &job.source_path, &source_path)
            .await?;
        info!(job_id = %job.job_id, bytes, "source file downloaded");

        // Step 2: Validate (5% → 10%)
        store::update_status(&self.store, &job.job_id, JobStatus::Processing, "validating", 10.0).await;
        info!(job_id = %job.job_id, step = "validate", "validating source file");
        let metadata = self.validator.validate(&source_path).await?;

        // Step 3: Transcode + Segment (15% → 75%, per-quality progress updated by transcoder)
        store::update_status(&self.store, &job.job_id, JobStatus::Processing, "transcoding", 15.0).await;
        info!(job_id = %job.job_id, step = "transcode", qualities = metadata.target_qualities.len(), "starting transcode");
        let transcode_results = self
            .transcoder
            .transcode_all(&source_path, &job.job_id, &metadata, &self.store)
            .await?;
        info!(job_id = %job.job_id, "transcode completed");

        // Step 4: Generate thumbnails (75% → 80%)
        store::update_status(&self.store, &job.job_id, JobStatus::Processing, "thumbnailing", 80.0).await;
        info!(job_id = %job.job_id, step = "thumbnail", "generating thumbnails");
        let thumb_dir = job_dir.join("thumbnails");
        let thumbnail_result = self
            .thumbnail_gen
            .generate(&source_path, metadata.duration_secs, &thumb_dir)
            .await?;

        // Step 5: Upload everything to MinIO (80% → 90%)
        store::update_status(&self.store, &job.job_id, JobStatus::Processing, "uploading", 90.0).await;
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
    #[tracing::instrument(skip_all, fields(job_id = %job.job_id, content_id = %job.content_id))]
    async fn process(&self, job: EncodingJob) -> Result<()> {
        let start = Instant::now();
        let job_dir = self.temp_dir.join(&job.job_id);

        metrics::ENCODING_ACTIVE_JOBS.inc();

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
                metrics::ENCODING_ACTIVE_JOBS.dec();
                metrics::ENCODING_JOBS_TOTAL.with_label_values(&["completed"]).inc();
                metrics::ENCODING_JOB_DURATION_SECONDS.with_label_values(&["all"]).observe(elapsed.as_secs_f64());
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
                metrics::ENCODING_ACTIVE_JOBS.dec();
                metrics::ENCODING_JOBS_TOTAL.with_label_values(&["failed"]).inc();
                metrics::ENCODING_JOB_DURATION_SECONDS.with_label_values(&["all"]).observe(elapsed.as_secs_f64());
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

#[cfg(test)]
mod tests {
    use super::*;
    use crate::config::{EncodingConfig, MinIOConfig};
    use crate::error::EncodingError;
    use crate::messaging::consumer::JobProcessor;
    use crate::messaging::models::{EncodingJob, EncodingResult};
    use crate::messaging::publisher::ResultPublisher;
    use crate::storage::StorageClient;
    use crate::store;
    use bytes::Bytes;
    use std::path::Path;
    use std::sync::Mutex;

    /// A storage client that always fails on download_to_file.
    struct FailingStorageClient;

    #[async_trait]
    impl StorageClient for FailingStorageClient {
        async fn download_to_file(&self, _: &str, _: &str, _: &Path) -> Result<u64> {
            Err(EncodingError::Storage("simulated download failure".into()))
        }
        async fn upload_from_file(&self, _: &str, _: &str, _: &Path, _: &str) -> Result<()> {
            Ok(())
        }
        async fn upload_bytes(&self, _: &str, _: &str, _: Bytes, _: &str) -> Result<()> {
            Ok(())
        }
        async fn delete(&self, _: &str, _: &str) -> Result<()> {
            Ok(())
        }
        async fn exists(&self, _: &str, _: &str) -> Result<bool> {
            Ok(false)
        }
        async fn health_check(&self, _: &str) -> Result<()> {
            Ok(())
        }
    }

    /// A publisher mock that records calls.
    struct MockPublisher {
        completed_calls: Mutex<Vec<EncodingResult>>,
        failed_calls: Mutex<Vec<EncodingResult>>,
    }

    impl MockPublisher {
        fn new() -> Self {
            Self {
                completed_calls: Mutex::new(Vec::new()),
                failed_calls: Mutex::new(Vec::new()),
            }
        }
    }

    #[async_trait]
    impl ResultPublisher for MockPublisher {
        async fn publish_completed(&self, result: EncodingResult) -> Result<()> {
            self.completed_calls.lock().unwrap().push(result);
            Ok(())
        }
        async fn publish_failed(&self, result: EncodingResult) -> Result<()> {
            self.failed_calls.lock().unwrap().push(result);
            Ok(())
        }
    }

    fn test_encoding_config() -> EncodingConfig {
        EncodingConfig {
            temp_dir: std::env::temp_dir()
                .join("encoding-test")
                .to_string_lossy()
                .into_owned(),
            ffmpeg_path: "ffmpeg".into(),
            ffprobe_path: "ffprobe".into(),
            max_concurrent_jobs: 1,
        }
    }

    fn test_minio_config() -> MinIOConfig {
        MinIOConfig {
            endpoint: "localhost:9000".into(),
            access_key: "test".into(),
            secret_key: "test".into(),
            use_ssl: false,
            raw_bucket: "streamvault-raw".into(),
            encoded_bucket: "streamvault-encoded".into(),
            thumbnails_bucket: "streamvault-thumbnails".into(),
            region: "us-east-1".into(),
        }
    }

    fn test_encoding_job() -> EncodingJob {
        EncodingJob {
            job_id: "test-job-001".into(),
            content_id: "movie-123".into(),
            source_path: "movie-123/original.mp4".into(),
            source_bucket: "streamvault-raw".into(),
            requested_by: "admin-user".into(),
            created_at: chrono::Utc::now(),
        }
    }

    async fn test_store() -> store::JobStore {
        store::JobStore::new("redis://localhost:6379/15")
            .await
            .expect("redis required for integration tests (DB 15)")
    }

    #[tokio::test]
    #[ignore]
    async fn process_inserts_job_into_store() {
        let store = test_store().await;
        let publisher: Arc<dyn ResultPublisher> = Arc::new(MockPublisher::new());
        let storage: Arc<dyn StorageClient> = Arc::new(FailingStorageClient);

        let orchestrator = PipelineOrchestrator::new(
            &test_encoding_config(),
            &test_minio_config(),
            storage,
            publisher,
            store.clone(),
        );

        let job = test_encoding_job();
        let _ = orchestrator.process(job.clone()).await;

        // Job should exist in store regardless of pipeline outcome
        let stored_job = store.get_job("test-job-001").await;
        assert!(stored_job.is_some(), "job should be inserted into store");
        assert_eq!(stored_job.unwrap().content_id, "movie-123");
    }

    #[tokio::test]
    #[ignore]
    async fn process_with_failing_download_publishes_failure() {
        let store = test_store().await;
        let mock_pub = Arc::new(MockPublisher::new());
        let publisher: Arc<dyn ResultPublisher> = Arc::clone(&mock_pub) as Arc<dyn ResultPublisher>;
        let storage: Arc<dyn StorageClient> = Arc::new(FailingStorageClient);

        let orchestrator = PipelineOrchestrator::new(
            &test_encoding_config(),
            &test_minio_config(),
            storage,
            publisher,
            store.clone(),
        );

        let job = test_encoding_job();
        let result = orchestrator.process(job).await;

        // Process should return an error
        assert!(result.is_err(), "process should fail when download fails");

        // Publisher should have received a failure event
        let failed = mock_pub.failed_calls.lock().unwrap();
        assert_eq!(failed.len(), 1, "should publish exactly one failure");
        assert_eq!(failed[0].status, "failed");
        assert_eq!(failed[0].job_id, "test-job-001");
        assert_eq!(failed[0].content_id, "movie-123");
        assert!(
            failed[0].error_message.is_some(),
            "failure should include error message"
        );
        assert_eq!(failed[0].event_type, "encoding.job.failed");

        // No completed events should be published
        let completed = mock_pub.completed_calls.lock().unwrap();
        assert_eq!(completed.len(), 0, "should not publish completed event");
    }

    #[tokio::test]
    #[ignore]
    async fn process_with_failing_download_marks_job_as_failed_in_store() {
        let store = test_store().await;
        let publisher: Arc<dyn ResultPublisher> = Arc::new(MockPublisher::new());
        let storage: Arc<dyn StorageClient> = Arc::new(FailingStorageClient);

        let orchestrator = PipelineOrchestrator::new(
            &test_encoding_config(),
            &test_minio_config(),
            storage,
            publisher,
            store.clone(),
        );

        let job = test_encoding_job();
        let _ = orchestrator.process(job).await;

        // Job in store should be marked as Failed
        let stored_job = store.get_job("test-job-001").await.unwrap();
        assert_eq!(
            stored_job.status,
            crate::domain::status::JobStatus::Failed,
            "job should be marked as Failed"
        );
        assert!(
            stored_job.error_message.is_some(),
            "failed job should have error message"
        );
    }
}
