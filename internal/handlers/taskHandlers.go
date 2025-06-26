package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go_final_project/internal/config"
	"go_final_project/internal/db"
	"go_final_project/internal/dto"
	"go_final_project/internal/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// Init настраивает маршрутизацию для веб-сервера. Функция регистрирует все необходимые
// маршруты API и их обработчики, используя переданные параметры:
// - маршрутизатор для регистрации endpoint'ов
// - подключение к базе данных для работы с данными
// - конфигурацию приложения для настройки обработчиков
func Init(r chi.Router, db *sql.DB, cfg *config.Config) {
	r.Get("/api/nextdate", nextDayHandler)
	r.Post("/api/task", auth(addTaskHandler(db), cfg))
	r.Get("/api/tasks", auth(getTasksHandler(db), cfg))
	r.Get("/api/task", auth(getTaskHandler(db), cfg))
	r.Put("/api/task", auth(updateTaskHandler(db), cfg))
	r.Post("/api/task/done", auth(taskDoneHandler(db), cfg))
	r.Delete("/api/task", auth(deleteTaskHandler(db), cfg))
	r.Post("/api/signin", authHandler(cfg))
}

// nextDayHandler обрабатывает запрос на вычисление следующей даты на основе правила повторения.
// Извлекает параметры запроса 'date', 'now' и 'repeat' для выполнения расчета.
// Возвращает следующую дату в формате "ГГГГММДД" или сообщение об ошибке при неверных входных данных.
func nextDayHandler(w http.ResponseWriter, r *http.Request) {
	originalDate := r.URL.Query().Get("date")
	nowTemp := r.URL.Query().Get("now")
	repeat := r.URL.Query().Get("repeat")

	now, err := time.Parse(utils.DateFormat, nowTemp)
	if err != nil {
		http.Error(w, "Неверный параметр now: "+err.Error(), http.StatusBadRequest)
		return
	}
	date, err := utils.NextDate(now, originalDate, repeat)
	if err != nil {
		http.Error(w, "Неверный параметр date: "+err.Error(), http.StatusBadRequest)
		return
	}
	s := fmt.Sprintf("%s", date)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(s))
	if err != nil {
		fmt.Printf("ошибка отправки ответа клиенту: %v", err)
	}
}

// addTaskHandler обрабатывает создание новой задачи путём декодирования запроса,
// проверки входных данных и взаимодействия с базой данных.
func addTaskHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.TaskRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondJsonError(w, http.StatusBadRequest, err)
			return
		}

		reqValid, err := utils.TaskValidator(req)
		if err != nil {
			utils.RespondJsonError(w, http.StatusBadRequest, err)
			return
		}
		task := dto.Task{
			Date:    reqValid.Date,
			Title:   reqValid.Title,
			Comment: reqValid.Comment,
			Repeat:  reqValid.Repeat,
		}
		result, err := db.AddTask(task.Date, task.Title, task.Comment, task.Repeat, database)
		if err != nil {
			utils.RespondJsonError(w, http.StatusBadRequest, "Ошибка вставки данных в дб"+err.Error())
			return
		}
		id, err := result.LastInsertId()
		idStr := strconv.FormatInt(id, 10)
		if err != nil {
			utils.RespondJsonError(w, http.StatusBadRequest, "Ошибка получения id из бд: "+err.Error())
			return
		}
		utils.RespondJson(w, http.StatusOK, dto.TaskResponse{ID: idStr})
	}
}

// getTasksHandler обрабатывает HTTP GET запросы для получения списка задач из базы данных на основе параметров поиска.
// Поддерживает фильтрацию по поисковым запросам или датам.
// При успешном выполнении возвращает JSON-объект со списком задач, в случае ошибки - сообщение об ошибке.
func getTasksHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")
		tasks, err := db.GetTasks(database, search, 20)
		if err != nil {
			utils.RespondJsonError(w, http.StatusInternalServerError, err)
			return
		}
		utils.RespondJson(w, http.StatusOK, map[string]any{"tasks": tasks})
	}
}

// updateTaskHandler обрабатывает HTTP-запросы на обновление существующей задачи в базе данных планировщика.
// Выполняет валидацию входных данных, проверяет существование задачи и обновляет базу данных.
// Отвечает соответствующим кодом состояния и сообщением в зависимости от успеха или неудачи операции.
func updateTaskHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var taskReq dto.Task
		err := json.NewDecoder(r.Body).Decode(&taskReq)
		if err != nil {
			utils.RespondJsonError(w, http.StatusBadRequest, err)
			return
		}
		err = db.TaskExists(database, taskReq.ID)
		if err != nil {
			utils.RespondJsonError(w, http.StatusBadRequest, err)
			return
		}
		req := dto.TaskRequest{
			Date:    taskReq.Date,
			Title:   taskReq.Title,
			Comment: taskReq.Comment,
			Repeat:  taskReq.Repeat,
		}
		taskValid, err := utils.TaskValidator(req)
		if err != nil {
			utils.RespondJsonError(w, http.StatusBadRequest, err)
			return
		}
		task := dto.Task{
			ID:      taskReq.ID,
			Date:    taskValid.Date,
			Title:   taskValid.Title,
			Comment: taskValid.Comment,
			Repeat:  taskValid.Repeat,
		}
		err = db.UpdateTask(database, task)
		if err != nil {
			utils.RespondJsonError(w, http.StatusInternalServerError, err)
			return
		}
		utils.RespondJson(w, http.StatusOK, new(dto.Task))
	}

}

// getTaskHandler обрабатывает GET-запросы для получения одиночной задачи по ID из базы данных
// и возвращает её в виде JSON-ответа.
func getTaskHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		task, err := db.GetTask(r, database)
		if err != nil {
			utils.RespondJsonError(w, http.StatusInternalServerError, err)
			return
		}
		utils.RespondJson(w, http.StatusOK, task)
	}
}

// taskDoneHandler обрабатывает задачу как выполненную и обрабатывает последующую логику
// на основе правила повторения задачи.
func taskDoneHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		task, err := db.GetTask(r, database)
		if err != nil {
			utils.RespondJsonError(w, http.StatusInternalServerError, err)
			return
		}
		if task.Repeat == "" {
			err = db.DeleteTask(database, task.ID)
			if err != nil {
				utils.RespondJsonError(w, http.StatusInternalServerError, err)
				return
			}
			utils.RespondJson(w, http.StatusOK, dto.Empty{})
			return
		}
		newDate, err := utils.NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			utils.RespondJsonError(w, http.StatusInternalServerError, err)
			return
		}
		task = dto.Task{
			ID:      task.ID,
			Date:    newDate,
			Title:   task.Title,
			Comment: task.Comment,
			Repeat:  task.Repeat,
		}
		err = db.UpdateTask(database, task)
		if err != nil {
			utils.RespondJsonError(w, http.StatusInternalServerError, err)
			return
		}
		utils.RespondJson(w, http.StatusOK, dto.Empty{})
	}
}

// deleteTaskHandler обрабатывает HTTP DELETE запросы на удаление задачи из базы данных по её ID.
// Находит задачу, удаляет её и отправляет соответствующий статус и JSON-ответ.
func deleteTaskHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		task, err := db.GetTask(r, database)
		if err != nil {
			utils.RespondJsonError(w, http.StatusInternalServerError, err)
			return
		}
		err = db.DeleteTask(database, task.ID)
		if err != nil {
			utils.RespondJsonError(w, http.StatusInternalServerError, err)
			return
		}
		utils.RespondJson(w, http.StatusOK, dto.Empty{})
	}
}
