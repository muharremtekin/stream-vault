use async_trait::async_trait;
use futures_lite::StreamExt;
use lapin::options::*;
use lapin::types::FieldTable;
use lapin::{Connection, ConnectionProperties};
use std::sync::Arc;
use std::time::Duration;
use tokio::sync::{watch, Semaphore};
use tokio::task::JoinSet;
use tracing::{error, info, warn, Instrument};
use tracing_opentelemetry::OpenTelemetrySpanExt;

use crate::config::RabbitMQConfig;
use crate::error::Result;
use crate::messaging::models::EncodingJob;
use crate::telemetry;

/// Trait for processing encoding jobs. Implement with actual pipeline in Part 2.
#[async_trait]
pub trait JobProcessor: Send + Sync {
    async fn process(&self, job: EncodingJob) -> Result<()>;
}

pub struct JobConsumer {
    config: RabbitMQConfig,
    runner: JobRunner,
    max_concurrent_jobs: usize,
    shutdown_rx: watch::Receiver<bool>,
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::error::EncodingError;
    use chrono::Utc;
    use std::collections::VecDeque;
    use std::sync::atomic::{AtomicUsize, Ordering};
    use tokio::sync::{mpsc, Mutex};

    fn job(id: usize) -> EncodingJob {
        EncodingJob {
            job_id: format!("job-{id}"),
            content_id: format!("content-{id}"),
            source_path: "source.mp4".into(),
            source_bucket: "raw".into(),
            requested_by: "tester".into(),
            created_at: Utc::now(),
        }
    }

    struct SequenceProcessor {
        outcomes: Mutex<VecDeque<bool>>,
        attempts: AtomicUsize,
    }

    impl SequenceProcessor {
        fn new(outcomes: impl IntoIterator<Item = bool>) -> Self {
            Self {
                outcomes: Mutex::new(outcomes.into_iter().collect()),
                attempts: AtomicUsize::new(0),
            }
        }
    }

    #[async_trait]
    impl JobProcessor for SequenceProcessor {
        async fn process(&self, _job: EncodingJob) -> Result<()> {
            self.attempts.fetch_add(1, Ordering::SeqCst);
            if self.outcomes.lock().await.pop_front().unwrap_or(false) {
                Ok(())
            } else {
                Err(EncodingError::Job("controlled failure".into()))
            }
        }
    }

    fn runner(processor: Arc<dyn JobProcessor>, max_retries: u32) -> JobRunner {
        JobRunner::new(processor, 1, max_retries, Duration::ZERO)
    }

    #[tokio::test]
    async fn succeeds_on_initial_attempt() {
        let processor = Arc::new(SequenceProcessor::new([true]));

        assert!(runner(processor.clone(), 3).run(job(1)).await.is_ok());
        assert_eq!(1, processor.attempts.load(Ordering::SeqCst));
    }

    #[tokio::test]
    async fn succeeds_after_retry() {
        let processor = Arc::new(SequenceProcessor::new([false, true]));

        assert!(runner(processor.clone(), 3).run(job(1)).await.is_ok());
        assert_eq!(2, processor.attempts.load(Ordering::SeqCst));
    }

    #[tokio::test]
    async fn returns_last_error_after_retries_are_exhausted() {
        let processor = Arc::new(SequenceProcessor::new([false, false, false]));

        assert!(runner(processor.clone(), 2).run(job(1)).await.is_err());
        assert_eq!(3, processor.attempts.load(Ordering::SeqCst));
    }

    #[tokio::test]
    async fn zero_retries_still_runs_the_initial_attempt_once() {
        let processor = Arc::new(SequenceProcessor::new([false]));

        assert!(runner(processor.clone(), 0).run(job(1)).await.is_err());
        assert_eq!(1, processor.attempts.load(Ordering::SeqCst));
    }

    struct BlockingProcessor {
        active: AtomicUsize,
        max_active: AtomicUsize,
        started: mpsc::UnboundedSender<String>,
        release: Semaphore,
    }

    #[async_trait]
    impl JobProcessor for BlockingProcessor {
        async fn process(&self, job: EncodingJob) -> Result<()> {
            let active = self.active.fetch_add(1, Ordering::SeqCst) + 1;
            self.max_active.fetch_max(active, Ordering::SeqCst);
            self.started.send(job.job_id).unwrap();

            let permit = self.release.acquire().await.unwrap();
            permit.forget();
            self.active.fetch_sub(1, Ordering::SeqCst);
            Ok(())
        }
    }

    #[tokio::test]
    async fn runs_jobs_in_parallel_without_exceeding_configured_limit() {
        let (started_tx, mut started_rx) = mpsc::unbounded_channel();
        let processor = Arc::new(BlockingProcessor {
            active: AtomicUsize::new(0),
            max_active: AtomicUsize::new(0),
            started: started_tx,
            release: Semaphore::new(0),
        });
        let runner = JobRunner::new(processor.clone(), 2, 0, Duration::ZERO);

        let handles: Vec<_> = (0..4)
            .map(|id| {
                let runner = runner.clone();
                tokio::spawn(async move { runner.run(job(id)).await })
            })
            .collect();

        // Two start events before any release prove actual parallel execution.
        for _ in 0..2 {
            tokio::time::timeout(Duration::from_secs(1), started_rx.recv())
                .await
                .expect("two jobs should start in parallel")
                .expect("start channel should remain open");
        }
        assert_eq!(2, processor.active.load(Ordering::SeqCst));

        processor.release.add_permits(2);
        for _ in 0..2 {
            tokio::time::timeout(Duration::from_secs(1), started_rx.recv())
                .await
                .expect("queued jobs should start after permits are released")
                .expect("start channel should remain open");
        }
        processor.release.add_permits(2);

        for handle in handles {
            handle.await.unwrap().unwrap();
        }
        assert_eq!(2, processor.max_active.load(Ordering::SeqCst));
    }

    #[test]
    fn prefetch_is_raised_to_the_configured_concurrency_limit() {
        assert_eq!(8, effective_prefetch(2, 8));
        assert_eq!(8, effective_prefetch(8, 2));
    }
}

fn effective_prefetch(configured_prefetch: u16, max_concurrent_jobs: usize) -> u16 {
    let concurrency_limit = u16::try_from(max_concurrent_jobs)
        .expect("validated max_concurrent_jobs must fit in a u16");
    configured_prefetch.max(concurrency_limit)
}

#[derive(Clone)]
struct JobRunner {
    processor: Arc<dyn JobProcessor>,
    permits: Arc<Semaphore>,
    max_retries: u32,
    retry_delay: Duration,
}

impl JobRunner {
    fn new(
        processor: Arc<dyn JobProcessor>,
        max_concurrent_jobs: usize,
        max_retries: u32,
        retry_delay: Duration,
    ) -> Self {
        assert!(max_concurrent_jobs > 0, "max_concurrent_jobs must be at least 1");
        Self {
            processor,
            permits: Arc::new(Semaphore::new(max_concurrent_jobs)),
            max_retries,
            retry_delay,
        }
    }

    async fn run(&self, job: EncodingJob) -> Result<()> {
        let _permit = self
            .permits
            .acquire()
            .await
            .map_err(|_| crate::error::EncodingError::Internal("job limiter closed".into()))?;

        // max_retries is the number of retries after the initial attempt.
        for retry in 0..=self.max_retries {
            match self.processor.process(job.clone()).await {
                Ok(()) => return Ok(()),
                Err(error) if retry < self.max_retries => {
                    warn!(
                        job_id = %job.job_id,
                        error = %error,
                        attempt = retry + 1,
                        max_retries = self.max_retries,
                        "job processing failed, retrying"
                    );
                    tokio::time::sleep(self.retry_delay).await;
                }
                Err(error) => return Err(error),
            }
        }

        unreachable!("the inclusive retry range always executes the initial attempt")
    }
}

impl JobConsumer {
    pub fn new(
        config: RabbitMQConfig,
        max_concurrent_jobs: usize,
        processor: Arc<dyn JobProcessor>,
        shutdown_rx: watch::Receiver<bool>,
    ) -> Self {
        let runner = JobRunner::new(
            processor,
            max_concurrent_jobs,
            config.max_retries,
            Duration::from_secs(config.retry_delay_secs),
        );
        Self {
            config,
            runner,
            max_concurrent_jobs,
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
        let prefetch = effective_prefetch(self.config.prefetch, self.max_concurrent_jobs);

        channel
            .basic_qos(prefetch, BasicQosOptions::default())
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

        info!(
            queue = %self.config.jobs_queue,
            configured_prefetch = self.config.prefetch,
            prefetch,
            max_concurrent_jobs = self.max_concurrent_jobs,
            "consumer started"
        );

        let mut in_flight = JoinSet::new();

        loop {
            tokio::select! {
                _ = self.shutdown_rx.changed() => {
                    info!("consumer stopping");
                    let _ = channel
                        .basic_cancel("encoding-consumer", BasicCancelOptions::default())
                        .await;
                    Self::drain_in_flight(&mut in_flight).await;
                    let _ = channel.close(200, "shutdown").await;
                    return Ok(());
                }
                completed = in_flight.join_next(), if !in_flight.is_empty() => {
                    if let Some(Err(join_error)) = completed {
                        error!(error = %join_error, "encoding job task failed");
                    }
                }
                delivery = consumer.next() => {
                    match delivery {
                        Some(Ok(delivery)) => {
                            let config = self.config.clone();
                            let runner = self.runner.clone();
                            in_flight.spawn(async move {
                                Self::handle_delivery(config, runner, delivery).await;
                            });
                        }
                        Some(Err(e)) => {
                            error!(error = %e, "delivery error");
                            Self::drain_in_flight(&mut in_flight).await;
                            return Err(e.into());
                        }
                        None => {
                            warn!("consumer channel closed");
                            Self::drain_in_flight(&mut in_flight).await;
                            return Err(crate::error::EncodingError::Internal(
                                "consumer channel closed".into(),
                            ));
                        }
                    }
                }
            }
        }
    }

    async fn drain_in_flight(in_flight: &mut JoinSet<()>) {
        while let Some(result) = in_flight.join_next().await {
            if let Err(join_error) = result {
                error!(error = %join_error, "encoding job task failed during shutdown");
            }
        }
    }

    async fn handle_delivery(
        config: RabbitMQConfig,
        runner: JobRunner,
        delivery: lapin::message::Delivery,
    ) {
        // Extract trace context from AMQP headers (links to upstream producer span)
        let headers = delivery
            .properties
            .headers()
            .as_ref()
            .cloned()
            .unwrap_or_default();
        let parent_cx = telemetry::extract_context(&headers);

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

        let consume_span = tracing::info_span!(
            "rabbitmq.consume",
            messaging.system = "rabbitmq",
            messaging.source.name = %config.jobs_queue,
            messaging.operation = "receive",
            job.id = %job.job_id,
            content.id = %job.content_id,
        );
        consume_span.set_parent(parent_cx);
        async move {
            info!(
                job_id = %job.job_id,
                content_id = %job.content_id,
                source = %job.source_path,
                "received encoding job"
            );

            match runner.run(job.clone()).await {
                Ok(()) => {
                    info!(job_id = %job.job_id, "job processed successfully");
                    let _ = delivery.ack(BasicAckOptions::default()).await;
                }
                Err(e) => {
                    error!(
                        job_id = %job.job_id,
                        error = %e,
                        "job failed after all retries, sending to DLX"
                    );
                    let _ = delivery
                        .nack(BasicNackOptions {
                            requeue: false,
                            multiple: false,
                        })
                        .await;
                }
            }
        }
        .instrument(consume_span)
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
