package main

import (
	"errors"
	"flag"
	"fmt"

	// Библиотека для миграций
	"github.com/golang-migrate/migrate/v4"
	// Драйвер для выполнения миграций SQLite 3
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	// Драйвер для получения миграций из файлов
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// go run ./cmd/migrator --storage-path=./storage/sso.db --migrations-path=./migrations

func main() {
	var storagePath, migrationsPath, migrationsTable string

	// Получаем необходимые значения из флагов запуска
	flag.StringVar(&storagePath, "storage-path", "", "path to storage")

	// Путь до папки с миграциями
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")

	// Таблица, в которой будет храниться информация о миграциях. Она нужна
	// для того, чтобы понимать, какие миграции уже применены, а какие нет
	// Дефолтное значение — 'migrations'
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")

	// Выполняем парсинг флагов
	flag.Parse()

	// Валидация параметров
	if storagePath == "" {
		panic("storage-path is required")
	}
	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	// Создаем объект мигратора, передав креды нашей БД
	m, err := migrate.New("file://"+migrationsPath,
		fmt.Sprintf("sqlite3://%s?x-migrations-table=%s", storagePath, migrationsTable),
	)
	if err != nil {
		panic(err)
	}

	// Выполняем миграции до последней версии
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}
		panic(err)
	}

	fmt.Println("migrations applied successfully")
}
