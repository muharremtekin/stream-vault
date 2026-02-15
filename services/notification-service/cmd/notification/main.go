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

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/streamvault/notification-service/internal/config"
	"github.com/streamvault/notification-service/internal/consumer"
	"github.com/streamvault/notification-service/internal/discovery"
	"github.com/streamvault/notification-service/internal/dispatcher"
	"github.com/streamvault/notification-service/internal/handler"
	"github.com/streamvault/notification-service/internal/middleware"
	"github.com/streamvault/notification-service/internal/store"
	"github.com/streamvault/notification-service/internal/telemetry"
	"github.com/streamvault/notification-service/internal/template"
	ws "github.com/streamvault/notification-service/internal/websocket"
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
		Msg("starting notification service")

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

	// Connect to MongoDB
	mongoClient, err := mongo.Connect(options.Client().ApplyURI(cfg.MongoDB.URI))
	if err != nil {
		log.Fatal().Err(err).Msg("mongodb connection failed")
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Error().Err(err).Msg("mongodb disconnect error")
		}
	}()

	pingCtx, pingCancel := context.WithTimeout(ctx, 10*time.Second)
	defer pingCancel()
	if err := mongoClient.Ping(pingCtx, nil); err != nil {
		log.Fatal().Err(err).Msg("mongodb ping failed")
	}
	log.Info().Msg("connected to mongodb")

	db := mongoClient.Database(cfg.MongoDB.Database)

	// Create stores and ensure indexes
	notifStore := store.NewMongoNotificationStore(db)
	if err := notifStore.EnsureIndexes(ctx); err != nil {
		log.Warn().Err(err).Msg("failed to ensure notification indexes")
	}

	prefStore := store.NewMongoPreferencesStore(db)
	if err := prefStore.EnsureIndexes(ctx); err != nil {
		log.Warn().Err(err).Msg("failed to ensure preferences indexes")
	}

	// Connect to RabbitMQ (non-blocking)
	var rabbitConn *amqp.Connection
	rabbitConn, err = amqp.Dial(cfg.RabbitMQ.URL)
	if err != nil {
		log.Warn().Err(err).Msg("rabbitmq connection failed, consumers will be disabled")
	} else {
		defer rabbitConn.Close()
		log.Info().Msg("connected to rabbitmq")
	}

	// Register with Consul (non-blocking)
	var consulClient *discovery.ConsulClient
	consulClient, err = discovery.NewConsulClient(cfg.Consul.Address)
	if err != nil {
		log.Warn().Err(err).Msg("consul unavailable, service discovery disabled")
	} else {
		if err := consulClient.Register("notification-service", cfg.Server.HTTPPort, cfg.Consul.HealthCheckInterval); err != nil {
			log.Warn().Err(err).Msg("failed to register with consul")
		}
	}

	// Initialize template engine
	tmplEngine, err := template.New()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize template engine")
	}

	// Initialize WebSocket hub
	hub := ws.NewHub(cfg.WebSocket.MaxConnectionsPerUser)
	go hub.Run()

	// Initialize senders and dispatcher
	emailSender := dispatcher.NewEmailSender("localhost", 1025)
	pushSender := dispatcher.NewPushSender()
	inAppSender := dispatcher.NewInAppSender(hub)
	dispatch := dispatcher.NewDispatcher(notifStore, prefStore, emailSender, pushSender, inAppSender)

	// Start RabbitMQ consumers (only if connected)
	var subConsumer *consumer.SubscriptionConsumer
	var encConsumer *consumer.EncodingConsumer
	var contentConsumer *consumer.ContentConsumer

	if rabbitConn != nil {
		subConsumer, err = consumer.NewSubscriptionConsumer(rabbitConn, cfg.RabbitMQ.SubscriptionQueue, cfg.RabbitMQ.Prefetch, dispatch, tmplEngine)
		if err != nil {
			log.Error().Err(err).Msg("failed to create subscription consumer")
		} else {
			go subConsumer.Start(ctx)
		}

		encConsumer, err = consumer.NewEncodingConsumer(rabbitConn, cfg.RabbitMQ.EncodingQueue, cfg.RabbitMQ.Prefetch, dispatch, tmplEngine)
		if err != nil {
			log.Error().Err(err).Msg("failed to create encoding consumer")
		} else {
			go encConsumer.Start(ctx)
		}

		contentConsumer, err = consumer.NewContentConsumer(rabbitConn, cfg.RabbitMQ.ContentQueue, cfg.RabbitMQ.Prefetch, dispatch, tmplEngine)
		if err != nil {
			log.Error().Err(err).Msg("failed to create content consumer")
		} else {
			go contentConsumer.Start(ctx)
		}
	}

	// Build HTTP handler
	rabbitCheckFn := func() bool {
		return rabbitConn != nil && !rabbitConn.IsClosed()
	}

	wsHandler := handler.NewWebSocketHandler(hub, cfg.JWT, cfg.WebSocket)

	mux := buildHTTPRouter(mongoClient, rabbitCheckFn, hub, wsHandler, notifStore, prefStore)
	httpHandler := applyMiddleware(mux,
		middleware.Recovery(),
		middleware.CorrelationID(),
		middleware.Logging(),
	)

	// Start HTTP server
	httpSrv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.HTTPPort),
		Handler:      otelhttp.NewHandler(httpHandler, "notification-service"),
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

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Info().Str("signal", sig.String()).Msg("shutting down gracefully")
	case err := <-httpErrCh:
		log.Error().Err(err).Msg("http server error")
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	ctxCancel()

	// Stop consumers
	if subConsumer != nil {
		subConsumer.Stop()
	}
	if encConsumer != nil {
		encConsumer.Stop()
	}
	if contentConsumer != nil {
		contentConsumer.Stop()
	}

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("forced http shutdown")
	}

	if consulClient != nil {
		if err := consulClient.Deregister(); err != nil {
			log.Error().Err(err).Msg("consul deregistration failed")
		}
	}

	log.Info().Msg("notification service stopped")
}

func buildHTTPRouter(
	mongoClient *mongo.Client,
	rabbitCheck func() bool,
	hub *ws.Hub,
	wsHandler *handler.WebSocketHandler,
	notifStore store.NotificationStore,
	prefStore store.PreferencesStore,
) *http.ServeMux {
	mux := http.NewServeMux()

	healthH := handler.NewHealthHandler(mongoClient, rabbitCheck)
	mux.HandleFunc("GET /health", healthH.ServeLive)
	mux.HandleFunc("GET /health/live", healthH.ServeLive)
	mux.HandleFunc("GET /health/ready", healthH.ServeReady)

	// WebSocket endpoint
	mux.HandleFunc("GET /ws/notifications", wsHandler.ServeWS)

	// Notification API endpoints
	handler.RegisterNotificationRoutes(mux, notifStore, prefStore)

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
