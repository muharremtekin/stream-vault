use tracing::info;

use crate::domain::status::JobStatus as DomainJobStatus;
use crate::encoding_proto;
use crate::store::JobStore;

pub struct EncodingGrpcService {
    store: JobStore,
}

impl EncodingGrpcService {
    pub fn new(store: JobStore) -> Self {
        Self { store }
    }
}

#[tonic::async_trait]
impl encoding_proto::encoding_service_server::EncodingService for EncodingGrpcService {
    async fn get_job_status(
        &self,
        request: tonic::Request<encoding_proto::GetJobStatusRequest>,
    ) -> std::result::Result<tonic::Response<encoding_proto::JobStatusResponse>, tonic::Status>
    {
        let job_id = &request.get_ref().job_id;
        info!(job_id, "get_job_status called");

        if job_id.is_empty() {
            return Err(tonic::Status::invalid_argument("job_id is required"));
        }

        let job = self
            .store
            .get_job(job_id)
            .await
            .ok_or_else(|| tonic::Status::not_found(format!("job {} not found", job_id)))?;

        Ok(tonic::Response::new(job_to_proto(&job)))
    }

    async fn list_jobs(
        &self,
        request: tonic::Request<encoding_proto::ListJobsRequest>,
    ) -> std::result::Result<tonic::Response<encoding_proto::ListJobsResponse>, tonic::Status>
    {
        let req = request.get_ref();
        info!("list_jobs called");

        let status_filter = if req.status_filter == encoding_proto::JobStatus::Unspecified as i32 {
            None
        } else {
            Some(proto_status_to_domain(req.status_filter))
        };

        let limit = if req.limit <= 0 { 50 } else { req.limit.min(100) } as usize;
        let offset = req.offset.max(0) as usize;

        let content_id_filter = if req.content_id_filter.is_empty() {
            None
        } else {
            Some(req.content_id_filter.as_str())
        };

        let (jobs, total_count) = self
            .store
            .list_jobs(status_filter, content_id_filter, limit, offset)
            .await;

        let response_jobs: Vec<_> = jobs.iter().map(job_to_proto).collect();

        Ok(tonic::Response::new(encoding_proto::ListJobsResponse {
            jobs: response_jobs,
            total_count: total_count as i64,
        }))
    }
}

fn domain_status_to_proto(status: DomainJobStatus) -> i32 {
    match status {
        DomainJobStatus::Queued => encoding_proto::JobStatus::Queued as i32,
        DomainJobStatus::Processing => encoding_proto::JobStatus::Processing as i32,
        DomainJobStatus::Completed => encoding_proto::JobStatus::Completed as i32,
        DomainJobStatus::Failed => encoding_proto::JobStatus::Failed as i32,
        DomainJobStatus::Cancelled => encoding_proto::JobStatus::Cancelled as i32,
    }
}

fn proto_status_to_domain(status: i32) -> DomainJobStatus {
    match status {
        1 => DomainJobStatus::Queued,
        2 => DomainJobStatus::Processing,
        3 => DomainJobStatus::Completed,
        4 => DomainJobStatus::Failed,
        5 => DomainJobStatus::Cancelled,
        _ => DomainJobStatus::Queued,
    }
}

fn datetime_to_timestamp(dt: chrono::DateTime<chrono::Utc>) -> prost_types::Timestamp {
    prost_types::Timestamp {
        seconds: dt.timestamp(),
        nanos: dt.timestamp_subsec_nanos() as i32,
    }
}

fn job_to_proto(job: &crate::domain::job::Job) -> encoding_proto::JobStatusResponse {
    encoding_proto::JobStatusResponse {
        job_id: job.job_id.clone(),
        content_id: job.content_id.clone(),
        status: domain_status_to_proto(job.status),
        progress_percentage: job.progress_percentage,
        current_step: job.current_step.clone(),
        outputs: job
            .outputs
            .iter()
            .map(|o| encoding_proto::EncodingOutput {
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
        error_message: job.error_message.clone().unwrap_or_default(),
        created_at: Some(datetime_to_timestamp(job.created_at)),
        started_at: job.started_at.map(datetime_to_timestamp),
        completed_at: job.completed_at.map(datetime_to_timestamp),
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::job::{Job, JobOutput};
    use crate::encoding_proto::encoding_service_server::EncodingService;
    use crate::store;
    use chrono::Utc;

    fn make_job(id: &str, content_id: &str, status: DomainJobStatus) -> Job {
        Job {
            job_id: id.to_string(),
            content_id: content_id.to_string(),
            source_path: "test.mp4".to_string(),
            source_bucket: "raw".to_string(),
            requested_by: "user-1".to_string(),
            status,
            progress_percentage: if status == DomainJobStatus::Completed {
                100.0
            } else {
                50.0
            },
            current_step: status.to_string(),
            target_qualities: Vec::new(),
            outputs: Vec::new(),
            error_message: if status == DomainJobStatus::Failed {
                Some("test error".to_string())
            } else {
                None
            },
            created_at: Utc::now(),
            started_at: Some(Utc::now()),
            completed_at: if status == DomainJobStatus::Completed || status == DomainJobStatus::Failed {
                Some(Utc::now())
            } else {
                None
            },
        }
    }

    #[test]
    fn test_domain_status_to_proto_all_variants() {
        assert_eq!(domain_status_to_proto(DomainJobStatus::Queued), 1);
        assert_eq!(domain_status_to_proto(DomainJobStatus::Processing), 2);
        assert_eq!(domain_status_to_proto(DomainJobStatus::Completed), 3);
        assert_eq!(domain_status_to_proto(DomainJobStatus::Failed), 4);
        assert_eq!(domain_status_to_proto(DomainJobStatus::Cancelled), 5);
    }

    #[test]
    fn test_proto_status_to_domain_all_values() {
        assert_eq!(proto_status_to_domain(1), DomainJobStatus::Queued);
        assert_eq!(proto_status_to_domain(2), DomainJobStatus::Processing);
        assert_eq!(proto_status_to_domain(3), DomainJobStatus::Completed);
        assert_eq!(proto_status_to_domain(4), DomainJobStatus::Failed);
        assert_eq!(proto_status_to_domain(5), DomainJobStatus::Cancelled);
        // Unknown falls back to Queued
        assert_eq!(proto_status_to_domain(0), DomainJobStatus::Queued);
        assert_eq!(proto_status_to_domain(99), DomainJobStatus::Queued);
    }

    #[test]
    fn test_datetime_to_timestamp() {
        let dt = chrono::DateTime::parse_from_rfc3339("2024-06-15T10:30:00Z")
            .unwrap()
            .with_timezone(&Utc);
        let ts = datetime_to_timestamp(dt);
        assert_eq!(ts.seconds, dt.timestamp());
        assert_eq!(ts.nanos, 0);
    }

    #[test]
    fn test_job_to_proto_basic() {
        let job = make_job("job-1", "content-1", DomainJobStatus::Processing);
        let proto = job_to_proto(&job);

        assert_eq!(proto.job_id, "job-1");
        assert_eq!(proto.content_id, "content-1");
        assert_eq!(proto.status, encoding_proto::JobStatus::Processing as i32);
        assert!((proto.progress_percentage - 50.0).abs() < f64::EPSILON);
        assert!(proto.error_message.is_empty());
        assert!(proto.created_at.is_some());
        assert!(proto.started_at.is_some());
        assert!(proto.completed_at.is_none());
    }

    #[test]
    fn test_job_to_proto_with_outputs() {
        let mut job = make_job("job-1", "c-1", DomainJobStatus::Completed);
        job.outputs = vec![JobOutput {
            quality: "720p".to_string(),
            width: 1280,
            height: 720,
            bitrate_kbps: 2800,
            file_size_bytes: 50_000,
            segment_count: 10,
            playlist_path: "c-1/720p/playlist.m3u8".to_string(),
            storage_path: "c-1/720p/".to_string(),
        }];

        let proto = job_to_proto(&job);
        assert_eq!(proto.outputs.len(), 1);
        assert_eq!(proto.outputs[0].quality, "720p");
        assert_eq!(proto.outputs[0].width, 1280);
        assert_eq!(proto.outputs[0].segment_count, 10);
    }

    #[test]
    fn test_job_to_proto_failed_with_error() {
        let job = make_job("job-1", "c-1", DomainJobStatus::Failed);
        let proto = job_to_proto(&job);

        assert_eq!(proto.status, encoding_proto::JobStatus::Failed as i32);
        assert_eq!(proto.error_message, "test error");
        assert!(proto.completed_at.is_some());
    }

    async fn test_store() -> store::JobStore {
        store::JobStore::new("redis://localhost:6379/15")
            .await
            .expect("redis required for integration tests (DB 15)")
    }

    #[tokio::test]
    #[ignore]
    async fn test_get_job_status_found() {
        let job_store = test_store().await;
        store::insert_job(&job_store, make_job("grpc-job-1", "c-1", DomainJobStatus::Processing)).await;

        let svc = EncodingGrpcService::new(job_store);
        let req = tonic::Request::new(encoding_proto::GetJobStatusRequest {
            job_id: "grpc-job-1".to_string(),
        });

        let resp = svc.get_job_status(req).await.unwrap();
        let body = resp.into_inner();
        assert_eq!(body.job_id, "grpc-job-1");
        assert_eq!(body.status, encoding_proto::JobStatus::Processing as i32);
    }

    #[tokio::test]
    #[ignore]
    async fn test_get_job_status_not_found() {
        let job_store = test_store().await;
        let svc = EncodingGrpcService::new(job_store);

        let req = tonic::Request::new(encoding_proto::GetJobStatusRequest {
            job_id: "nonexistent-grpc".to_string(),
        });

        let err = svc.get_job_status(req).await.unwrap_err();
        assert_eq!(err.code(), tonic::Code::NotFound);
    }

    #[tokio::test]
    #[ignore]
    async fn test_get_job_status_empty_id() {
        let job_store = test_store().await;
        let svc = EncodingGrpcService::new(job_store);

        let req = tonic::Request::new(encoding_proto::GetJobStatusRequest {
            job_id: "".to_string(),
        });

        let err = svc.get_job_status(req).await.unwrap_err();
        assert_eq!(err.code(), tonic::Code::InvalidArgument);
    }

    #[tokio::test]
    #[ignore]
    async fn test_list_jobs_empty() {
        let job_store = test_store().await;
        let svc = EncodingGrpcService::new(job_store);

        let req = tonic::Request::new(encoding_proto::ListJobsRequest {
            status_filter: 0,
            content_id_filter: String::new(),
            limit: 0,
            offset: 0,
        });

        let resp = svc.list_jobs(req).await.unwrap().into_inner();
        // May have data from other tests, just verify it works
        assert!(resp.total_count >= 0);
    }

    #[tokio::test]
    #[ignore]
    async fn test_list_jobs_with_status_filter() {
        let job_store = test_store().await;
        store::insert_job(&job_store, make_job("grpc-j1", "c1", DomainJobStatus::Processing)).await;
        store::insert_job(&job_store, make_job("grpc-j2", "c2", DomainJobStatus::Completed)).await;
        store::insert_job(&job_store, make_job("grpc-j3", "c3", DomainJobStatus::Processing)).await;

        let svc = EncodingGrpcService::new(job_store);
        let req = tonic::Request::new(encoding_proto::ListJobsRequest {
            status_filter: encoding_proto::JobStatus::Processing as i32,
            content_id_filter: String::new(),
            limit: 50,
            offset: 0,
        });

        let resp = svc.list_jobs(req).await.unwrap().into_inner();
        assert!(resp.total_count >= 2);
        for j in &resp.jobs {
            assert_eq!(j.status, encoding_proto::JobStatus::Processing as i32);
        }
    }

    #[tokio::test]
    #[ignore]
    async fn test_list_jobs_with_content_id_filter() {
        let job_store = test_store().await;
        store::insert_job(&job_store, make_job("grpc-cid-j1", "grpc-c1", DomainJobStatus::Processing)).await;
        store::insert_job(&job_store, make_job("grpc-cid-j2", "grpc-c1", DomainJobStatus::Completed)).await;
        store::insert_job(&job_store, make_job("grpc-cid-j3", "grpc-c2", DomainJobStatus::Processing)).await;

        let svc = EncodingGrpcService::new(job_store);
        let req = tonic::Request::new(encoding_proto::ListJobsRequest {
            status_filter: 0,
            content_id_filter: "grpc-c1".to_string(),
            limit: 50,
            offset: 0,
        });

        let resp = svc.list_jobs(req).await.unwrap().into_inner();
        assert_eq!(resp.total_count, 2);
        for j in &resp.jobs {
            assert_eq!(j.content_id, "grpc-c1");
        }
    }

    #[tokio::test]
    #[ignore]
    async fn test_list_jobs_pagination() {
        let job_store = test_store().await;
        for i in 0..5 {
            store::insert_job(
                &job_store,
                make_job(&format!("grpc-page-j{}", i), "grpc-page-c1", DomainJobStatus::Processing),
            )
            .await;
        }

        let svc = EncodingGrpcService::new(job_store);
        let req = tonic::Request::new(encoding_proto::ListJobsRequest {
            status_filter: 0,
            content_id_filter: "grpc-page-c1".to_string(),
            limit: 2,
            offset: 1,
        });

        let resp = svc.list_jobs(req).await.unwrap().into_inner();
        assert_eq!(resp.total_count, 5);
        assert_eq!(resp.jobs.len(), 2);
    }
}
