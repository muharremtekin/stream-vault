use async_trait::async_trait;
use futures_lite::StreamExt;
use lapin::options::*;
use lapin::types::FieldTable;
use lapin::{Channel, Connection, ConnectionProperties};
use std::sync::Arc;
use std::time::Duration;
use tokio::sync::watch;
use tracing::{error, info, warn};

use crate::config::RabbitMQConfig;
use crate::error::Result;
use crate::messaging::models::EncodingJob;

/// Trait for processing encoding jobs. Implement with actual pipeline in Part 2.
#[async_trait]
pub trait JobProcessor: Send + Sync {
    async fn process(&self, job: EncodingJob) -> Result<()>;
}

pub struct JobConsumer {
    config: RabbitMQConfig,
    processor: Arc<dyn JobProcessor>,
    shutdown_rx: watch::Receiver<bool>,
}

impl JobConsumer {
    pub fn new(
        config: RabbitMQConfig,
        processor: Arc<dyn JobProcessor>,
        shutdown_rx: watch::Receiver<bool>,
    ) -> Self {
        Self {
            config,
            processor,
            shutdown_rx,
        }
    }

    pub async fn start(&mut self) {
        loop {
            if *self.shutdown_rx.borrow() {
                info!("consumer stopping (shutdown signal)");
                return;
            }

            match self.consume_loop().await {
                Ok(_) => {
                    info!("consumer loop exited cleanly");
                    return;
                }
                Err(e) => {
                    error!(error = %e, "consumer loop error");
                    if *self.shutdown_rx.borrow() {
                        return;
                    }
                    self.reconnect_backoff().await;
                }
            }
        }
    }

    async fn consume_loop(&mut self) -> Result<()> {
        let connection =
            Connection::connect(&self.config.url, ConnectionProperties::default()).await?;
        let channel = connection.create_channel().await?;

        channel
            .basic_qos(self.config.prefetch, BasicQosOptions::default())
            .await?;

        let mut consumer = channel
            .basic_consume(
                &self.config.jobs_queue,
                "encoding-consumer",
                BasicConsumeOptions {
                    no_ack: false,
                    ..Default::default()
                },
                FieldTable::default(),
            )
            .await?;

        info!(queue = %self.config.jobs_queue, prefetch = self.config.prefetch, "consumer started");

        loop {
            tokio::select! {
                _ = self.shutdown_rx.changed() => {
                    info!("consumer stopping");
                    let _ = channel.close(200, "shutdown").await;
                    return Ok(());
                }
                delivery = consumer.next() => {
                    match delivery {
                        Some(Ok(delivery)) => {
                            self.handle_delivery(&channel, delivery).await;
                        }
                        Some(Err(e)) => {
                            error!(error = %e, "delivery error");
                            return Err(e.into());
                        }
                        None => {
                            warn!("consumer channel closed");
                            return Err(crate::error::EncodingError::Internal(
                                "consumer channel closed".into(),
                            ));
                        }
                    }
                }
            }
        }
    }

    async fn handle_delivery(&self, _channel: &Channel, delivery: lapin::message::Delivery) {
        let job = match serde_json::from_slice::<EncodingJob>(&delivery.data) {
            Ok(job) => job,
            Err(e) => {
                error!(error = %e, "failed to deserialize job, sending to DLX");
                let _ = delivery
                    .nack(BasicNackOptions {
                        requeue: false,
                        multiple: false,
                    })
                    .await;
                return;
            }
        };

        info!(
            job_id = %job.job_id,
            content_id = %job.content_id,
            source = %job.source_path,
            "received encoding job"
        );

        let mut last_err = None;
        for attempt in 1..=self.config.max_retries {
            match self.processor.process(job.clone()).await {
                Ok(_) => {
                    info!(job_id = %job.job_id, "job processed successfully");
                    let _ = delivery.ack(BasicAckOptions::default()).await;
                    return;
                }
                Err(e) => {
                    if attempt < self.config.max_retries {
                        warn!(
                            job_id = %job.job_id,
                            error = %e,
                            attempt,
                            max_retries = self.config.max_retries,
                            "job processing failed, retrying"
                        );
                        tokio::time::sleep(Duration::from_secs(self.config.retry_delay_secs))
                            .await;
                    }
                    last_err = Some(e);
                }
            }
        }

        error!(
            job_id = %job.job_id,
            error = %last_err.unwrap(),
            "job failed after all retries, sending to DLX"
        );
        let _ = delivery
            .nack(BasicNackOptions {
                requeue: false,
                multiple: false,
            })
            .await;
    }

    async fn reconnect_backoff(&self) {
        let mut backoff = Duration::from_secs(1);
        let max_backoff = Duration::from_secs(30);

        loop {
            if *self.shutdown_rx.borrow() {
                return;
            }

            warn!(backoff_secs = backoff.as_secs(), "waiting before reconnect");
            tokio::time::sleep(backoff).await;

            match Connection::connect(&self.config.url, ConnectionProperties::default()).await {
                Ok(_) => {
                    info!("reconnect probe succeeded");
                    return;
                }
                Err(e) => {
                    error!(error = %e, "reconnect probe failed");
                    backoff = (backoff * 2).min(max_backoff);
                }
            }
        }
    }
}
