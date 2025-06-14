package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	WebDir     string
	DBPath     string
	SchemaPath string
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
	cfg := &Config{
		Port:       os.Getenv("TODO_PORT"),
		WebDir:     os.Getenv("WEB_DIR"),
		DBPath:     os.Getenv("TODO_DBFILE"),
		SchemaPath: os.Getenv("TODO_SCHEMA_PATH"),
	}
	if cfg.Port == "" {
		cfg.Port = "7540"
	}
	if cfg.DBPath == "" {
		cfg.DBPath = "./internal/data/scheduler.db"
	}
	if cfg.WebDir == "" {
		cfg.WebDir = "./web"
	}
	if cfg.SchemaPath == "" {
		cfg.SchemaPath = "./internal/db/schema.sql"
	}

	return cfg, nil
}

func GetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("Переменная окружения " + key + " не установлена")
	}
	return value
}
