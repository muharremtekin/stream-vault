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
/// Includes event envelope fields per Rule 3.4.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EncodingResult {
    // Event envelope fields (Rule 3.4)
    pub event_id: String,
    pub event_type: String,
    pub timestamp: String,
    pub source: String,
    pub correlation_id: String,

    // Data fields
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
            event_id: "550e8400-e29b-41d4-a716-446655440000".into(),
            event_type: "encoding.job.completed".into(),
            timestamp: "2024-01-15T10:35:00Z".into(),
            source: "encoding-service".into(),
            correlation_id: "abc-123".into(),
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
        assert!(json.contains("event_id"));
        assert!(json.contains("encoding.job.completed"));
        assert!(json.contains("encoding-service"));
    }

    #[test]
    fn test_encoding_result_serialize_failed() {
        let result = EncodingResult {
            event_id: "660e8400-e29b-41d4-a716-446655440001".into(),
            event_type: "encoding.job.failed".into(),
            timestamp: "2024-01-15T10:35:00Z".into(),
            source: "encoding-service".into(),
            correlation_id: "abc-123".into(),
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
        assert!(json.contains("encoding.job.failed"));
    }

    #[test]
    fn test_encoding_result_deserialize_with_envelope() {
        let json = r#"{
            "event_id": "550e8400-e29b-41d4-a716-446655440000",
            "event_type": "encoding.job.completed",
            "timestamp": "2024-01-15T10:35:00Z",
            "source": "encoding-service",
            "correlation_id": "abc-123",
            "job_id": "abc-123",
            "content_id": "movie-456",
            "status": "completed",
            "outputs": [],
            "duration_seconds": 120,
            "completed_at": "2024-01-15T10:35:00Z"
        }"#;

        let result: EncodingResult = serde_json::from_str(json).unwrap();
        assert_eq!(result.event_id, "550e8400-e29b-41d4-a716-446655440000");
        assert_eq!(result.event_type, "encoding.job.completed");
        assert_eq!(result.source, "encoding-service");
        assert_eq!(result.correlation_id, "abc-123");
        assert_eq!(result.job_id, "abc-123");
    }
}
