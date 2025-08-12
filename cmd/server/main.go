package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/jackc/pgx/v4/pgxpool"
	pb "github.com/ogozo/proto-definitions/gen/go/user"
	"github.com/ogozo/service-user/internal/config"
	"github.com/ogozo/service-user/internal/observability"
	"github.com/ogozo/service-user/internal/user"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

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

	dbpool, err := pgxpool.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()
	log.Println("Database connection successful.")

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
