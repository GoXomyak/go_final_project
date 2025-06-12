package main

import (
	"fmt"
	"go_final_project/internal/config"
	"go_final_project/internal/server"
	"log"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		panic("Ошибка загрузки конфигурации: " + err.Error())
	}
	s := server.New(cfg)
	fmt.Println("сервер запущен на порту:", cfg.Port)
	log.Fatal(s.Start())
}
