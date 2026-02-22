use chrono::Utc;
use redis::AsyncCommands;
use tracing::warn;

use crate::domain::job::{Job, JobOutput};
use crate::domain::status::JobStatus;

/// Key prefixes for Redis storage.
const JOB_KEY_PREFIX: &str = "enc:job:";
const JOB_INDEX_KEY: &str = "enc:jobs";
const CONTENT_INDEX_PREFIX: &str = "enc:content:";

/// TTL for completed/failed jobs: 7 days.
const TERMINAL_JOB_TTL_SECS: i64 = 7 * 24 * 60 * 60;

/// Redis-backed job store. Wraps a `ConnectionManager` for automatic
/// reconnection and is cheaply cloneable (`Arc` under the hood).
#[derive(Clone)]
pub struct JobStore {
    conn: redis::aio::ConnectionManager,
}

impl JobStore {
    pub async fn new(redis_url: &str) -> anyhow::Result<Self> {
        let client = redis::Client::open(redis_url)?;
        let conn = redis::aio::ConnectionManager::new(client).await?;
        Ok(Self { conn })
    }

    /// Get a single job by ID.
    pub async fn get_job(&self, job_id: &str) -> Option<Job> {
        let key = format!("{}{}", JOB_KEY_PREFIX, job_id);
        let data: Option<String> = self.conn.clone().get(&key).await.ok()?;
        data.and_then(|json| serde_json::from_str(&json).ok())
    }

    /// List jobs with optional status and content_id filters, sorted by created_at DESC.
    pub async fn list_jobs(
        &self,
        status_filter: Option<JobStatus>,
        content_id_filter: Option<&str>,
        limit: usize,
        offset: usize,
    ) -> (Vec<Job>, usize) {
        let job_ids: Vec<String> = if let Some(cid) = content_id_filter {
            // Use content index set for filtered queries
            let content_key = format!("{}{}", CONTENT_INDEX_PREFIX, cid);
            let ids: Vec<String> = self
                .conn
                .clone()
                .smembers(&content_key)
                .await
                .unwrap_or_default();
            ids
        } else {
            // Fetch all job IDs from sorted set (newest first)
            self.conn
                .clone()
                .zrevrange(JOB_INDEX_KEY, 0, -1)
                .await
                .unwrap_or_default()
        };

        // Fetch all jobs in bulk via MGET
        if job_ids.is_empty() {
            return (Vec::new(), 0);
        }

        let keys: Vec<String> = job_ids
            .iter()
            .map(|id| format!("{}{}", JOB_KEY_PREFIX, id))
            .collect();

        let values: Vec<Option<String>> =
            redis::cmd("MGET")
                .arg(&keys)
                .query_async(&mut self.conn.clone())
                .await
                .unwrap_or_else(|_| vec![None; keys.len()]);

        let mut jobs: Vec<Job> = values
            .into_iter()
            .flatten()
            .filter_map(|json| serde_json::from_str(&json).ok())
            .filter(|j: &Job| status_filter.map_or(true, |s| j.status == s))
            .collect();

        // Sort by created_at descending (content index doesn't preserve order)
        jobs.sort_by(|a, b| b.created_at.cmp(&a.created_at));

        let total_count = jobs.len();
        let page: Vec<Job> = jobs.into_iter().skip(offset).take(limit).collect();

        (page, total_count)
    }
}

pub async fn insert_job(store: &JobStore, job: Job) {
    let key = format!("{}{}", JOB_KEY_PREFIX, job.job_id);
    let score = job.created_at.timestamp_millis() as f64;
    let content_key = format!("{}{}", CONTENT_INDEX_PREFIX, job.content_id);

    let json = match serde_json::to_string(&job) {
        Ok(j) => j,
        Err(e) => {
            warn!(error = %e, job_id = %job.job_id, "failed to serialize job");
            return;
        }
    };

    let job_id = job.job_id.clone();
    let mut conn = store.conn.clone();

    // Pipeline: SET job data + ZADD to index + SADD to content index
    let result: Result<(), redis::RedisError> = redis::pipe()
        .atomic()
        .set(&key, &json)
        .zadd(JOB_INDEX_KEY, &job_id, score)
        .sadd(&content_key, &job_id)
        .query_async(&mut conn)
        .await;

    if let Err(e) = result {
        warn!(error = %e, job_id = %job_id, "failed to insert job into redis");
    }
}

pub async fn update_status(
    store: &JobStore,
    job_id: &str,
    status: JobStatus,
    step: &str,
    progress: f64,
) {
    let key = format!("{}{}", JOB_KEY_PREFIX, job_id);
    let mut conn = store.conn.clone();

    let data: Option<String> = match conn.get(&key).await {
        Ok(d) => d,
        Err(e) => {
            warn!(error = %e, job_id, "failed to read job for status update");
            return;
        }
    };

    if let Some(json) = data {
        if let Ok(mut job) = serde_json::from_str::<Job>(&json) {
            job.status = status;
            job.current_step = step.to_string();
            job.progress_percentage = progress;

            if let Ok(updated_json) = serde_json::to_string(&job) {
                let _: Result<(), _> = conn.set(&key, &updated_json).await;
            }
        }
    }
}

pub async fn complete_job(store: &JobStore, job_id: &str, outputs: Vec<JobOutput>) {
    let key = format!("{}{}", JOB_KEY_PREFIX, job_id);
    let mut conn = store.conn.clone();

    let data: Option<String> = match conn.get(&key).await {
        Ok(d) => d,
        Err(e) => {
            warn!(error = %e, job_id, "failed to read job for completion");
            return;
        }
    };

    if let Some(json) = data {
        if let Ok(mut job) = serde_json::from_str::<Job>(&json) {
            job.status = JobStatus::Completed;
            job.progress_percentage = 100.0;
            job.current_step = "completed".to_string();
            job.outputs = outputs;
            job.completed_at = Some(Utc::now());

            if let Ok(updated_json) = serde_json::to_string(&job) {
                let _: Result<(), _> = conn.set_ex(&key, &updated_json, TERMINAL_JOB_TTL_SECS as u64).await;
            }
        }
    }
}

pub async fn fail_job(store: &JobStore, job_id: &str, error: &str) {
    let key = format!("{}{}", JOB_KEY_PREFIX, job_id);
    let mut conn = store.conn.clone();

    let data: Option<String> = match conn.get(&key).await {
        Ok(d) => d,
        Err(e) => {
            warn!(error = %e, job_id, "failed to read job for failure update");
            return;
        }
    };

    if let Some(json) = data {
        if let Ok(mut job) = serde_json::from_str::<Job>(&json) {
            job.status = JobStatus::Failed;
            job.current_step = "failed".to_string();
            job.error_message = Some(error.to_string());
            job.completed_at = Some(Utc::now());

            if let Ok(updated_json) = serde_json::to_string(&job) {
                let _: Result<(), _> = conn.set_ex(&key, &updated_json, TERMINAL_JOB_TTL_SECS as u64).await;
            }
        }
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

    #[test]
    fn test_job_serialization_roundtrip() {
        let job = make_test_job("job-1", "content-1");
        let json = serde_json::to_string(&job).unwrap();
        let deserialized: Job = serde_json::from_str(&json).unwrap();
        assert_eq!(deserialized.job_id, "job-1");
        assert_eq!(deserialized.content_id, "content-1");
        assert_eq!(deserialized.status, JobStatus::Processing);
    }

    #[test]
    fn test_key_format() {
        assert_eq!(
            format!("{}{}", JOB_KEY_PREFIX, "abc-123"),
            "enc:job:abc-123"
        );
        assert_eq!(
            format!("{}{}", CONTENT_INDEX_PREFIX, "movie-456"),
            "enc:content:movie-456"
        );
    }
}
