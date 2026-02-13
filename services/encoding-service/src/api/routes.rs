use axum::routing::get;
use axum::Router;

use crate::AppState;

pub fn encoding_routes() -> Router<AppState> {
    Router::new()
        .route("/api/encoding/jobs", get(super::handlers::list_jobs))
        .route("/api/encoding/jobs/{job_id}", get(super::handlers::get_job))
}
