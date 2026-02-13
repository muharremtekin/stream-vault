package main

import (
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/streamvault/streaming-service/internal/config"
	"github.com/streamvault/streaming-service/internal/grpcserver"
	"github.com/streamvault/streaming-service/internal/progress"
	"github.com/streamvault/streaming-service/internal/storage"
	streamingv1 "github.com/streamvault/streaming-service/proto/streaming/v1"
)

func newGRPCServer(store storage.Storage, progressSvc *progress.Service, redisClient *redis.Client, minioCfg config.MinIOConfig) *grpcserver.StreamingServer {
	return grpcserver.NewStreamingServer(store, progressSvc, redisClient, minioCfg)
}

func registerGRPC(srv *grpc.Server, streamingSrv *grpcserver.StreamingServer) {
	streamingv1.RegisterStreamingServiceServer(srv, streamingSrv)
	reflection.Register(srv)
}
