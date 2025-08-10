package user

import (
	"context"
	"time"

	pb "github.com/ogozo/proto-definitions/gen/go/user"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var jwtSecret = []byte("super-secret-key")

// Handler, gRPC isteklerini yönetir.
type Handler struct {
	pb.UnimplementedUserServiceServer
	service *Service
}

// NewHandler, yeni bir handler örneği oluşturur.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Register, gRPC Register isteğini işler.
func (h *Handler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, err := h.service.Register(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not register user: %v", err)
	}

	return &pb.RegisterResponse{User: &pb.User{Id: user.ID, Email: user.Email, Role: user.Role}}, nil
}

// Login, gRPC Login isteğini işler ve JWT döndürür.
func (h *Handler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := h.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "login failed: %v", err)
	}

	// JWT oluşturma
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not create token: %v", err)
	}

	return &pb.LoginResponse{AccessToken: tokenString}, nil
}