package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"github.com/DenisBochko/yandex_SSO/internal/domain/models"
	"github.com/DenisBochko/yandex_SSO/internal/storage"

	"github.com/jackc/pgx/v5"
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
		case "23505": // Нарушение уникальности
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
		case "22P02": // такого email не существует
			return models.User{}, storage.ErrUserNotFound
		default:
			return models.User{}, fmt.Errorf("failed to get user: %w", err)
		}
	}

	return user, nil
}

// Users возвращает список пользователей по списку id
func (s *Storage) Users(ctx context.Context, ids []string) ([]models.User, error) {
	var users []models.User

	rows, err := s.db.Query(ctx, "SELECT id, name, email, pass_hash, verify, avatar FROM users WHERE id = ANY($1)", ids)
	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case "22P02": // такого id не существует
			return nil, storage.ErrUserNotFound
		default:
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
	}

	defer rows.Close()

	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.PassHash, &user.Verified, &user.Avatar); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		switch err.(*pgconn.PgError).Code {
		case "22P02":
			return nil, storage.ErrUserNotFound
		default:
			return nil, fmt.Errorf("failed to iterate over users: %w", err)
		}
	}

	return users, nil
}

func (s *Storage) UserById(ctx context.Context, id string) (models.User, error) {
	var user models.User

	err := s.db.QueryRow(ctx, `
        SELECT id, name, email, pass_hash, verify, avatar
        FROM users WHERE id = $1
    `, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PassHash,
		&user.Verified,
		&user.Avatar,
	)

	// пользователь не найден
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, storage.ErrUserNotFound
	}

	// некорректный UUID (например, невалидная строка)
	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case "22P02": // invalid_text_representation (например, плохой uuid)
			return models.User{}, storage.ErrUserNotFound
		default:
			return models.User{}, fmt.Errorf("failed to get user: %w", err)
		}
	}

	if err != nil {
		return models.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// Обновляет данные пользователя по id
func (s *Storage) UpdateUser(ctx context.Context, user models.User) (bool, error) {
	_, err := s.db.Exec(ctx, "UPDATE users SET name = $1, email = $2, verify = $3, avatar = $4 WHERE id = $5",
		user.Name,
		user.Email,
		user.Verified,
		user.Avatar,
		user.ID,
	)

	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case "22P02": // такого id не существует
			return false, storage.ErrUserNotFound
		default:
			return false, fmt.Errorf("failed to update user: %w", err)
		}
	}

	return true, nil
}

func (s *Storage) DeleteUser(ctx context.Context, id string) (bool, error) {
	_, err := s.db.Exec(ctx, "DELETE FROM users WHERE id = $1", id)

	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case "22P02": // такого id не существует
			return false, storage.ErrUserNotFound
		default:
			return false, fmt.Errorf("failed to delete user: %w", err)
		}
	}

	return true, nil
}

func (s *Storage) CreateVerificationToken(ctx context.Context, userID string, token string, expiresAt time.Time) (bool, error) {
	_, err := s.db.Exec(ctx, "INSERT INTO verification_tokens(user_id, token, expires_at) VALUES($1, $2, $3)", userID, token, expiresAt)
	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case "23505": // Нарушение уникальности
			return false, storage.ErrTokenExists
		default:
			return false, fmt.Errorf("failed to create verification token: %w", err)
		}
	}

	return true, nil
}

func (s *Storage) VerifyToken(ctx context.Context, token string) (bool, error) {
	var userID string
	var expiresAt time.Time

	err := s.db.QueryRow(ctx, "SELECT user_id, expires_at FROM verification_tokens WHERE token = $1", token).Scan(&userID, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, storage.ErrTokenNotFound
		}
		if pgErr, ok := err.(*pgconn.PgError); ok {
			return false, fmt.Errorf("postgres error on token lookup: %s (%s)", pgErr.Message, pgErr.Code)
		}
		return false, fmt.Errorf("failed to get verification token: %w", err)
	}

	if time.Now().UTC().After(expiresAt) {
		return false, storage.ErrTokenExpired
	}

	_, err = s.db.Exec(ctx, "UPDATE users SET verify = true WHERE id = $1", userID)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "22P02" {
				return false, storage.ErrUserNotFound
			}
			return false, fmt.Errorf("postgres error on user update: %s (%s)", pgErr.Message, pgErr.Code)
		}
		return false, fmt.Errorf("failed to update user: %w", err)
	}

	return true, nil
}
