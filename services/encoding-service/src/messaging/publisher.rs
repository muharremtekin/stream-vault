use async_trait::async_trait;
use lapin::options::BasicPublishOptions;
use lapin::types::FieldTable;
use lapin::{BasicProperties, Channel, Connection, ConnectionProperties};
use std::sync::Arc;
use tokio::sync::Mutex;
use tracing::{error, info};

use crate::config::RabbitMQConfig;
use crate::error::{EncodingError, Result};
use crate::messaging::models::EncodingResult;

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

    async fn publish(&self, result: &EncodingResult, routing_key: &str) -> Result<()> {
        let payload = serde_json::to_vec(result)?;
        let properties = BasicProperties::default()
            .with_content_type("application/json".into())
            .with_delivery_mode(2); // persistent

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
