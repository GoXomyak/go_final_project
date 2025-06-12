package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port   string
	WebDir string
}

func Load(path string) (*Config, error) {
	// Здесь должна быть логика загрузки конфигурации из файла или переменных окружения
	// Например, можно использовать os.Getenv для получения переменных окружения
	if path == "" {
		path = ".env"
	}
	err := godotenv.Load(path)
	if err != nil {
		return nil, err
	}
	return &Config{
		Port:   os.Getenv("TODO_PORT"),
		WebDir: os.Getenv("WEB_DIR"),
	}, nil
}

func GetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("Переменная окружения " + key + " не установлена")
	}
	return value
}
