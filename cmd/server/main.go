package main

import (
	"context"
	"log"
	"net"

	"github.com/jackc/pgx/v4/pgxpool"
	pb "github.com/ogozo/proto-definitions/gen/go/user"
	"github.com/ogozo/service-user/internal/user"
	"google.golang.org/grpc"
)

const (
	// Bu adresleri daha sonra config'den alacağız
	dbURL    = "postgres://admin:secret@localhost:5432/ecommerce"
	grpcPort = ":50051"
)

func main() {
	// Veritabanı bağlantısı
	dbpool, err := pgxpool.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()
	log.Println("Database connection successful.")

	// Bağımlılıkların enjeksiyonu (Dependency Injection)
	userRepo := user.NewRepository(dbpool)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	// gRPC sunucusunu başlatma
	lis, err := net.Listen("tcp", grpcPort)
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
