package main

import (
	"fmt"
	"go_final_project/internal/config"
	"go_final_project/internal/db"
	"go_final_project/internal/handlers"
	"go_final_project/internal/server"
	"log"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		panic("Ошибка загрузки конфигурации: " + err.Error())
	}
	database, err := db.Connect(cfg)
	if err != nil {
		panic("Невозможно создать подключение к бд: " + err.Error())
	}
	defer func() {
		if err := database.Close(); err != nil {
			fmt.Printf("Ошибка закрытия базы данных: " + err.Error())
		}
	}()
	s := server.New(cfg)
	handlers.Init(s.Router, database, cfg)
	fmt.Println("сервер запущен на порту:", cfg.Port)
	log.Fatal(s.Start())
}
