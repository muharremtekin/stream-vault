use thiserror::Error;

#[derive(Error, Debug)]
pub enum EncodingError {
    #[error("configuration error: {0}")]
    Config(String),

    #[error("storage error: {0}")]
    Storage(String),

    #[error("rabbitmq error: {0}")]
    RabbitMQ(#[from] lapin::Error),

    #[error("consul error: {0}")]
    Consul(String),

    #[error("serialization error: {0}")]
    Serialization(#[from] serde_json::Error),

    #[error("io error: {0}")]
    Io(#[from] std::io::Error),

    #[error("http error: {0}")]
    Http(#[from] reqwest::Error),

    #[error("encoding job error: {0}")]
    Job(String),

    #[error("ffmpeg not found or not executable")]
    FFmpegNotFound,

    #[error("internal error: {0}")]
    Internal(String),
}

pub type Result<T> = std::result::Result<T, EncodingError>;
