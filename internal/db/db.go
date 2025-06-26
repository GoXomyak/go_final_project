package db

import (
	"database/sql"
	"errors"
	"fmt"
	"go_final_project/internal/config"
	"go_final_project/internal/dto"
	"go_final_project/internal/utils"
	"net/http"
	"os"
	"path/filepath"
	"time"

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

func Connect(cfg *config.Config) (*sql.DB, error) {
	absPath, err := filepath.Abs(cfg.DBPath)
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

	if err = initSchema(db, cfg.SchemaPath); err != nil {
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

func GetTasks(db *sql.DB, search string, limit int) ([]dto.Task, error) {
	tasks := make([]dto.Task, 0)
	var rows *sql.Rows
	var err error
	if search == "" {
		rows, err = db.Query(`SELECT * FROM scheduler
ORDER BY date
LIMIT :limit`, sql.Named("limit", limit))
	} else if date, parseErr := time.Parse(utils.DateSearchFormat, search); parseErr == nil {
		formatted := date.Format(utils.DateFormat)
		rows, err = db.Query(`SELECT * FROM scheduler 
WHERE date = :formatted
LIMIT :limit`, sql.Named("formatted", formatted), sql.Named("limit", limit))
	} else {
		pattern := "%" + search + "%"
		rows, err = db.Query(`SELECT * FROM scheduler
WHERE title LIKE (:pattern) OR comment LIKE (:pattern)
ORDER BY date
LIMIT :limit`, sql.Named("pattern", pattern), sql.Named("limit", limit))
	}

	if err != nil {
		return []dto.Task{}, errors.New("Ошибка получения задач из базы данных: " + err.Error())
	}
	defer func() {
		if err = rows.Close(); err != nil {
			fmt.Printf("Ошибка закрытия rows: %v", err)
		}
	}()

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

func TaskExists(db *sql.DB, id string) error {
	var exists int

	err := db.QueryRow(`SELECT 1 FROM scheduler WHERE id = :id`, sql.Named("id", id)).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("задача с таким ID не найдена")
		}
		return errors.New("Ошибка чтения бд: " + err.Error())
	}

	return nil
}

func UpdateTask(db *sql.DB, task dto.Task) error {
	_, err := db.Exec(`UPDATE scheduler 
SET date = :date, title = :title, comment = :comment, repeat = :repeat
WHERE id = :id`, sql.Named("id", task.ID), sql.Named("date", task.Date), sql.Named("title", task.Title), sql.Named("comment", task.Comment), sql.Named("repeat", task.Repeat))
	if err != nil {
		return errors.New("ошибка обновления бд" + err.Error())
	}
	return nil
}

func GetTask(r *http.Request, database *sql.DB) (dto.Task, error) {
	var task dto.Task
	id := r.URL.Query().Get("id")
	if id == "" {
		return task, errors.New("ID не указан")

	}
	err := TaskExists(database, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return task, err
		}
		return task, err
	}

	row := database.QueryRow(`SELECT * FROM scheduler WHERE id = :id`, sql.Named("id", id))
	err = row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return dto.Task{}, errors.New("ошибка получения задачи из бд: " + err.Error())
	}
	return task, nil
}

func DeleteTask(db *sql.DB, id string) error {
	_, err := db.Exec("DELETE FROM scheduler WHERE id = :id", sql.Named("id", id))
	if err != nil {
		return errors.New("Ошибка удаления задачи из бд: " + err.Error())
	}
	return nil
}
