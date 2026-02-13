use async_trait::async_trait;
use aws_credential_types::Credentials;
use aws_sdk_s3::config::{BehaviorVersion, Region};
use aws_sdk_s3::Client;
use aws_smithy_types::byte_stream::ByteStream;
use bytes::Bytes;
use std::path::Path;
use tokio::fs::File;
use tokio::io::AsyncWriteExt;
use tracing::{debug, info};

use crate::config::MinIOConfig;
use crate::error::{EncodingError, Result};
use crate::storage::StorageClient;

pub struct MinIOClient {
    client: Client,
}

impl MinIOClient {
    pub async fn new(config: &MinIOConfig) -> Result<Self> {
        let credentials = Credentials::new(
            &config.access_key,
            &config.secret_key,
            None,
            None,
            "static",
        );

        let endpoint = format!(
            "{}://{}",
            if config.use_ssl { "https" } else { "http" },
            config.endpoint
        );

        let s3_config = aws_sdk_s3::Config::builder()
            .endpoint_url(&endpoint)
            .region(Region::new(config.region.clone()))
            .credentials_provider(credentials)
            .force_path_style(true)
            .behavior_version(BehaviorVersion::latest())
            .build();

        let client = Client::from_conf(s3_config);

        info!(endpoint = %config.endpoint, "minio client initialized");
        Ok(Self { client })
    }
}

#[async_trait]
impl StorageClient for MinIOClient {
    async fn download_to_file(
        &self,
        bucket: &str,
        key: &str,
        dest: &Path,
    ) -> Result<u64> {
        let resp = self
            .client
            .get_object()
            .bucket(bucket)
            .key(key)
            .send()
            .await
            .map_err(|e| EncodingError::Storage(format!("get_object {}/{}: {}", bucket, key, e)))?;

        let mut file = File::create(dest).await?;
        let mut stream = resp.body.into_async_read();
        let bytes_written = tokio::io::copy(&mut stream, &mut file).await?;
        file.flush().await?;

        debug!(bucket, key, bytes = bytes_written, dest = ?dest, "downloaded object");
        Ok(bytes_written)
    }

    async fn upload_from_file(
        &self,
        bucket: &str,
        key: &str,
        file_path: &Path,
        content_type: &str,
    ) -> Result<()> {
        let body = ByteStream::from_path(file_path)
            .await
            .map_err(|e| EncodingError::Storage(format!("read file {}: {}", file_path.display(), e)))?;

        self.client
            .put_object()
            .bucket(bucket)
            .key(key)
            .body(body)
            .content_type(content_type)
            .send()
            .await
            .map_err(|e| EncodingError::Storage(format!("put_object {}/{}: {}", bucket, key, e)))?;

        debug!(bucket, key, path = ?file_path, "uploaded file");
        Ok(())
    }

    async fn upload_bytes(
        &self,
        bucket: &str,
        key: &str,
        data: Bytes,
        content_type: &str,
    ) -> Result<()> {
        let size = data.len();
        let body = ByteStream::from(data);

        self.client
            .put_object()
            .bucket(bucket)
            .key(key)
            .body(body)
            .content_type(content_type)
            .send()
            .await
            .map_err(|e| EncodingError::Storage(format!("put_object {}/{}: {}", bucket, key, e)))?;

        debug!(bucket, key, size, "uploaded bytes");
        Ok(())
    }

    async fn delete(&self, bucket: &str, key: &str) -> Result<()> {
        self.client
            .delete_object()
            .bucket(bucket)
            .key(key)
            .send()
            .await
            .map_err(|e| EncodingError::Storage(format!("delete {}/{}: {}", bucket, key, e)))?;

        debug!(bucket, key, "deleted object");
        Ok(())
    }

    async fn exists(&self, bucket: &str, key: &str) -> Result<bool> {
        match self.client.head_object().bucket(bucket).key(key).send().await {
            Ok(_) => Ok(true),
            Err(e) => {
                let err_str = format!("{}", e);
                if err_str.contains("NotFound") || err_str.contains("404") || err_str.contains("NoSuchKey") {
                    Ok(false)
                } else {
                    Err(EncodingError::Storage(format!(
                        "head_object {}/{}: {}",
                        bucket, key, e
                    )))
                }
            }
        }
    }

    async fn health_check(&self, bucket: &str) -> Result<()> {
        self.client
            .head_bucket()
            .bucket(bucket)
            .send()
            .await
            .map_err(|e| EncodingError::Storage(format!("head_bucket {}: {}", bucket, e)))?;
        Ok(())
    }
}

/// Stub storage client for unit tests (no real MinIO connection).
#[cfg(test)]
pub struct StubStorageClient;

#[cfg(test)]
#[async_trait]
impl StorageClient for StubStorageClient {
    async fn download_to_file(&self, _: &str, _: &str, _: &Path) -> Result<u64> {
        Ok(0)
    }
    async fn upload_from_file(&self, _: &str, _: &str, _: &Path, _: &str) -> Result<()> {
        Ok(())
    }
    async fn upload_bytes(&self, _: &str, _: &str, _: Bytes, _: &str) -> Result<()> {
        Ok(())
    }
    async fn delete(&self, _: &str, _: &str) -> Result<()> {
        Ok(())
    }
    async fn exists(&self, _: &str, _: &str) -> Result<bool> {
        Ok(false)
    }
    async fn health_check(&self, _: &str) -> Result<()> {
        Ok(())
    }
}
