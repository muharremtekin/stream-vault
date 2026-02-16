use async_trait::async_trait;
use lapin::options::BasicPublishOptions;
use lapin::types::FieldTable;
use lapin::{BasicProperties, Channel, Connection, ConnectionProperties};
use std::sync::Arc;
use tokio::sync::Mutex;
use tracing::{error, info, warn};

use crate::config::RabbitMQConfig;
use crate::error::{EncodingError, Result};
use crate::messaging::models::EncodingResult;
use crate::telemetry;

#[async_trait]
pub trait ResultPublisher: Send + Sync {
    async fn publish_completed(&self, result: EncodingResult) -> Result<()>;
    async fn publish_failed(&self, result: EncodingResult) -> Result<()>;
}

pub struct RabbitMQPublisher {
    connection: Connection,
    channel: Arc<Mutex<Channel>>,
    config: RabbitMQConfig,
}

impl RabbitMQPublisher {
    pub async fn new(config: RabbitMQConfig) -> Result<Self> {
        let connection =
            Connection::connect(&config.url, ConnectionProperties::default()).await?;
        let channel = connection.create_channel().await?;

        // Declare exchange to be safe (idempotent if already exists)
        channel
            .exchange_declare(
                &config.exchange,
                lapin::ExchangeKind::Topic,
                lapin::options::ExchangeDeclareOptions {
                    durable: true,
                    passive: false,
                    ..Default::default()
                },
                FieldTable::default(),
            )
            .await?;

        info!(exchange = %config.exchange, "rabbitmq publisher ready");

        Ok(Self {
            connection,
            channel: Arc::new(Mutex::new(channel)),
            config,
        })
    }

    #[tracing::instrument(
        skip_all,
        fields(
            messaging.system = "rabbitmq",
            messaging.destination.name = %self.config.exchange,
            messaging.rabbitmq.routing_key = %routing_key,
            messaging.message.id = %result.event_id,
            job.id = %result.job_id,
        ),
        name = "rabbitmq.publish"
    )]
    async fn publish(&self, result: &EncodingResult, routing_key: &str) -> Result<()> {
        let payload = serde_json::to_vec(result)?;
        let headers = telemetry::inject_context();
        let properties = BasicProperties::default()
            .with_content_type("application/json".into())
            .with_delivery_mode(2) // persistent
            .with_headers(headers);

        let publish_result = {
            let channel = self.channel.lock().await;
            channel
                .basic_publish(
                    &self.config.exchange,
                    routing_key,
                    BasicPublishOptions::default(),
                    &payload,
                    properties.clone(),
                )
                .await
        };

        match publish_result {
            Ok(confirm) => {
                confirm.await.map_err(|e| {
                    EncodingError::Internal(format!("publish confirm error: {}", e))
                })?;
                info!(
                    job_id = %result.job_id,
                    status = %result.status,
                    routing_key,
                    "published result"
                );
                Ok(())
            }
            Err(e) => {
                error!(error = ?e, "publish failed, recreating channel");
                self.recreate_channel().await?;

                let channel = self.channel.lock().await;
                channel
                    .basic_publish(
                        &self.config.exchange,
                        routing_key,
                        BasicPublishOptions::default(),
                        &payload,
                        properties,
                    )
                    .await?
                    .await
                    .map_err(|e| {
                        EncodingError::Internal(format!("publish confirm error: {}", e))
                    })?;

                info!(
                    job_id = %result.job_id,
                    status = %result.status,
                    "published result (after channel recreation)"
                );
                Ok(())
            }
        }
    }

    async fn recreate_channel(&self) -> Result<()> {
        let new_channel = self.connection.create_channel().await?;
        let mut channel = self.channel.lock().await;
        *channel = new_channel;
        info!("publisher channel recreated");
        Ok(())
    }

    pub fn connection(&self) -> &Connection {
        &self.connection
    }

    pub async fn close(&self) {
        let channel = self.channel.lock().await;
        if let Err(e) = channel.close(200, "shutdown").await {
            warn!(error = %e, "error closing publisher channel");
        }
        if let Err(e) = self.connection.close(200, "shutdown").await {
            warn!(error = %e, "error closing publisher connection");
        }
        info!("rabbitmq publisher connection closed");
    }
}

#[async_trait]
impl ResultPublisher for RabbitMQPublisher {
    async fn publish_completed(&self, result: EncodingResult) -> Result<()> {
        self.publish(&result, &self.config.completed_routing_key)
            .await
    }

    async fn publish_failed(&self, result: EncodingResult) -> Result<()> {
        self.publish(&result, &self.config.failed_routing_key).await
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::messaging::models::{EncodingOutput, EncodingResult};
    use std::sync::Mutex;

    /// Mock publisher that records all publish calls for assertion.
    struct TrackingPublisher {
        completed: Mutex<Vec<EncodingResult>>,
        failed: Mutex<Vec<EncodingResult>>,
    }

    impl TrackingPublisher {
        fn new() -> Self {
            Self {
                completed: Mutex::new(Vec::new()),
                failed: Mutex::new(Vec::new()),
            }
        }
    }

    #[async_trait]
    impl ResultPublisher for TrackingPublisher {
        async fn publish_completed(&self, result: EncodingResult) -> Result<()> {
            self.completed.lock().unwrap().push(result);
            Ok(())
        }
        async fn publish_failed(&self, result: EncodingResult) -> Result<()> {
            self.failed.lock().unwrap().push(result);
            Ok(())
        }
    }

    fn make_result(status: &str) -> EncodingResult {
        EncodingResult {
            event_id: "evt-001".into(),
            event_type: format!("encoding.job.{}", status),
            timestamp: "2024-01-15T10:35:00Z".into(),
            source: "encoding-service".into(),
            correlation_id: "job-001".into(),
            job_id: "job-001".into(),
            content_id: "movie-456".into(),
            status: status.into(),
            outputs: if status == "completed" {
                vec![EncodingOutput {
                    quality: "720p".into(),
                    width: 1280,
                    height: 720,
                    bitrate_kbps: 2800,
                    segment_count: 12,
                    playlist_path: "movie-456/720p/playlist.m3u8".into(),
                }]
            } else {
                vec![]
            },
            duration_seconds: 120,
            error_message: if status == "failed" {
                Some("codec not supported".into())
            } else {
                None
            },
            completed_at: "2024-01-15T10:35:00Z".into(),
        }
    }

    #[tokio::test]
    async fn publish_completed_records_result() {
        let publisher = TrackingPublisher::new();
        let result = make_result("completed");

        publisher.publish_completed(result).await.unwrap();

        let calls = publisher.completed.lock().unwrap();
        assert_eq!(calls.len(), 1);
        assert_eq!(calls[0].status, "completed");
        assert_eq!(calls[0].job_id, "job-001");
        assert_eq!(calls[0].content_id, "movie-456");
        assert_eq!(calls[0].outputs.len(), 1);
        assert_eq!(calls[0].outputs[0].quality, "720p");
        assert!(calls[0].error_message.is_none());
    }

    #[tokio::test]
    async fn publish_failed_records_result() {
        let publisher = TrackingPublisher::new();
        let result = make_result("failed");

        publisher.publish_failed(result).await.unwrap();

        let calls = publisher.failed.lock().unwrap();
        assert_eq!(calls.len(), 1);
        assert_eq!(calls[0].status, "failed");
        assert_eq!(calls[0].job_id, "job-001");
        assert_eq!(calls[0].outputs.len(), 0);
        assert_eq!(
            calls[0].error_message.as_deref(),
            Some("codec not supported")
        );
    }

    #[tokio::test]
    async fn publish_completed_and_failed_are_independent() {
        let publisher = TrackingPublisher::new();

        publisher
            .publish_completed(make_result("completed"))
            .await
            .unwrap();
        publisher
            .publish_failed(make_result("failed"))
            .await
            .unwrap();

        assert_eq!(publisher.completed.lock().unwrap().len(), 1);
        assert_eq!(publisher.failed.lock().unwrap().len(), 1);
    }

    #[tokio::test]
    async fn publish_result_event_envelope_fields() {
        let publisher = TrackingPublisher::new();
        let result = make_result("completed");

        publisher.publish_completed(result).await.unwrap();

        let calls = publisher.completed.lock().unwrap();
        assert_eq!(calls[0].event_id, "evt-001");
        assert_eq!(calls[0].event_type, "encoding.job.completed");
        assert_eq!(calls[0].source, "encoding-service");
        assert_eq!(calls[0].correlation_id, "job-001");
    }
}
