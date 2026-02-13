use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

/// Incoming encoding job message from the streaming service.
/// Consumed from the `encoding.jobs` queue.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EncodingJob {
    pub job_id: String,
    pub content_id: String,
    pub source_path: String,
    pub source_bucket: String,
    pub requested_by: String,
    pub created_at: DateTime<Utc>,
}

/// Outgoing encoding result published to the `encoding` exchange.
/// Consumed by streaming-service and catalog-service from their respective queues.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EncodingResult {
    pub job_id: String,
    pub content_id: String,
    /// "completed" or "failed"
    pub status: String,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub outputs: Vec<EncodingOutput>,
    pub duration_seconds: i64,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error_message: Option<String>,
    pub completed_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EncodingOutput {
    pub quality: String,
    pub width: i32,
    pub height: i32,
    pub bitrate_kbps: i32,
    pub segment_count: i32,
    pub playlist_path: String,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_encoding_job_deserialize() {
        let json = r#"{
            "job_id": "abc-123",
            "content_id": "movie-456",
            "source_path": "abc-123/original.mp4",
            "source_bucket": "streamvault-raw",
            "requested_by": "user-789",
            "created_at": "2024-01-15T10:30:00Z"
        }"#;

        let job: EncodingJob = serde_json::from_str(json).unwrap();
        assert_eq!(job.job_id, "abc-123");
        assert_eq!(job.content_id, "movie-456");
        assert_eq!(job.source_bucket, "streamvault-raw");
    }

    #[test]
    fn test_encoding_result_serialize_completed() {
        let result = EncodingResult {
            job_id: "abc-123".into(),
            content_id: "movie-456".into(),
            status: "completed".into(),
            outputs: vec![EncodingOutput {
                quality: "720p".into(),
                width: 1280,
                height: 720,
                bitrate_kbps: 2800,
                segment_count: 12,
                playlist_path: "movie-456/720p/playlist.m3u8".into(),
            }],
            duration_seconds: 120,
            error_message: None,
            completed_at: "2024-01-15T10:35:00Z".into(),
        };

        let json = serde_json::to_string(&result).unwrap();
        assert!(!json.contains("error_message"));
        assert!(json.contains("720p"));
    }

    #[test]
    fn test_encoding_result_serialize_failed() {
        let result = EncodingResult {
            job_id: "abc-123".into(),
            content_id: "movie-456".into(),
            status: "failed".into(),
            outputs: vec![],
            duration_seconds: 0,
            error_message: Some("unsupported codec".into()),
            completed_at: "2024-01-15T10:35:00Z".into(),
        };

        let json = serde_json::to_string(&result).unwrap();
        assert!(json.contains("error_message"));
        assert!(!json.contains("outputs"));
    }
}
