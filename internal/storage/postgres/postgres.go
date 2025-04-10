package postgresql

import (
	"context"
	"fmt"
	"yandex-sso/internal/domain/models"
	"yandex-sso/internal/storage"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

// Конструктор Storage
func New(db *pgxpool.Pool) *Storage {
	return &Storage{db: db}
}

// SaveUser сохраняет пользователя в БД
func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	var id int64

	err := s.db.QueryRow(ctx, "INSERT INTO users(email, pass_hash) VALUES($1, $2) RETURNING id", email, passHash).Scan(&id)
	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case "23505":
			return -1, storage.ErrUserExists
		default:
			return -1, fmt.Errorf("failed to get user: %w", err)
		}
	}
	
	return id, nil
}

// User returns user by email
func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	var user models.User

	err := s.db.QueryRow(ctx, "SELECT id, email, pass_hash FROM users WHERE email = $1",
		email).Scan(&user.ID, &user.Email, &user.PassHash)

	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case "23505":
			fmt.Println("User already exists")
			return models.User{}, storage.ErrUserExists
		default:
			return models.User{}, fmt.Errorf("failed to get user: %w", err)
		}
	}

	return user, nil
}

func (s *Storage) App(ctx context.Context, id int) (models.App, error) {
	return models.App{}, fmt.Errorf("not implemented")
}
