use serde::{Deserialize, Serialize};
use std::fmt;
use std::str::FromStr;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum JobStatus {
    Queued,
    Processing,
    Completed,
    Failed,
    Cancelled,
}

impl fmt::Display for JobStatus {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            JobStatus::Queued => write!(f, "queued"),
            JobStatus::Processing => write!(f, "processing"),
            JobStatus::Completed => write!(f, "completed"),
            JobStatus::Failed => write!(f, "failed"),
            JobStatus::Cancelled => write!(f, "cancelled"),
        }
    }
}

impl FromStr for JobStatus {
    type Err = String;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.to_lowercase().as_str() {
            "queued" => Ok(JobStatus::Queued),
            "processing" => Ok(JobStatus::Processing),
            "completed" => Ok(JobStatus::Completed),
            "failed" => Ok(JobStatus::Failed),
            "cancelled" => Ok(JobStatus::Cancelled),
            _ => Err(format!("unknown job status: {}", s)),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_display() {
        assert_eq!(JobStatus::Queued.to_string(), "queued");
        assert_eq!(JobStatus::Processing.to_string(), "processing");
        assert_eq!(JobStatus::Completed.to_string(), "completed");
        assert_eq!(JobStatus::Failed.to_string(), "failed");
        assert_eq!(JobStatus::Cancelled.to_string(), "cancelled");
    }

    #[test]
    fn test_from_str_lowercase() {
        assert_eq!("queued".parse::<JobStatus>().unwrap(), JobStatus::Queued);
        assert_eq!("processing".parse::<JobStatus>().unwrap(), JobStatus::Processing);
        assert_eq!("completed".parse::<JobStatus>().unwrap(), JobStatus::Completed);
        assert_eq!("failed".parse::<JobStatus>().unwrap(), JobStatus::Failed);
        assert_eq!("cancelled".parse::<JobStatus>().unwrap(), JobStatus::Cancelled);
    }

    #[test]
    fn test_from_str_case_insensitive() {
        assert_eq!("PROCESSING".parse::<JobStatus>().unwrap(), JobStatus::Processing);
        assert_eq!("Completed".parse::<JobStatus>().unwrap(), JobStatus::Completed);
        assert_eq!("FAILED".parse::<JobStatus>().unwrap(), JobStatus::Failed);
    }

    #[test]
    fn test_from_str_invalid() {
        assert!("invalid".parse::<JobStatus>().is_err());
        assert!("".parse::<JobStatus>().is_err());
    }

    #[test]
    fn test_serde_roundtrip() {
        let status = JobStatus::Processing;
        let json = serde_json::to_string(&status).unwrap();
        assert_eq!(json, "\"processing\"");

        let deserialized: JobStatus = serde_json::from_str(&json).unwrap();
        assert_eq!(deserialized, status);
    }

    #[test]
    fn test_serde_all_variants() {
        let variants = vec![
            (JobStatus::Queued, "\"queued\""),
            (JobStatus::Processing, "\"processing\""),
            (JobStatus::Completed, "\"completed\""),
            (JobStatus::Failed, "\"failed\""),
            (JobStatus::Cancelled, "\"cancelled\""),
        ];
        for (status, expected_json) in variants {
            let json = serde_json::to_string(&status).unwrap();
            assert_eq!(json, expected_json);
            let back: JobStatus = serde_json::from_str(&json).unwrap();
            assert_eq!(back, status);
        }
    }
}
