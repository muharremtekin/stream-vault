package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"

	"github.com/streamvault/streaming-service/internal/config"
)

type MinIOStorage struct {
	client    *minio.Client
	rawBucket string
}

func NewMinIOStorage(cfg config.MinIOConfig) (*MinIOStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("creating minio client: %w", err)
	}

	log.Info().Str("endpoint", cfg.Endpoint).Msg("connected to minio")

	return &MinIOStorage{
		client:    client,
		rawBucket: cfg.RawBucket,
	}, nil
}

func (s *MinIOStorage) Upload(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error {
	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}

	_, err := s.client.PutObject(ctx, bucket, key, reader, size, opts)
	if err != nil {
		return fmt.Errorf("uploading object %s/%s: %w", bucket, key, err)
	}

	log.Debug().Str("bucket", bucket).Str("key", key).Int64("size", size).Msg("uploaded object")
	return nil
}

func (s *MinIOStorage) Download(ctx context.Context, bucket, key string) (io.ReadCloser, *ObjectInfo, error) {
	obj, err := s.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("getting object %s/%s: %w", bucket, key, err)
	}

	stat, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, nil, fmt.Errorf("stat object %s/%s: %w", bucket, key, err)
	}

	info := &ObjectInfo{
		Key:          stat.Key,
		Size:         stat.Size,
		ContentType:  stat.ContentType,
		LastModified: stat.LastModified,
	}

	return obj, info, nil
}

func (s *MinIOStorage) List(ctx context.Context, bucket, prefix string) ([]ObjectInfo, error) {
	var objects []ObjectInfo

	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}

	for obj := range s.client.ListObjects(ctx, bucket, opts) {
		if obj.Err != nil {
			return nil, fmt.Errorf("listing objects in %s/%s: %w", bucket, prefix, obj.Err)
		}
		objects = append(objects, ObjectInfo{
			Key:          obj.Key,
			Size:         obj.Size,
			ContentType:  obj.ContentType,
			LastModified: obj.LastModified,
		})
	}

	return objects, nil
}

func (s *MinIOStorage) Delete(ctx context.Context, bucket, key string) error {
	err := s.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("deleting object %s/%s: %w", bucket, key, err)
	}
	return nil
}

func (s *MinIOStorage) Exists(ctx context.Context, bucket, key string) (bool, error) {
	_, err := s.client.StatObject(ctx, bucket, key, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("checking object %s/%s: %w", bucket, key, err)
	}
	return true, nil
}

func (s *MinIOStorage) HealthCheck(ctx context.Context) error {
	_, err := s.client.BucketExists(ctx, s.rawBucket)
	if err != nil {
		return fmt.Errorf("minio health check: %w", err)
	}
	return nil
}
