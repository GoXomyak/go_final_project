package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"go_final_project/internal/config"
	"go_final_project/internal/dto"
	"go_final_project/utils"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func authHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var cred dto.Credentials
		err := json.NewDecoder(r.Body).Decode(&cred)
		if err != nil {
			utils.RespondeJsonError(w, http.StatusBadRequest, errors.New("ошибка декодирования json"))
			return
		}
		if cfg.Password == "" {
			utils.RespondeJsonError(w, http.StatusBadRequest, "пароль для входа не установлен")
			return
		}

		if cred.Password != cfg.Password {
			utils.RespondeJsonError(w, http.StatusUnauthorized, errors.New("неверный пароль"))
			return
		}
		expirationTime := time.Now().Add(time.Hour * 24)
		claims := dto.Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(cfg.SecretJWT))
		if err != nil {
			utils.RespondeJsonError(w, http.StatusInternalServerError, errors.New("невозможно создать токен"))
			return
		}
		fmt.Println(tokenString)
		utils.RespondeJson(w, http.StatusOK, dto.Token{AccessToken: tokenString})
	}
}

func auth(next http.HandlerFunc, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(cfg.Password) > 0 {
			var jwtString string
			cookie, err := r.Cookie("token")
			if err != nil {
				if errors.Is(err, http.ErrNoCookie) {
					utils.RespondeJsonError(w, http.StatusUnauthorized, errors.New("пользователь не авторизован"))
					return
				}
				utils.RespondeJsonError(w, http.StatusUnauthorized, errors.New("ошибка получения cookie"))
				return
			}
			jwtString = cookie.Value
			ok, err := validateJWT(jwtString, cfg)
			if !ok {
				utils.RespondeJsonError(w, http.StatusUnauthorized, err)
				return
			}
			next(w, r)
		}
	}
}

func validateJWT(jwtString string, cfg *config.Config) (bool, error) {
	claims := &dto.Claims{}
	token, err := jwt.ParseWithClaims(jwtString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("несоответствие метода хеширования token.Header[\"alg\"]")
		}
		return []byte(cfg.SecretJWT), nil
	})
	if err != nil {
		return false, err
	}
	if !token.Valid {
		return false, errors.New("токен не валиден")
	}
	return true, nil
}
