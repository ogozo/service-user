// service-user/cmd/server/main.go
package main

import (
	"context"
	"log"
	"net"

	"github.com/jackc/pgx/v4/pgxpool"
	pb "github.com/ogozo/proto-definitions/gen/go/user"
	"github.com/ogozo/service-user/internal/config"
	"github.com/ogozo/service-user/internal/user"
	"google.golang.org/grpc"
)

func main() {
	config.LoadConfig()
	cfg := config.AppConfig

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
	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, userHandler)

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
