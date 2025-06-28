// Package config предоставляет функционал для управления конфигурацией приложения.
//
// Пакет обеспечивает:
// - Загрузку конфигурационных параметров
// - Валидацию настроек
// - Доступ к конфигурационным данным
//
// Основные параметры конфигурации:
// - Путь к файлу базы данных SQLite
// - Путь к файлу схемы базы данных
// - Настройки HTTP-сервера
//
// Конфигурация может быть загружена из различных источников
// (файлы, переменные окружения) и валидируется перед использованием.
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
	Password   string
	SecretJWT  string
}

// Load загружает конфигурацию приложения из переменных окружения или указанного .env файла.
// При отсутствии пути к файлу по умолчанию используется ".env".
//
// Загружаемые параметры:
// - TODO_PORT: порт HTTP-сервера (по умолчанию 7540)
// - WEB_DIR: путь к директории с веб-файлами (по умолчанию "./web")
// - TODO_DBFILE: путь к файлу SQLite БД (по умолчанию "./internal/data/scheduler.db")
// - TODO_SCHEMA_PATH: путь к файлу схемы БД (по умолчанию "./internal/db/schema.sql")
// - TODO_PASSWORD: пароль для аутентификации (если пустой, аутентификация отключена)
// - TODO_JWT_SECRET: секретный ключ для подписи JWT токенов
func Load(path string) (*Config, error) {
	if path == "" {
		path = ".env"
	}

	if os.Getenv("RUNNING_IN_DOCKER") != "true" {
		err := godotenv.Load(path)
		if err != nil {
			return nil, err
		}
	}

	cfg := &Config{
		Port:       getEnv("TODO_PORT", "7540"),
		WebDir:     getEnv("WEB_DIR", "./web"),
		DBPath:     getEnv("TODO_DBFILE", "./internal/data/scheduler.db"),
		SchemaPath: getEnv("TODO_SCHEMA_PATH", "./internal/db/schema.sql"),
		Password:   getEnv("TODO_PASSWORD", ""),
		SecretJWT:  getEnv("TODO_JWT_SECRET", "lineForSignature"),
	}

	return cfg, nil
}

// getEnv получает значение переменной окружения по указанному ключу.
// Если переменная не установлена, возвращает значение по умолчанию (fallback).
func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
