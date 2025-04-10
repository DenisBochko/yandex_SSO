# yandex_SSO

**При локальном запуске указываем путь к конфигу через флаг (либо прокидываем переменную окружения CONFIG_PATH=./path/to/config.yaml)**
``` go
go run cmd/sso/main.go --config=./config/config.yaml
```

```
sso
├── cmd.............. Команды для запуска приложения и утилит
│	└── sso......... Основная точка входа в сервис SSO
├── config........... Конфигурационные yaml-файлы
├── internal......... Внутренности проекта
│	├── app.......... Код для запуска различных компонентов приложения
│	│	└── grpc.... Запуск gRPC-сервера
│	├── config....... Загрузка конфигурации
│	├── domain
│	│	└── models.. Структуры данных и модели домена
│	├── grpcHandlers
│	│	└── auth.... gRPC-хэндлеры сервиса Auth
│	├── lib.......... Общие вспомогательные утилиты и функции
│	├── services..... Сервисный слой (бизнес-логика)
│	│	├── auth
│	│	└── permissions
│	└── storage...... Слой хранения данных
│	    └── postgres.. Реализация PostgreSQL
├── pkg .............. Общие реализации различных инструментов (Postgres, Redis, kafka)
└── db
    └──migrations....... Миграции для базы данных
```