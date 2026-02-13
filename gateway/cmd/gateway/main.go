package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/streamvault/gateway/internal/config"
	"github.com/streamvault/gateway/internal/discovery"
	"github.com/streamvault/gateway/internal/health"
	"github.com/streamvault/gateway/internal/middleware"
	"github.com/streamvault/gateway/internal/proxy"
)

func main() {
	configPath := flag.String("config", "", "path to config.yaml")
	flag.Parse()

	// ---- Load Configuration ----
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// ---- Configure Logging ----
	setupLogging(cfg.Logging)
	log.Info().Int("port", cfg.Server.Port).Msg("starting streamvault api gateway")

	// ---- Service Discovery ----
	consulClient, err := discovery.NewConsulClient(cfg.Consul.Address)
	if err != nil {
		log.Warn().Err(err).Msg("consul unavailable, using static service addresses only")
	}

	resolver := discovery.NewResolver(consulClient, cfg.Services)

	// ---- Upstream Proxy ----
	upstreamManager := proxy.NewUpstreamManager()
	routes := proxy.DefaultRoutes()
	router := proxy.NewRouter(routes, resolver, upstreamManager)

	// ---- Health Handler ----
	healthHandler := health.NewHandler(cfg.Services, resolver)

	// ---- Build Top-Level Mux ----
	topMux := http.NewServeMux()
	topMux.Handle("/health", healthHandler)
	topMux.Handle("/", router.Handler())

	// ---- Rate Limiter ----
	var rateLimitMiddleware func(http.Handler) http.Handler
	rateLimiter, err := middleware.NewRateLimiter(
		cfg.RateLimit.RedisURL,
		cfg.RateLimit.RequestsPerSecond,
		cfg.RateLimit.Burst,
	)
	if err != nil {
		log.Warn().Err(err).Msg("redis unavailable, rate limiting disabled")
		rateLimitMiddleware = middleware.NewNoOpRateLimiter()
	} else {
		defer rateLimiter.Close()
		rateLimitMiddleware = rateLimiter.Middleware()
	}

	// ---- Middleware Chain ----
	// Order: recovery -> correlation -> logging -> cors -> ratelimit -> auth -> proxy
	// The outermost middleware executes first.
	handler := applyMiddleware(topMux,
		middleware.Recovery(),
		middleware.CorrelationID(),
		middleware.Logging(),
		middleware.CORS(middleware.DefaultCORSOptions()),
		rateLimitMiddleware,
		middleware.Auth(cfg.JWT.Secret, cfg.JWT.Issuer),
	)

	// ---- HTTP Server ----
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// ---- Graceful Shutdown ----
	errCh := make(chan error, 1)
	go func() {
		log.Info().Str("addr", srv.Addr).Msg("listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Info().Str("signal", sig.String()).Msg("shutting down gracefully")
	case err := <-errCh:
		log.Error().Err(err).Msg("server error")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("forced shutdown")
	}

	// Deregister from Consul if it was connected.
	if consulClient != nil {
		_ = consulClient.DeregisterService("gateway")
	}

	log.Info().Msg("gateway stopped")
}

// applyMiddleware wraps the handler with each middleware in the order provided.
// The first middleware in the slice becomes the outermost layer.
func applyMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	// Apply in reverse so that the first middleware listed wraps the outermost layer.
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// setupLogging configures zerolog based on the logging configuration.
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
	// Default format is JSON which is zerolog's default.

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}
