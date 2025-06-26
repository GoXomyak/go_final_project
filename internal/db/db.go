// Package db предоставляет функционал для работы с SQLite базой данных.
//
// Пакет обеспечивает:
// - Подключение к SQLite базе данных
// - Инициализацию схемы базы данных
// - CRUD операции для задач планировщика
// - Проверку существования и валидацию данных
//
// Основные функции:
//
// Подключение к БД:
//   - Connect(*config.Config) - устанавливает соединение с БД
//   - CheckAndCreate(string) - создаёт необходимые директории и файл БД
//
// Работа с задачами:
//   - AddTask(date, title, comment, repeat string, *sql.DB) - добавление новой задачи
//   - GetTasks(*sql.DB, search string, limit int) - получение списка задач
//   - GetTask(*http.Request, *sql.DB) - получение одной задачи по ID
//   - UpdateTask(*sql.DB, dto.Task) - обновление существующей задачи
//   - DeleteTask(*sql.DB, id string) - удаление задачи

// Все операции с базой данных защищены от SQL-инъекций благодаря
// использованию подготовленных выражений и параметризованных запросов.
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

// CheckAndCreate проверяет существование директории и файла по указанному пути dbPath.
// Если они отсутствуют, создаёт их.
//
// Выполняет следующие операции:
// 1. Проверяет существование директории для БД
// 2. Создаёт директорию с правами 0755, если она отсутствует
// 3. Проверяет существование файла БД
// 4. Создаёт пустой файл БД, если он отсутствует
func CheckAndCreate(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		createErr := os.MkdirAll(dir, 0755)
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

// Connect инициализирует подключение к SQLite базе данных используя предоставленную конфигурацию.
//
// Функция выполняет следующие действия:
// 1. Проверяет и создаёт необходимые директории и файл БД
// 2. Устанавливает соединение с БД
// 3. Загружает и применяет схему БД из файла, указанного в конфигурации
// 4. Настраивает параметры соединения
func Connect(cfg *config.Config) (*sql.DB, error) {
	absDBPath, err := filepath.Abs(cfg.DBPath)
	if err != nil {
		return nil, err
	}

	if err = CheckAndCreate(absDBPath); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", absDBPath)
	if err != nil {
		return nil, err
	}
	absSchemaPath, err := filepath.Abs(cfg.SchemaPath)
	if err != nil {
		return nil, err
	}
	if err = initSchema(db, absSchemaPath); err != nil {
		return nil, err
	}

	return db, nil
}

// initSchema инициализирует схему базы данных, выполняя команды SQL из файла схемы.
// Принимает подключение к базе данных и путь к файлу схемы в качестве параметров.
// Возвращает ошибку, если чтение файла схемы или выполнение схемы завершается неудачей.
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

// AddTask вставляет новую задачу в таблицу scheduler с указанной датой, заголовком, комментарием и правилом повторения.
// Возвращает результат выполнения SQL и любую обнаруженную ошибку.
func AddTask(date, title, comment, repeat string, db *sql.DB) (sql.Result, error) {
	return db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)",
		sql.Named("date", date),
		sql.Named("title", title),
		sql.Named("comment", comment),
		sql.Named("repeat", repeat))
}

// GetTasks извлекает задачи из базы данных на основе предоставленного поискового запроса и лимита.
// Он поддерживает поиск по пустому запросу, точной дате или нечеткому соответствию в полях заголовка и комментария.
// Возвращает структуру dto.Task и ошибку, если какая-либо операция не удалась.
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

// TaskExists проверяет существование задачи с указанным ID в базе данных scheduler
// и возвращает ошибку, если задача не найдена или произошла ошибка при выполнении запроса.
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

// UpdateTask обновляет существующую задачу в базе данных scheduler предоставленными данными и возвращает ошибку, если таковая возникает.
func UpdateTask(db *sql.DB, task dto.Task) error {
	_, err := db.Exec(`UPDATE scheduler 
SET date = :date, title = :title, comment = :comment, repeat = :repeat
WHERE id = :id`, sql.Named("id", task.ID), sql.Named("date", task.Date), sql.Named("title", task.Title), sql.Named("comment", task.Comment), sql.Named("repeat", task.Repeat))
	if err != nil {
		return errors.New("ошибка обновления бд" + err.Error())
	}
	return nil
}

// GetTask извлекает задачу из базы данных, используя идентификатор из HTTP-запроса, и возвращает ее или ошибку в случае сбоя.
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

// DeleteTask удаляет задачу из базы данных scheduler по ее ID и возвращает ошибку в случае сбоя.
func DeleteTask(db *sql.DB, id string) error {
	_, err := db.Exec("DELETE FROM scheduler WHERE id = :id", sql.Named("id", id))
	if err != nil {
		return errors.New("Ошибка удаления задачи из бд: " + err.Error())
	}
	return nil
}
