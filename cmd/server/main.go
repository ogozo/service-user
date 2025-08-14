package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"

	"net/http"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	pb "github.com/ogozo/proto-definitions/gen/go/user"
	"github.com/ogozo/service-user/internal/config"
	"github.com/ogozo/service-user/internal/observability"
	"github.com/ogozo/service-user/internal/user"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
)

func startMetricsServer(port string) {
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		log.Printf("Metrics server listening on port %s", port)
		if err := http.ListenAndServe(port, mux); err != nil {
			log.Fatalf("failed to start metrics server: %v", err)
		}
	}()
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	config.LoadConfig()
	cfg := config.AppConfig

	shutdown, err := observability.InitTracerProvider(ctx, cfg.OtelServiceName, cfg.OtelExporterEndpoint)
	if err != nil {
		log.Fatalf("failed to initialize tracer provider: %v", err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown tracer provider: %v", err)
		}
	}()

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to parse pgx config: %v", err)
	}

	poolConfig.ConnConfig.Tracer = otelpgx.NewTracer()

	dbpool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbpool.Close()

	if err := otelpgx.RecordStats(dbpool, otelpgx.WithStatsMeterProvider(otel.GetMeterProvider())); err != nil {
		log.Printf("WARN: unable to record database stats: %v", err)
	}

	log.Println("Database connection successful for user service, with OTel instrumentation.")
	startMetricsServer(cfg.MetricsPort)

	userRepo := user.NewRepository(dbpool)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService, cfg.JWTSecretKey)

	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	pb.RegisterUserServiceServer(s, userHandler)

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
