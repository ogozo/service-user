package user

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Repository, veritabanı işlemlerini yönetir.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository, yeni bir repository örneği oluşturur.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateUser, veritabanına yeni bir kullanıcı ekler.
func (r *Repository) CreateUser(ctx context.Context, email, hashedPassword string) (string, error) {
	var id string
	query := `INSERT INTO users (email, password, role) VALUES ($1, $2, 'CUSTOMER') RETURNING id`
	err := r.db.QueryRow(ctx, query, email, hashedPassword).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

// GetUserByEmail, email adresine göre kullanıcıyı getirir.
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	query := `SELECT id, email, password, role FROM users WHERE email = $1`
	err := r.db.QueryRow(ctx, query, email).Scan(&u.ID, &u.Email, &u.Password, &u.Role)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// User veri modeli
type User struct {
	ID       string
	Email    string
	Password string
	Role     string
}