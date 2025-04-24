# yandex_SSO

## При локальном запуске указываем путь к конфигу через флаг (либо прокидываем переменную окружения CONFIG_PATH=./path/to/config.yaml)

``` go
go run cmd/sso/main.go --config=./config/config.yaml
```

## Запуск через docker

``` bash
docker-compose up
```
