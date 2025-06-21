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

func New(cfg *config.Config) *Server {
	// Создает новый HTTP сервер с заданным портом и директорией для статических файлов
	r := chi.NewRouter()
	return &Server{
		cfg:    cfg,
		Router: r,
	}
}

func (s *Server) Start() error {
	// Запускает HTTP сервер на заданном порту
	s.Router.Handle("/*", handlers.StaticHandler(s.cfg))
	return http.ListenAndServe(":"+s.cfg.Port, s.Router)
}
