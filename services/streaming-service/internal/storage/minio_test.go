package storage

import (
	"errors"
	"fmt"
	"testing"

	"github.com/minio/minio-go/v7"
)

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "NoSuchKey error",
			err:  minio.ErrorResponse{Code: "NoSuchKey"},
			want: true,
		},
		{
			name: "wrapped NoSuchKey error",
			err:  fmt.Errorf("stat object bucket/key: %w", minio.ErrorResponse{Code: "NoSuchKey"}),
			want: true,
		},
		{
			name: "double wrapped NoSuchKey error",
			err:  fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", minio.ErrorResponse{Code: "NoSuchKey"})),
			want: true,
		},
		{
			name: "different MinIO error code",
			err:  minio.ErrorResponse{Code: "AccessDenied"},
			want: false,
		},
		{
			name: "wrapped different MinIO error",
			err:  fmt.Errorf("getting object: %w", minio.ErrorResponse{Code: "InternalError"}),
			want: false,
		},
		{
			name: "generic error",
			err:  errors.New("connection refused"),
			want: false,
		},
		{
			name: "wrapped generic error",
			err:  fmt.Errorf("stat object: %w", errors.New("timeout")),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNotFound(tt.err)
			if got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}
