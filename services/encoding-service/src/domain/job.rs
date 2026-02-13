use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

use super::profile::Quality;
use super::status::JobStatus;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Job {
    pub job_id: String,
    pub content_id: String,
    pub source_path: String,
    pub source_bucket: String,
    pub requested_by: String,
    pub status: JobStatus,
    pub progress_percentage: f64,
    pub current_step: String,
    pub target_qualities: Vec<Quality>,
    pub outputs: Vec<JobOutput>,
    pub error_message: Option<String>,
    pub created_at: DateTime<Utc>,
    pub started_at: Option<DateTime<Utc>>,
    pub completed_at: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct JobOutput {
    pub quality: String,
    pub width: i32,
    pub height: i32,
    pub bitrate_kbps: i32,
    pub file_size_bytes: i64,
    pub segment_count: i32,
    pub playlist_path: String,
    pub storage_path: String,
}
