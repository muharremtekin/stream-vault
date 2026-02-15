use serde::Deserialize;

#[derive(Debug, Clone, Deserialize)]
pub struct Config {
    pub server: ServerConfig,
    pub consul: ConsulConfig,
    pub minio: MinIOConfig,
    pub rabbitmq: RabbitMQConfig,
    pub encoding: EncodingConfig,
    pub logging: LoggingConfig,
    pub telemetry: TelemetryConfig,
}

#[derive(Debug, Clone, Deserialize)]
pub struct ServerConfig {
    pub http_port: u16,
    pub grpc_port: u16,
}

#[derive(Debug, Clone, Deserialize)]
pub struct ConsulConfig {
    pub address: String,
    pub service_name: String,
    pub health_check_interval_secs: u64,
}

#[derive(Debug, Clone, Deserialize)]
pub struct MinIOConfig {
    pub endpoint: String,
    pub access_key: String,
    pub secret_key: String,
    pub use_ssl: bool,
    pub raw_bucket: String,
    pub encoded_bucket: String,
    pub thumbnails_bucket: String,
    pub region: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct RabbitMQConfig {
    pub url: String,
    pub jobs_queue: String,
    pub exchange: String,
    pub completed_routing_key: String,
    pub failed_routing_key: String,
    pub prefetch: u16,
    pub max_retries: u32,
    pub retry_delay_secs: u64,
}

#[derive(Debug, Clone, Deserialize)]
pub struct EncodingConfig {
    pub temp_dir: String,
    pub ffmpeg_path: String,
    pub ffprobe_path: String,
    pub max_concurrent_jobs: usize,
}

#[derive(Debug, Clone, Deserialize)]
pub struct LoggingConfig {
    pub level: String,
    pub format: String,
}

#[derive(Debug, Clone, Deserialize)]
pub struct TelemetryConfig {
    pub enabled: bool,
    pub endpoint: String,
    pub service_name: String,
    pub insecure: bool,
}

pub fn load() -> anyhow::Result<Config> {
    let cfg = config::Config::builder()
        // Defaults
        .set_default("server.http_port", 5004)?
        .set_default("server.grpc_port", 50052)?
        .set_default("consul.address", "localhost:8500")?
        .set_default("consul.service_name", "encoding-service")?
        .set_default("consul.health_check_interval_secs", 10)?
        .set_default("minio.endpoint", "localhost:9000")?
        .set_default("minio.access_key", "")?
        .set_default("minio.secret_key", "")?
        .set_default("minio.use_ssl", false)?
        .set_default("minio.raw_bucket", "streamvault-raw")?
        .set_default("minio.encoded_bucket", "streamvault-encoded")?
        .set_default("minio.thumbnails_bucket", "streamvault-thumbnails")?
        .set_default("minio.region", "us-east-1")?
        .set_default("rabbitmq.url", "amqp://localhost:5672/")?
        .set_default("rabbitmq.jobs_queue", "encoding.jobs")?
        .set_default("rabbitmq.exchange", "encoding")?
        .set_default("rabbitmq.completed_routing_key", "job.completed")?
        .set_default("rabbitmq.failed_routing_key", "job.failed")?
        .set_default("rabbitmq.prefetch", 2)?
        .set_default("rabbitmq.max_retries", 3)?
        .set_default("rabbitmq.retry_delay_secs", 5)?
        .set_default("encoding.temp_dir", "/tmp/encoding")?
        .set_default("encoding.ffmpeg_path", "/usr/bin/ffmpeg")?
        .set_default("encoding.ffprobe_path", "/usr/bin/ffprobe")?
        .set_default("encoding.max_concurrent_jobs", 2)?
        .set_default("logging.level", "info")?
        .set_default("logging.format", "json")?
        .set_default("telemetry.enabled", true)?
        .set_default("telemetry.endpoint", "http://localhost:4318")?
        .set_default("telemetry.service_name", "encoding-service")?
        .set_default("telemetry.insecure", true)?
        // File source
        .add_source(config::File::with_name("config").required(false))
        // Environment variables: ENCODING_MINIO__ACCESS_KEY, ENCODING_RABBITMQ__URL, etc.
        // Use "__" (double underscore) as hierarchy separator so single underscores
        // in field names (access_key, secret_key, etc.) are preserved.
        .add_source(
            config::Environment::with_prefix("ENCODING")
                .prefix_separator("_")
                .separator("__")
                .try_parsing(true),
        )
        .build()?;

    cfg.try_deserialize().map_err(Into::into)
}
