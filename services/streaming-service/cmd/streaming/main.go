package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/grpc"

	"github.com/streamvault/streaming-service/internal/config"
	"github.com/streamvault/streaming-service/internal/metrics"
	"github.com/streamvault/streaming-service/internal/discovery"
	"github.com/streamvault/streaming-service/internal/handler"
	"github.com/streamvault/streaming-service/internal/messaging"
	"github.com/streamvault/streaming-service/internal/middleware"
	"github.com/streamvault/streaming-service/internal/progress"
	"github.com/streamvault/streaming-service/internal/storage"
	"github.com/streamvault/streaming-service/internal/telemetry"
)

func main() {
	configPath := flag.String("config", "", "path to config.yaml")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Configure logging
	setupLogging(cfg.Logging, cfg.OTel.ServiceName)
	log.Info().
		Int("http_port", cfg.Server.HTTPPort).
		Int("grpc_port", cfg.Server.GRPCPort).
		Msg("starting streaming service")

	// OpenTelemetry
	otelShutdown, err := telemetry.InitTracer(context.Background(), telemetry.OTelConfig{
		Enabled:     cfg.OTel.Enabled,
		Endpoint:    cfg.OTel.Endpoint,
		ServiceName: cfg.OTel.ServiceName,
		Insecure:    cfg.OTel.Insecure,
	})
	if err != nil {
		log.Warn().Err(err).Msg("failed to initialize OpenTelemetry, tracing disabled")
	} else {
		defer func() {
			if err := otelShutdown(context.Background()); err != nil {
				log.Error().Err(err).Msg("error shutting down OTel tracer provider")
			}
		}()
		log.Info().Str("endpoint", cfg.OTel.Endpoint).Msg("OpenTelemetry tracing initialized")
	}

	// Connect to Redis
	redisOpts, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid redis URL")
	}
	redisClient := redis.NewClient(redisOpts)
	defer redisClient.Close()

	if err := redisotel.InstrumentTracing(redisClient); err != nil {
		log.Warn().Err(err).Msg("failed to instrument redis with OTel tracing")
	}
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatal().Err(err).Msg("redis connection failed")
	}
	log.Info().Msg("connected to redis")

	// Connect to MinIO
	minioStore, err := storage.NewMinIOStorage(cfg.MinIO)
	if err != nil {
		log.Fatal().Err(err).Msg("minio connection failed")
	}
	log.Info().Msg("connected to minio")

	// Connect to RabbitMQ
	rabbitConn, err := amqp.Dial(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("rabbitmq connection failed")
	}
	defer rabbitConn.Close()
	log.Info().Msg("connected to rabbitmq")

	// Create publisher
	publisher, err := messaging.NewPublisher(rabbitConn, cfg.RabbitMQ)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create rabbitmq publisher")
	}
	defer publisher.Close()

	// Create progress components
	progressRepo := progress.NewRedisRepository(redisClient, cfg.Redis.ProgressTTL)
	watchPub := progress.NewWatchCompletedPublisher(publisher, redisClient, 24*time.Hour)
	progressSvc := progress.NewService(progressRepo, watchPub)

	// Start RabbitMQ consumer
	consumer, err := messaging.NewConsumer(rabbitConn, cfg.RabbitMQ, redisClient)
	if err != nil {
		log.Warn().Err(err).Msg("failed to create rabbitmq consumer, encoding results won't be consumed")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if consumer != nil {
		go consumer.Start(ctx)
	}

	// Register with Consul
	consulClient, err := discovery.NewConsulClient(cfg.Consul.Address)
	if err != nil {
		log.Warn().Err(err).Msg("consul unavailable, service discovery disabled")
	} else {
		if err := consulClient.Register("streaming-service", cfg.Server.HTTPPort, cfg.Consul.HealthCheckInterval); err != nil {
			log.Warn().Err(err).Msg("failed to register with consul")
		}
	}

	// Startup tracking
	var startupReady atomic.Bool

	// Build HTTP handler
	rabbitCheckFn := func() bool {
		return rabbitConn != nil && !rabbitConn.IsClosed()
	}

	mux := buildHTTPRouter(cfg, minioStore, publisher, progressSvc, redisClient, rabbitCheckFn, &startupReady)
	httpHandler := applyMiddleware(mux,
		middleware.Recovery(),
		middleware.CorrelationID(),
		metrics.Middleware(),
		middleware.Logging(),
	)

	// Start HTTP server
	httpSrv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.HTTPPort),
		Handler:      otelhttp.NewHandler(httpHandler, "streaming-service"),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	httpErrCh := make(chan error, 1)
	go func() {
		log.Info().Str("addr", httpSrv.Addr).Msg("http server listening")
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			httpErrCh <- err
		}
		close(httpErrCh)
	}()

	// Start gRPC server
	grpcSrv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(metrics.GRPCUnaryInterceptor()),
	)
	streamingSrv := newGRPCServer(minioStore, progressSvc, redisClient, cfg.MinIO)
	registerGRPC(grpcSrv, streamingSrv)

	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		log.Fatal().Err(err).Int("port", cfg.Server.GRPCPort).Msg("failed to listen for grpc")
	}

	grpcErrCh := make(chan error, 1)
	go func() {
		log.Info().Int("port", cfg.Server.GRPCPort).Msg("grpc server listening")
		if err := grpcSrv.Serve(grpcLis); err != nil {
			grpcErrCh <- err
		}
		close(grpcErrCh)
	}()

	startupReady.Store(true)
	log.Info().Msg("startup complete, all systems ready")

	// Wait for shutdown signal
	select {
	case <-ctx.Done():
		log.Info().Msg("shutdown signal received, shutting down gracefully")
	case err := <-httpErrCh:
		stop()
		log.Error().Err(err).Msg("http server error")
	case err := <-grpcErrCh:
		stop()
		log.Error().Err(err).Msg("grpc server error")
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if consumer != nil {
		consumer.Stop()
	}

	grpcDone := make(chan struct{})
	go func() {
		grpcSrv.GracefulStop()
		close(grpcDone)
	}()
	select {
	case <-grpcDone:
		log.Info().Msg("grpc server stopped gracefully")
	case <-shutdownCtx.Done():
		log.Warn().Msg("grpc graceful stop timed out, forcing stop")
		grpcSrv.Stop()
	}

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("forced http shutdown")
	}

	if consulClient != nil {
		if err := consulClient.Deregister(); err != nil {
			log.Error().Err(err).Msg("consul deregistration failed")
		}
	}

	log.Info().Msg("streaming service stopped")
}

func buildHTTPRouter(cfg *config.Config, store storage.Storage, pub messaging.Publisher,
	progressSvc *progress.Service, redisClient *redis.Client, rabbitCheck func() bool,
	startupReady *atomic.Bool) *http.ServeMux {

	mux := http.NewServeMux()

	// Health
	healthH := handler.NewHealthHandler(store, redisClient, rabbitCheck, startupReady)
	mux.HandleFunc("GET /health", healthH.ServeDetailed)
	mux.HandleFunc("GET /health/live", healthH.ServeLive)
	mux.HandleFunc("GET /health/ready", healthH.ServeReady)
	mux.HandleFunc("GET /health/startup", healthH.ServeStartup)
	mux.Handle("GET /metrics", promhttp.Handler())

	// Upload (admin only)
	uploadH := handler.NewUploadHandler(store, pub, cfg.Upload, cfg.MinIO)
	mux.Handle("POST /api/stream/upload", middleware.Admin()(uploadH))

	// HLS Streaming with subscription + concurrent middleware
	manifestH := handler.NewManifestHandler(store, cfg.MinIO)
	segmentH := handler.NewSegmentHandler(store, cfg.MinIO)

	streamingMW := func(h http.Handler) http.Handler {
		return middleware.Subscription()(
			middleware.ConcurrentStreams(redisClient, cfg.Redis.ConcurrentTTL)(h))
	}

	mux.Handle("GET /stream/{contentId}/manifest.m3u8",
		streamingMW(http.HandlerFunc(manifestH.MasterPlaylist)))
	mux.Handle("GET /stream/{contentId}/{quality}/playlist.m3u8",
		streamingMW(http.HandlerFunc(manifestH.MediaPlaylist)))
	mux.Handle("GET /stream/{contentId}/{quality}/{segment}",
		streamingMW(http.HandlerFunc(segmentH.ServeSegment)))

	// Progress
	progressH := handler.NewProgressHandler(progressSvc)
	mux.HandleFunc("GET /api/stream/continue-watching", progressH.ContinueWatching)
	mux.HandleFunc("GET /api/stream/{contentId}/progress", progressH.GetProgress)
	mux.HandleFunc("POST /api/stream/{contentId}/progress", progressH.SaveProgress)

	return mux
}

func applyMiddleware(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func setupLogging(logCfg config.LoggingConfig, serviceName string) {
	switch logCfg.Level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if logCfg.Format == "console" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).
			With().Str("service", serviceName).Logger()
	} else {
		log.Logger = log.With().Str("service", serviceName).Logger()
	}
}
