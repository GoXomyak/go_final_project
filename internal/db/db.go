package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func CheckAndCreate(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		createErr := os.Mkdir(dir, 0755)
		if createErr != nil {
			return fmt.Errorf("Ошибка создания директоррии: %s\n", createErr)
		}
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		file, createErr := os.Create(dbPath)
		if createErr != nil {
			return fmt.Errorf("Ошибка создания файла: %s\n", createErr)
		}
		defer func() {
			if err = file.Close(); err != nil {
				fmt.Printf("Ошибка создания файла: %s\n", createErr)
			}
		}()
		fmt.Printf("Файл базы данных создан: %s\n", dbPath)
	}
	return nil
}

func Connect(dbPath string, schemaPath string) (*sql.DB, error) {
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, err
	}

	if err = CheckAndCreate(absPath); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", absPath)
	if err != nil {
		return nil, err
	}

	if err = initSchema(db, schemaPath); err != nil {
		return nil, err
	}

	return db, nil
}

func initSchema(db *sql.DB, schemaPath string) error {
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("Не удалось прочитать schema: %s\n", err)
	}
	_, err = db.Exec(string(schemaBytes))
	if err != nil {
		return fmt.Errorf("Не удалось создать schema: %s\n", err)
	}
	return nil
}
