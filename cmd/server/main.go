package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ogozo/proto-definitions/gen/go/user"
	"github.com/ogozo/service-user/internal/config"
	"github.com/ogozo/service-user/internal/healthcheck"
	"github.com/ogozo/service-user/internal/logging"
	"github.com/ogozo/service-user/internal/observability"
	internalUser "github.com/ogozo/service-user/internal/user"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func startMetricsServer(l *zap.Logger, port string) {
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		l.Info("metrics server started", zap.String("port", port))
		if err := http.ListenAndServe(port, mux); err != nil {
			l.Fatal("failed to start metrics server", zap.Error(err))
		}
	}()
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var cfg config.UserConfig
	config.LoadConfig(&cfg)

	logging.Init(cfg.OtelServiceName)
	defer logging.Sync()

	logger := logging.FromContext(ctx)

	shutdown, err := observability.InitTracerProvider(ctx, cfg.OtelServiceName, cfg.OtelExporterEndpoint, logger)
	if err != nil {
		logger.Fatal("failed to initialize tracer provider", zap.Error(err))
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			logger.Fatal("failed to shutdown tracer provider", zap.Error(err))
		}
	}()

	startMetricsServer(logger, cfg.MetricsPort)

	var dbpool *pgxpool.Pool
	healthcheck.ConnectWithRetry(ctx, "PostgreSQL", 5, 2*time.Second, func() error {
		var err error
		poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
		if err != nil {
			return fmt.Errorf("failed to parse pgx config: %w", err)
		}
		poolConfig.ConnConfig.Tracer = otelpgx.NewTracer()

		dbpool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			return err
		}
		return dbpool.Ping(ctx)
	})
	defer dbpool.Close()

	if err := otelpgx.RecordStats(dbpool, otelpgx.WithStatsMeterProvider(otel.GetMeterProvider())); err != nil {
		logger.Error("unable to record database stats", zap.Error(err))
	}

	userRepo := internalUser.NewRepository(dbpool)
	userService := internalUser.NewService(userRepo)
	userHandler := internalUser.NewHandler(userService, cfg.JWTSecretKey)

	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err), zap.String("port", cfg.GRPCPort))
	}

	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	user.RegisterUserServiceServer(s, userHandler)

	logger.Info("gRPC server listening", zap.String("address", lis.Addr().String()))
	if err := s.Serve(lis); err != nil {
		logger.Fatal("failed to serve gRPC", zap.Error(err))
	}
}
