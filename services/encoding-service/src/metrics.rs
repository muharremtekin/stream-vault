use axum::http::StatusCode;
use axum::response::{IntoResponse, Response};
use lazy_static::lazy_static;
use prometheus::{
    register_gauge, register_histogram_vec, register_int_counter_vec, register_int_gauge,
    Encoder, Gauge, HistogramVec, IntCounterVec, IntGauge, TextEncoder,
};

lazy_static! {
    // RED metrics
    pub static ref HTTP_REQUESTS_TOTAL: IntCounterVec = register_int_counter_vec!(
        "http_requests_total",
        "Total number of HTTP requests.",
        &["method", "path", "status"]
    )
    .unwrap();
    pub static ref HTTP_REQUEST_ERRORS_TOTAL: IntCounterVec = register_int_counter_vec!(
        "http_request_errors_total",
        "Total number of HTTP requests resulting in errors (4xx/5xx).",
        &["method", "path", "status"]
    )
    .unwrap();
    pub static ref HTTP_REQUEST_DURATION_SECONDS: HistogramVec = register_histogram_vec!(
        "http_request_duration_seconds",
        "HTTP request duration in seconds.",
        &["method", "path"],
        vec![0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0]
    )
    .unwrap();

    // Business metrics
    pub static ref ENCODING_JOBS_TOTAL: IntCounterVec = register_int_counter_vec!(
        "encoding_jobs_total",
        "Total encoding jobs processed by status.",
        &["status"]
    )
    .unwrap();
    pub static ref ENCODING_JOB_DURATION_SECONDS: HistogramVec = register_histogram_vec!(
        "encoding_job_duration_seconds",
        "Encoding job duration in seconds.",
        &["quality"],
        vec![10.0, 30.0, 60.0, 120.0, 300.0, 600.0, 1800.0]
    )
    .unwrap();
    pub static ref ENCODING_ACTIVE_JOBS: IntGauge =
        register_int_gauge!("encoding_active_jobs", "Number of currently processing encoding jobs.")
            .unwrap();
    pub static ref ENCODING_QUEUE_DEPTH: IntGauge =
        register_int_gauge!("encoding_queue_depth", "Number of jobs waiting in the encoding queue.")
            .unwrap();
    pub static ref ENCODING_FFMPEG_CPU_USAGE: Gauge = register_gauge!(
        "encoding_ffmpeg_cpu_usage",
        "Current FFmpeg CPU usage percentage."
    )
    .unwrap();
}

/// Axum handler that returns Prometheus metrics in text format.
pub async fn metrics_handler() -> Response {
    let encoder = TextEncoder::new();
    let metric_families = prometheus::gather();
    let mut buffer = Vec::new();
    match encoder.encode(&metric_families, &mut buffer) {
        Ok(()) => (
            StatusCode::OK,
            [(
                axum::http::header::CONTENT_TYPE,
                encoder.format_type().to_string(),
            )],
            buffer,
        )
            .into_response(),
        Err(e) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            format!("metrics encoding error: {}", e),
        )
            .into_response(),
    }
}

/// Axum middleware that records RED metrics for every HTTP request.
pub async fn metrics_middleware(
    req: axum::extract::Request,
    next: axum::middleware::Next,
) -> Response {
    let method = req.method().to_string();
    let path = normalize_path(req.uri().path());

    if path == "/metrics" {
        return next.run(req).await;
    }

    let start = std::time::Instant::now();
    let response = next.run(req).await;
    let duration = start.elapsed().as_secs_f64();

    let status = response.status().as_u16().to_string();

    HTTP_REQUESTS_TOTAL
        .with_label_values(&[&method, &path, &status])
        .inc();
    HTTP_REQUEST_DURATION_SECONDS
        .with_label_values(&[&method, &path])
        .observe(duration);

    if response.status().is_client_error() || response.status().is_server_error() {
        HTTP_REQUEST_ERRORS_TOTAL
            .with_label_values(&[&method, &path, &status])
            .inc();
    }

    response
}

fn normalize_path(path: &str) -> String {
    if path == "/health" || path == "/health/live" || path == "/health/ready" {
        return "/health".to_string();
    }
    if path == "/metrics" {
        return "/metrics".to_string();
    }
    if path.starts_with("/api/encoding/jobs/") && path.len() > "/api/encoding/jobs/".len() {
        return "/api/encoding/jobs/{id}".to_string();
    }
    if path.starts_with("/api/encoding/jobs") {
        return "/api/encoding/jobs".to_string();
    }
    path.to_string()
}
