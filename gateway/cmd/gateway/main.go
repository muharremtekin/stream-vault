package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/streamvault/gateway/internal/config"
	"github.com/streamvault/gateway/internal/discovery"
	"github.com/streamvault/gateway/internal/health"
	"github.com/streamvault/gateway/internal/metrics"
	"github.com/streamvault/gateway/internal/middleware"
	"github.com/streamvault/gateway/internal/proxy"
	"github.com/streamvault/gateway/internal/resilience"
	"github.com/streamvault/gateway/internal/telemetry"
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
	setupLogging(cfg.Logging, cfg.OTel.ServiceName)
	log.Info().Int("port", cfg.Server.Port).Msg("starting streamvault api gateway")

	// ---- OpenTelemetry ----
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

	// ---- Service Discovery ----
	consulClient, err := discovery.NewConsulClient(cfg.Consul.Address)
	if err != nil {
		log.Warn().Err(err).Msg("consul unavailable, using static service addresses only")
	}

	resolver := discovery.NewResolver(consulClient, cfg.Services)

	// ---- Upstream Proxy ----
	upstreamManager := proxy.NewUpstreamManager(cfg.Resilience.Retry)
	routes := proxy.DefaultRoutes()

	// ---- Resilience (Circuit Breaker + Bulkhead + Timeout) ----
	serviceNames := collectServiceNames(routes)
	res := resilience.New(cfg.Resilience, serviceNames)

	router := proxy.NewRouter(routes, resolver, upstreamManager, res)

	// ---- Startup Tracking ----
	var startupReady atomic.Bool

	// Pre-declare rateLimiter so the health check closure can capture it.
	var rateLimiter *middleware.RateLimiter

	// ---- Health Handler ----
	redisCheckFn := func() bool {
		if rateLimiter == nil {
			return false
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return rateLimiter.Ping(ctx) == nil
	}
	healthHandler := health.NewHandler(cfg.Services, resolver, &startupReady, redisCheckFn)

	// ---- Build Top-Level Mux ----
	topMux := http.NewServeMux()
	topMux.Handle("/health", healthHandler)
	topMux.HandleFunc("/health/live", healthHandler.ServeLive)
	topMux.HandleFunc("/health/ready", healthHandler.ServeReady)
	topMux.HandleFunc("/health/startup", healthHandler.ServeStartup)
	topMux.Handle("/metrics", promhttp.Handler())
	topMux.Handle("/", router.Handler())

	// ---- Rate Limiter ----
	var rateLimitMiddleware func(http.Handler) http.Handler
	rl, err := middleware.NewRateLimiter(
		cfg.RateLimit.RedisURL,
		cfg.RateLimit.RequestsPerSecond,
		cfg.RateLimit.Burst,
	)
	if err != nil {
		log.Warn().Err(err).Msg("redis unavailable, rate limiting disabled")
		rateLimitMiddleware = middleware.NewNoOpRateLimiter()
	} else {
		rateLimiter = rl
		defer rateLimiter.Close()
		rateLimitMiddleware = rateLimiter.Middleware()
	}

	// ---- Middleware Chain ----
	// Order: recovery -> correlation -> logging -> cors -> ratelimit -> auth -> proxy
	// The outermost middleware executes first.
	handler := applyMiddleware(topMux,
		middleware.Recovery(),
		middleware.CorrelationID(),
		metrics.Middleware(),
		middleware.Logging(),
		middleware.CORS(middleware.DefaultCORSOptions()),
		rateLimitMiddleware,
		middleware.Auth(cfg.JWT.Secret, cfg.JWT.Issuer),
		middleware.StreamingAuth(),
	)

	// ---- HTTP Server ----
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      otelhttp.NewHandler(handler, "gateway"),
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

	startupReady.Store(true)
	log.Info().Msg("startup complete, all systems ready")

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

// collectServiceNames returns unique service names from the route table.
func collectServiceNames(routes []proxy.Route) []string {
	seen := make(map[string]struct{})
	var names []string
	for _, r := range routes {
		if _, ok := seen[r.ServiceName]; !ok {
			seen[r.ServiceName] = struct{}{}
			names = append(names, r.ServiceName)
		}
	}
	return names
}

// setupLogging configures zerolog based on the logging configuration.
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
