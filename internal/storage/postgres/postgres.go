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
func (s *Storage) SaveUser(ctx context.Context, name string, email string, passHash []byte) (string, error) {
	var id string

	err := s.db.QueryRow(ctx, "INSERT INTO users(name, email, pass_hash) VALUES($1, $2, $3) RETURNING id", name, email, passHash).Scan(&id)
	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case "23505":
			return "", storage.ErrUserExists
		default:
			return "", fmt.Errorf("failed to get user: %w", err)
		}
	}

	return id, nil
}

// User возвращает пользователя по email
func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	var user models.User

	err := s.db.QueryRow(ctx, "SELECT id, name, email, pass_hash, verify, avatar FROM users WHERE email = $1",
		email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PassHash,
		&user.Verified,
		&user.Avatar,
	)

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

// Users возвращает пользователей по списку id
func (s *Storage) Users(ctx context.Context, ids []string) ([]models.User, error) {
	var users []models.User

	rows, err := s.db.Query(ctx, "SELECT id, name, email, pass_hash, verify, avatar FROM users WHERE id = ANY($1)", ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		// НЕ ДОБАВЛЯЕМ АВАТАР
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.PassHash, &user.Verified, &user.Avatar); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over users: %w", err)
	}

	return users, nil
}
