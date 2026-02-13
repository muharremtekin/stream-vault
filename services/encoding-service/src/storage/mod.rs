pub mod minio;

use async_trait::async_trait;
use bytes::Bytes;
use std::path::Path;

use crate::error::Result;

#[async_trait]
pub trait StorageClient: Send + Sync {
    /// Download an object from a bucket to a local file. Returns bytes written.
    async fn download_to_file(
        &self,
        bucket: &str,
        key: &str,
        dest: &Path,
    ) -> Result<u64>;

    /// Upload a local file to a bucket.
    async fn upload_from_file(
        &self,
        bucket: &str,
        key: &str,
        file_path: &Path,
        content_type: &str,
    ) -> Result<()>;

    /// Upload raw bytes to a bucket.
    async fn upload_bytes(
        &self,
        bucket: &str,
        key: &str,
        data: Bytes,
        content_type: &str,
    ) -> Result<()>;

    /// Delete an object from a bucket.
    async fn delete(&self, bucket: &str, key: &str) -> Result<()>;

    /// Check if an object exists.
    async fn exists(&self, bucket: &str, key: &str) -> Result<bool>;

    /// Verify connectivity by checking a bucket exists.
    async fn health_check(&self, bucket: &str) -> Result<()>;
}
