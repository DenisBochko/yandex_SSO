package main

import (
	"fmt"
	"yandex-sso/internal/config"
)

func main() {
	cfg := config.MustLoad()

	fmt.Println(cfg)
}
