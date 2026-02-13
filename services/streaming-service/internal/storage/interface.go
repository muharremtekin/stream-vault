package storage

import (
	"context"
	"io"
	"time"
)

type ObjectInfo struct {
	Key          string
	Size         int64
	ContentType  string
	LastModified time.Time
}

type Storage interface {
	Upload(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error
	Download(ctx context.Context, bucket, key string) (io.ReadCloser, *ObjectInfo, error)
	List(ctx context.Context, bucket, prefix string) ([]ObjectInfo, error)
	Delete(ctx context.Context, bucket, key string) error
	Exists(ctx context.Context, bucket, key string) (bool, error)
	HealthCheck(ctx context.Context) error
}
