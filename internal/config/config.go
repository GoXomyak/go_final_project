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

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
