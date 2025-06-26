# ypfinalproject

API для управления задачами с возможностью повторения и хранением в SQLite

## Выполненные задания

Все задания выполнены в т.ч. повышенной сложности [*]

## Локальный запуск

Скопируйте `.env.example` в `.env` в корень проекта и заполните необходимые параметры. 

Файл .env.example уже заполнен дефолтными параметрами.

Запустите:
   ```bash
   go run main.go
   ```
Откройте в браузере: http://localhost:7540 (если порт стандартный)

## Тестирование

Запуск тестов с флагом без кэширования:
```bash
go test -count=1 ./tests
```
В tests/settings.go использовались следующие параметры:
```
var Port = 7540
var DBFile = "../internal/data/scheduler.db"
var FullNextDate = true
var Search = true
var Token = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTEwMzM3Njh9.WvKeyUjHFhqpOzaRzfyCoKmnjBcARayZTGkM4latCdw`
```

## Запуск через Docker

1. Соберите образ
```bash
docker build -t ypfinalproject .
```
2. Запустите контейнер с .env
```bash
docker run --env-file .env -e RUNNING_IN_DOCKER=true -p 7540:7540 ypfinalproject
```
3. Откройте в браузере: http://localhost:7540
4. Default password = myPassword