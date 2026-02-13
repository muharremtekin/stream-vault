use axum::extract::{Path, Query, State};
use axum::http::StatusCode;
use axum::Json;
use serde::{Deserialize, Serialize};

use crate::domain::job::Job;
use crate::domain::status::JobStatus;
use crate::AppState;

#[derive(Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct JobResponse {
    pub job_id: String,
    pub content_id: String,
    pub status: String,
    pub progress_percentage: f64,
    pub current_step: String,
    pub outputs: Vec<OutputResponse>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error_message: Option<String>,
    pub created_at: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub started_at: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub completed_at: Option<String>,
}

#[derive(Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct OutputResponse {
    pub quality: String,
    pub width: i32,
    pub height: i32,
    pub bitrate_kbps: i32,
    pub file_size_bytes: i64,
    pub segment_count: i32,
    pub playlist_path: String,
    pub storage_path: String,
}

#[derive(Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ListJobsResponse {
    pub jobs: Vec<JobResponse>,
    pub total_count: usize,
}

#[derive(Serialize)]
pub struct ErrorResponse {
    pub error: String,
}

#[derive(Deserialize)]
pub struct ListJobsQuery {
    pub status: Option<String>,
    pub content_id: Option<String>,
    pub limit: Option<usize>,
    pub offset: Option<usize>,
}

pub async fn get_job(
    State(state): State<AppState>,
    Path(job_id): Path<String>,
) -> Result<Json<JobResponse>, (StatusCode, Json<ErrorResponse>)> {
    let store = state.store.read().await;
    match store.get(&job_id) {
        Some(job) => Ok(Json(job_to_response(job))),
        None => Err((
            StatusCode::NOT_FOUND,
            Json(ErrorResponse {
                error: format!("job {} not found", job_id),
            }),
        )),
    }
}

pub async fn list_jobs(
    State(state): State<AppState>,
    Query(params): Query<ListJobsQuery>,
) -> Json<ListJobsResponse> {
    let store = state.store.read().await;

    let status_filter: Option<JobStatus> = params.status.as_deref().and_then(|s| s.parse().ok());
    let limit = params.limit.unwrap_or(50).min(100);
    let offset = params.offset.unwrap_or(0);

    let mut jobs: Vec<&Job> = store
        .values()
        .filter(|j| {
            status_filter.map_or(true, |s| j.status == s)
                && params
                    .content_id
                    .as_ref()
                    .map_or(true, |c| j.content_id == *c)
        })
        .collect();

    jobs.sort_by(|a, b| b.created_at.cmp(&a.created_at));

    let total_count = jobs.len();
    let jobs: Vec<JobResponse> = jobs
        .into_iter()
        .skip(offset)
        .take(limit)
        .map(job_to_response)
        .collect();

    Json(ListJobsResponse { jobs, total_count })
}

fn job_to_response(job: &Job) -> JobResponse {
    JobResponse {
        job_id: job.job_id.clone(),
        content_id: job.content_id.clone(),
        status: job.status.to_string(),
        progress_percentage: job.progress_percentage,
        current_step: job.current_step.clone(),
        outputs: job
            .outputs
            .iter()
            .map(|o| OutputResponse {
                quality: o.quality.clone(),
                width: o.width,
                height: o.height,
                bitrate_kbps: o.bitrate_kbps,
                file_size_bytes: o.file_size_bytes,
                segment_count: o.segment_count,
                playlist_path: o.playlist_path.clone(),
                storage_path: o.storage_path.clone(),
            })
            .collect(),
        error_message: job.error_message.clone(),
        created_at: job.created_at.to_rfc3339(),
        started_at: job.started_at.map(|t| t.to_rfc3339()),
        completed_at: job.completed_at.map(|t| t.to_rfc3339()),
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::job::JobOutput;
    use crate::store;
    use chrono::Utc;

    fn make_job(id: &str, content_id: &str, status: JobStatus) -> Job {
        Job {
            job_id: id.to_string(),
            content_id: content_id.to_string(),
            source_path: "test.mp4".to_string(),
            source_bucket: "raw".to_string(),
            requested_by: "user-1".to_string(),
            status,
            progress_percentage: 50.0,
            current_step: status.to_string(),
            target_qualities: Vec::new(),
            outputs: Vec::new(),
            error_message: None,
            created_at: Utc::now(),
            started_at: Some(Utc::now()),
            completed_at: None,
        }
    }

    #[test]
    fn test_job_to_response_basic() {
        let job = make_job("j1", "c1", JobStatus::Processing);
        let resp = job_to_response(&job);

        assert_eq!(resp.job_id, "j1");
        assert_eq!(resp.content_id, "c1");
        assert_eq!(resp.status, "processing");
        assert!((resp.progress_percentage - 50.0).abs() < f64::EPSILON);
        assert_eq!(resp.current_step, "processing");
        assert!(resp.outputs.is_empty());
        assert!(resp.error_message.is_none());
        assert!(resp.started_at.is_some());
        assert!(resp.completed_at.is_none());
    }

    #[test]
    fn test_job_to_response_with_outputs() {
        let mut job = make_job("j1", "c1", JobStatus::Completed);
        job.outputs = vec![JobOutput {
            quality: "720p".to_string(),
            width: 1280,
            height: 720,
            bitrate_kbps: 2800,
            file_size_bytes: 50_000,
            segment_count: 12,
            playlist_path: "c1/720p/playlist.m3u8".to_string(),
            storage_path: "c1/720p/".to_string(),
        }];

        let resp = job_to_response(&job);
        assert_eq!(resp.outputs.len(), 1);
        assert_eq!(resp.outputs[0].quality, "720p");
        assert_eq!(resp.outputs[0].width, 1280);
        assert_eq!(resp.outputs[0].bitrate_kbps, 2800);
    }

    #[test]
    fn test_job_to_response_with_error() {
        let mut job = make_job("j1", "c1", JobStatus::Failed);
        job.error_message = Some("codec not supported".to_string());

        let resp = job_to_response(&job);
        assert_eq!(resp.status, "failed");
        assert_eq!(resp.error_message, Some("codec not supported".to_string()));
    }

    #[test]
    fn test_job_response_serialization_camel_case() {
        let job = make_job("j1", "c1", JobStatus::Processing);
        let resp = job_to_response(&job);
        let json = serde_json::to_string(&resp).unwrap();

        assert!(json.contains("\"jobId\""));
        assert!(json.contains("\"contentId\""));
        assert!(json.contains("\"progressPercentage\""));
        assert!(json.contains("\"currentStep\""));
        assert!(json.contains("\"createdAt\""));
        assert!(json.contains("\"startedAt\""));
        // Optional None fields should be omitted
        assert!(!json.contains("\"errorMessage\""));
        assert!(!json.contains("\"completedAt\""));
    }

    #[test]
    fn test_list_jobs_response_serialization() {
        let resp = ListJobsResponse {
            jobs: vec![],
            total_count: 0,
        };
        let json = serde_json::to_string(&resp).unwrap();
        assert!(json.contains("\"totalCount\""));
        assert!(json.contains("\"jobs\""));
    }

    #[test]
    fn test_error_response_serialization() {
        let resp = ErrorResponse {
            error: "not found".to_string(),
        };
        let json = serde_json::to_string(&resp).unwrap();
        assert_eq!(json, r#"{"error":"not found"}"#);
    }

    // Integration tests using Axum test helpers
    use axum::body::Body;
    use axum::http::Request;
    use axum::Router;
    use http_body_util::BodyExt;
    use tower::ServiceExt;

    fn test_app(job_store: store::JobStore) -> Router {
        use crate::api::routes::encoding_routes;
        use std::sync::Arc;

        let config = crate::config::load().unwrap();
        let app_state = AppState {
            storage: Arc::new(crate::storage::minio::StubStorageClient),
            config,
            store: job_store,
        };

        encoding_routes().with_state(app_state)
    }

    #[tokio::test]
    async fn test_http_list_jobs_empty() {
        let job_store = store::new_job_store();
        let app = test_app(job_store);

        let resp = app
            .oneshot(
                Request::builder()
                    .uri("/api/encoding/jobs")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();

        assert_eq!(resp.status(), 200);
        let body = resp.into_body().collect().await.unwrap().to_bytes();
        let result: ListJobsResponse = serde_json::from_slice(&body).unwrap();
        assert_eq!(result.total_count, 0);
        assert!(result.jobs.is_empty());
    }

    #[tokio::test]
    async fn test_http_get_job_not_found() {
        let job_store = store::new_job_store();
        let app = test_app(job_store);

        let resp = app
            .oneshot(
                Request::builder()
                    .uri("/api/encoding/jobs/nonexistent")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();

        assert_eq!(resp.status(), 404);
    }

    #[tokio::test]
    async fn test_http_get_job_found() {
        let job_store = store::new_job_store();
        store::insert_job(&job_store, make_job("job-1", "content-1", JobStatus::Processing)).await;
        let app = test_app(job_store);

        let resp = app
            .oneshot(
                Request::builder()
                    .uri("/api/encoding/jobs/job-1")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();

        assert_eq!(resp.status(), 200);
        let body = resp.into_body().collect().await.unwrap().to_bytes();
        let result: JobResponse = serde_json::from_slice(&body).unwrap();
        assert_eq!(result.job_id, "job-1");
        assert_eq!(result.status, "processing");
    }

    #[tokio::test]
    async fn test_http_list_jobs_with_filter() {
        let job_store = store::new_job_store();
        store::insert_job(&job_store, make_job("j1", "c1", JobStatus::Processing)).await;
        store::insert_job(&job_store, make_job("j2", "c2", JobStatus::Completed)).await;
        let app = test_app(job_store);

        let resp = app
            .oneshot(
                Request::builder()
                    .uri("/api/encoding/jobs?status=processing")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();

        assert_eq!(resp.status(), 200);
        let body = resp.into_body().collect().await.unwrap().to_bytes();
        let result: ListJobsResponse = serde_json::from_slice(&body).unwrap();
        assert_eq!(result.total_count, 1);
        assert_eq!(result.jobs[0].status, "processing");
    }

    #[tokio::test]
    async fn test_http_list_jobs_pagination() {
        let job_store = store::new_job_store();
        for i in 0..5 {
            store::insert_job(
                &job_store,
                make_job(&format!("j{}", i), "c1", JobStatus::Processing),
            )
            .await;
        }
        let app = test_app(job_store);

        let resp = app
            .oneshot(
                Request::builder()
                    .uri("/api/encoding/jobs?limit=2&offset=1")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();

        assert_eq!(resp.status(), 200);
        let body = resp.into_body().collect().await.unwrap().to_bytes();
        let result: ListJobsResponse = serde_json::from_slice(&body).unwrap();
        assert_eq!(result.total_count, 5);
        assert_eq!(result.jobs.len(), 2);
    }
}
