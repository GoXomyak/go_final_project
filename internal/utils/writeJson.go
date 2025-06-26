package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func RespondeJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		fmt.Println(err)
	}

}

func RespondeJsonError(w http.ResponseWriter, status int, err any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	var errStr string
	switch e := err.(type) {
	case error:
		errStr = e.Error()
	case string:
		errStr = e
	default:
		errStr = fmt.Sprint(e)
	}
	errJson := json.NewEncoder(w).Encode(map[string]any{
		"error": errStr,
	})
	if errJson != nil {
		fmt.Println(errJson)
	}
}
