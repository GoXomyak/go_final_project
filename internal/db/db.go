package db

import (
	"database/sql"
	"errors"
	"fmt"
	"go_final_project/internal/dto"
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

func AddTask(date, title, comment, repeat string, db *sql.DB) (sql.Result, error) {
	return db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)",
		sql.Named("date", date),
		sql.Named("title", title),
		sql.Named("comment", comment),
		sql.Named("repeat", repeat))
}

func GetTasks(db *sql.DB, limit int) ([]dto.Task, error) {
	rows, err := db.Query(`SELECT id, date, title, comment, repeat FROM scheduler
ORDER BY date
LIMIT :limit`, sql.Named("limit", limit))
	if err != nil {
		return []dto.Task{}, errors.New("Ошибка получения задач из базы данных: " + err.Error())
	}
	defer func() {
		if err = rows.Close(); err != nil {
			fmt.Printf("Ошибка закрытия rows: %v", err)
		}
	}()
	tasks := make([]dto.Task, 0)
	for rows.Next() {
		var task dto.Task
		if err = rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return []dto.Task{}, errors.New("Ошибка сканирования задачи: " + err.Error())
		}
		tasks = append(tasks, task)
	}
	if err = rows.Err(); err != nil {
		return []dto.Task{}, errors.New("Ошибка сканирования задачи: " + err.Error())
	}
	return tasks, nil
}
