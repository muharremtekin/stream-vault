use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

use super::profile::Quality;
use super::status::JobStatus;
use crate::messaging::models::EncodingJob;

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

impl Job {
    pub fn from_encoding_job(msg: &EncodingJob) -> Self {
        Self {
            job_id: msg.job_id.clone(),
            content_id: msg.content_id.clone(),
            source_path: msg.source_path.clone(),
            source_bucket: msg.source_bucket.clone(),
            requested_by: msg.requested_by.clone(),
            status: JobStatus::Processing,
            progress_percentage: 0.0,
            current_step: "starting".to_string(),
            target_qualities: Vec::new(),
            outputs: Vec::new(),
            error_message: None,
            created_at: msg.created_at,
            started_at: Some(Utc::now()),
            completed_at: None,
        }
    }
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

impl From<&crate::messaging::models::EncodingOutput> for JobOutput {
    fn from(o: &crate::messaging::models::EncodingOutput) -> Self {
        Self {
            quality: o.quality.clone(),
            width: o.width,
            height: o.height,
            bitrate_kbps: o.bitrate_kbps,
            file_size_bytes: o.file_size_bytes,
            segment_count: o.segment_count,
            playlist_path: o.playlist_path.clone(),
            storage_path: String::new(),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::messaging::models;

    #[test]
    fn test_from_encoding_job() {
        let msg = EncodingJob {
            job_id: "job-123".to_string(),
            content_id: "movie-456".to_string(),
            source_path: "job-123/original.mp4".to_string(),
            source_bucket: "streamvault-raw".to_string(),
            requested_by: "user-789".to_string(),
            created_at: Utc::now(),
        };

        let job = Job::from_encoding_job(&msg);

        assert_eq!(job.job_id, "job-123");
        assert_eq!(job.content_id, "movie-456");
        assert_eq!(job.source_path, "job-123/original.mp4");
        assert_eq!(job.source_bucket, "streamvault-raw");
        assert_eq!(job.requested_by, "user-789");
        assert_eq!(job.status, JobStatus::Processing);
        assert!((job.progress_percentage - 0.0).abs() < f64::EPSILON);
        assert_eq!(job.current_step, "starting");
        assert!(job.target_qualities.is_empty());
        assert!(job.outputs.is_empty());
        assert!(job.error_message.is_none());
        assert!(job.started_at.is_some());
        assert!(job.completed_at.is_none());
    }

    #[test]
    fn test_job_output_from_encoding_output() {
        let msg_output = models::EncodingOutput {
            quality: "720p".to_string(),
            width: 1280,
            height: 720,
            bitrate_kbps: 2800,
            file_size_bytes: 50_000_000,
            segment_count: 15,
            playlist_path: "movie-456/720p/playlist.m3u8".to_string(),
        };

        let output = JobOutput::from(&msg_output);

        assert_eq!(output.quality, "720p");
        assert_eq!(output.width, 1280);
        assert_eq!(output.height, 720);
        assert_eq!(output.bitrate_kbps, 2800);
        assert_eq!(output.file_size_bytes, 50_000_000);
        assert_eq!(output.segment_count, 15);
        assert_eq!(output.playlist_path, "movie-456/720p/playlist.m3u8");
        assert!(output.storage_path.is_empty());
    }

    #[test]
    fn test_job_serialization() {
        let job = Job {
            job_id: "j1".to_string(),
            content_id: "c1".to_string(),
            source_path: "s.mp4".to_string(),
            source_bucket: "raw".to_string(),
            requested_by: "u1".to_string(),
            status: JobStatus::Completed,
            progress_percentage: 100.0,
            current_step: "completed".to_string(),
            target_qualities: Vec::new(),
            outputs: vec![JobOutput {
                quality: "360p".to_string(),
                width: 640,
                height: 360,
                bitrate_kbps: 800,
                file_size_bytes: 1000,
                segment_count: 5,
                playlist_path: "c1/360p/playlist.m3u8".to_string(),
                storage_path: "c1/360p/".to_string(),
            }],
            error_message: None,
            created_at: Utc::now(),
            started_at: Some(Utc::now()),
            completed_at: Some(Utc::now()),
        };

        let json = serde_json::to_string(&job).unwrap();
        let back: Job = serde_json::from_str(&json).unwrap();
        assert_eq!(back.job_id, "j1");
        assert_eq!(back.status, JobStatus::Completed);
        assert_eq!(back.outputs.len(), 1);
        assert_eq!(back.outputs[0].quality, "360p");
    }
}
