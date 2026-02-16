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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/streamvault/recommendation-service/internal/batch"
	"github.com/streamvault/recommendation-service/internal/cache"
	"github.com/streamvault/recommendation-service/internal/config"
	"github.com/streamvault/recommendation-service/internal/consumer"
	"github.com/streamvault/recommendation-service/internal/discovery"
	"github.com/streamvault/recommendation-service/internal/engine"
	"github.com/streamvault/recommendation-service/internal/grpcserver"
	"github.com/streamvault/recommendation-service/internal/handler"
	"github.com/streamvault/recommendation-service/internal/metrics"
	"github.com/streamvault/recommendation-service/internal/middleware"
	"github.com/streamvault/recommendation-service/internal/repository"
	"github.com/streamvault/recommendation-service/internal/telemetry"
	recv1 "github.com/streamvault/recommendation-service/proto/recommendation/v1"
)

func main() {
	configPath := flag.String("config", "", "path to config.yaml")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	setupLogging(cfg.Logging, cfg.OTel.ServiceName)
	log.Info().
		Int("http_port", cfg.Server.HTTPPort).
		Int("grpc_port", cfg.Server.GRPCPort).
		Msg("starting recommendation service")

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

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	// Connect to PostgreSQL
	poolConfig, err := pgxpool.ParseConfig(cfg.Postgres.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid postgres URL")
	}
	poolConfig.MaxConns = cfg.Postgres.MaxConns
	poolConfig.MinConns = cfg.Postgres.MinConns

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("postgres connection failed")
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("postgres ping failed")
	}
	log.Info().Msg("connected to postgres")

	// Run migrations
	if cfg.Postgres.RunMigrations {
		if err := repository.RunMigrations(ctx, pool); err != nil {
			log.Fatal().Err(err).Msg("database migration failed")
		}
	}

	// Connect to Redis
	redisOpts, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid redis URL")
	}
	redisClient := redis.NewClient(redisOpts)
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("redis connection failed")
	}
	if err := redisotel.InstrumentTracing(redisClient); err != nil {
		log.Warn().Err(err).Msg("failed to instrument redis with OTel tracing")
	}
	log.Info().Msg("connected to redis")

	// Connect to RabbitMQ
	var rabbitConn *amqp.Connection
	rabbitConn, err = amqp.Dial(cfg.RabbitMQ.URL)
	if err != nil {
		log.Warn().Err(err).Msg("rabbitmq connection failed, consumers disabled")
	} else {
		defer rabbitConn.Close()
		log.Info().Msg("connected to rabbitmq")
	}

	// Create repository, cache, and recommendation engine
	repo := repository.New(pool)
	featureCache := cache.New(redisClient)
	recEngine := engine.New(repo, featureCache)

	// Start RabbitMQ consumers
	var watchCons *consumer.WatchConsumer
	var ratingCons *consumer.RatingConsumer
	var catalogCons *consumer.CatalogConsumer
	if rabbitConn != nil {
		watchCons, err = consumer.NewWatchConsumer(rabbitConn, cfg.RabbitMQ, repo, featureCache)
		if err != nil {
			log.Warn().Err(err).Msg("failed to create watch consumer")
		} else {
			go watchCons.Start(ctx)
		}

		ratingCons, err = consumer.NewRatingConsumer(rabbitConn, cfg.RabbitMQ, repo, featureCache)
		if err != nil {
			log.Warn().Err(err).Msg("failed to create rating consumer")
		} else {
			go ratingCons.Start(ctx)
		}

		catalogCons, err = consumer.NewCatalogConsumer(rabbitConn, cfg.RabbitMQ, repo, featureCache)
		if err != nil {
			log.Warn().Err(err).Msg("failed to create catalog consumer")
		} else {
			go catalogCons.Start(ctx)
		}
	}

	// Start batch scheduler (recompute profiles and similarity every 6 hours)
	batchScheduler := batch.NewScheduler(repo, featureCache, recEngine.ContentBased(), 6*time.Hour)
	go batchScheduler.Start(ctx)

	// Register with Consul
	var consulClient *discovery.ConsulClient
	consulClient, err = discovery.NewConsulClient(cfg.Consul.Address)
	if err != nil {
		log.Warn().Err(err).Msg("consul unavailable, service discovery disabled")
	} else {
		if err := consulClient.Register("recommendation-service", cfg.Server.HTTPPort, cfg.Consul.HealthCheckInterval); err != nil {
			log.Warn().Err(err).Msg("failed to register with consul")
		}
	}

	// Build HTTP handler
	rabbitCheckFn := func() bool {
		return rabbitConn != nil && !rabbitConn.IsClosed()
	}

	startupReady := &atomic.Bool{}
	mux := buildHTTPRouter(pool, redisClient, rabbitCheckFn, startupReady, recEngine)
	httpHandler := applyMiddleware(mux,
		middleware.Recovery(),
		middleware.CorrelationID(),
		metrics.Middleware(),
		middleware.Logging(),
	)

	// Start HTTP server
	httpSrv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.HTTPPort),
		Handler:      otelhttp.NewHandler(httpHandler, "recommendation-service"),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	startupReady.Store(true)
	log.Info().Msg("recommendation service startup complete")

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
	recGRPC := grpcserver.NewRecommendationServer(recEngine)
	recv1.RegisterRecommendationServiceServer(grpcSrv, recGRPC)
	reflection.Register(grpcSrv)

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

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Info().Str("signal", sig.String()).Msg("shutting down gracefully")
	case err := <-httpErrCh:
		log.Error().Err(err).Msg("http server error")
	case err := <-grpcErrCh:
		log.Error().Err(err).Msg("grpc server error")
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	ctxCancel()

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

	// Stop consumers
	if watchCons != nil {
		watchCons.Stop()
	}
	if ratingCons != nil {
		ratingCons.Stop()
	}
	if catalogCons != nil {
		catalogCons.Stop()
	}

	if consulClient != nil {
		if err := consulClient.Deregister(); err != nil {
			log.Error().Err(err).Msg("consul deregistration failed")
		}
	}

	log.Info().Msg("recommendation service stopped")
}

func buildHTTPRouter(pool *pgxpool.Pool, redisClient *redis.Client,
	rabbitCheck func() bool, startupReady *atomic.Bool, recommender engine.Recommender) *http.ServeMux {

	mux := http.NewServeMux()

	healthH := handler.NewHealthHandler(pool, redisClient, rabbitCheck, startupReady)
	mux.HandleFunc("GET /health", healthH.ServeLive)
	mux.HandleFunc("GET /health/live", healthH.ServeLive)
	mux.HandleFunc("GET /health/ready", healthH.ServeReady)
	mux.HandleFunc("GET /health/startup", healthH.ServeStartup)
	mux.Handle("GET /metrics", promhttp.Handler())

	// Register recommendation routes
	handler.RegisterRecommendationRoutes(mux, recommender)

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
