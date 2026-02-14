package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/streamvault/search-service/internal/cache"
	"github.com/streamvault/search-service/internal/config"
	"github.com/streamvault/search-service/internal/consumer"
	"github.com/streamvault/search-service/internal/discovery"
	"github.com/streamvault/search-service/internal/elasticsearch"
	"github.com/streamvault/search-service/internal/grpcserver"
	"github.com/streamvault/search-service/internal/handler"
	"github.com/streamvault/search-service/internal/middleware"
	"github.com/streamvault/search-service/internal/trending"
	searchv1 "github.com/streamvault/search-service/proto/search/v1"
)

func main() {
	configPath := flag.String("config", "", "path to config.yaml")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	setupLogging(cfg.Logging)
	log.Info().
		Int("http_port", cfg.Server.HTTPPort).
		Int("grpc_port", cfg.Server.GRPCPort).
		Msg("starting search service")

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

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
	log.Info().Msg("connected to redis")

	// Connect to RabbitMQ
	rabbitConn, err := amqp.Dial(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("rabbitmq connection failed")
	}
	defer rabbitConn.Close()
	log.Info().Msg("connected to rabbitmq")

	// Connect to Elasticsearch
	esClient, err := elasticsearch.NewClient(cfg.Elasticsearch)
	if err != nil {
		log.Fatal().Err(err).Msg("elasticsearch client creation failed")
	}

	if err := esClient.HealthCheck(ctx); err != nil {
		log.Fatal().Err(err).Msg("elasticsearch health check failed")
	}
	log.Info().Msg("connected to elasticsearch")

	if err := esClient.EnsureIndex(ctx); err != nil {
		log.Warn().Err(err).Msg("failed to verify elasticsearch index")
	}

	// Create search components
	indexer := elasticsearch.NewIndexer(esClient)
	searcher := elasticsearch.NewSearcher(esClient)

	// Create cache and cached searcher
	searchCache := cache.NewCache(redisClient)
	cachedSearcher := cache.NewCachedSearcher(searcher, searchCache)

	// Create trending service
	trendingSvc := trending.NewService(redisClient)
	go trendingSvc.StartRotation(ctx)

	// Start RabbitMQ consumers
	catalogConsumer, err := consumer.NewCatalogConsumer(rabbitConn, cfg.RabbitMQ, indexer, trendingSvc)
	if err != nil {
		log.Warn().Err(err).Msg("failed to create catalog consumer")
	} else {
		go catalogConsumer.Start(ctx)
	}

	watchConsumer, err := consumer.NewWatchConsumer(rabbitConn, cfg.RabbitMQ, indexer, redisClient, trendingSvc)
	if err != nil {
		log.Warn().Err(err).Msg("failed to create watch consumer")
	} else {
		go watchConsumer.Start(ctx)
	}

	// Register with Consul
	var consulClient *discovery.ConsulClient
	consulClient, err = discovery.NewConsulClient(cfg.Consul.Address)
	if err != nil {
		log.Warn().Err(err).Msg("consul unavailable, service discovery disabled")
	} else {
		if err := consulClient.Register("search-service", cfg.Server.HTTPPort, cfg.Consul.HealthCheckInterval); err != nil {
			log.Warn().Err(err).Msg("failed to register with consul")
		}
	}

	// Build HTTP handler
	rabbitCheckFn := func() bool {
		return rabbitConn != nil && !rabbitConn.IsClosed()
	}

	mux := buildHTTPRouter(esClient, redisClient, cachedSearcher, trendingSvc, searchCache, rabbitCheckFn)
	httpHandler := applyMiddleware(mux,
		middleware.Recovery(),
		middleware.CorrelationID(),
		middleware.Logging(),
	)

	// Start HTTP server
	httpSrv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.HTTPPort),
		Handler:      httpHandler,
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
	grpcSrv := grpc.NewServer()
	searchGRPC := grpcserver.NewSearchServer(cachedSearcher, trendingSvc)
	searchv1.RegisterSearchServiceServer(grpcSrv, searchGRPC)
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

	if catalogConsumer != nil {
		catalogConsumer.Stop()
	}
	if watchConsumer != nil {
		watchConsumer.Stop()
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

	log.Info().Msg("search service stopped")
}

func buildHTTPRouter(esClient *elasticsearch.Client, redisClient *redis.Client,
	cachedSearcher handler.Searcher, trendingSvc *trending.Service,
	searchCache *cache.Cache, rabbitCheck func() bool) *http.ServeMux {

	mux := http.NewServeMux()

	// Health
	healthH := handler.NewHealthHandler(esClient, redisClient, rabbitCheck)
	mux.HandleFunc("GET /health", healthH.ServeLive)
	mux.HandleFunc("GET /health/live", healthH.ServeLive)
	mux.HandleFunc("GET /health/ready", healthH.ServeReady)

	// Search
	searchH := handler.NewSearchHandler(cachedSearcher)
	mux.Handle("GET /api/search", searchH)

	// Autocomplete
	autocompleteH := handler.NewAutocompleteHandler(cachedSearcher)
	mux.Handle("GET /api/search/autocomplete", autocompleteH)

	// Trending
	trendingH := handler.NewTrendingHandler(trendingSvc, searchCache)
	mux.Handle("GET /api/search/trending", trendingH)

	return mux
}

func applyMiddleware(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func setupLogging(logCfg config.LoggingConfig) {
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

	if logCfg.Format == "console" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}
