package user

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

// Service, kullanıcı ile ilgili iş mantığını içerir.
type Service struct {
	repo *Repository
}

// NewService, yeni bir service örneği oluşturur.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Register, yeni bir kullanıcı kaydeder.
func (s *Service) Register(ctx context.Context, email, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	id, err := s.repo.CreateUser(ctx, email, string(hashedPassword))
	if err != nil {
		return nil, err
	}

	return &User{ID: id, Email: email, Role: "CUSTOMER"}, nil
}

// Login, kullanıcıyı doğrular.
func (s *Service) Login(ctx context.Context, email, password string) (*User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}