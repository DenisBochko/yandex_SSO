package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"yandex-sso/internal/domain/models"
	"yandex-sso/internal/storage"

	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

// Конструктор Storage
func New(storagePath string) (*Storage, error) {
	// Указываем путь до файла БД
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &Storage{db: db}, nil
}

// SaveUser saves user to db
func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	// Простой запрос на добавление пользователя
	stmt, err := s.db.Prepare("INSERT INTO users(email, pass_hash) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}

	// Выполняем запрос, передав параметры
	res, err := stmt.ExecContext(ctx, email, passHash)
	if err != nil {
		var sqliteErr sqlite3.Error

		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", storage.ErrUserExists, err)
		}

		return 0, fmt.Errorf("failed to execute statement: %w", err)
	}

	// Получаем ID созданной записи
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

// User returns user by email
func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	stmt, err := s.db.Prepare("SELECT id, email, pass_hash FROM users WHERE email = ?")
	if err != nil {
		return models.User{}, fmt.Errorf("failed to prepare statement: %w", err)
	}

	row := stmt.QueryRowContext(ctx, email)

	var user models.User
	err = row.Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", storage.ErrUserNotFound, err)
		}

		return models.User{}, fmt.Errorf("failed to scan row: %w", err)
	}

	return user, nil
}

// App returns app by id
func (s *Storage) App(ctx context.Context, id int) (models.App, error) {
    stmt, err := s.db.Prepare("SELECT id, name, secret FROM apps WHERE id = ?")
    if err != nil {
        return models.App{}, fmt.Errorf("failed to prepare statement: %w", err)
    }

    row := stmt.QueryRowContext(ctx, id)

    var app models.App
    err = row.Scan(&app.ID, &app.Name, &app.Secret)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return models.App{}, fmt.Errorf("%s: %w", storage.ErrAppNotFound, err)
        }

        return models.App{}, fmt.Errorf("failed to scan row: %w", err)
    }

    return app, nil
}