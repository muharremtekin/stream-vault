use std::collections::HashMap;
use std::sync::Arc;

use chrono::Utc;
use tokio::sync::RwLock;

use crate::domain::job::{Job, JobOutput};
use crate::domain::status::JobStatus;

pub type JobStore = Arc<RwLock<HashMap<String, Job>>>;

pub fn new_job_store() -> JobStore {
    Arc::new(RwLock::new(HashMap::new()))
}

pub async fn insert_job(store: &JobStore, job: Job) {
    let mut w = store.write().await;
    w.insert(job.job_id.clone(), job);
}

pub async fn update_status(
    store: &JobStore,
    job_id: &str,
    status: JobStatus,
    step: &str,
    progress: f64,
) {
    let mut w = store.write().await;
    if let Some(job) = w.get_mut(job_id) {
        job.status = status;
        job.current_step = step.to_string();
        job.progress_percentage = progress;
    }
}

pub async fn complete_job(store: &JobStore, job_id: &str, outputs: Vec<JobOutput>) {
    let mut w = store.write().await;
    if let Some(job) = w.get_mut(job_id) {
        job.status = JobStatus::Completed;
        job.progress_percentage = 100.0;
        job.current_step = "completed".to_string();
        job.outputs = outputs;
        job.completed_at = Some(Utc::now());
    }
}

pub async fn fail_job(store: &JobStore, job_id: &str, error: &str) {
    let mut w = store.write().await;
    if let Some(job) = w.get_mut(job_id) {
        job.status = JobStatus::Failed;
        job.current_step = "failed".to_string();
        job.error_message = Some(error.to_string());
        job.completed_at = Some(Utc::now());
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::job::Job;

    fn make_test_job(id: &str, content_id: &str) -> Job {
        Job {
            job_id: id.to_string(),
            content_id: content_id.to_string(),
            source_path: "test/source.mp4".to_string(),
            source_bucket: "raw".to_string(),
            requested_by: "user-1".to_string(),
            status: JobStatus::Processing,
            progress_percentage: 0.0,
            current_step: "starting".to_string(),
            target_qualities: Vec::new(),
            outputs: Vec::new(),
            error_message: None,
            created_at: Utc::now(),
            started_at: Some(Utc::now()),
            completed_at: None,
        }
    }

    #[tokio::test]
    async fn test_new_job_store_is_empty() {
        let store = new_job_store();
        let r = store.read().await;
        assert!(r.is_empty());
    }

    #[tokio::test]
    async fn test_insert_and_read_job() {
        let store = new_job_store();
        let job = make_test_job("job-1", "content-1");

        insert_job(&store, job).await;

        let r = store.read().await;
        assert_eq!(r.len(), 1);
        let j = r.get("job-1").unwrap();
        assert_eq!(j.content_id, "content-1");
        assert_eq!(j.status, JobStatus::Processing);
    }

    #[tokio::test]
    async fn test_update_status() {
        let store = new_job_store();
        insert_job(&store, make_test_job("job-1", "c-1")).await;

        update_status(&store, "job-1", JobStatus::Processing, "transcoding", 50.0).await;

        let r = store.read().await;
        let j = r.get("job-1").unwrap();
        assert_eq!(j.current_step, "transcoding");
        assert!((j.progress_percentage - 50.0).abs() < f64::EPSILON);
    }

    #[tokio::test]
    async fn test_update_status_nonexistent_job_is_noop() {
        let store = new_job_store();
        // Should not panic
        update_status(&store, "nonexistent", JobStatus::Processing, "step", 10.0).await;
        let r = store.read().await;
        assert!(r.is_empty());
    }

    #[tokio::test]
    async fn test_complete_job() {
        let store = new_job_store();
        insert_job(&store, make_test_job("job-1", "c-1")).await;

        let outputs = vec![JobOutput {
            quality: "720p".to_string(),
            width: 1280,
            height: 720,
            bitrate_kbps: 2800,
            file_size_bytes: 50_000_000,
            segment_count: 12,
            playlist_path: "c-1/720p/playlist.m3u8".to_string(),
            storage_path: "c-1/720p/".to_string(),
        }];

        complete_job(&store, "job-1", outputs).await;

        let r = store.read().await;
        let j = r.get("job-1").unwrap();
        assert_eq!(j.status, JobStatus::Completed);
        assert!((j.progress_percentage - 100.0).abs() < f64::EPSILON);
        assert_eq!(j.current_step, "completed");
        assert_eq!(j.outputs.len(), 1);
        assert_eq!(j.outputs[0].quality, "720p");
        assert!(j.completed_at.is_some());
    }

    #[tokio::test]
    async fn test_fail_job() {
        let store = new_job_store();
        insert_job(&store, make_test_job("job-1", "c-1")).await;

        fail_job(&store, "job-1", "unsupported codec").await;

        let r = store.read().await;
        let j = r.get("job-1").unwrap();
        assert_eq!(j.status, JobStatus::Failed);
        assert_eq!(j.current_step, "failed");
        assert_eq!(j.error_message.as_deref(), Some("unsupported codec"));
        assert!(j.completed_at.is_some());
    }

    #[tokio::test]
    async fn test_multiple_jobs() {
        let store = new_job_store();
        insert_job(&store, make_test_job("job-1", "c-1")).await;
        insert_job(&store, make_test_job("job-2", "c-2")).await;
        insert_job(&store, make_test_job("job-3", "c-1")).await;

        let r = store.read().await;
        assert_eq!(r.len(), 3);

        // Complete one, fail another
        drop(r);
        complete_job(&store, "job-1", vec![]).await;
        fail_job(&store, "job-2", "error").await;

        let r = store.read().await;
        assert_eq!(r.get("job-1").unwrap().status, JobStatus::Completed);
        assert_eq!(r.get("job-2").unwrap().status, JobStatus::Failed);
        assert_eq!(r.get("job-3").unwrap().status, JobStatus::Processing);
    }
}
