use serde::Serialize;
use std::time::Duration;
use tracing::{info, warn};

use crate::config::ConsulConfig;
use crate::error::{EncodingError, Result};

pub struct ConsulClient {
    client: reqwest::Client,
    base_url: String,
    service_id: Option<String>,
}

#[derive(Serialize)]
struct ServiceRegistration {
    #[serde(rename = "ID")]
    id: String,
    #[serde(rename = "Name")]
    name: String,
    #[serde(rename = "Address")]
    address: String,
    #[serde(rename = "Port")]
    port: u16,
    #[serde(rename = "Tags")]
    tags: Vec<String>,
    #[serde(rename = "Check")]
    check: HealthCheck,
}

#[derive(Serialize)]
struct HealthCheck {
    #[serde(rename = "HTTP")]
    http: String,
    #[serde(rename = "Interval")]
    interval: String,
    #[serde(rename = "Timeout")]
    timeout: String,
    #[serde(rename = "DeregisterCriticalServiceAfter")]
    deregister_critical_service_after: String,
}

impl ConsulClient {
    pub fn new(address: &str) -> Result<Self> {
        let client = reqwest::Client::builder()
            .timeout(Duration::from_secs(5))
            .build()
            .map_err(EncodingError::Http)?;

        Ok(Self {
            client,
            base_url: format!("http://{}", address),
            service_id: None,
        })
    }

    pub async fn register(&mut self, config: &ConsulConfig, http_port: u16) -> Result<()> {
        let service_id = config.service_name.clone();

        let registration = ServiceRegistration {
            id: service_id.clone(),
            name: config.service_name.clone(),
            address: config.service_name.clone(),
            port: http_port,
            tags: vec!["encoding".into(), "api".into(), "v1".into()],
            check: HealthCheck {
                http: format!(
                    "http://{}:{}/health",
                    config.service_name, http_port
                ),
                interval: format!("{}s", config.health_check_interval_secs),
                timeout: "5s".into(),
                deregister_critical_service_after: "30s".into(),
            },
        };

        let url = format!("{}/v1/agent/service/register", self.base_url);
        let response = self
            .client
            .put(&url)
            .json(&registration)
            .send()
            .await
            .map_err(EncodingError::Http)?;

        if response.status().is_success() {
            self.service_id = Some(service_id.clone());
            info!(service_id = %service_id, "registered with consul");
            Ok(())
        } else {
            let status = response.status();
            let body = response.text().await.unwrap_or_default();
            Err(EncodingError::Consul(format!(
                "registration failed: {} - {}",
                status, body
            )))
        }
    }

    pub async fn deregister(&self) {
        let Some(service_id) = &self.service_id else {
            return;
        };

        let url = format!(
            "{}/v1/agent/service/deregister/{}",
            self.base_url, service_id
        );

        match self.client.put(&url).send().await {
            Ok(resp) if resp.status().is_success() => {
                info!(service_id = %service_id, "deregistered from consul");
            }
            Ok(resp) => {
                warn!(status = %resp.status(), "consul deregistration returned error");
            }
            Err(e) => {
                warn!(error = %e, "consul deregistration failed");
            }
        }
    }
}
