package dto

import "github.com/golang-jwt/jwt/v5"

type TaskRequest struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type TaskResponse struct {
	ID string `json:"id"`
}

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type Empty struct {
}

type Credentials struct {
	Password string `json:"password"`
}

type Claims struct {
	jwt.RegisteredClaims
}

type Token struct {
	AccessToken string `json:"token"`
}
