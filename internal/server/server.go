// Package server реализует HTTP-сервер для планировщика задач.
//
// Предоставляет API для:
// - создания новых задач
// - получения списка задач
// - обновления существующих задач
// - удаления задач
// - отметки задач как выполненных
//
// Сервер поддерживает обработку правил повторения задач и
// взаимодействует с базой данных для постоянного хранения информации.
package server

import (
	"go_final_project/internal/config"
	"go_final_project/internal/handlers"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	cfg    *config.Config
	Router *chi.Mux
}

// New создает новый экземпляр HTTP сервера с конфигурацией и маршрутизатором chi.
func New(cfg *config.Config) *Server {
	// Создает новый HTTP сервер с заданным портом и директорией для статических файлов
	r := chi.NewRouter()
	return &Server{
		cfg:    cfg,
		Router: r,
	}
}

// Start запускает HTTP сервер на порту, указанном в конфигурации, с настроенным маршрутизатором.
func (s *Server) Start() error {
	// Запускает HTTP сервер на заданном порту
	s.Router.Handle("/*", handlers.StaticHandler(s.cfg))
	return http.ListenAndServe(":"+s.cfg.Port, s.Router)
}
