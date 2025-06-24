package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go_final_project/internal/db"
	"go_final_project/internal/dto"
	"go_final_project/internal/validators"
	"go_final_project/utils"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func Init(r chi.Router, db *sql.DB) {
	r.Get("/api/nextdate", nextDayHandler)
	r.Post("/api/task", addTaskHandler(db))
	r.Get("/api/tasks", getTasksHandler(db))
	r.Get("/api/task", getTaskHandler(db))
	r.Put("/api/task", updateTaskHandler(db))
	r.Post("/api/task/done", taskDoneHandler(db))
	r.Delete("/api/task", deleteTaskHandler(db))
}

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
func addTaskHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.TaskRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondeJsonError(w, http.StatusBadRequest, err)
			return
		}

		reqValid, err := validators.TaskValidator(req)
		if err != nil {
			utils.RespondeJsonError(w, http.StatusBadRequest, err)
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
			utils.RespondeJsonError(w, http.StatusBadRequest, "Ошибка вставки данных в дб"+err.Error())
			return
		}
		id, err := result.LastInsertId()
		if err != nil {
			utils.RespondeJsonError(w, http.StatusBadRequest, "Ошибка получения id из бд: "+err.Error())
		}
		utils.RespondeJson(w, http.StatusOK, dto.TaskResponse{ID: id})
	}
}

func getTasksHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")
		tasks, err := db.GetTasks(database, search, 20)
		if err != nil {
			utils.RespondeJsonError(w, http.StatusInternalServerError, err)
			return
		}
		utils.RespondeJson(w, http.StatusOK, map[string]any{"tasks": tasks})
	}
}

func updateTaskHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var taskReq dto.Task
		err := json.NewDecoder(r.Body).Decode(&taskReq)
		if err != nil {
			utils.RespondeJsonError(w, http.StatusBadRequest, err)
			return
		}
		err = db.TaskExists(database, taskReq.ID)
		if err != nil {
			utils.RespondeJsonError(w, http.StatusBadRequest, err)
			return
		}
		req := dto.TaskRequest{
			Date:    taskReq.Date,
			Title:   taskReq.Title,
			Comment: taskReq.Comment,
			Repeat:  taskReq.Repeat,
		}
		taskValid, err := validators.TaskValidator(req)
		if err != nil {
			utils.RespondeJsonError(w, http.StatusBadRequest, err)
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
			utils.RespondeJsonError(w, http.StatusInternalServerError, err)
			return
		}
		utils.RespondeJson(w, http.StatusOK, new(dto.Task))
	}

}

func getTaskHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		task, err := db.GetTask(r, database)
		if err != nil {
			utils.RespondeJsonError(w, http.StatusInternalServerError, err)
			return
		}
		utils.RespondeJson(w, http.StatusOK, task)
	}
}

func taskDoneHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		task, err := db.GetTask(r, database)
		if err != nil {
			utils.RespondeJsonError(w, http.StatusInternalServerError, err)
			return
		}
		if task.Repeat == "" {
			err = db.DeleteTask(database, task.ID)
			if err != nil {
				utils.RespondeJsonError(w, http.StatusInternalServerError, err)
				return
			}
			utils.RespondeJson(w, http.StatusOK, dto.Empty{})
			return
		}
		newDate, err := utils.NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			utils.RespondeJsonError(w, http.StatusInternalServerError, err)
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
			utils.RespondeJsonError(w, http.StatusInternalServerError, err)
			return
		}
		utils.RespondeJson(w, http.StatusOK, dto.Empty{})
	}
}

func deleteTaskHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		task, err := db.GetTask(r, database)
		if err != nil {
			utils.RespondeJsonError(w, http.StatusInternalServerError, err)
			return
		}
		err = db.DeleteTask(database, task.ID)
		if err != nil {
			utils.RespondeJsonError(w, http.StatusInternalServerError, err)
			return
		}
		utils.RespondeJson(w, http.StatusOK, dto.Empty{})
	}
}
