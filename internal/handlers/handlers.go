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

const dateFormat = "20060102"

func Init(r chi.Router, db *sql.DB) {
	r.Get("/api/nextdate", nextDayHandler)
	r.Post("/api/task", addTaskHandler(db))
}

func nextDayHandler(w http.ResponseWriter, r *http.Request) {
	originalDate := r.URL.Query().Get("date")
	nowTemp := r.URL.Query().Get("now")
	repeat := r.URL.Query().Get("repeat")

	now, err := time.Parse(dateFormat, nowTemp)
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
