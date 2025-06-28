package handlers

import (
	"go_final_project/internal/config"
	"net/http"
	"os"
	"path/filepath"
)

// StaticHandler возвращает HTTP-обработчик для раздачи статических файлов из настроенной веб-директории.
// Если запрашиваемый файл не существует, возвращает файл "index.html" в качестве запасного варианта.
func StaticHandler(cfg *config.Config) http.HandlerFunc {
	// Возвращает обработчик для статических файлов
	fs := http.FileServer(http.Dir(cfg.WebDir))

	return func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(cfg.WebDir, r.URL.Path)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			// Если файл не существует - отдать index.html
			http.ServeFile(w, r, filepath.Join(cfg.WebDir, "index.html"))
			return
		}
		fs.ServeHTTP(w, r)
	}
}
