package utils

import (
	"fmt"
	"go_final_project/internal/dto"
	"strings"
	"time"
)

func TaskValidator(req dto.TaskRequest) (dto.TaskRequest, error) {
	now := time.Now().Format(DateFormat)
	var newDate string
	var err error
	if strings.TrimSpace(req.Date) == "" {
		req.Date = now
	}
	if len(req.Date) != len(DateFormat) {
		return dto.TaskRequest{}, fmt.Errorf("неверный формат даты")
	}
	_, ok := CheckDate(req.Date)
	if !ok {
		return dto.TaskRequest{}, fmt.Errorf("некорректная дата")
	}
	if strings.TrimSpace(req.Repeat) != "" {
		newDate, err = NextDate(time.Now(), req.Date, req.Repeat)
		if err != nil {
			return dto.TaskRequest{}, err
		}
	}

	if req.Date < now {
		if strings.TrimSpace(req.Repeat) == "" {
			req.Date = now
		} else {
			// Если дата меньше текущей, то используем следующую дату по правилам повторения
			req.Date = newDate
		}
	}

	if strings.TrimSpace(req.Title) == "" {
		return dto.TaskRequest{}, fmt.Errorf("поле Title обязательно к заполнению")
	}

	if now > req.Date {
		return dto.TaskRequest{}, fmt.Errorf("указанное время меньше текущего")
	}

	return req, nil

}
